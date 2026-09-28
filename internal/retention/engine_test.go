package retention

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"safora/internal/models"
)

func TestRetentionLock_FailedRun(t *testing.T) {
	eng := NewEngine()
	job := &models.Job{RetentionPolicy: "keep 1 runs"}
	pruned, err := eng.Prune(context.Background(), job, "failed")
	if err == nil {
		t.Errorf("Expected retention lock error, got nil")
	}
	if len(pruned) != 0 {
		t.Errorf("Expected 0 pruned paths, got %d", len(pruned))
	}
}

func TestRetentionLock_SingleCopy(t *testing.T) {
	tempDir := t.TempDir()
	os.MkdirAll(filepath.Join(tempDir, "copy1"), 0755)

	job := &models.Job{
		RetentionPolicy: "keep 0 runs", // Force aggressive delete
		Destinations: []models.Destination{
			{Path: filepath.Join(tempDir, "{date}")},
		},
	}

	eng := NewEngine()
	pruned, err := eng.Prune(context.Background(), job, "success")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(pruned) != 0 {
		t.Errorf("Retention Lock Rule 2 failed: pruned single copy")
	}

	if _, err := os.Stat(filepath.Join(tempDir, "copy1")); os.IsNotExist(err) {
		t.Errorf("Single copy was deleted")
	}
}

func TestRetention_KeepRuns(t *testing.T) {
	tempDir := t.TempDir()
	
	d1 := filepath.Join(tempDir, "copy1")
	d2 := filepath.Join(tempDir, "copy2")
	d3 := filepath.Join(tempDir, "copy3")
	
	os.MkdirAll(d1, 0755)
	os.MkdirAll(d2, 0755)
	os.MkdirAll(d3, 0755)

	// Sleep slightly or use Chtimes to ensure determinism
	now := time.Now()
	os.Chtimes(d1, now.Add(-3*time.Hour), now.Add(-3*time.Hour))
	os.Chtimes(d2, now.Add(-2*time.Hour), now.Add(-2*time.Hour))
	os.Chtimes(d3, now.Add(-1*time.Hour), now.Add(-1*time.Hour)) // Newest

	job := &models.Job{
		RetentionPolicy: "keep 2 runs",
		Destinations: []models.Destination{
			{Path: filepath.Join(tempDir, "{date}")},
		},
	}

	eng := NewEngine()
	pruned, err := eng.Prune(context.Background(), job, "success")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(pruned) != 1 {
		t.Fatalf("Expected 1 pruned path, got %d", len(pruned))
	}

	if filepath.Base(pruned[0]) != "copy1" {
		t.Errorf("Expected copy1 to be pruned, got %s", pruned[0])
	}
}

func TestRetention_KeepDays(t *testing.T) {
	tempDir := t.TempDir()
	
	d1 := filepath.Join(tempDir, "copy_old")
	d2 := filepath.Join(tempDir, "copy_new")
	
	os.MkdirAll(d1, 0755)
	os.MkdirAll(d2, 0755)

	now := time.Now()
	os.Chtimes(d1, now.AddDate(0, 0, -10), now.AddDate(0, 0, -10)) // 10 days old
	os.Chtimes(d2, now.AddDate(0, 0, -1), now.AddDate(0, 0, -1))   // 1 day old

	job := &models.Job{
		RetentionPolicy: "keep 5 days",
		Destinations: []models.Destination{
			{Path: filepath.Join(tempDir, "{date}")},
		},
	}

	eng := NewEngine()
	pruned, err := eng.Prune(context.Background(), job, "success")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(pruned) != 1 {
		t.Fatalf("Expected 1 pruned path, got %d", len(pruned))
	}

	if filepath.Base(pruned[0]) != "copy_old" {
		t.Errorf("Expected copy_old to be pruned, got %s", pruned[0])
	}
}
