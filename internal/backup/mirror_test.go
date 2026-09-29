package backup

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"safora/internal/database"
	"safora/internal/models"
)

func TestDateStampedMirroring(t *testing.T) {
	// Setup in-memory DB
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	defer db.Close()

	// Setup temp dirs
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")

	os.MkdirAll(filepath.Join(srcDir, "folder1"), 0755)
	os.MkdirAll(filepath.Join(srcDir, "123LAUDOS123"), 0755)

	os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(srcDir, "folder1", "file2.tmp"), []byte("temp file"), 0644)
	os.WriteFile(filepath.Join(srcDir, "123LAUDOS123", "secret.txt"), []byte("secret"), 0644)

	job := &models.Job{
		Name:            "Test Job",
		StorageStrategy: "Date-Stamped Mirroring",
		Sources: []models.Source{
			{Path: srcDir, ExclusionRules: "DIR:123LAUDOS123;FILE:*.tmp"},
		},
		Destinations: []models.Destination{
			{Path: dstDir},
		},
		RetryCount: 1,
		RetryWait:  0,
	}

	database.SaveJob(db, job)

	engine := NewDefaultEngine(db)
	run, err := engine.Run(context.Background(), job)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if run.Status != "success" {
		t.Errorf("Expected status success, got %s", run.Status)
	}

	if run.FilesProcessed != 1 {
		t.Errorf("Expected 1 file processed (file1.txt), got %d", run.FilesProcessed)
	}

	if run.BytesTransferred != 5 {
		t.Errorf("Expected 5 bytes transferred, got %d", run.BytesTransferred)
	}

	// Verify file exists
	if _, err := os.Stat(filepath.Join(dstDir, "file1.txt")); os.IsNotExist(err) {
		t.Errorf("file1.txt was not copied")
	}

	// Verify exclusions
	if _, err := os.Stat(filepath.Join(dstDir, "folder1", "file2.tmp")); !os.IsNotExist(err) {
		t.Errorf("file2.tmp should have been excluded")
	}

	if _, err := os.Stat(filepath.Join(dstDir, "123LAUDOS123")); !os.IsNotExist(err) {
		t.Errorf("123LAUDOS123 should have been excluded")
	}
}
