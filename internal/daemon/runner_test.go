package daemon

import (
	"context"
	"testing"
	"time"

	"safora/internal/database"
	"safora/internal/models"
)

func TestRunner_ExecuteJob(t *testing.T) {
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	defer db.Close()

	job := &models.Job{
		Name:            "Test Job",
		StorageStrategy: "Date-Stamped Mirroring",
		Sources:         []models.Source{{Path: "src"}},
		Destinations:    []models.Destination{{Path: "dst"}},
	}

	database.SaveJob(db, job)

	runner := NewRunner(db)

	// Execute synchronously for test
	runner.executeJob(job.ID)

	runs, err := database.GetAllRuns(db)
	if err != nil {
		t.Fatalf("Failed to get runs: %v", err)
	}

	if len(runs) != 1 {
		t.Fatalf("Expected 1 run, got %d", len(runs))
	}

	if runs[0].Status != "success" && runs[0].Status != "warning" && runs[0].Status != "failed" {
		t.Errorf("Invalid status: %s", runs[0].Status)
	}
}

func TestRunner_StartStop(t *testing.T) {
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	defer db.Close()

	runner := NewRunner(db)

	ctx, cancel := context.WithCancel(context.Background())

	go runner.Start(ctx)

	// Let it run for a short bit
	time.Sleep(100 * time.Millisecond)

	cancel()

	// Just verify it doesn't panic
}
