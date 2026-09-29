package retention

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"safora/internal/models"
)

var ErrLocked = errors.New("retention lock active: last run failed")

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

// Prune analyzes the destination and removes old backups based on the retention policy.
func (e *Engine) Prune(ctx context.Context, job *models.Job, lastRunStatus string) ([]string, error) {
	// Retention Lock Rule 1: Do not prune if the current run failed
	if lastRunStatus == "failed" || lastRunStatus == "cancelled" {
		return nil, ErrLocked
	}

	if job.RetentionPolicy == "" {
		return nil, nil // No policy defined
	}

	var prunedPaths []string

	for _, dest := range job.Destinations {
		// D:\Backups\Daily\{yesterday:DD-MM-YY} -> Parent is D:\Backups\Daily
		idx := strings.Index(dest.Path, "{")
		if idx == -1 {
			// Not a templated path, cannot do automatic retention reliably
			continue
		}

		parentDir := dest.Path[:idx]

		if _, err := os.Stat(parentDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(parentDir)
		if err != nil {
			return prunedPaths, err
		}

		var candidates []os.FileInfo
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			info, err := entry.Info()
			if err == nil {
				candidates = append(candidates, info)
			}
		}

		// Sort candidates by modification time descending (newest first)
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].ModTime().After(candidates[j].ModTime())
		})

		// Retention Lock Rule 2: We must have at least one valid copy remaining
		if len(candidates) <= 1 {
			continue
		}

		toDelete := e.evaluatePolicy(job.RetentionPolicy, candidates)

		// Filter out the newest copy from toDelete to enforce Retention Lock Rule 2
		newest := candidates[0].Name()
		var finalDeletes []string
		for _, del := range toDelete {
			if del != newest {
				finalDeletes = append(finalDeletes, del)
			}
		}

		for _, name := range finalDeletes {
			fullPath := filepath.Join(parentDir, name)
			if err := os.RemoveAll(fullPath); err == nil {
				prunedPaths = append(prunedPaths, fullPath)
			}
		}
	}

	return prunedPaths, nil
}

func (e *Engine) evaluatePolicy(policy string, candidates []os.FileInfo) []string {
	var toDelete []string

	parts := strings.Fields(strings.ToLower(policy))
	if len(parts) == 2 {
		parts = append(parts, "runs") // legacy "KEEP 5" saved by older dashboards
	}
	if len(parts) >= 3 && parts[0] == "keep" {
		val, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil
		}

		unit := parts[2]
		if unit == "runs" || unit == "copies" {
			if len(candidates) > val {
				for i := val; i < len(candidates); i++ {
					toDelete = append(toDelete, candidates[i].Name())
				}
			}
		} else if unit == "days" {
			cutoff := time.Now().AddDate(0, 0, -val)
			for _, c := range candidates {
				if c.ModTime().Before(cutoff) {
					toDelete = append(toDelete, c.Name())
				}
			}
		}
	}
	return toDelete
}
