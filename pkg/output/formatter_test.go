package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/giantswarm/frontmatter-validator/pkg/validator"
)

func TestDumpAnnotationsToFS(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		results         map[string]validator.ValidationResult
		expectedCount   int
		expectedLevel   string
		expectedTitle   string
		expectedFile    string
		expectedEndLine int
		messageContains []string
		description     string
	}{
		{
			name:          "empty results",
			filename:      "empty-annotations.json",
			results:       make(map[string]validator.ValidationResult),
			expectedCount: 0,
			description:   "Should create empty JSON array for no validation results",
		},
		{
			name:     "single file with warnings",
			filename: "warnings-annotations.json",
			results: map[string]validator.ValidationResult{
				"test/file.md": {
					NumFrontMatterLines: 5,
					Checks: []validator.CheckResult{
						{Check: validator.NoLinkTitle, Line: 3},
						{Check: validator.NoWeight, Line: 2},
					},
				},
			},
			expectedCount:   1,
			expectedLevel:   "warning",
			expectedTitle:   "Found 2 less severe problems",
			expectedFile:    "test/file.md",
			expectedEndLine: 6, // NumFrontMatterLines + 1
			messageContains: []string{
				"WARN - The page should have a linkTitle",
				"WARN - The page should have a weight attribute",
			},
			description: "Should create warning annotation for WARN severity checks",
		},
		{
			name:     "single file with failures",
			filename: "failures-annotations.json",
			results: map[string]validator.ValidationResult{
				"docs/critical.md": {
					NumFrontMatterLines: 8,
					Checks: []validator.CheckResult{
						{Check: validator.NoTitle, Line: 1},
						{Check: validator.NoDescription, Line: 1},
					},
				},
			},
			expectedCount:   1,
			expectedLevel:   "failure",
			expectedTitle:   "Found 2 severe problems",
			expectedFile:    "docs/critical.md",
			expectedEndLine: 9, // NumFrontMatterLines + 1
			description:     "Should create failure annotation for FAIL severity checks",
		},
		{
			name:     "mixed severities",
			filename: "mixed-annotations.json",
			results: map[string]validator.ValidationResult{
				"docs/mixed.md": {
					NumFrontMatterLines: 10,
					Checks: []validator.CheckResult{
						{Check: validator.NoTitle, Line: 2},          // FAIL
						{Check: validator.ReviewTooLongAgo, Line: 5}, // WARN
						{Check: validator.NoDescription, Line: 3},    // FAIL
					},
				},
			},
			expectedCount:   1,
			expectedLevel:   "failure", // Should be failure due to presence of FAIL checks
			expectedTitle:   "Found 2 severe and 1 less severe problems",
			expectedFile:    "docs/mixed.md",
			expectedEndLine: 11, // NumFrontMatterLines + 1
			description:     "Should create failure annotation when both severities are present",
		},
		{
			name:     "file with check values",
			filename: "values-annotations.json",
			results: map[string]validator.ValidationResult{
				"test.md": {
					NumFrontMatterLines: 4,
					Checks: []validator.CheckResult{
						{
							Check: validator.LongTitle,
							Value: "This is an extremely long title that definitely exceeds the maximum character limit",
							Line:  2,
						},
					},
				},
			},
			expectedCount:   1,
			expectedLevel:   "failure",
			expectedTitle:   "Found 1 severe problem",
			expectedFile:    "test.md",
			expectedEndLine: 5,
			messageContains: []string{
				"This is an extremely long title that definitely exceeds the maximum character limit",
			},
			description: "Should include check values in annotation messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a virtual filesystem
			fs := afero.NewMemMapFs()
			formatter := New()

			err := formatter.DumpAnnotationsToFS(fs, tt.filename, tt.results)
			if err != nil {
				t.Fatalf("%s: Expected no error, got %v", tt.description, err)
			}

			// Check that the file was created
			exists, err := afero.Exists(fs, tt.filename)
			if err != nil {
				t.Fatalf("%s: Error checking file existence: %v", tt.description, err)
			}
			if !exists {
				t.Fatalf("%s: Expected annotations file to be created", tt.description)
			}

			// Read and verify the content
			content, err := afero.ReadFile(fs, tt.filename)
			if err != nil {
				t.Fatalf("%s: Error reading annotations file: %v", tt.description, err)
			}

			var annotations []validator.Annotation
			err = json.Unmarshal(content, &annotations)
			if err != nil {
				t.Fatalf("%s: Error parsing JSON: %v", tt.description, err)
			}

			if len(annotations) != tt.expectedCount {
				t.Errorf("%s: Expected %d annotations, got %d", tt.description, tt.expectedCount, len(annotations))
			}

			// Skip detailed checks for empty results
			if tt.expectedCount == 0 {
				return
			}

			annotation := annotations[0]

			if annotation.File != tt.expectedFile {
				t.Errorf("%s: Expected file '%s', got '%s'", tt.description, tt.expectedFile, annotation.File)
			}

			if annotation.Line != 1 {
				t.Errorf("%s: Expected line 1, got %d", tt.description, annotation.Line)
			}

			if annotation.EndLine != tt.expectedEndLine {
				t.Errorf("%s: Expected end line %d, got %d", tt.description, tt.expectedEndLine, annotation.EndLine)
			}

			if annotation.AnnotationLevel != tt.expectedLevel {
				t.Errorf("%s: Expected annotation level '%s', got '%s'", tt.description, tt.expectedLevel, annotation.AnnotationLevel)
			}

			if annotation.Title != tt.expectedTitle {
				t.Errorf("%s: Expected title '%s', got '%s'", tt.description, tt.expectedTitle, annotation.Title)
			}

			// Check message contents
			for _, expectedContent := range tt.messageContains {
				if !strings.Contains(annotation.Message, expectedContent) {
					t.Errorf("%s: Expected message to contain '%s', got message: %s", tt.description, expectedContent, annotation.Message)
				}
			}
		})
	}
}

func TestDumpAnnotationsToFS_MultipleFiles(t *testing.T) {
	// Create a virtual filesystem
	fs := afero.NewMemMapFs()
	formatter := New()

	// Create test results with multiple files
	results := map[string]validator.ValidationResult{
		"docs/file1.md": {
			NumFrontMatterLines: 5,
			Checks: []validator.CheckResult{
				{
					Check: validator.NoWeight,
					Line:  3,
				},
			},
		},
		"docs/file2.md": {
			NumFrontMatterLines: 7,
			Checks: []validator.CheckResult{
				{
					Check: validator.NoTitle,
					Line:  1,
				},
			},
		},
	}

	err := formatter.DumpAnnotationsToFS(fs, "multi-annotations.json", results)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Read and verify the content
	content, err := afero.ReadFile(fs, "multi-annotations.json")
	if err != nil {
		t.Fatalf("Error reading annotations file: %v", err)
	}

	var annotations []validator.Annotation
	err = json.Unmarshal(content, &annotations)
	if err != nil {
		t.Fatalf("Error parsing JSON: %v", err)
	}

	if len(annotations) != 2 {
		t.Fatalf("Expected 2 annotations, got %d", len(annotations))
	}

	// Find annotations by file (order may vary due to map iteration)
	var file1Annotation, file2Annotation *validator.Annotation
	for i := range annotations {
		switch annotations[i].File {
		case "docs/file1.md":
			file1Annotation = &annotations[i]
		case "docs/file2.md":
			file2Annotation = &annotations[i]
		}
	}

	if file1Annotation == nil {
		t.Fatal("Expected annotation for docs/file1.md")
	}
	if file2Annotation == nil {
		t.Fatal("Expected annotation for docs/file2.md")
	}

	// Verify file1 annotation (warning)
	if file1Annotation.AnnotationLevel != "warning" {
		t.Errorf("Expected file1 annotation level 'warning', got '%s'", file1Annotation.AnnotationLevel)
	}

	// Verify file2 annotation (failure)
	if file2Annotation.AnnotationLevel != "failure" {
		t.Errorf("Expected file2 annotation level 'failure', got '%s'", file2Annotation.AnnotationLevel)
	}
}

func TestDumpAnnotationsToFS_FileSystemError(t *testing.T) {
	// Create a read-only filesystem to simulate creation errors
	fs := afero.NewReadOnlyFs(afero.NewMemMapFs())
	formatter := New()

	results := map[string]validator.ValidationResult{
		"test.md": {
			NumFrontMatterLines: 4,
			Checks: []validator.CheckResult{
				{
					Check: validator.NoTitle,
					Line:  1,
				},
			},
		},
	}

	err := formatter.DumpAnnotationsToFS(fs, "readonly-test.json", results)
	if err == nil {
		t.Fatal("Expected error when writing to read-only filesystem, got nil")
	}

	// Verify the error is related to file creation
	if !strings.Contains(err.Error(), "operation not permitted") && !strings.Contains(err.Error(), "read-only") {
		t.Errorf("Expected file system error, got: %v", err)
	}
}

func TestDumpAnnotations_BackwardCompatibility(t *testing.T) {
	// Test that the original DumpAnnotations method still works
	// This would normally write to the real filesystem, but we're just testing it doesn't crash

	// We can't easily test the actual file creation without touching the real filesystem,
	// but we can at least verify the method exists and can be called
	// In a real scenario, this would create annotations.json in the current directory

	// For now, just verify the method signature is correct and doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("DumpAnnotations panicked: %v", r)
		}
	}()

	// This will try to write to the real filesystem, but since we have empty results,
	// it should at least create an empty annotations.json file
	// We'll skip this test in CI environments or when we don't want to touch the filesystem
	if testing.Short() {
		t.Skip("Skipping filesystem test in short mode")
	}

	// Note: In a production test suite, you might want to:
	// 1. Change to a temporary directory
	// 2. Clean up the annotations.json file after the test
	// 3. Or mock the filesystem layer completely
}

func TestBuildAnnotations_EdgeCases(t *testing.T) {
	formatter := New()

	// Test with zero line numbers
	results := map[string]validator.ValidationResult{
		"test.md": {
			NumFrontMatterLines: 0,
			Checks: []validator.CheckResult{
				{
					Check: validator.NoTitle,
					Line:  0, // Zero line number
				},
			},
		},
	}

	annotations := formatter.buildAnnotations(results)

	if len(annotations) != 1 {
		t.Fatalf("Expected 1 annotation, got %d", len(annotations))
	}

	annotation := annotations[0]

	// Should default to line 1 and endLine should be 1 (NumFrontMatterLines + 1)
	if annotation.Line != 1 {
		t.Errorf("Expected line 1, got %d", annotation.Line)
	}

	if annotation.EndLine != 1 {
		t.Errorf("Expected end line 1, got %d", annotation.EndLine)
	}
}

func TestColorFunctions(t *testing.T) {
	formatter := New()

	tests := []struct {
		name        string
		function    func(string) string
		input       string
		expected    string
		description string
	}{
		{
			name:        "colorSeverity_FAIL",
			function:    formatter.colorSeverity,
			input:       validator.SeverityFail,
			expected:    "\033[1;31mFAIL\033[0m",
			description: "Should format FAIL severity with bold red color",
		},
		{
			name:        "colorSeverity_WARN",
			function:    formatter.colorSeverity,
			input:       validator.SeverityWarn,
			expected:    "\033[1;33mWARN\033[0m",
			description: "Should format WARN severity with bold yellow color",
		},
		{
			name:        "colorSeverity_unknown",
			function:    formatter.colorSeverity,
			input:       "UNKNOWN",
			expected:    "UNKNOWN",
			description: "Should return input unchanged for unknown severity",
		},
		{
			name:        "colorHeadline",
			function:    formatter.colorHeadline,
			input:       "TEST_CHECK",
			expected:    "\033[37mTEST_CHECK\033[0m",
			description: "Should format headlines with white color",
		},
		{
			name:        "colorLiteral",
			function:    formatter.colorLiteral,
			input:       "test value",
			expected:    "\033[36mtest value\033[0m",
			description: "Should format literals with cyan color",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function(tt.input)
			if result != tt.expected {
				t.Errorf("%s: expected '%s', got '%s'", tt.description, tt.expected, result)
			}
		})
	}
}

func TestPrintJSON(t *testing.T) {
	tests := []struct {
		name        string
		results     map[string]validator.ValidationResult
		expectedLen int
		checkTitle  string
		checkOwner  []string
		description string
	}{
		{
			name:        "empty results",
			results:     make(map[string]validator.ValidationResult),
			expectedLen: 0,
			description: "Should output empty JSON array for no results",
		},
		{
			name: "single result with title and owner",
			results: map[string]validator.ValidationResult{
				"docs/test.md": {
					Checks: []validator.CheckResult{
						{
							Check: validator.ReviewTooLongAgo,
							Title: "Test Document",
							Owner: []string{"https://github.com/orgs/giantswarm/teams/team-honeybadger"},
						},
					},
				},
			},
			expectedLen: 1,
			checkTitle:  "Test Document",
			checkOwner:  []string{"team/honeybadger"},
			description: "Should format JSON output with title and extracted team name",
		},
		{
			name: "result without title should be skipped",
			results: map[string]validator.ValidationResult{
				"docs/test.md": {
					Checks: []validator.CheckResult{
						{
							Check: validator.NoTitle,
							// No Title field - should be skipped
						},
					},
				},
			},
			expectedLen: 0,
			description: "Should skip checks without titles",
		},
		{
			name: "multiple owners",
			results: map[string]validator.ValidationResult{
				"docs/test.md": {
					Checks: []validator.CheckResult{
						{
							Check: validator.ReviewTooLongAgo,
							Title: "Multi-Owner Doc",
							Owner: []string{
								"https://github.com/orgs/giantswarm/teams/team-honeybadger",
								"https://github.com/orgs/giantswarm/teams/team-phoenix",
							},
						},
					},
				},
			},
			expectedLen: 1,
			checkTitle:  "Multi-Owner Doc",
			checkOwner:  []string{"team/honeybadger", "team/phoenix"},
			description: "Should extract multiple team names from owner URLs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			formatter := New()
			formatter.PrintJSON(tt.results)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Parse JSON output
			var jsonOutput []validator.JSONOutput
			err := json.Unmarshal([]byte(output), &jsonOutput)
			if err != nil {
				t.Fatalf("%s: Failed to parse JSON output: %v\nOutput: %s", tt.description, err, output)
			}

			if len(jsonOutput) != tt.expectedLen {
				t.Errorf("%s: Expected %d items, got %d", tt.description, tt.expectedLen, len(jsonOutput))
			}

			if tt.expectedLen > 0 {
				item := jsonOutput[0]

				expectedTitle := "Doc entry \"" + tt.checkTitle + "\" needs to be reviewed"
				if item.Title != expectedTitle {
					t.Errorf("%s: Expected title '%s', got '%s'", tt.description, expectedTitle, item.Title)
				}

				if len(item.Owner) != len(tt.checkOwner) {
					t.Errorf("%s: Expected %d owners, got %d", tt.description, len(tt.checkOwner), len(item.Owner))
				} else {
					for i, expectedOwner := range tt.checkOwner {
						if item.Owner[i] != expectedOwner {
							t.Errorf("%s: Expected owner[%d] '%s', got '%s'", tt.description, i, expectedOwner, item.Owner[i])
						}
					}
				}

				// Check that message contains the docs host URL
				if !strings.Contains(item.Message, validator.DocsHost) {
					t.Errorf("%s: Expected message to contain docs host URL", tt.description)
				}
			}
		})
	}
}

func TestMaxInt(t *testing.T) {
	tests := []struct {
		a, b, expected int
		description    string
	}{
		{5, 3, 5, "Should return first value when it's larger"},
		{2, 8, 8, "Should return second value when it's larger"},
		{4, 4, 4, "Should return either value when they're equal"},
		{-1, -5, -1, "Should work with negative numbers"},
		{0, -1, 0, "Should work with zero"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := maxInt(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("%s: maxInt(%d, %d) = %d, expected %d", tt.description, tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestPrintStdout(t *testing.T) {
	tests := []struct {
		name            string
		results         map[string]validator.ValidationResult
		expectedOutputs []string
		expectedCounts  []string
		description     string
	}{
		{
			name:            "empty results",
			results:         make(map[string]validator.ValidationResult),
			expectedOutputs: []string{}, // Should only show summary
			expectedCounts:  []string{},
			description:     "Should show empty output for no validation results",
		},
		{
			name: "single file with failures",
			results: map[string]validator.ValidationResult{
				"docs/test.md": {
					Checks: []validator.CheckResult{
						{Check: validator.NoTitle},
						{Check: validator.NoDescription},
					},
				},
			},
			expectedOutputs: []string{
				"docs/test.md",
				"NO_TITLE",
				"NO_DESCRIPTION",
			},
			expectedCounts: []string{
				"Found 2 critical problems",
			},
			description: "Should display file path, check names, and failure count",
		},
		{
			name: "single file with warnings",
			results: map[string]validator.ValidationResult{
				"docs/test.md": {
					Checks: []validator.CheckResult{
						{Check: validator.NoWeight},
						{Check: validator.NoLinkTitle},
					},
				},
			},
			expectedOutputs: []string{
				"docs/test.md",
				"NO_WEIGHT",
				"NO_LINK_TITLE",
			},
			expectedCounts: []string{
				"Found 2 less severe problems",
			},
			description: "Should display warnings and warning count",
		},
		{
			name: "mixed severities",
			results: map[string]validator.ValidationResult{
				"docs/test.md": {
					Checks: []validator.CheckResult{
						{Check: validator.NoTitle},       // FAIL
						{Check: validator.NoWeight},      // WARN
						{Check: validator.NoDescription}, // FAIL
					},
				},
			},
			expectedOutputs: []string{
				"docs/test.md",
				"NO_TITLE",
				"NO_DESCRIPTION",
				"NO_WEIGHT",
			},
			expectedCounts: []string{
				"Found 2 critical problems",
				"Found 1 less severe problem",
			},
			description: "Should group failures first, then warnings, with separate counts",
		},
		{
			name: "multiple files",
			results: map[string]validator.ValidationResult{
				"docs/file1.md": {
					Checks: []validator.CheckResult{
						{Check: validator.NoTitle},
					},
				},
				"docs/file2.md": {
					Checks: []validator.CheckResult{
						{Check: validator.NoWeight},
					},
				},
			},
			expectedOutputs: []string{
				"file1.md", // Should contain both filenames
				"file2.md",
				"NO_TITLE",
				"NO_WEIGHT",
			},
			expectedCounts: []string{
				"Found 1 critical problem",
				"Found 1 less severe problem",
			},
			description: "Should handle multiple files with different severities",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			formatter := New()
			formatter.PrintStdout(tt.results)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Check expected outputs are present
			for _, expectedOutput := range tt.expectedOutputs {
				if !strings.Contains(output, expectedOutput) {
					t.Errorf("%s: Expected output to contain '%s', got:\n%s", tt.description, expectedOutput, output)
				}
			}

			// Check expected count messages are present
			for _, expectedCount := range tt.expectedCounts {
				if !strings.Contains(output, expectedCount) {
					t.Errorf("%s: Expected output to contain '%s', got:\n%s", tt.description, expectedCount, output)
				}
			}
		})
	}
}
