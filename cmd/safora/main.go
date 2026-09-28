package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"safora/internal/backup"
	"safora/internal/database"
	"safora/internal/importer"
	"safora/internal/pathresolver"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  safora template resolve \"<pattern>\"")
		fmt.Println("  safora initdb <datasource>")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
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
		db, err := database.InitDB("safora.db")
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

		db, err := database.InitDB("safora.db")
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
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
