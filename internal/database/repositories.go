package database

import (
	"database/sql"
	"fmt"

	"safora/internal/models"
)

// SaveJob inserts a Job and its related sources and destinations into the database.
func SaveJob(db *sql.DB, job *models.Job) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
		INSERT INTO jobs (name, description, storage_strategy, retention_policy, retry_count, retry_wait, log_output)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		job.Name, job.Description, job.StorageStrategy, job.RetentionPolicy, job.RetryCount, job.RetryWait, job.LogOutput,
	)
	if err != nil {
		return fmt.Errorf("insert job: %w", err)
	}

	jobID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	job.ID = jobID

	for i := range job.Sources {
		src := &job.Sources[i]
		_, err := tx.Exec(`
			INSERT INTO sources (job_id, path, exclusion_rules)
			VALUES (?, ?, ?)`,
			jobID, src.Path, src.ExclusionRules,
		)
		if err != nil {
			return fmt.Errorf("insert source: %w", err)
		}
	}

	for i := range job.Destinations {
		dst := &job.Destinations[i]
		_, err := tx.Exec(`
			INSERT INTO destinations (job_id, path)
			VALUES (?, ?)`,
			jobID, dst.Path,
		)
		if err != nil {
			return fmt.Errorf("insert destination: %w", err)
		}
	}

	return tx.Commit()
}
