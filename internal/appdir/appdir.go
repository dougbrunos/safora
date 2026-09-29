// Package appdir decides where Safora keeps its database and API token (ADR 0004).
package appdir

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Resolve returns the data directory, creating it if needed:
//   - SAFORA_DATA_DIR, when set;
//   - the executable's directory, in portable mode (--portable);
//   - otherwise the system location: %ProgramData%\Safora on Windows,
//     /Library/Application Support/Safora on macOS, /var/lib/safora elsewhere.
func Resolve(portable bool) (string, error) {
	dir := os.Getenv("SAFORA_DATA_DIR")
	switch {
	case dir != "":
	case portable:
		exe, err := os.Executable()
		if err != nil {
			return "", err
		}
		dir = filepath.Dir(exe)
	default:
		dir = systemDir()
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("cannot create data directory %s: %w (run as administrator/root, or use --portable)", dir, err)
	}
	return dir, nil
}

func systemDir() string {
	switch runtime.GOOS {
	case "windows":
		base := os.Getenv("ProgramData")
		if base == "" {
			base = `C:\ProgramData`
		}
		return filepath.Join(base, "Safora")
	case "darwin":
		return "/Library/Application Support/Safora"
	default:
		return "/var/lib/safora"
	}
}
