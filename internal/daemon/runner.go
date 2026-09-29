package daemon

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/robfig/cron/v3"

	"safora/internal/backup"
	"safora/internal/database"
)

const drivePollInterval = 10 * time.Second

type Runner struct {
	db     *sql.DB
	cron   *cron.Cron
	engine *backup.DefaultEngine

	mu      sync.Mutex
	entries map[int64]cron.EntryID // job ID -> cron entry
}

func NewRunner(db *sql.DB, engine *backup.DefaultEngine) *Runner {
	return &Runner{db: db, cron: cron.New(), engine: engine, entries: map[int64]cron.EntryID{}}
}

func (r *Runner) Start(ctx context.Context) {
	r.SyncSchedules()
	r.cron.Start()

	go r.watchDrives(ctx)

	<-ctx.Done()
	r.cron.Stop()
}

// SyncSchedules makes the cron entries match the Jobs in the database. Call it
// whenever a Job is created, updated or deleted.
func (r *Runner) SyncSchedules() {
	jobs, err := database.GetAllJobs(r.db)
	if err != nil {
		log.Printf("Sync schedules failed: %v", err)
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, id := range r.entries {
		r.cron.Remove(id)
	}
	r.entries = map[int64]cron.EntryID{}

	for _, j := range jobs {
		if j.Schedule == "" {
			continue
		}
		jobID := j.ID
		entry, err := r.cron.AddFunc(j.Schedule, func() { r.executeJob(jobID) })
		if err != nil {
			log.Printf("Job %d has an invalid schedule %q: %v", jobID, j.Schedule, err)
			continue
		}
		r.entries[jobID] = entry
	}
}

func (r *Runner) watchDrives(ctx context.Context) {
	state := map[string]bool{} // removable root -> mounted at last poll
	r.pollDrives(state)

	tick := time.NewTicker(drivePollInterval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			r.pollDrives(state)
		}
	}
}

// pollDrives triggers the Jobs of a removable Destination that went from
// unmounted to mounted since the last poll. A root seen for the first time only
// records its state, so drives already present at startup do not trigger.
func (r *Runner) pollDrives(state map[string]bool) {
	dests, err := database.GetAllDestinations(r.db)
	if err != nil {
		return
	}

	jobsByRoot := map[string]map[int64]bool{}
	for _, d := range dests {
		root := removableRoot(d.Path)
		if root == "" {
			continue
		}
		if jobsByRoot[root] == nil {
			jobsByRoot[root] = map[int64]bool{}
		}
		jobsByRoot[root][d.JobID] = true
	}

	for root, jobIDs := range jobsByRoot {
		_, statErr := os.Stat(root)
		mounted := statErr == nil
		wasMounted, seen := state[root]
		state[root] = mounted

		if seen && !wasMounted && mounted {
			for id := range jobIDs {
				log.Printf("Drive/Mount %s arrived, triggering job %d", root, id)
				go r.executeJob(id)
			}
		}
	}
}

// removableRoot returns the path whose existence means the volume holding path
// is mounted (a Windows drive root, or /media/<user>/<label>, /mnt/<name>), or
// "" when path is not on a removable volume.
func removableRoot(path string) string {
	if len(path) >= 2 && path[1] == ':' {
		return path[:2] + `\`
	}
	for prefix, depth := range map[string]int{"/media/": 3, "/mnt/": 2} {
		if strings.HasPrefix(path, prefix) {
			parts := strings.Split(strings.Trim(path, "/"), "/")
			if len(parts) > depth {
				parts = parts[:depth]
			}
			return "/" + strings.Join(parts, "/")
		}
	}
	return ""
}

func (r *Runner) executeJob(jobID int64) {
	job, err := database.GetJobByID(r.db, jobID)
	if err != nil {
		log.Printf("Execute job failed: %v", err)
		return
	}

	run, err := r.engine.Run(context.Background(), job)
	if errors.Is(err, backup.ErrJobBusy) {
		return
	}

	if err != nil || run.Status == "failed" {
		beeep.Alert("Safora Backup Failed", fmt.Sprintf("Job %s failed: %v", job.Name, err), "")
	} else {
		beeep.Notify("Safora Backup Success", fmt.Sprintf("Job %s completed successfully.", job.Name), "")
	}
}
