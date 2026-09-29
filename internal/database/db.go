package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

// pragmas are applied on every new connection.
const pragmas = "_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"

// InitDB initializes the database and runs migrations.
func InitDB(dataSourceName string) (*sql.DB, error) {
	sep := "?"
	if strings.Contains(dataSourceName, "?") {
		sep = "&"
	}
	db, err := sql.Open("sqlite", dataSourceName+sep+pragmas)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// migrations are applied in order; the index+1 is the schema version stored in
// PRAGMA user_version. Never edit an applied migration, append a new one.
var migrations = []string{
	// 1: initial schema (idempotent so pre-versioning databases adopt it as-is)
	`CREATE TABLE IF NOT EXISTS jobs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		storage_strategy TEXT NOT NULL,
		retention_policy TEXT,
		retry_count INTEGER DEFAULT 0,
		retry_wait INTEGER DEFAULT 0,
		log_output TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS sources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		job_id INTEGER NOT NULL,
		path TEXT NOT NULL,
		exclusion_rules TEXT,
		FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS destinations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		job_id INTEGER NOT NULL,
		path TEXT NOT NULL,
		FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS runs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		job_id INTEGER NOT NULL,
		status TEXT NOT NULL,
		started_at DATETIME NOT NULL,
		completed_at DATETIME,
		duration_seconds INTEGER,
		bytes_transferred INTEGER,
		FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id INTEGER NOT NULL,
		level TEXT NOT NULL,
		message TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (run_id) REFERENCES runs(id) ON DELETE CASCADE
	);`,
	// 2: persist files processed per run
	`ALTER TABLE runs ADD COLUMN files_processed INTEGER DEFAULT 0;`,
	// 3: per-job cron schedule (empty = manual / drive triggers only)
	`ALTER TABLE jobs ADD COLUMN schedule TEXT NOT NULL DEFAULT '';`,
	// 4: opt-in post-copy integrity verification
	`ALTER TABLE jobs ADD COLUMN verify_integrity INTEGER NOT NULL DEFAULT 0;`,
	// 5: opt-in removal from the Destination of files deleted from the Source
	`ALTER TABLE jobs ADD COLUMN sync_deletions INTEGER NOT NULL DEFAULT 0;`,
}

func migrate(db *sql.DB) error {
	// A single connection keeps user_version reads/writes consistent with the DDL.
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	var version int
	if err := conn.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		return err
	}

	for v := version; v < len(migrations); v++ {
		tx, err := conn.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(migrations[v]); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d: %w", v+1, err)
		}
		// PRAGMA does not accept bound parameters.
		if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", v+1)); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
