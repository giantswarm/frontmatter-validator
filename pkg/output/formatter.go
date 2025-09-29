package output

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/afero"

	"github.com/giantswarm/frontmatter-validator/pkg/validator"
)

// Formatter handles different output formats
type Formatter struct {
	checksMap map[string]validator.Check
}

// New creates a new Formatter instance
func New() *Formatter {
	checksMap := make(map[string]validator.Check)
	for _, check := range validator.GetChecks() {
		checksMap[check.ID] = check
	}

	return &Formatter{
		checksMap: checksMap,
	}
}

// PrintStdout prints validation results to stdout with colored output
func (f *Formatter) PrintStdout(results map[string]validator.ValidationResult) {
	nFails := 0
	nWarnings := 0

	for filePath, result := range results {
		fmt.Printf("\n%s\n", filePath)

		var fails []validator.CheckResult
		var warnings []validator.CheckResult

		for _, check := range result.Checks {
			if f.checksMap[check.Check].Severity == validator.SeverityFail {
				fails = append(fails, check)
				nFails++
			} else {
				warnings = append(warnings, check)
				nWarnings++
			}
		}

		// Print failures first
		for _, check := range fails {
			f.printCheckResult(check, validator.SeverityFail)
		}

		// Then warnings
		for _, check := range warnings {
			f.printCheckResult(check, validator.SeverityWarn)
		}
	}

	fmt.Println()
	if nFails > 0 {
		fmt.Printf("Found %d critical problems, marked with %s.\n", nFails, f.colorSeverity(validator.SeverityFail))
	}
	if nWarnings > 0 {
		fmt.Printf("Found %d less severe problems, marked with %s.\n", nWarnings, f.colorSeverity(validator.SeverityWarn))
	}
}

// PrintJSON prints validation results as JSON for issue tracking
func (f *Formatter) PrintJSON(results map[string]validator.ValidationResult) {
	var output []validator.JSONOutput

	for filePath, result := range results {
		for _, check := range result.Checks {
			title := check.Title
			if title == "" {
				continue
			}

			description := f.checksMap[check.Check].Description
			var owners []string

			if len(check.Owner) > 0 {
				for _, owner := range check.Owner {
					re := regexp.MustCompile(`/.*\/([^\/]+)\/?$`)
					matches := re.FindStringSubmatch(owner)
					if len(matches) > 1 {
						teamName := matches[1]
						teamLabel := strings.ReplaceAll(teamName, "-", "/")
						owners = append(owners, teamLabel)
					}
				}
			}

			output = append(output, validator.JSONOutput{
				Title:   fmt.Sprintf("Doc entry \"%s\" needs to be reviewed", title),
				Message: fmt.Sprintf("%s for [this document](%s%s).", description, validator.DocsHost, filePath),
				Owner:   owners,
			})
		}
	}

	jsonBytes, _ := json.Marshal(output)
	fmt.Println(string(jsonBytes))
}

// DumpAnnotations creates GitHub Actions annotations file using the OS filesystem
func (f *Formatter) DumpAnnotations(results map[string]validator.ValidationResult) error {
	return f.DumpAnnotationsToFS(afero.NewOsFs(), "annotations.json", results)
}

// DumpAnnotationsToFS creates GitHub Actions annotations file using the provided filesystem
func (f *Formatter) DumpAnnotationsToFS(fs afero.Fs, filename string, results map[string]validator.ValidationResult) error {
	annotations := f.buildAnnotations(results)

	file, err := fs.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(annotations)
}

// buildAnnotations creates the annotations data structure from validation results
func (f *Formatter) buildAnnotations(results map[string]validator.ValidationResult) []validator.Annotation {
	var annotations []validator.Annotation

	for filePath, result := range results {
		level := "warning"
		endLine := 1
		nWarnings := 0
		nFails := 0
		var message strings.Builder

		for _, check := range result.Checks {
			if check.Line > 0 {
				endLine = maxInt(endLine, check.Line)
			}
			if check.EndLine > 0 {
				endLine = maxInt(endLine, check.EndLine)
			} else {
				endLine = maxInt(endLine, result.NumFrontMatterLines+1)
			}

			checkInfo := f.checksMap[check.Check]
			if checkInfo.Severity == validator.SeverityFail {
				level = "failure"
				nFails++
			} else {
				nWarnings++
			}

			message.WriteString(fmt.Sprintf("%s - %s\n", checkInfo.Severity, checkInfo.Description))
			if checkInfo.HasValue && check.Value != nil && check.Value != "" {
				message.WriteString(fmt.Sprintf(": %v\n", check.Value))
			}
			message.WriteString("\n")
		}

		var headline string
		if nWarnings > 0 && nFails > 0 {
			headline = fmt.Sprintf("Found %d severe and %d less severe problems", nFails, nWarnings)
		} else if nFails > 0 {
			headline = fmt.Sprintf("Found %d severe problems", nFails)
		} else {
			headline = fmt.Sprintf("Found %d less severe problems", nWarnings)
		}

		annotations = append(annotations, validator.Annotation{
			File:            filePath,
			Line:            1,
			EndLine:         endLine,
			Title:           headline,
			Message:         message.String(),
			AnnotationLevel: level,
		})
	}

	return annotations
}

// printCheckResult prints a single check result with formatting
func (f *Formatter) printCheckResult(check validator.CheckResult, severity string) {
	checkInfo := f.checksMap[check.Check]
	line := fmt.Sprintf(" - %s - %s - %s",
		f.colorSeverity(severity),
		f.colorHeadline(check.Check),
		checkInfo.Description)

	if checkInfo.HasValue && check.Value != nil && check.Value != "" {
		line += fmt.Sprintf(": %s", f.colorLiteral(fmt.Sprintf("%v", check.Value)))
	}

	fmt.Println(line)
}

// Color functions (simplified - could be enhanced with actual ANSI colors)
func (f *Formatter) colorSeverity(severity string) string {
	switch severity {
	case validator.SeverityFail:
		return fmt.Sprintf("\033[1;31m%s\033[0m", severity) // Bold red
	case validator.SeverityWarn:
		return fmt.Sprintf("\033[1;33m%s\033[0m", severity) // Bold yellow
	default:
		return severity
	}
}

func (f *Formatter) colorHeadline(text string) string {
	return fmt.Sprintf("\033[37m%s\033[0m", text) // White
}

func (f *Formatter) colorLiteral(text string) string {
	return fmt.Sprintf("\033[36m%s\033[0m", text) // Cyan
}

// maxInt returns the maximum of two integers
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
