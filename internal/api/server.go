package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"safora/internal/backup"
	"safora/internal/database"
	"safora/internal/importer"
	"safora/internal/models"
	"safora/ui"
)

type Server struct {
	db     *sql.DB
	broker *Broker
	token  string
	engine *backup.DefaultEngine

	// OnJobsChanged, if set, runs after a Job is created, updated or deleted.
	OnJobsChanged func()
}

func (s *Server) jobsChanged() {
	if s.OnJobsChanged != nil {
		s.OnJobsChanged()
	}
}

func NewServer(db *sql.DB, tokenPath string) (*Server, error) {
	token, err := loadOrCreateToken(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("api token: %w", err)
	}
	s := &Server{db: db, broker: NewBroker(), token: token, engine: backup.NewDefaultEngine(db)}
	// Every Run, however triggered, streams to the dashboard.
	s.engine.SetLogCallback(func(level, msg string) {
		s.broker.Broadcast(fmt.Sprintf("[%s] %s", level, msg))
	})
	return s, nil
}

// Engine returns the shared backup engine, wired to the Live Stream.
func (s *Server) Engine() *backup.DefaultEngine { return s.engine }

// Start serves the dashboard and API on addr (a loopback address).
func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.Handler(addr))
}

// Handler builds the routes wrapped in the Host/Origin/token checks for addr.
func (s *Server) Handler(addr string) http.Handler {

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/jobs", s.handleGetJobs)
	mux.HandleFunc("POST /api/jobs", s.handleCreateJob)
	mux.HandleFunc("PUT /api/jobs/{id}", s.handleUpdateJob)
	mux.HandleFunc("GET /api/jobs/{id}", s.handleGetJobByID)
	mux.HandleFunc("DELETE /api/jobs/{id}", s.handleDeleteJob)
	mux.HandleFunc("POST /api/jobs/{id}/run", s.handleRunJob)
	mux.HandleFunc("POST /api/jobs/{id}/cancel", s.handleCancelJob)

	mux.HandleFunc("POST /api/importer/parse", s.handleImporterParse)

	mux.HandleFunc("GET /api/runs", s.handleGetRuns)
	mux.HandleFunc("GET /api/runs/{id}", s.handleGetRunByID)

	mux.HandleFunc("GET /api/fs/list", s.handleFsList)

	mux.HandleFunc("GET /api/stream", s.broker.ServeHTTP)
	mux.Handle("/", http.FileServer(ui.GetStaticFS()))

	return s.secure(addr, mux)
}

// writeErr maps ErrNotFound to 404 and everything else to 500.
func writeErr(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	if errors.Is(err, database.ErrNotFound) {
		code = http.StatusNotFound
	}
	http.Error(w, err.Error(), code)
}

func (s *Server) handleGetJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := database.GetAllJobs(s.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(jobs)
}

func (s *Server) handleCreateJob(w http.ResponseWriter, r *http.Request) {
	var job models.Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := job.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := models.ValidateSchedule(job.Schedule); err != nil {
		http.Error(w, "invalid schedule: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := database.SaveJob(s.db, &job); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.jobsChanged()
	json.NewEncoder(w).Encode(job)
}

func (s *Server) handleGetJobByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	job, err := database.GetJobByID(s.db, id)
	if err != nil {
		writeErr(w, err)
		return
	}
	json.NewEncoder(w).Encode(job)
}

func (s *Server) handleDeleteJob(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := database.DeleteJob(s.db, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.jobsChanged()
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleRunJob(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	job, err := database.GetJobByID(s.db, id)
	if err != nil {
		writeErr(w, err)
		return
	}

	if s.engine.Busy(id) {
		http.Error(w, backup.ErrJobBusy.Error(), http.StatusConflict)
		return
	}

	go func() {
		_, _ = s.engine.Run(context.Background(), job)
	}()

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "{\"message\": \"Job %d triggered in background\"}", id)
}

// handleCancelJob stops the Run of the Job in progress: 204 when stopped, 409 when nothing is running.
func (s *Server) handleCancelJob(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if !s.engine.Cancel(id) {
		http.Error(w, "job is not running", http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleImporterParse(w http.ResponseWriter, r *http.Request) {
	// For MVP, we can write body to a temp file and use ParseBatchScript, or refactor to accept io.Reader
	// Let's create a temp file.
	f, err := os.CreateTemp("", "safora-import-*.bat")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(f.Name())

	if _, err := io.Copy(f, r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	f.Close()

	job, err := importer.ParseBatchScript(f.Name())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(job)
}

func (s *Server) handleGetRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := database.GetAllRuns(s.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(runs)
}

func (s *Server) handleGetRunByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	run, err := database.GetRunByID(s.db, id)
	if err != nil {
		writeErr(w, err)
		return
	}

	logs, _ := database.GetLogsForRun(s.db, id)

	response := struct {
		*models.Run
		Logs []models.Log `json:"logs"`
	}{
		Run:  run,
		Logs: logs,
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleUpdateJob(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var job models.Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	job.ID = id

	if err := job.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := models.ValidateSchedule(job.Schedule); err != nil {
		http.Error(w, "invalid schedule: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := database.UpdateJob(s.db, &job); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.jobsChanged()
	json.NewEncoder(w).Encode(job)
}
