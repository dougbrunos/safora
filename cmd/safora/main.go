package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"safora/internal/api"
	"safora/internal/appdir"
	"safora/internal/backup"
	"safora/internal/daemon"
	"safora/internal/database"
	"safora/internal/importer"
	"safora/internal/pathresolver"
)

// version is set at build time: -ldflags "-X main.version=1.2.3".
var version = "dev"

var (
	portable bool
	dataDir  string
)

// dataPaths resolves the data directory once and returns the database and token paths.
func dataPaths() (db, token string) {
	if dataDir == "" {
		dir, err := appdir.Resolve(portable)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		dataDir = dir
		if _, err := os.Stat(filepath.Join(dir, "safora.db")); os.IsNotExist(err) {
			if _, err := os.Stat("safora.db"); err == nil {
				fmt.Fprintf(os.Stderr, "Notice: found safora.db in the current directory, but data now lives in %s.\nUse --portable to keep using the local file, or copy it to that directory.\n", dir)
			}
		}
	}
	return filepath.Join(dataDir, "safora.db"), filepath.Join(dataDir, api.TokenFile)
}

func main() {
	// --portable may appear anywhere; strip it before the commands parse positional arguments.
	args := []string{os.Args[0]}
	for _, a := range os.Args[1:] {
		if a == "--portable" {
			portable = true
			continue
		}
		args = append(args, a)
	}
	os.Args = args

	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  safora template resolve \"<pattern>\"")
		fmt.Println("  safora initdb <datasource>")
		fmt.Println("  safora job import <file.bat>")
		fmt.Println("  safora run <job-id>")
		fmt.Println("  safora serve [port]")
		fmt.Println("  safora service [install | uninstall | start | stop | status | run]")
		fmt.Println("  safora version")
		fmt.Println("Add --portable to keep data next to the executable instead of the system data directory.")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "version":
		fmt.Println("safora", version)
	case "template":
		if len(os.Args) < 4 || os.Args[2] != "resolve" {
			fmt.Println("Usage: safora template resolve \"<pattern>\"")
			os.Exit(1)
		}
		pattern := os.Args[3]
		resolved, err := pathresolver.Resolve(pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving pattern: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(resolved)
	case "initdb":
		if len(os.Args) < 3 {
			fmt.Println("Usage: safora initdb <datasource>")
			os.Exit(1)
		}
		dsn := os.Args[2]
		_, err := database.InitDB(dsn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Database initialized successfully.")
	case "job":
		if len(os.Args) < 4 || os.Args[2] != "import" {
			fmt.Println("Usage: safora job import <file.bat>")
			os.Exit(1)
		}
		batFile := os.Args[3]
		job, err := importer.ParseBatchScript(batFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing script: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Parsed Job: %+v\n", job)

		// Persist to SQLite
		db, err := database.InitDB(mustDB())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}

		if err := database.SaveJob(db, job); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving job: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully imported job ID: %d\n", job.ID)
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Usage: safora run <job-id>")
			os.Exit(1)
		}
		jobID, err := strconv.ParseInt(os.Args[2], 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid job ID: %v\n", err)
			os.Exit(1)
		}

		db, err := database.InitDB(mustDB())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}

		job, err := database.GetJobByID(db, jobID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting job: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("▶ Starting Backup Job: %s (ID: %d)\n", job.Name, job.ID)
		fmt.Println("--------------------------------------------------")

		eng := backup.NewDefaultEngine(db)
		run, err := eng.Run(context.Background(), job)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Job execution failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("--------------------------------------------------")
		fmt.Printf("✓ Backup Completed: %s\n", run.Status)
		fmt.Printf("  Files Copied: %d\n", run.FilesProcessed)
		fmt.Printf("  Transferred:  %d bytes\n", run.BytesTransferred)
		fmt.Printf("  Duration:     %ds\n", run.DurationSeconds)
	case "serve":
		db, err := database.InitDB(mustDB())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}

		if _, err := database.RecoverOrphanedRuns(db); err != nil {
			fmt.Fprintf(os.Stderr, "Error recovering interrupted runs: %v\n", err)
			os.Exit(1)
		}

		server, err := api.NewServer(db, mustToken())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating server: %v\n", err)
			os.Exit(1)
		}
		port := "3434"
		if len(os.Args) >= 3 {
			port = os.Args[2]
		}

		addr := "127.0.0.1:" + port
		fmt.Printf("Starting Safora API server on %s\n", addr)
		if err := server.Start(addr); err != nil {
			fmt.Fprintf(os.Stderr, "Server failed: %v\n", err)
			os.Exit(1)
		}
	case "service":
		if len(os.Args) < 3 {
			fmt.Println("Usage: safora service [install | uninstall | start | stop | status | run]")
			os.Exit(1)
		}
		action := os.Args[2]

		db, err := database.InitDB(mustDB())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}

		if err := daemon.ManageService(action, db, mustToken(), portable); err != nil {
			fmt.Fprintf(os.Stderr, "Service error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func mustDB() string    { db, _ := dataPaths(); return db }
func mustToken() string { _, token := dataPaths(); return token }
