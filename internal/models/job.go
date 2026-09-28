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
