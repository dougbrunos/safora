package importer

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"safora/internal/models"
)

var (
	// A run of date parts joined by at most one separator: %DD%-%MM%-%YY%.
	dateRunRe = regexp.MustCompile(`(?i)%(?:DD|MM|YYYY|YY)%(?:[^%\\/:*?"<>| ]?%(?:DD|MM|YYYY|YY)%)*`)
	dateTokRe = regexp.MustCompile(`(?i)%(DD|MM|YYYY|YY)%`)
	// DateAdd("d", -1, Date): how many days the script shifts the date by.
	dateAddRe = regexp.MustCompile(`(?i)DateAdd\(\s*"d"\s*,\s*(-?\d+)`)
	varRe     = regexp.MustCompile(`%([A-Za-z0-9_]+)%`)
)

func isDatePart(name string) bool {
	switch name {
	case "DD", "MM", "YY", "YYYY":
		return true
	}
	return false
}

// ParseBatchScript reads a Windows batch script and attempts to extract a Safora Job configuration.
func ParseBatchScript(filePath string) (*models.Job, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open script: %w", err)
	}
	defer file.Close()

	job := &models.Job{
		Name:            "Imported Job",
		StorageStrategy: "Date-Stamped Mirroring",
		// Empty means "never delete": a script that does not prune old copies must not start doing so.
	}

	var (
		vars       = map[string]string{}
		dayOffset  int
		yearDigits = 2
		robocopy   []string
	)

	for _, line := range logicalLines(file) {
		lower := strings.ToLower(line)

		if m := dateAddRe.FindStringSubmatch(expand(line, vars)); m != nil {
			dayOffset, _ = strconv.Atoi(m[1])
		}

		switch {
		case strings.HasPrefix(lower, "set "):
			if name, value, ok := parseSet(line); ok {
				vars[name] = value
				// set "YY=%result:~0,4%" takes four characters: a four-digit year.
				if (name == "YY" || name == "YYYY") && strings.Contains(value, "~0,4") {
					yearDigits = 4
				}
			}
		case strings.HasPrefix(lower, "robocopy "):
			robocopy = append(robocopy, line)
		}
	}

	for _, cmd := range robocopy {
		cmd = withDateTemplates(expandCommand(cmd, vars), dayOffset, yearDigits)
		if err := parseRobocopyCommand(cmd, job); err != nil {
			return nil, err
		}
	}

	if len(job.Sources) == 0 {
		return nil, fmt.Errorf("no valid robocopy commands found to parse")
	}

	return job, nil
}

// logicalLines reads the script, joining lines that end with the ^ continuation character.
func logicalLines(f *os.File) []string {
	var lines []string
	var pending string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasSuffix(line, "^") {
			pending += strings.TrimSpace(strings.TrimSuffix(line, "^")) + " "
			continue
		}
		lines = append(lines, pending+line)
		pending = ""
	}
	if pending != "" {
		lines = append(lines, strings.TrimSpace(pending))
	}
	return lines
}

// parseSet reads `set NAME=value` and `set "NAME=value"`, ignoring set /a and set /p.
func parseSet(line string) (name, value string, ok bool) {
	rest := strings.TrimSpace(line[4:])
	if strings.HasPrefix(rest, "/") {
		return "", "", false
	}
	if len(rest) >= 2 && strings.HasPrefix(rest, `"`) && strings.HasSuffix(rest, `"`) {
		rest = rest[1 : len(rest)-1]
	}
	name, value, found := strings.Cut(rest, "=")
	if !found {
		return "", "", false
	}
	return strings.ToUpper(strings.TrimSpace(name)), strings.Trim(strings.TrimSpace(value), `"`), true
}

// expand substitutes %NAME% with the script's variables (recursively). Date parts
// (DD, MM, YY, YYYY) are left alone: they become Safora path templates later.
func expand(s string, vars map[string]string) string {
	for i := 0; i < 10; i++ {
		next := varRe.ReplaceAllStringFunc(s, func(tok string) string {
			name := strings.ToUpper(tok[1 : len(tok)-1])
			if v, ok := vars[name]; ok && !isDatePart(name) {
				return v
			}
			return tok
		})
		if next == s {
			break
		}
		s = next
	}
	return s
}

// expandCommand expands variables in a robocopy command line. Values containing
// spaces are quoted (unless already inside quotes) so the paths stay one argument.
func expandCommand(cmd string, vars map[string]string) string {
	var out strings.Builder
	last := 0
	for _, loc := range varRe.FindAllStringSubmatchIndex(cmd, -1) {
		name := strings.ToUpper(cmd[loc[2]:loc[3]])
		v, ok := vars[name]
		if !ok || isDatePart(name) {
			continue
		}
		out.WriteString(cmd[last:loc[0]])
		v = expand(v, vars)
		insideQuotes := strings.Count(out.String(), `"`)%2 == 1
		if strings.ContainsAny(v, " \t") && !insideQuotes {
			v = `"` + v + `"`
		}
		out.WriteString(v)
		last = loc[1]
	}
	out.WriteString(cmd[last:])
	return out.String()
}

// withDateTemplates turns runs of date parts into Safora templates:
// %DD%-%MM%-%YY% becomes {yesterday:DD-MM-YY} when the script shifted the date by -1 day.
func withDateTemplates(s string, dayOffset, yearDigits int) string {
	kind := "today"
	switch {
	case dayOffset == -1:
		kind = "yesterday"
	case dayOffset != 0:
		kind = "offset:" + strconv.Itoa(dayOffset)
	}
	return dateRunRe.ReplaceAllStringFunc(s, func(run string) string {
		format := dateTokRe.ReplaceAllStringFunc(run, func(tok string) string {
			part := strings.ToUpper(tok[1 : len(tok)-1])
			if part == "YY" && yearDigits == 4 {
				return "YYYY"
			}
			return part
		})
		return "{" + kind + ":" + format + "}"
	})
}

func parseRobocopyCommand(cmd string, job *models.Job) error {
	// Simple tokenization honoring quotes
	tokens := tokenize(cmd)
	if len(tokens) < 3 {
		return fmt.Errorf("invalid robocopy command: not enough arguments")
	}

	// robocopy <source> <dest> [options...]
	srcPath := tokens[1]
	dstPath := tokens[2]

	src := models.Source{
		Path: srcPath,
	}
	dst := models.Destination{
		Path: dstPath,
	}

	var xdArgs []string
	var xfArgs []string

	for i := 3; i < len(tokens); i++ {
		token := tokens[i]
		upperToken := strings.ToUpper(token)

		if upperToken == "/XD" {
			for j := i + 1; j < len(tokens); j++ {
				if strings.HasPrefix(tokens[j], "/") {
					break
				}
				xdArgs = append(xdArgs, tokens[j])
				i = j
			}
		} else if upperToken == "/XF" {
			for j := i + 1; j < len(tokens); j++ {
				if strings.HasPrefix(tokens[j], "/") {
					break
				}
				xfArgs = append(xfArgs, tokens[j])
				i = j
			}
		} else if upperToken == "/MIR" || upperToken == "/PURGE" {
			// robocopy deletes destination files that left the source.
			job.SyncDeletions = true
		} else if strings.HasPrefix(upperToken, "/R:") {
			val, _ := strconv.Atoi(strings.TrimPrefix(upperToken, "/R:"))
			job.RetryCount = val
		} else if strings.HasPrefix(upperToken, "/W:") {
			val, _ := strconv.Atoi(strings.TrimPrefix(upperToken, "/W:"))
			job.RetryWait = val
		} else if strings.HasPrefix(upperToken, "/LOG:") {
			job.LogOutput = strings.TrimPrefix(token, "/LOG:") // Keep original case for path
		} else if strings.HasPrefix(upperToken, "/LOG+:") {
			job.LogOutput = strings.TrimPrefix(token, "/LOG+:")
		}
	}

	// Combine exclusions
	var exclusions []string
	if len(xdArgs) > 0 {
		exclusions = append(exclusions, "DIR:"+strings.Join(xdArgs, ","))
	}
	if len(xfArgs) > 0 {
		exclusions = append(exclusions, "FILE:"+strings.Join(xfArgs, ","))
	}
	if len(exclusions) > 0 {
		src.ExclusionRules = strings.Join(exclusions, ";")
	}

	job.Sources = append(job.Sources, src)
	job.Destinations = append(job.Destinations, dst)

	return nil
}

// tokenize splits a string by spaces, but keeps quoted substrings intact.
func tokenize(s string) []string {
	var tokens []string
	var current strings.Builder
	inQuotes := false

	for _, char := range s {
		if char == '"' {
			inQuotes = !inQuotes
		} else if char == ' ' && !inQuotes {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(char)
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}
