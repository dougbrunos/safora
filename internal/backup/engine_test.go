package backup

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"safora/internal/database"
	"safora/internal/models"
)

// newDB opens a file-backed database: the Engine and the test use it from several goroutines.
func newDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := database.InitDB(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// saved stores a Job with the given sources and one destination.
func saved(t *testing.T, db *sql.DB, dst string, srcs []string, mod func(*models.Job)) *models.Job {
	t.Helper()
	job := &models.Job{Name: "test", StorageStrategy: "Date-Stamped Mirroring", Destinations: []models.Destination{{Path: dst}}}
	for _, s := range srcs {
		job.Sources = append(job.Sources, models.Source{Path: s})
	}
	if mod != nil {
		mod(job)
	}
	if err := database.SaveJob(db, job); err != nil {
		t.Fatalf("save job: %v", err)
	}
	return job
}

func run(t *testing.T, e *DefaultEngine, job *models.Job) *models.Run {
	t.Helper()
	r, err := e.Run(context.Background(), job)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	return r
}

func TestIncrementalCopySkipsUnchangedFiles(t *testing.T) {
	db, src, dst := newDB(t), t.TempDir(), t.TempDir()
	write(t, filepath.Join(src, "a.txt"), "hello")
	write(t, filepath.Join(src, "sub", "b.txt"), "world")
	job := saved(t, db, dst, []string{src}, nil)
	e := NewDefaultEngine(db)

	if r := run(t, e, job); r.FilesProcessed != 2 || r.Status != "success" {
		t.Fatalf("first run: want 2 files, success; got %d, %s", r.FilesProcessed, r.Status)
	}
	if r := run(t, e, job); r.FilesProcessed != 0 {
		t.Errorf("second run should skip unchanged files, copied %d", r.FilesProcessed)
	}

	write(t, filepath.Join(src, "a.txt"), "hello again") // different size
	if r := run(t, e, job); r.FilesProcessed != 1 {
		t.Errorf("third run should copy only the changed file, copied %d", r.FilesProcessed)
	}
	if got, _ := os.ReadFile(filepath.Join(dst, "a.txt")); string(got) != "hello again" {
		t.Errorf("destination not updated: %q", got)
	}
}

func TestMirrorDeletions(t *testing.T) {
	for _, sync := range []bool{true, false} {
		db, src, dst := newDB(t), t.TempDir(), t.TempDir()
		write(t, filepath.Join(src, "keep.txt"), "k")
		write(t, filepath.Join(src, "gone.txt"), "g")
		write(t, filepath.Join(src, "olddir", "x.txt"), "x")
		job := saved(t, db, dst, []string{src}, func(j *models.Job) { j.SyncDeletions = sync })
		e := NewDefaultEngine(db)
		run(t, e, job)

		os.Remove(filepath.Join(src, "gone.txt"))
		os.RemoveAll(filepath.Join(src, "olddir"))
		run(t, e, job)

		if !exists(filepath.Join(dst, "keep.txt")) {
			t.Errorf("sync=%v: keep.txt must stay", sync)
		}
		if got := !exists(filepath.Join(dst, "gone.txt")); got != sync {
			t.Errorf("sync=%v: gone.txt removed=%v", sync, got)
		}
		if got := !exists(filepath.Join(dst, "olddir")); got != sync {
			t.Errorf("sync=%v: olddir removed=%v", sync, got)
		}
	}
}

func TestMirrorDeletionsSafeguards(t *testing.T) {
	t.Run("empty source deletes nothing", func(t *testing.T) {
		db, src, dst := newDB(t), t.TempDir(), t.TempDir()
		write(t, filepath.Join(dst, "precious.txt"), "p")
		job := saved(t, db, dst, []string{src}, func(j *models.Job) { j.SyncDeletions = true })
		run(t, NewDefaultEngine(db), job)
		if !exists(filepath.Join(dst, "precious.txt")) {
			t.Error("empty source must not wipe the destination")
		}
	})

	t.Run("a failed copy deletes nothing", func(t *testing.T) {
		db, src, dst := newDB(t), t.TempDir(), t.TempDir()
		write(t, filepath.Join(src, "a.txt"), "a")
		write(t, filepath.Join(dst, "a.txt", "blocker"), "x") // a.txt is a directory: the copy fails
		write(t, filepath.Join(dst, "extra.txt"), "e")
		job := saved(t, db, dst, []string{src}, func(j *models.Job) { j.SyncDeletions = true })
		r := run(t, NewDefaultEngine(db), job)
		if r.Status != "warning" {
			t.Errorf("a file that fails to copy should make the run a warning, got %s", r.Status)
		}
		if !exists(filepath.Join(dst, "extra.txt")) {
			t.Error("nothing may be deleted when a file failed to copy")
		}
	})

	t.Run("excluded items are never deleted", func(t *testing.T) {
		db, src, dst := newDB(t), t.TempDir(), t.TempDir()
		write(t, filepath.Join(src, "a.txt"), "a")
		write(t, filepath.Join(dst, "node_modules", "lib.js"), "js")
		write(t, filepath.Join(dst, "debug.log"), "log")
		job := saved(t, db, dst, []string{src}, func(j *models.Job) {
			j.SyncDeletions = true
			j.Sources[0].ExclusionRules = "DIR:node_modules;FILE:*.log"
		})
		run(t, NewDefaultEngine(db), job)
		if !exists(filepath.Join(dst, "node_modules", "lib.js")) || !exists(filepath.Join(dst, "debug.log")) {
			t.Error("excluded destination items must survive mirroring")
		}
	})
}

func TestMultipleSourcesGetSeparateSubfolders(t *testing.T) {
	db, dst := newDB(t), t.TempDir()
	base := t.TempDir()
	a, b := filepath.Join(base, "one", "data"), filepath.Join(base, "two", "data")
	write(t, filepath.Join(a, "same.txt"), "from a")
	write(t, filepath.Join(b, "same.txt"), "from b")
	job := saved(t, db, dst, []string{a, b}, nil)
	run(t, NewDefaultEngine(db), job)

	if got, _ := os.ReadFile(filepath.Join(dst, "data", "same.txt")); string(got) != "from a" {
		t.Errorf("first source: %q", got)
	}
	if got, _ := os.ReadFile(filepath.Join(dst, "data-2", "same.txt")); string(got) != "from b" {
		t.Errorf("colliding source must not overwrite the first: %q", got)
	}
}

func TestSingleSourceKeepsFlatLayout(t *testing.T) {
	db, src, dst := newDB(t), t.TempDir(), t.TempDir()
	write(t, filepath.Join(src, "a.txt"), "a")
	run(t, NewDefaultEngine(db), saved(t, db, dst, []string{src}, nil))
	if !exists(filepath.Join(dst, "a.txt")) {
		t.Error("a single source is copied straight into the destination")
	}
}

func TestVerifyCopy(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "src"), "abcdef")
	write(t, filepath.Join(dir, "same"), "abcdef")
	write(t, filepath.Join(dir, "flipped"), "abcdeX") // same size, different content
	write(t, filepath.Join(dir, "shorter"), "abc")

	if err := verifyCopy(filepath.Join(dir, "src"), filepath.Join(dir, "same")); err != nil {
		t.Errorf("identical files: %v", err)
	}
	if err := verifyCopy(filepath.Join(dir, "src"), filepath.Join(dir, "flipped")); err == nil {
		t.Error("a flipped byte must be detected")
	}
	if err := verifyCopy(filepath.Join(dir, "src"), filepath.Join(dir, "shorter")); err == nil {
		t.Error("a size difference must be detected")
	}
}

func TestParseExclusions(t *testing.T) {
	cases := []struct {
		in         string
		dirs, file []string
	}{
		{"", nil, nil},
		{"DIR:a,b;FILE:*.x,*.y", []string{"a", "b"}, []string{"*.x", "*.y"}},
		{"DIR:temp,FILE:*.log", []string{"temp"}, []string{"*.log"}},
		{"DIR:temp,cache", []string{"temp", "cache"}, nil},
	}
	for _, c := range cases {
		dirs, files := parseExclusions(c.in)
		if len(dirs) != len(c.dirs) || len(files) != len(c.file) {
			t.Errorf("%q: got dirs=%v files=%v", c.in, dirs, files)
			continue
		}
		for i := range dirs {
			if dirs[i] != c.dirs[i] {
				t.Errorf("%q: dirs=%v", c.in, dirs)
			}
		}
		for i := range files {
			if files[i] != c.file[i] {
				t.Errorf("%q: files=%v", c.in, files)
			}
		}
	}
}

// blockingJob is a Job whose only file can never be copied (its destination is a
// directory), so the Run sits in the wait between retries until it is cancelled.
func blockingJob(t *testing.T, db *sql.DB) *models.Job {
	t.Helper()
	src, dst := t.TempDir(), t.TempDir()
	write(t, filepath.Join(src, "a.txt"), "a")
	write(t, filepath.Join(dst, "a.txt", "blocker"), "x")
	return saved(t, db, dst, []string{src}, func(j *models.Job) { j.RetryCount = 5; j.RetryWait = 60 })
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestCancelStopsARunInItsRetryWait(t *testing.T) {
	db := newDB(t)
	e := NewDefaultEngine(db)
	job := blockingJob(t, db)

	done := make(chan *models.Run, 1)
	go func() {
		r, _ := e.Run(context.Background(), job)
		done <- r
	}()
	waitFor(t, "the run to start", func() bool { return e.Busy(job.ID) })
	time.Sleep(100 * time.Millisecond) // let it reach the retry wait

	if !e.Cancel(job.ID) {
		t.Fatal("Cancel should report a run in progress")
	}
	select {
	case r := <-done:
		if r.Status != "cancelled" {
			t.Errorf("status = %s, want cancelled", r.Status)
		}
		stored, err := database.GetRunByID(db, r.ID)
		if err != nil || stored.Status != "cancelled" {
			t.Errorf("stored run: %+v, %v", stored, err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("the run did not stop after Cancel")
	}
	if e.Busy(job.ID) {
		t.Error("the job must be free again after a cancelled run")
	}
	if e.Cancel(job.ID) {
		t.Error("Cancel with nothing running must report false")
	}
}

func TestSecondRunOfTheSameJobIsRefused(t *testing.T) {
	db := newDB(t)
	e := NewDefaultEngine(db)
	job := blockingJob(t, db)

	go e.Run(context.Background(), job)
	waitFor(t, "the run to start", func() bool { return e.Busy(job.ID) })
	defer e.Cancel(job.ID)

	if _, err := e.Run(context.Background(), job); !errors.Is(err, ErrJobBusy) {
		t.Errorf("second run error = %v, want ErrJobBusy", err)
	}
}
