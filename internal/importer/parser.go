package importer

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"safora/internal/models"
)

// ParseBatchScript reads a Windows batch script and attempts to extract a Safora Job configuration.
func ParseBatchScript(filePath string) (*models.Job, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open script: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	
	job := &models.Job{
		Name:            "Imported Job",
		StorageStrategy: "Date-Stamped Mirroring",
		RetentionPolicy: "keep 30 days", // Default
	}

	var hasYesterday bool
	var dateFormat string
	var dynamicVarName string
	
	vars := make(map[string]string)

	robocopyCmds := []string{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		lineExpanded := line
		for k, v := range vars {
			lineExpanded = strings.ReplaceAll(lineExpanded, "%"+k+"%", v)
		}
		
		// Detect yesterday logic
		if strings.Contains(strings.ToLower(lineExpanded), "dateadd") && strings.Contains(lineExpanded, "-1") {
			hasYesterday = true
		}

		// Detect set VAR=VALUE
		if strings.HasPrefix(strings.ToLower(line), "set ") {
			parts := strings.SplitN(line[4:], "=", 2)
			if len(parts) == 2 {
				varName := parts[0]
				varVal := parts[1]
				
				// Keep track of all variables
				vars[varName] = varVal

				if strings.Contains(varVal, "%DD%") || strings.Contains(varVal, "%MM%") {
					dynamicVarName = varName
					dateFormat = strings.ReplaceAll(varVal, "%DD%", "DD")
					dateFormat = strings.ReplaceAll(dateFormat, "%MM%", "MM")
					dateFormat = strings.ReplaceAll(dateFormat, "%YY%", "YY")
					dateFormat = strings.ReplaceAll(dateFormat, "%YYYY%", "YYYY")
				}
			}
		}

		// Collect robocopy commands
		if strings.HasPrefix(strings.ToLower(line), "robocopy ") {
			// Resolve any local batch variables in the command line first if we can,
			// or at least replace the date variable with our Safora template.
			if dynamicVarName != "" {
				templateVar := "{today:" + dateFormat + "}"
				if hasYesterday {
					templateVar = "{yesterday:" + dateFormat + "}"
				}
				vars[dynamicVarName] = templateVar
			}
			
			// Simple variable expansion
			for k, v := range vars {
				// Also try expanding inside other variables recursively if needed, but simple is fine
				line = strings.ReplaceAll(line, "%"+k+"%", v)
			}
			
			// Run a second pass to expand nested variables like SOURCE which uses YESTERDAY
			for k, v := range vars {
				vExpanded := v
				for k2, v2 := range vars {
					vExpanded = strings.ReplaceAll(vExpanded, "%"+k2+"%", v2)
				}
				line = strings.ReplaceAll(line, "%"+k+"%", vExpanded)
			}

			robocopyCmds = append(robocopyCmds, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	for _, cmd := range robocopyCmds {
		if err := parseRobocopyCommand(cmd, job); err != nil {
			return nil, err
		}
	}

	if len(job.Sources) == 0 {
		return nil, fmt.Errorf("no valid robocopy commands found to parse")
	}

	return job, nil
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
