package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"safora/internal/backup"
	"safora/internal/database"
	"safora/internal/importer"
	"safora/internal/models"
	"safora/ui"
)

type Server struct {
	db     *sql.DB
	broker *Broker
}

func NewServer(db *sql.DB) *Server {
	return &Server{
		db:     db,
		broker: NewBroker(),
	}
}

func (s *Server) Start(addr string) error {

	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("GET /api/jobs", s.handleGetJobs)
	mux.HandleFunc("POST /api/jobs", s.handleCreateJob)
	mux.HandleFunc("GET /api/jobs/", s.handleGetJobByID) // Needs manual ID parsing since Go 1.24 router is not fully regex in ServeMux unless using 1.22+ features correctly
	mux.HandleFunc("DELETE /api/jobs/", s.handleDeleteJob)
	mux.HandleFunc("POST /api/jobs/", s.handleRunJob) // /api/jobs/:id/run

	mux.HandleFunc("POST /api/importer/parse", s.handleImporterParse)

	mux.HandleFunc("GET /api/runs", s.handleGetRuns)
	mux.HandleFunc("GET /api/runs/", s.handleGetRunByID)

	mux.HandleFunc("GET /api/stream", s.broker.ServeHTTP)

	// Since Go 1.22, ServeMux supports method and path variables
	// Let's redefine with Go 1.22+ routing syntax
	mux22 := http.NewServeMux()
	mux22.HandleFunc("GET /api/jobs", s.handleGetJobs)
	mux22.HandleFunc("POST /api/jobs", s.handleCreateJob)
	mux22.HandleFunc("PUT /api/jobs/{id}", s.handleUpdateJob)
	mux22.HandleFunc("GET /api/jobs/{id}", s.handleGetJobByID)
	mux22.HandleFunc("DELETE /api/jobs/{id}", s.handleDeleteJob)
	mux22.HandleFunc("POST /api/jobs/{id}/run", s.handleRunJob)

	mux22.HandleFunc("POST /api/importer/parse", s.handleImporterParse)

	mux22.HandleFunc("GET /api/runs", s.handleGetRuns)
	mux22.HandleFunc("GET /api/runs/{id}", s.handleGetRunByID)

	mux22.HandleFunc("GET /api/stream", s.broker.ServeHTTP)

	// Serve UI
	importUI := func() http.Handler {
		return http.FileServer(ui.GetStaticFS())
	}
	mux22.Handle("/", importUI())

	return http.ListenAndServe(addr, s.authMiddleware(mux22))
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Localhost bind security with optional authorization token support
		if !strings.HasPrefix(r.RemoteAddr, "127.0.0.1") && !strings.HasPrefix(r.RemoteAddr, "[::1]") {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			// Validate token (dummy validation for MVP)
			if token != "Bearer secret-token" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// Handlers implementation

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
	if err := database.SaveJob(s.db, &job); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
		http.Error(w, err.Error(), http.StatusNotFound)
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
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	eng := backup.NewDefaultEngine(s.db)

	// Hook the engine logger to SSE broker
	eng.SetLogCallback(func(level, msg string) {
		s.broker.Broadcast(fmt.Sprintf("[%s] %s", level, msg))
	})

	go func() {
		// Run in background
		_, _ = eng.Run(context.Background(), job)
	}()

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "{\"message\": \"Job %d triggered in background\"}", id)
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
		http.Error(w, err.Error(), http.StatusNotFound)
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

	if err := database.UpdateJob(s.db, &job); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(job)
}
