package database

import (
	"database/sql"
	"errors"
	"fmt"

	"safora/internal/models"
)

var ErrNotFound = errors.New("not found")

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

	if err := insertChildren(tx, jobID, job); err != nil {
		return err
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
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

func GetAllJobs(db *sql.DB) ([]models.Job, error) {
	rows, err := db.Query("SELECT id, name, description, storage_strategy, retention_policy, retry_count, retry_wait, log_output, created_at FROM jobs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []models.Job
	for rows.Next() {
		var job models.Job
		if err := rows.Scan(&job.ID, &job.Name, &job.Description, &job.StorageStrategy, &job.RetentionPolicy, &job.RetryCount, &job.RetryWait, &job.LogOutput, &job.CreatedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func DeleteJob(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM jobs WHERE id = ?", id)
	return err
}

func GetAllRuns(db *sql.DB) ([]models.Run, error) {
	rows, err := db.Query("SELECT id, job_id, status, started_at, completed_at, duration_seconds, bytes_transferred FROM runs ORDER BY started_at DESC LIMIT 100")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []models.Run
	for rows.Next() {
		var r models.Run
		if err := rows.Scan(&r.ID, &r.JobID, &r.Status, &r.StartedAt, &r.CompletedAt, &r.DurationSeconds, &r.BytesTransferred); err != nil {
			return nil, err
		}
		runs = append(runs, r)
	}
	return runs, nil
}

func GetRunByID(db *sql.DB, id int64) (*models.Run, error) {
	r := &models.Run{}
	row := db.QueryRow("SELECT id, job_id, status, started_at, completed_at, duration_seconds, bytes_transferred FROM runs WHERE id = ?", id)
	err := row.Scan(&r.ID, &r.JobID, &r.Status, &r.StartedAt, &r.CompletedAt, &r.DurationSeconds, &r.BytesTransferred)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return r, nil
}

func GetLogsForRun(db *sql.DB, runID int64) ([]models.Log, error) {
	rows, err := db.Query("SELECT id, run_id, level, message, created_at FROM logs WHERE run_id = ? ORDER BY created_at ASC", runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.Log
	for rows.Next() {
		var l models.Log
		if err := rows.Scan(&l.ID, &l.RunID, &l.Level, &l.Message, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// UpdateJob updates a Job and its related sources and destinations.
func UpdateJob(db *sql.DB, job *models.Job) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE jobs SET name = ?, description = ?, storage_strategy = ?, retention_policy = ?, retry_count = ?, retry_wait = ?, log_output = ?
		WHERE id = ?`,
		job.Name, job.Description, job.StorageStrategy, job.RetentionPolicy, job.RetryCount, job.RetryWait, job.LogOutput, job.ID,
	)
	if err != nil {
		return fmt.Errorf("update job: %w", err)
	}

	_, err = tx.Exec(`DELETE FROM sources WHERE job_id = ?`, job.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM destinations WHERE job_id = ?`, job.ID)
	if err != nil {
		return err
	}

	if err := insertChildren(tx, job.ID, job); err != nil {
		return err
	}

	return tx.Commit()
}

func insertChildren(tx *sql.Tx, jobID int64, job *models.Job) error {
	for _, src := range job.Sources {
		if _, err := tx.Exec(`INSERT INTO sources (job_id, path, exclusion_rules) VALUES (?, ?, ?)`,
			jobID, src.Path, src.ExclusionRules); err != nil {
			return fmt.Errorf("insert source: %w", err)
		}
	}
	for _, dst := range job.Destinations {
		if _, err := tx.Exec(`INSERT INTO destinations (job_id, path) VALUES (?, ?)`, jobID, dst.Path); err != nil {
			return fmt.Errorf("insert destination: %w", err)
		}
	}
	return nil
}
