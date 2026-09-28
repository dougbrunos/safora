package backup

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"safora/internal/database"
	"safora/internal/models"
	"safora/internal/pathresolver"
)

type DateStampedMirroring struct {
	db    *sql.DB
	runID int64
}

func NewDateStampedMirroring(db *sql.DB, runID int64) *DateStampedMirroring {
	return &DateStampedMirroring{db: db, runID: runID}
}

func (s *DateStampedMirroring) Run(ctx context.Context, job *models.Job) (*models.Run, error) {
	result := &models.Run{
		Status: "success",
	}

	for _, dst := range job.Destinations {
		resolvedDst, err := pathresolver.Resolve(dst.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve destination %s: %w", dst.Path, err)
		}

		if err := os.MkdirAll(resolvedDst, 0755); err != nil {
			return nil, fmt.Errorf("failed to create destination %s: %w", resolvedDst, err)
		}

		for _, src := range job.Sources {
			resolvedSrc, err := pathresolver.Resolve(src.Path)
			if err != nil {
				s.log("ERROR", fmt.Errorf("Failed to resolve source %s: %w", src.Path, err))
				result.Status = "warning"
				continue
			}

			// Actually run the copy logic
			bytes, files, err := s.mirror(ctx, resolvedSrc, resolvedDst, src.ExclusionRules, job.RetryCount, job.RetryWait)
			result.BytesTransferred += bytes
			result.FilesProcessed += files
			if err != nil {
				s.log("ERROR", fmt.Errorf("Mirroring failed for %s -> %s: %w", resolvedSrc, resolvedDst, err))
				result.Status = "failed"
				// We don't return early to allow other sources/destinations to try
			}
		}
	}

	return result, nil
}

func (s *DateStampedMirroring) mirror(ctx context.Context, src, dst, exclusions string, retryCount, retryWait int) (int64, int64, error) {
	var totalBytes int64
	var totalFiles int64

	excDirs, excFiles := parseExclusions(exclusions)

	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		
		if relPath == "." {
			return nil
		}

		// Check directory exclusions
		if d.IsDir() {
			for _, ed := range excDirs {
				if d.Name() == ed {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Check file exclusions
		for _, ef := range excFiles {
			matched, _ := filepath.Match(ef, d.Name())
			if matched {
				return nil
			}
		}

		targetPath := filepath.Join(dst, relPath)
		targetDir := filepath.Dir(targetPath)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return err
		}

		copiedBytes, err := s.copyFileWithRetry(path, targetPath, retryCount, retryWait)
		if err != nil {
			s.log("WARNING", fmt.Errorf("Failed to copy %s: %w", path, err))
			// Continue with next files
		} else {
			totalBytes += copiedBytes
			totalFiles++
			s.log("INFO", fmt.Errorf("Copied %s -> %s (%d bytes)", path, targetPath, copiedBytes))
		}

		return nil
	})

	return totalBytes, totalFiles, err
}

func (s *DateStampedMirroring) copyFileWithRetry(src, dst string, retries, waitSecs int) (int64, error) {
	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(waitSecs) * time.Second)
		}
		
		bytes, err := copyFile(src, dst)
		if err == nil {
			return bytes, nil
		}
		lastErr = err
	}
	return 0, lastErr
}

func copyFile(src, dst string) (int64, error) {
	sourceFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destFile.Close()

	return io.Copy(destFile, sourceFile)
}

func parseExclusions(excl string) ([]string, []string) {
	var dirs, files []string
	if excl == "" {
		return dirs, files
	}
	
	parts := strings.Split(excl, ";")
	for _, p := range parts {
		if strings.HasPrefix(p, "DIR:") {
			dirs = append(dirs, strings.Split(strings.TrimPrefix(p, "DIR:"), ",")...)
		} else if strings.HasPrefix(p, "FILE:") {
			files = append(files, strings.Split(strings.TrimPrefix(p, "FILE:"), ",")...)
		}
	}
	return dirs, files
}

func (s *DateStampedMirroring) log(level string, err error) {
	_ = database.SaveLog(s.db, &models.Log{
		RunID:     s.runID,
		Level:     level,
		Message:   err.Error(),
		CreatedAt: time.Now(),
	})
}
