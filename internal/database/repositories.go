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

func GetJobByID(db *sql.DB, id int64) (*models.Job, error) {
	job := &models.Job{}
	row := db.QueryRow(`
		SELECT id, name, description, storage_strategy, retention_policy, retry_count, retry_wait, log_output, created_at
		FROM jobs WHERE id = ?`, id)
	err := row.Scan(&job.ID, &job.Name, &job.Description, &job.StorageStrategy, &job.RetentionPolicy, &job.RetryCount, &job.RetryWait, &job.LogOutput, &job.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job not found")
		}
		return nil, err
	}

	rows, err := db.Query("SELECT id, job_id, path, exclusion_rules FROM sources WHERE job_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var src models.Source
		if err := rows.Scan(&src.ID, &src.JobID, &src.Path, &src.ExclusionRules); err != nil {
			return nil, err
		}
		job.Sources = append(job.Sources, src)
	}

	dRows, err := db.Query("SELECT id, job_id, path FROM destinations WHERE job_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer dRows.Close()
	for dRows.Next() {
		var dst models.Destination
		if err := dRows.Scan(&dst.ID, &dst.JobID, &dst.Path); err != nil {
			return nil, err
		}
		job.Destinations = append(job.Destinations, dst)
	}

	return job, nil
}

func SaveRun(db *sql.DB, run *models.Run) error {
	if run.ID == 0 {
		res, err := db.Exec(`
			INSERT INTO runs (job_id, status, started_at)
			VALUES (?, ?, ?)`, run.JobID, run.Status, run.StartedAt)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		run.ID = id
		return nil
	}

	_, err := db.Exec(`
		UPDATE runs SET status = ?, completed_at = ?, duration_seconds = ?, bytes_transferred = ?
		WHERE id = ?`, run.Status, run.CompletedAt, run.DurationSeconds, run.BytesTransferred, run.ID)
	return err
}

func SaveLog(db *sql.DB, log *models.Log) error {
	_, err := db.Exec(`
		INSERT INTO logs (run_id, level, message, created_at)
		VALUES (?, ?, ?, ?)`, log.RunID, log.Level, log.Message, log.CreatedAt)
	return err
}
