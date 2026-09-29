package importer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseBatchScript(t *testing.T) {
	fixturePath := filepath.Join("..", "..", ".scratch", "mvp-foundation", "fixtures", "sample.bat")

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

	expectedExclusions := "DIR:123LAUDOS123,temp;FILE:*.tmp,*.bak"
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
