package models

import "time"

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
	FilesProcessed   int64 // Not in DB schema but useful in model, we can alter db if needed or just keep it here
}

type Log struct {
	ID        int64
	RunID     int64
	Level     string
	Message   string
	CreatedAt time.Time
}
