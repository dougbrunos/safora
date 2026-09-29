package models

import (
	"errors"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

type Job struct {
	ID              int64
	Name            string
	Description     string
	StorageStrategy string
	RetentionPolicy string
	CreatedAt       time.Time

	Sources      []Source
	Destinations []Destination

	RetryCount int
	RetryWait  int
	LogOutput  string

	// Schedule is a standard five-field cron expression; empty means the Job
	// only runs manually or when the Drive Watcher triggers it.
	Schedule string
	// VerifyIntegrity compares size and SHA-256 of every copied file.
	VerifyIntegrity bool
	// SyncDeletions removes from the Destination the files that no longer exist
	// in the Source. Off by default because it deletes backup data.
	SyncDeletions bool
}

// Validate checks the fields every Job needs before it can be saved.
func (j *Job) Validate() error {
	if strings.TrimSpace(j.Name) == "" {
		return errors.New("name is required")
	}
	if len(j.Sources) == 0 {
		return errors.New("at least one source is required")
	}
	if len(j.Destinations) == 0 {
		return errors.New("at least one destination is required")
	}
	for _, s := range j.Sources {
		if strings.TrimSpace(s.Path) == "" {
			return errors.New("source path cannot be empty")
		}
	}
	for _, d := range j.Destinations {
		if strings.TrimSpace(d.Path) == "" {
			return errors.New("destination path cannot be empty")
		}
	}
	return nil
}

// ValidateSchedule accepts an empty schedule or a standard cron expression.
func ValidateSchedule(s string) error {
	if s == "" {
		return nil
	}
	_, err := cron.ParseStandard(s)
	return err
}

type Source struct {
	ID             int64
	JobID          int64
	Path           string
	ExclusionRules string
}

type Destination struct {
	ID    int64
	JobID int64
	Path  string
}

type Run struct {
	ID               int64
	JobID            int64
	Status           string // "running", "success", "warning", "failed"
	StartedAt        time.Time
	CompletedAt      *time.Time
	DurationSeconds  int64
	BytesTransferred int64
	FilesProcessed   int64
}

type Log struct {
	ID        int64
	RunID     int64
	Level     string
	Message   string
	CreatedAt time.Time
}
