package pathresolver

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Now is a variable to allow mocking in tests
var Now = time.Now

// translateFormat converts simple date format strings to Go's time layout.
func translateFormat(f string) string {
	f = strings.ReplaceAll(f, "YYYY", "2006")
	f = strings.ReplaceAll(f, "YY", "06")
	f = strings.ReplaceAll(f, "MM", "01")
	f = strings.ReplaceAll(f, "DD", "02")
	return f
}

// Resolve takes a path pattern and replaces variables with their runtime values.
func Resolve(pattern string) (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	re := regexp.MustCompile(`\{([^}]+)\}`)

	var resolveErr error

	result := re.ReplaceAllStringFunc(pattern, func(match string) string {
		if resolveErr != nil {
			return match // Stop processing if we already have an error
		}

		inner := match[1 : len(match)-1]
		parts := strings.Split(inner, ":")
		cmd := parts[0]

		if cmd == "hostname" {
			return hostname
		}

		formatStr := "YYYY-MM-DD" // Default format
		var offset int = 0

		if cmd == "today" {
			if len(parts) > 1 {
				formatStr = strings.Join(parts[1:], ":")
			}
			offset = 0
		} else if cmd == "yesterday" {
			if len(parts) > 1 {
				formatStr = strings.Join(parts[1:], ":")
			}
			offset = -1
		} else if cmd == "offset" {
			if len(parts) < 2 {
				resolveErr = fmt.Errorf("offset requires a value: %s", match)
				return match
			}
			val, err := strconv.Atoi(parts[1])
			if err != nil {
				resolveErr = fmt.Errorf("invalid offset value in %s: %w", match, err)
				return match
			}
			offset = val
			if len(parts) > 2 {
				formatStr = strings.Join(parts[2:], ":")
			}
		} else {
			resolveErr = fmt.Errorf("unknown template variable: %s", cmd)
			return match
		}

		targetTime := Now().AddDate(0, 0, offset)
		goFormat := translateFormat(formatStr)

		return targetTime.Format(goFormat)
	})

	if resolveErr != nil {
		return "", resolveErr
	}

	return result, nil
}
