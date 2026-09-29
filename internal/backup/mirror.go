package backup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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
	db          *sql.DB
	runID       int64
	logCallback func(level, msg string)

	verify        bool // set from the Job: compare size and SHA-256 after each copy
	syncDeletions bool // set from the Job: remove Destination files gone from the Source
	mismatches    int
	failedFiles   int // files that could not be copied after all retries
	skipped       int // files already up to date in the Destination
	deleted       int // Destination items removed by syncDeletions
}

func NewDateStampedMirroring(db *sql.DB, runID int64, cb func(level, msg string)) *DateStampedMirroring {
	return &DateStampedMirroring{db: db, runID: runID, logCallback: cb}
}

func (s *DateStampedMirroring) Run(ctx context.Context, job *models.Job) (*models.Run, error) {
	result := &models.Run{
		Status: "success",
	}
	s.verify, s.syncDeletions = job.VerifyIntegrity, job.SyncDeletions
	s.mismatches, s.failedFiles, s.skipped, s.deleted = 0, 0, 0, 0

	for _, dst := range job.Destinations {
		if ctx.Err() != nil {
			return result, nil
		}
		resolvedDst, err := pathresolver.Resolve(dst.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve destination %s: %w", dst.Path, err)
		}

		if err := os.MkdirAll(resolvedDst, 0755); err != nil {
			return nil, fmt.Errorf("failed to create destination %s: %w", resolvedDst, err)
		}

		usedNames := map[string]bool{}
		for _, src := range job.Sources {
			if ctx.Err() != nil {
				return result, nil
			}
			resolvedSrc, err := pathresolver.Resolve(src.Path)
			if err != nil {
				s.log("ERROR", fmt.Errorf("Failed to resolve source %s: %w", src.Path, err))
				result.Status = "warning"
				continue
			}

			// One Source keeps the flat layout; several get a subfolder each so
			// same-named files from different Sources never overwrite each other.
			target := resolvedDst
			if len(job.Sources) > 1 {
				target = filepath.Join(resolvedDst, subfolderName(resolvedSrc, usedNames))
			}

			bytes, files, err := s.mirror(ctx, resolvedSrc, target, src.ExclusionRules, job.RetryCount, job.RetryWait)
			result.BytesTransferred += bytes
			result.FilesProcessed += files
			if err != nil && ctx.Err() == nil {
				s.log("ERROR", fmt.Errorf("Mirroring failed for %s -> %s: %w", resolvedSrc, target, err))
				result.Status = "failed"
				// We don't return early to allow other sources/destinations to try
			}
		}
	}

	if (s.mismatches > 0 || s.failedFiles > 0) && result.Status == "success" {
		result.Status = "warning"
	}
	s.log("INFO", fmt.Errorf("Summary: %d copied, %d skipped (unchanged), %d deleted", result.FilesProcessed, s.skipped, s.deleted))
	return result, nil
}

// subfolderName names the Destination subfolder of a Source after its last path
// segment, adding -2, -3... when two Sources would collide.
func subfolderName(src string, used map[string]bool) string {
	base := strings.Trim(filepath.Base(filepath.Clean(src)), `\/:.`)
	if base == "" {
		base = "source"
	}
	name := base
	for i := 2; used[strings.ToLower(name)]; i++ {
		name = fmt.Sprintf("%s-%d", base, i)
	}
	used[strings.ToLower(name)] = true
	return name
}

func (s *DateStampedMirroring) mirror(ctx context.Context, src, dst, exclusions string, retryCount, retryWait int) (int64, int64, error) {
	var totalBytes int64
	var totalFiles int64
	failedBefore := s.failedFiles
	seen := map[string]bool{}    // relative paths of Source files (copied or unchanged)
	srcDirs := map[string]bool{} // relative paths of Source directories

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
			if matchesName(excDirs, d.Name(), false) {
				return filepath.SkipDir
			}
			srcDirs[relPath] = true
			return nil
		}

		// Check file exclusions
		if matchesName(excFiles, d.Name(), true) {
			return nil
		}
		seen[relPath] = true

		targetPath := filepath.Join(dst, relPath)
		info, err := d.Info()
		if err != nil {
			s.failedFiles++
			s.log("WARNING", fmt.Errorf("Failed to read %s: %w", path, err))
			return nil
		}
		if unchanged(info, targetPath) {
			s.skipped++
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		copiedBytes, err := s.copyFileWithRetry(ctx, path, targetPath, retryCount, retryWait)
		if err != nil {
			s.failedFiles++
			s.log("WARNING", fmt.Errorf("Failed to copy %s: %w", path, err))
			// Continue with next files
		} else {
			totalBytes += copiedBytes
			totalFiles++
			s.log("INFO", fmt.Errorf("Copied %s -> %s (%d bytes)", path, targetPath, copiedBytes))
			if s.verify {
				if err := verifyCopy(path, targetPath); err != nil {
					s.mismatches++
					s.log("ERROR", fmt.Errorf("Integrity check failed for %s: %w", targetPath, err))
				}
			}
		}

		return nil
	})

	if err == nil && s.syncDeletions {
		switch {
		case s.failedFiles > failedBefore:
			s.log("WARNING", fmt.Errorf("Deletions skipped for %s: some files failed to copy", dst))
		case len(seen) == 0:
			s.log("WARNING", fmt.Errorf("Deletions skipped for %s: source %s is empty", dst, src))
		default:
			s.deleteExtraneous(dst, seen, srcDirs, excDirs, excFiles)
		}
	}

	return totalBytes, totalFiles, err
}

// unchanged reports whether dst already holds the same file as src: same size
// and modification time (2s tolerance for FAT/exFAT drives).
func unchanged(src fs.FileInfo, dst string) bool {
	di, err := os.Stat(dst)
	if err != nil || di.IsDir() || di.Size() != src.Size() {
		return false
	}
	diff := src.ModTime().Sub(di.ModTime())
	if diff < 0 {
		diff = -diff
	}
	return diff <= 2*time.Second
}

// deleteExtraneous removes Destination files and directories that are not in
// the Source. Items covered by Exclusion Rules are never touched.
func (s *DateStampedMirroring) deleteExtraneous(dst string, seen, srcDirs map[string]bool, excDirs, excFiles []string) {
	_ = filepath.WalkDir(dst, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, relErr := filepath.Rel(dst, path)
		if relErr != nil || rel == "." {
			return nil
		}

		if d.IsDir() {
			if matchesName(excDirs, d.Name(), false) {
				return filepath.SkipDir
			}
			if !srcDirs[rel] {
				s.remove(path, "directory", os.RemoveAll)
				return filepath.SkipDir
			}
			return nil
		}

		if !matchesName(excFiles, d.Name(), true) && !seen[rel] {
			s.remove(path, "file", os.Remove)
		}
		return nil
	})
}

func (s *DateStampedMirroring) remove(path, kind string, rm func(string) error) {
	if err := rm(path); err != nil {
		s.log("WARNING", fmt.Errorf("Failed to delete %s %s: %w", kind, path, err))
		return
	}
	s.deleted++
	s.log("INFO", fmt.Errorf("Deleted %s %s (no longer in source)", kind, path))
}

// matchesName reports whether name equals one of the rules (or, for files, matches it as a glob).
func matchesName(rules []string, name string, glob bool) bool {
	for _, r := range rules {
		if glob {
			if ok, _ := filepath.Match(r, name); ok {
				return true
			}
		} else if name == r {
			return true
		}
	}
	return false
}

func (s *DateStampedMirroring) copyFileWithRetry(ctx context.Context, src, dst string, retries, waitSecs int) (int64, error) {
	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			case <-time.After(time.Duration(waitSecs) * time.Second):
			}
		}

		bytes, err := copyFile(ctx, src, dst)
		if err == nil {
			return bytes, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			break
		}
	}
	return 0, lastErr
}

// ctxReader makes a copy in progress stop as soon as the Run is cancelled.
type ctxReader struct {
	ctx context.Context
	r   io.Reader
}

func (c ctxReader) Read(p []byte) (int, error) {
	if err := c.ctx.Err(); err != nil {
		return 0, err
	}
	return c.r.Read(p)
}

// copyFile copies src to dst and gives dst the modification time of src, which
// is what unchanged() compares on the next Run. A cancelled copy leaves no partial file.
func copyFile(ctx context.Context, src, dst string) (int64, error) {
	in, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return 0, err
	}

	out, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	n, err := io.Copy(out, ctxReader{ctx, in})
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		if ctx.Err() != nil {
			os.Remove(dst)
		}
		return n, err
	}
	return n, os.Chtimes(dst, info.ModTime(), info.ModTime())
}

// verifyCopy returns an error unless src and dst have the same size and SHA-256.
func verifyCopy(src, dst string) error {
	srcSum, srcSize, err := hashFile(src)
	if err != nil {
		return err
	}
	dstSum, dstSize, err := hashFile(dst)
	if err != nil {
		return err
	}
	if srcSize != dstSize {
		return fmt.Errorf("size mismatch: source %d bytes, copy %d bytes", srcSize, dstSize)
	}
	if srcSum != dstSum {
		return fmt.Errorf("SHA-256 mismatch: source %s, copy %s", srcSum, dstSum)
	}
	return nil
}

func hashFile(path string) (sum string, size int64, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	h := sha256.New()
	size, err = io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil)), size, err
}

// parseExclusions reads "DIR:a,b;FILE:*.x,*.y". Items without a prefix belong to
// the most recent DIR:/FILE: group, so "DIR:a,FILE:*.x" also works.
func parseExclusions(excl string) ([]string, []string) {
	var dirs, files []string
	isDir := false
	for _, item := range strings.FieldsFunc(excl, func(r rune) bool { return r == ';' || r == ',' }) {
		item = strings.TrimSpace(item)
		switch {
		case strings.HasPrefix(item, "DIR:"):
			isDir, item = true, strings.TrimPrefix(item, "DIR:")
		case strings.HasPrefix(item, "FILE:"):
			isDir, item = false, strings.TrimPrefix(item, "FILE:")
		}
		if item == "" {
			continue
		}
		if isDir {
			dirs = append(dirs, item)
		} else {
			files = append(files, item)
		}
	}
	return dirs, files
}

func (s *DateStampedMirroring) log(level string, err error) {
	msg := err.Error()
	if s.logCallback != nil {
		s.logCallback(level, msg)
	}
	_ = database.SaveLog(s.db, &models.Log{
		RunID:     s.runID,
		Level:     level,
		Message:   msg,
		CreatedAt: time.Now(),
	})
}
