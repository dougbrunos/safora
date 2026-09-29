package daemon

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"safora/internal/backup"
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

	runner := NewRunner(db, backup.NewDefaultEngine(db))

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

	runner := NewRunner(db, backup.NewDefaultEngine(db))

	ctx, cancel := context.WithCancel(context.Background())

	go runner.Start(ctx)

	// Let it run for a short bit
	time.Sleep(100 * time.Millisecond)

	cancel()

	// Just verify it doesn't panic
}

func TestRemovableRoot(t *testing.T) {
	cases := map[string]string{
		`E:\Backups\Daily`:               `E:\`,
		`f:/x`:                           `f:\`,
		`/media/ana/USB DISK/backups/db`: "/media/ana/USB DISK",
		`/media/ana`:                     "/media/ana",
		`/mnt/nas/share/data`:            "/mnt/nas",
		`/home/ana/backup`:               "",
		`/srv/data`:                      "",
		`\\server\share\dir`:             "",
	}
	for path, want := range cases {
		if got := removableRoot(path); got != want {
			t.Errorf("removableRoot(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestSyncSchedulesFollowsTheJobs(t *testing.T) {
	db, err := database.InitDB(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	scheduled := &models.Job{Name: "a", StorageStrategy: "Date-Stamped Mirroring", Schedule: "0 2 * * *",
		Sources: []models.Source{{Path: "s"}}, Destinations: []models.Destination{{Path: "d"}}}
	manual := &models.Job{Name: "b", StorageStrategy: "Date-Stamped Mirroring",
		Sources: []models.Source{{Path: "s"}}, Destinations: []models.Destination{{Path: "d"}}}
	database.SaveJob(db, scheduled)
	database.SaveJob(db, manual)

	r := NewRunner(db, backup.NewDefaultEngine(db))
	r.SyncSchedules()
	if len(r.entries) != 1 || r.entries[scheduled.ID] == 0 {
		t.Fatalf("only the scheduled job gets a cron entry, got %v", r.entries)
	}

	// The schedule is removed from the job and added to the other one.
	scheduled.Schedule = ""
	database.UpdateJob(db, scheduled)
	manual.Schedule = "*/15 * * * *"
	database.UpdateJob(db, manual)
	r.SyncSchedules()
	if len(r.entries) != 1 || r.entries[manual.ID] == 0 {
		t.Fatalf("entries must follow the edited schedules, got %v", r.entries)
	}
	if len(r.cron.Entries()) != 1 {
		t.Errorf("stale cron entries were not removed: %d", len(r.cron.Entries()))
	}

	database.DeleteJob(db, manual.ID)
	r.SyncSchedules()
	if len(r.entries) != 0 {
		t.Errorf("a deleted job keeps no entry, got %v", r.entries)
	}
}
