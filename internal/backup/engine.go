package backup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"safora/internal/database"
	"safora/internal/models"
	"safora/internal/retention"
)

// ErrJobBusy is returned when a Job is started while a Run of it is in progress.
var ErrJobBusy = errors.New("job already running")

type DefaultEngine struct {
	db          *sql.DB
	logCallback func(level, msg string)
	running     sync.Map // job ID -> struct{}
}

func NewDefaultEngine(db *sql.DB) *DefaultEngine {
	return &DefaultEngine{db: db}
}

func (e *DefaultEngine) SetLogCallback(cb func(level, msg string)) {
	e.logCallback = cb
}

// Busy reports whether a Run of the Job is in progress.
func (e *DefaultEngine) Busy(jobID int64) bool {
	_, ok := e.running.Load(jobID)
	return ok
}

func (e *DefaultEngine) Run(ctx context.Context, job *models.Job) (*models.Run, error) {
	if _, loaded := e.running.LoadOrStore(job.ID, struct{}{}); loaded {
		msg := fmt.Sprintf("Job %d (%s) is already running; start refused", job.ID, job.Name)
		log.Print(msg)
		if e.logCallback != nil {
			e.logCallback("WARNING", msg)
		}
		return nil, ErrJobBusy
	}
	defer e.running.Delete(job.ID)

	run := &models.Run{
		JobID:     job.ID,
		Status:    "running",
		StartedAt: time.Now(),
	}
	if err := database.SaveRun(e.db, run); err != nil {
		return nil, fmt.Errorf("failed to save run state: %w", err)
	}

	e.log(run.ID, "INFO", fmt.Sprintf("Starting job %d: %s", job.ID, job.Name))

	if job.StorageStrategy != "Date-Stamped Mirroring" && job.StorageStrategy != "" {
		err := fmt.Errorf("unsupported storage strategy: %s", job.StorageStrategy)
		e.failRun(run, err)
		return run, err
	}

	result, err := NewDateStampedMirroring(e.db, run.ID, e.logCallback).Run(ctx, job)
	if err != nil {
		e.failRun(run, err)
		return run, err
	}

	now := time.Now()
	run.CompletedAt = &now
	run.DurationSeconds = int64(now.Sub(run.StartedAt).Seconds())
	run.Status = result.Status
	run.BytesTransferred = result.BytesTransferred
	run.FilesProcessed = result.FilesProcessed

	if err := database.SaveRun(e.db, run); err != nil {
		e.log(run.ID, "ERROR", fmt.Sprintf("failed to save final run state: %v", err))
	}

	e.log(run.ID, "INFO", fmt.Sprintf("Job finished with status %s", run.Status))

	// Post-run retention hook
	pruned, err := retention.NewEngine().Prune(ctx, job, run.Status)
	if err != nil {
		if errors.Is(err, retention.ErrLocked) {
			e.log(run.ID, "WARNING", fmt.Sprintf("Retention bypassed: %v", err))
		} else {
			e.log(run.ID, "ERROR", fmt.Sprintf("Retention error: %v", err))
		}
	} else if len(pruned) > 0 {
		e.log(run.ID, "INFO", fmt.Sprintf("Retention pruned %d historical copies", len(pruned)))
	}

	return run, nil
}

func (e *DefaultEngine) failRun(run *models.Run, err error) {
	now := time.Now()
	run.CompletedAt = &now
	run.DurationSeconds = int64(now.Sub(run.StartedAt).Seconds())
	run.Status = "failed"
	e.log(run.ID, "ERROR", err.Error())
	_ = database.SaveRun(e.db, run)
}

func (e *DefaultEngine) log(runID int64, level, msg string) {
	if e.logCallback != nil {
		e.logCallback(level, msg)
	}
	_ = database.SaveLog(e.db, &models.Log{
		RunID:     runID,
		Level:     level,
		Message:   msg,
		CreatedAt: time.Now(),
	})
}
