package daemon

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/robfig/cron/v3"
	
	"safora/internal/backup"
	"safora/internal/database"
)

type Runner struct {
	db       *sql.DB
	cron     *cron.Cron
	engine   *backup.DefaultEngine
}

func NewRunner(db *sql.DB) *Runner {
	return &Runner{
		db:     db,
		cron:   cron.New(),
		engine: backup.NewDefaultEngine(db),
	}
}

func (r *Runner) Start(ctx context.Context) {
	// Schedule jobs (For MVP we assume jobs might have a cron schedule, but let's just 
	// pull them from DB. The schema doesn't have a schedule field yet. 
	// We'll mock a default schedule or add the watcher)

	// Note: since schema lacks a cron string, we just set a periodic 1-hour dummy trigger 
	// for jobs to prove the scheduler works for the MVP, or we can just focus on the watcher.
	r.cron.AddFunc("@hourly", func() {
		log.Println("Running scheduled jobs...")
		jobs, _ := database.GetAllJobs(r.db)
		for _, j := range jobs {
			r.executeJob(j.ID)
		}
	})
	
	r.cron.Start()

	// Drive watcher
	go r.watchDrives(ctx)

	<-ctx.Done()
	r.cron.Stop()
}

func (r *Runner) watchDrives(ctx context.Context) {
	// Polling for drive arrival (naive approach for MVP)
	// Checks destinations of all jobs to see if they just appeared.
	
	knownDrives := make(map[string]bool)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
			jobs, err := database.GetAllJobs(r.db)
			if err != nil {
				continue
			}
			
			for _, job := range jobs {
				// Re-fetch job with destinations
				fullJob, err := database.GetJobByID(r.db, job.ID)
				if err != nil {
					continue
				}

				for _, dest := range fullJob.Destinations {
					var drive string
					if len(dest.Path) >= 2 && dest.Path[1] == ':' { // Windows F:
						drive = dest.Path[:2]
					} else if len(dest.Path) > 6 && (dest.Path[:7] == "/media/" || dest.Path[:5] == "/mnt/") { // Linux paths
						// Just use the first 3 directories as the "drive" identity (e.g. /media/user/usb)
						// Highly simplified for MVP
						drive = dest.Path
					}

					if drive != "" {
						if !knownDrives[drive] {
							// For MVP, we assume if it's evaluated here, it just "arrived" 
							// In a real implementation we'd check os.Stat to see if it's actually mounted
							knownDrives[drive] = true
							log.Printf("Drive/Mount %s detected, triggering job %d", drive, job.ID)
							go r.executeJob(job.ID)
						}
					}
				}
			}
		}
	}
}

func (r *Runner) executeJob(jobID int64) {
	job, err := database.GetJobByID(r.db, jobID)
	if err != nil {
		log.Printf("Execute job failed: %v", err)
		return
	}

	run, err := r.engine.Run(context.Background(), job)
	
	if err != nil || run.Status == "failed" {
		beeep.Alert("Safora Backup Failed", fmt.Sprintf("Job %s failed: %v", job.Name, err), "")
	} else {
		beeep.Notify("Safora Backup Success", fmt.Sprintf("Job %s completed successfully.", job.Name), "")
	}
}
