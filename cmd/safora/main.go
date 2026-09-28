package main

import (
	"fmt"
	"os"

	"safora/internal/database"
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
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
