package backup

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"safora/internal/database"
	"safora/internal/models"
	"safora/internal/retention"
)

type Engine interface {
	Run(ctx context.Context, job *models.Job) (*models.Run, error)
}

type DefaultEngine struct {
	db *sql.DB
}

func NewDefaultEngine(db *sql.DB) *DefaultEngine {
	return &DefaultEngine{db: db}
}

func (e *DefaultEngine) Run(ctx context.Context, job *models.Job) (*models.Run, error) {
	run := &models.Run{
		JobID:     job.ID,
		Status:    "running",
		StartedAt: time.Now(),
	}
	if err := database.SaveRun(e.db, run); err != nil {
		return nil, fmt.Errorf("failed to save run state: %w", err)
	}

	e.log(run.ID, "INFO", fmt.Errorf("Starting job %d: %s", job.ID, job.Name))

	var strategy Engine
	if job.StorageStrategy == "Date-Stamped Mirroring" || job.StorageStrategy == "" {
		strategy = NewDateStampedMirroring(e.db, run.ID)
	} else {
		err := fmt.Errorf("unsupported storage strategy: %s", job.StorageStrategy)
		e.failRun(run, err)
		return run, err
	}

	result, err := strategy.Run(ctx, job)
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
		e.log(run.ID, "ERROR", fmt.Errorf("failed to save final run state: %w", err))
	}

	e.log(run.ID, "INFO", fmt.Errorf("Job finished with status %s", run.Status))

	// Post-run retention hook
	retEngine := retention.NewEngine()
	pruned, err := retEngine.Prune(ctx, job, run.Status)
	if err != nil {
		if err.Error() == "retention lock active: last run failed" {
			e.log(run.ID, "WARNING", fmt.Errorf("Retention bypassed: %w", err))
		} else {
			e.log(run.ID, "ERROR", fmt.Errorf("Retention error: %w", err))
		}
	} else if len(pruned) > 0 {
		e.log(run.ID, "INFO", fmt.Errorf("Retention pruned %d historical copies", len(pruned)))
	}

	return run, nil
}

func (e *DefaultEngine) failRun(run *models.Run, err error) {
	now := time.Now()
	run.CompletedAt = &now
	run.DurationSeconds = int64(now.Sub(run.StartedAt).Seconds())
	run.Status = "failed"
	e.log(run.ID, "ERROR", err)
	_ = database.SaveRun(e.db, run)
}

func (e *DefaultEngine) log(runID int64, level string, err error) {
	_ = database.SaveLog(e.db, &models.Log{
		RunID:     runID,
		Level:     level,
		Message:   err.Error(),
		CreatedAt: time.Now(),
	})
}
