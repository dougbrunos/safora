package database

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"safora/internal/models"
)

func openFile(t *testing.T) (*sql.DB, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := InitDB(path)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, path
}

func count(t *testing.T, db *sql.DB, table string) int {
	t.Helper()
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func sampleJob() *models.Job {
	return &models.Job{
		Name: "nightly", StorageStrategy: "Date-Stamped Mirroring", RetentionPolicy: "keep 5 runs",
		RetryCount: 3, RetryWait: 30, Schedule: "0 2 * * *", VerifyIntegrity: true, SyncDeletions: true,
		Sources:      []models.Source{{Path: "/src/a", ExclusionRules: "DIR:tmp"}, {Path: "/src/b"}},
		Destinations: []models.Destination{{Path: "/dst/{today}"}},
	}
}

func TestSchemaIsVersionedAndReopenable(t *testing.T) {
	db, path := openFile(t)

	var version int
	db.QueryRow("PRAGMA user_version").Scan(&version)
	if version != len(migrations) {
		t.Errorf("user_version = %d, want %d", version, len(migrations))
	}
	var fk int
	db.QueryRow("PRAGMA foreign_keys").Scan(&fk)
	if fk != 1 {
		t.Error("foreign keys must be enforced")
	}
	var mode string
	db.QueryRow("PRAGMA journal_mode").Scan(&mode)
	if mode != "wal" {
		t.Errorf("journal_mode = %s, want wal", mode)
	}

	db.Close()
	again, err := InitDB(path)
	if err != nil {
		t.Fatalf("reopening must not re-run migrations: %v", err)
	}
	again.Close()
}

func TestUpgradeFromUnversionedDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "old.db")
	old, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	// The schema as created before versioning existed, plus a job that must survive.
	if _, err := old.Exec(migrations[0]); err != nil {
		t.Fatal(err)
	}
	if _, err := old.Exec(`INSERT INTO jobs (name, storage_strategy) VALUES ('legacy', 'Date-Stamped Mirroring')`); err != nil {
		t.Fatal(err)
	}
	old.Close()

	db, err := InitDB(path)
	if err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}
	defer db.Close()

	jobs, err := GetAllJobs(db)
	if err != nil || len(jobs) != 1 || jobs[0].Name != "legacy" {
		t.Fatalf("legacy job lost in the upgrade: %v, %v", jobs, err)
	}
	if jobs[0].Schedule != "" || jobs[0].SyncDeletions {
		t.Errorf("new columns must default to empty/off: %+v", jobs[0])
	}
}

func TestJobRoundTrip(t *testing.T) {
	db, _ := openFile(t)
	job := sampleJob()
	if err := SaveJob(db, job); err != nil {
		t.Fatal(err)
	}
	got, err := GetJobByID(db, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Schedule != "0 2 * * *" || !got.VerifyIntegrity || !got.SyncDeletions || got.RetentionPolicy != "keep 5 runs" {
		t.Errorf("fields lost: %+v", got)
	}
	if len(got.Sources) != 2 || got.Sources[0].Path != "/src/a" || got.Sources[1].Path != "/src/b" {
		t.Errorf("sources = %+v (order must follow insertion)", got.Sources)
	}

	got.Name, got.Schedule, got.Sources = "renamed", "", got.Sources[:1]
	if err := UpdateJob(db, got); err != nil {
		t.Fatal(err)
	}
	updated, _ := GetJobByID(db, job.ID)
	if updated.Name != "renamed" || updated.Schedule != "" || len(updated.Sources) != 1 {
		t.Errorf("update not applied: %+v", updated)
	}
}

func TestGetAllJobsIncludesSourcesAndDestinations(t *testing.T) {
	db, _ := openFile(t)
	SaveJob(db, sampleJob())
	second := sampleJob()
	second.Name, second.Sources, second.Destinations = "second", second.Sources[:1], []models.Destination{{Path: "/other"}}
	SaveJob(db, second)

	jobs, err := GetAllJobs(db)
	if err != nil || len(jobs) != 2 {
		t.Fatalf("jobs = %d, err = %v", len(jobs), err)
	}
	if len(jobs[0].Sources) != 2 || len(jobs[0].Destinations) != 1 || len(jobs[1].Sources) != 1 || jobs[1].Destinations[0].Path != "/other" {
		t.Errorf("list must carry each job's own sources and destinations (the edit form relies on it): %+v", jobs)
	}
}

func TestEmptyListsAreNotNil(t *testing.T) {
	db, _ := openFile(t)
	jobs, _ := GetAllJobs(db)
	runs, _ := GetAllRuns(db)
	if jobs == nil || runs == nil {
		t.Error("empty results must be empty slices so the API returns [] and not null")
	}
}

func TestNotFound(t *testing.T) {
	db, _ := openFile(t)
	if _, err := GetJobByID(db, 42); !errors.Is(err, ErrNotFound) {
		t.Errorf("job: %v", err)
	}
	if _, err := GetRunByID(db, 42); !errors.Is(err, ErrNotFound) {
		t.Errorf("run: %v", err)
	}
}

func TestDeleteJobCascades(t *testing.T) {
	db, _ := openFile(t)
	job := sampleJob()
	SaveJob(db, job)
	run := &models.Run{JobID: job.ID, Status: "success", StartedAt: time.Now()}
	SaveRun(db, run)
	SaveLog(db, &models.Log{RunID: run.ID, Level: "INFO", Message: "m", CreatedAt: time.Now()})

	if err := DeleteJob(db, job.ID); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"jobs", "sources", "destinations", "runs", "logs"} {
		if n := count(t, db, table); n != 0 {
			t.Errorf("%s still has %d rows after deleting the job", table, n)
		}
	}
}

func TestRunningRunIsReadable(t *testing.T) {
	db, _ := openFile(t)
	job := sampleJob()
	SaveJob(db, job)
	SaveRun(db, &models.Run{JobID: job.ID, Status: "running", StartedAt: time.Now()}) // duration/bytes are NULL
	runs, err := GetAllRuns(db)
	if err != nil || len(runs) != 1 || runs[0].Status != "running" {
		t.Fatalf("a run in progress must be listable: %v, %v", runs, err)
	}
}

func TestRecoverOrphanedRuns(t *testing.T) {
	db, _ := openFile(t)
	job := sampleJob()
	SaveJob(db, job)
	stuck := &models.Run{JobID: job.ID, Status: "running", StartedAt: time.Now()}
	done := &models.Run{JobID: job.ID, Status: "success", StartedAt: time.Now()}
	SaveRun(db, stuck)
	SaveRun(db, done)

	n, err := RecoverOrphanedRuns(db)
	if err != nil || n != 1 {
		t.Fatalf("recovered %d, err %v", n, err)
	}
	if got, _ := GetRunByID(db, stuck.ID); got.Status != "failed" || got.CompletedAt == nil {
		t.Errorf("stuck run: %+v", got)
	}
	if got, _ := GetRunByID(db, done.ID); got.Status != "success" {
		t.Errorf("finished run must be untouched: %+v", got)
	}
	logs, _ := GetLogsForRun(db, stuck.ID)
	if len(logs) != 1 || logs[0].Level != "ERROR" {
		t.Errorf("the interruption must be explained in the log: %+v", logs)
	}
}

func TestRunFilesProcessedIsPersisted(t *testing.T) {
	db, _ := openFile(t)
	job := sampleJob()
	SaveJob(db, job)
	run := &models.Run{JobID: job.ID, Status: "running", StartedAt: time.Now()}
	SaveRun(db, run)
	now := time.Now()
	run.Status, run.CompletedAt, run.FilesProcessed, run.BytesTransferred = "success", &now, 7, 1234
	SaveRun(db, run)
	got, _ := GetRunByID(db, run.ID)
	if got.FilesProcessed != 7 || got.BytesTransferred != 1234 {
		t.Errorf("run stats lost: %+v", got)
	}
}
