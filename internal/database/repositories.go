package database

import (
	"database/sql"
	"errors"
	"fmt"

	"safora/internal/models"
)

var ErrNotFound = errors.New("not found")

// Nullable text columns are COALESCEd: one NULL must not make the whole list unreadable.
const jobColumns = "id, name, COALESCE(description, ''), storage_strategy, COALESCE(retention_policy, ''), COALESCE(retry_count, 0), COALESCE(retry_wait, 0), COALESCE(log_output, ''), schedule, verify_integrity, sync_deletions, created_at"

func scanJob(row interface{ Scan(...any) error }, j *models.Job) error {
	return row.Scan(&j.ID, &j.Name, &j.Description, &j.StorageStrategy, &j.RetentionPolicy, &j.RetryCount, &j.RetryWait, &j.LogOutput, &j.Schedule, &j.VerifyIntegrity, &j.SyncDeletions, &j.CreatedAt)
}

// SaveJob inserts a Job and its related sources and destinations into the database.
func SaveJob(db *sql.DB, job *models.Job) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
		INSERT INTO jobs (name, description, storage_strategy, retention_policy, retry_count, retry_wait, log_output, schedule, verify_integrity, sync_deletions)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		job.Name, job.Description, job.StorageStrategy, job.RetentionPolicy, job.RetryCount, job.RetryWait, job.LogOutput, job.Schedule, job.VerifyIntegrity, job.SyncDeletions,
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
	err := scanJob(db.QueryRow("SELECT "+jobColumns+" FROM jobs WHERE id = ?", id), job)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	rows, err := db.Query("SELECT id, job_id, path, COALESCE(exclusion_rules, '') FROM sources WHERE job_id = ? ORDER BY id", id)
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

	dRows, err := db.Query("SELECT id, job_id, path FROM destinations WHERE job_id = ? ORDER BY id", id)
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
		UPDATE runs SET status = ?, completed_at = ?, duration_seconds = ?, bytes_transferred = ?, files_processed = ?
		WHERE id = ?`, run.Status, run.CompletedAt, run.DurationSeconds, run.BytesTransferred, run.FilesProcessed, run.ID)
	return err
}

// RecoverOrphanedRuns marks Runs still "running" as failed. Call it at startup,
// before any Trigger can start a new Run: a Run left running by a previous
// process can never finish.
func RecoverOrphanedRuns(db *sql.DB) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		INSERT INTO logs (run_id, level, message)
		SELECT id, 'ERROR', 'Run interrupted: Safora stopped before it finished'
		FROM runs WHERE status = 'running'`); err != nil {
		return 0, err
	}
	res, err := tx.Exec(`
		UPDATE runs SET status = 'failed', completed_at = CURRENT_TIMESTAMP
		WHERE status = 'running'`)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, tx.Commit()
}

func SaveLog(db *sql.DB, log *models.Log) error {
	_, err := db.Exec(`
		INSERT INTO logs (run_id, level, message, created_at)
		VALUES (?, ?, ?, ?)`, log.RunID, log.Level, log.Message, log.CreatedAt)
	return err
}

func GetAllJobs(db *sql.DB) ([]models.Job, error) {
	rows, err := db.Query("SELECT " + jobColumns + " FROM jobs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := []models.Job{}
	for rows.Next() {
		var job models.Job
		if err := scanJob(rows, &job); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close() // release the connection before the follow-up queries

	// Attach Sources and Destinations so list consumers (the edit form, the
	// schedule sync) see complete Jobs without one query per Job.
	srcRows, err := db.Query("SELECT id, job_id, path, COALESCE(exclusion_rules, '') FROM sources ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer srcRows.Close()
	sources := map[int64][]models.Source{}
	for srcRows.Next() {
		var src models.Source
		if err := srcRows.Scan(&src.ID, &src.JobID, &src.Path, &src.ExclusionRules); err != nil {
			return nil, err
		}
		sources[src.JobID] = append(sources[src.JobID], src)
	}

	dests, err := GetAllDestinations(db)
	if err != nil {
		return nil, err
	}
	destinations := map[int64][]models.Destination{}
	for _, d := range dests {
		destinations[d.JobID] = append(destinations[d.JobID], d)
	}

	for i := range jobs {
		jobs[i].Sources = sources[jobs[i].ID]
		jobs[i].Destinations = destinations[jobs[i].ID]
	}
	return jobs, nil
}

// GetAllDestinations returns the Destinations of every Job in a single query.
func GetAllDestinations(db *sql.DB) ([]models.Destination, error) {
	rows, err := db.Query("SELECT id, job_id, path FROM destinations ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dests := []models.Destination{}
	for rows.Next() {
		var d models.Destination
		if err := rows.Scan(&d.ID, &d.JobID, &d.Path); err != nil {
			return nil, err
		}
		dests = append(dests, d)
	}
	return dests, rows.Err()
}

func DeleteJob(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM jobs WHERE id = ?", id)
	return err
}

func GetAllRuns(db *sql.DB) ([]models.Run, error) {
	rows, err := db.Query("SELECT id, job_id, status, started_at, completed_at, COALESCE(duration_seconds, 0), COALESCE(bytes_transferred, 0), COALESCE(files_processed, 0) FROM runs ORDER BY started_at DESC LIMIT 100")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := []models.Run{}
	for rows.Next() {
		var r models.Run
		if err := rows.Scan(&r.ID, &r.JobID, &r.Status, &r.StartedAt, &r.CompletedAt, &r.DurationSeconds, &r.BytesTransferred, &r.FilesProcessed); err != nil {
			return nil, err
		}
		runs = append(runs, r)
	}
	return runs, nil
}

func GetRunByID(db *sql.DB, id int64) (*models.Run, error) {
	r := &models.Run{}
	row := db.QueryRow("SELECT id, job_id, status, started_at, completed_at, COALESCE(duration_seconds, 0), COALESCE(bytes_transferred, 0), COALESCE(files_processed, 0) FROM runs WHERE id = ?", id)
	err := row.Scan(&r.ID, &r.JobID, &r.Status, &r.StartedAt, &r.CompletedAt, &r.DurationSeconds, &r.BytesTransferred, &r.FilesProcessed)
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

	logs := []models.Log{}
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
		UPDATE jobs SET name = ?, description = ?, storage_strategy = ?, retention_policy = ?, retry_count = ?, retry_wait = ?, log_output = ?, schedule = ?, verify_integrity = ?, sync_deletions = ?
		WHERE id = ?`,
		job.Name, job.Description, job.StorageStrategy, job.RetentionPolicy, job.RetryCount, job.RetryWait, job.LogOutput, job.Schedule, job.VerifyIntegrity, job.SyncDeletions, job.ID,
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
