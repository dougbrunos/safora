package importer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseBatchScript(t *testing.T) {
	fixturePath := filepath.Join("testdata", "sample.bat")

	// Ensure the fixture exists
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skipf("Fixture not found at %s", fixturePath)
	}

	job, err := ParseBatchScript(fixturePath)
	if err != nil {
		t.Fatalf("Failed to parse script: %v", err)
	}

	if job.Name != "Imported Job" {
		t.Errorf("Expected Name 'Imported Job', got '%s'", job.Name)
	}

	if job.RetryCount != 5 {
		t.Errorf("Expected RetryCount 5, got %d", job.RetryCount)
	}

	if job.RetryWait != 5 {
		t.Errorf("Expected RetryWait 5, got %d", job.RetryWait)
	}

	if job.LogOutput != "C:\\logs\\backup.log" {
		t.Errorf("Expected LogOutput 'C:\\logs\\backup.log', got '%s'", job.LogOutput)
	}

	if len(job.Sources) != 1 {
		t.Fatalf("Expected 1 Source, got %d", len(job.Sources))
	}

	src := job.Sources[0]
	expectedSrcPath := "C:\\Data\\Production\\{yesterday:DD-MM-YY}"
	if src.Path != expectedSrcPath {
		t.Errorf("Expected Source Path '%s', got '%s'", expectedSrcPath, src.Path)
	}

	expectedExclusions := "DIR:Cache,temp;FILE:*.tmp,*.bak"
	if src.ExclusionRules != expectedExclusions {
		t.Errorf("Expected ExclusionRules '%s', got '%s'", expectedExclusions, src.ExclusionRules)
	}

	if len(job.Destinations) != 1 {
		t.Fatalf("Expected 1 Destination, got %d", len(job.Destinations))
	}

	dst := job.Destinations[0]
	expectedDstPath := "D:\\Backups\\Daily\\{yesterday:DD-MM-YY}"
	if dst.Path != expectedDstPath {
		t.Errorf("Expected Destination Path '%s', got '%s'", expectedDstPath, dst.Path)
	}
}

func TestParseBatchScript_MirrorFlagTurnsOnSyncDeletions(t *testing.T) {
	job, err := ParseBatchScript(filepath.Join("testdata", "sample.bat")) // uses /MIR
	if err != nil {
		t.Fatal(err)
	}
	if !job.SyncDeletions {
		t.Error("/MIR means the destination mirrors the source: SyncDeletions should be on")
	}
	if job.RetentionPolicy != "" {
		t.Errorf("an imported job must not start deleting old copies, got %q", job.RetentionPolicy)
	}
}

// date_parts.bat builds the date from separate DD, MM and YY variables (four-digit
// year), keeps values with spaces in variables and does not use /MIR.
func TestParseBatchScript_DatePartsInSeparateVariables(t *testing.T) {
	job, err := ParseBatchScript(filepath.Join("testdata", "date_parts.bat"))
	if err != nil {
		t.Fatal(err)
	}

	if len(job.Sources) != 1 || len(job.Destinations) != 1 {
		t.Fatalf("want 1 source and 1 destination, got %d and %d", len(job.Sources), len(job.Destinations))
	}
	if got, want := job.Sources[0].Path, `D:\BACKUPCHAPECO\Incremental\{yesterday:DD-MM-YYYY}`; got != want {
		t.Errorf("source = %q, want %q", got, want)
	}
	if got, want := job.Destinations[0].Path, `F:\Backup Incremental CH\{yesterday:DD-MM-YYYY}\{yesterday:DD-MM-YYYY}CH`; got != want {
		t.Errorf("destination = %q, want %q", got, want)
	}
	if got, want := job.Sources[0].ExclusionRules, "DIR:Cache"; got != want {
		t.Errorf("exclusions = %q, want %q", got, want)
	}
	if job.RetryCount != 5 || job.RetryWait != 5 {
		t.Errorf("retries = %d/%d, want 5/5", job.RetryCount, job.RetryWait)
	}
	if job.SyncDeletions {
		t.Error("/E alone must not turn on Mirror deletions")
	}
}

func TestParseBatchScript_LineContinuationAndOffset(t *testing.T) {
	script := filepath.Join(t.TempDir(), "s.bat")
	content := "set qty=-7\r\n" +
		"echo x = DateAdd(\"d\",%qty%,Date)\r\n" +
		"set \"DD=%%a\"\r\n" +
		"robocopy C:\\src\\%DD% D:\\dst ^\r\n  /XD tmp /R:2\r\n"
	if err := os.WriteFile(script, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	job, err := ParseBatchScript(script)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := job.Sources[0].Path, `C:\src\{offset:-7:DD}`; got != want {
		t.Errorf("source = %q, want %q", got, want)
	}
	if job.Sources[0].ExclusionRules != "DIR:tmp" || job.RetryCount != 2 {
		t.Errorf("continued line not parsed: %+v retries=%d", job.Sources[0], job.RetryCount)
	}
}

func TestParseBatchScript_NoRobocopy(t *testing.T) {
	script := filepath.Join(t.TempDir(), "s.bat")
	os.WriteFile(script, []byte("@echo off\r\necho hello\r\n"), 0o644)
	if _, err := ParseBatchScript(script); err == nil {
		t.Error("a script without robocopy must be an error")
	}
}
