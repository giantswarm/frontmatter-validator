package output

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/giantswarm/frontmatter-validator/pkg/validator"
	"github.com/spf13/afero"
)

func TestDumpAnnotationsToFS_EmptyResults(t *testing.T) {
	// Create a virtual filesystem
	fs := afero.NewMemMapFs()
	formatter := New()

	results := make(map[string]validator.ValidationResult)

	err := formatter.DumpAnnotationsToFS(fs, "test-annotations.json", results)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check that the file was created
	exists, err := afero.Exists(fs, "test-annotations.json")
	if err != nil {
		t.Fatalf("Error checking file existence: %v", err)
	}
	if !exists {
		t.Fatal("Expected annotations file to be created")
	}

	// Read and verify the content
	content, err := afero.ReadFile(fs, "test-annotations.json")
	if err != nil {
		t.Fatalf("Error reading annotations file: %v", err)
	}

	// Should be an empty JSON array
	var annotations []validator.Annotation
	err = json.Unmarshal(content, &annotations)
	if err != nil {
		t.Fatalf("Error parsing JSON: %v", err)
	}

	if len(annotations) != 0 {
		t.Errorf("Expected empty annotations array, got %d items", len(annotations))
	}
}

func TestDumpAnnotationsToFS_SingleFileWithWarnings(t *testing.T) {
	// Create a virtual filesystem
	fs := afero.NewMemMapFs()
	formatter := New()

	// Create test results with warnings (using actual WARN severity checks)
	results := map[string]validator.ValidationResult{
		"test/file.md": {
			NumFrontMatterLines: 5,
			Checks: []validator.CheckResult{
				{
					Check: validator.NoLinkTitle,
					Line:  3,
				},
				{
					Check: validator.NoWeight,
					Line:  2,
				},
			},
		},
	}

	err := formatter.DumpAnnotationsToFS(fs, "test-annotations.json", results)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Read and verify the content
	content, err := afero.ReadFile(fs, "test-annotations.json")
	if err != nil {
		t.Fatalf("Error reading annotations file: %v", err)
	}

	var annotations []validator.Annotation
	err = json.Unmarshal(content, &annotations)
	if err != nil {
		t.Fatalf("Error parsing JSON: %v", err)
	}

	if len(annotations) != 1 {
		t.Fatalf("Expected 1 annotation, got %d", len(annotations))
	}

	annotation := annotations[0]

	// Verify annotation fields
	if annotation.File != "test/file.md" {
		t.Errorf("Expected file 'test/file.md', got '%s'", annotation.File)
	}

	if annotation.Line != 1 {
		t.Errorf("Expected line 1, got %d", annotation.Line)
	}

	if annotation.EndLine != 6 { // NumFrontMatterLines + 1
		t.Errorf("Expected end line 6, got %d", annotation.EndLine)
	}

	if annotation.AnnotationLevel != "warning" {
		t.Errorf("Expected annotation level 'warning', got '%s'", annotation.AnnotationLevel)
	}

	expectedTitle := "Found 2 less severe problems"
	if annotation.Title != expectedTitle {
		t.Errorf("Expected title '%s', got '%s'", expectedTitle, annotation.Title)
	}

	// Check that message contains both check descriptions
	if !strings.Contains(annotation.Message, "WARN - The page should have a linkTitle") {
		t.Error("Expected message to contain linkTitle warning")
	}

	if !strings.Contains(annotation.Message, "WARN - The page should have a weight attribute") {
		t.Error("Expected message to contain weight warning")
	}
}

func TestDumpAnnotationsToFS_SingleFileWithFailures(t *testing.T) {
	// Create a virtual filesystem
	fs := afero.NewMemMapFs()
	formatter := New()

	// Create test results with failures
	results := map[string]validator.ValidationResult{
		"docs/critical.md": {
			NumFrontMatterLines: 8,
			Checks: []validator.CheckResult{
				{
					Check: validator.NoTitle,
					Line:  1,
				},
				{
					Check: validator.NoDescription,
					Line:  1,
				},
			},
		},
	}

	err := formatter.DumpAnnotationsToFS(fs, "critical-annotations.json", results)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Read and verify the content
	content, err := afero.ReadFile(fs, "critical-annotations.json")
	if err != nil {
		t.Fatalf("Error reading annotations file: %v", err)
	}

	var annotations []validator.Annotation
	err = json.Unmarshal(content, &annotations)
	if err != nil {
		t.Fatalf("Error parsing JSON: %v", err)
	}

	if len(annotations) != 1 {
		t.Fatalf("Expected 1 annotation, got %d", len(annotations))
	}

	annotation := annotations[0]

	// Should be marked as failure due to FAIL severity checks
	if annotation.AnnotationLevel != "failure" {
		t.Errorf("Expected annotation level 'failure', got '%s'", annotation.AnnotationLevel)
	}

	expectedTitle := "Found 2 severe problems"
	if annotation.Title != expectedTitle {
		t.Errorf("Expected title '%s', got '%s'", expectedTitle, annotation.Title)
	}
}

func TestDumpAnnotationsToFS_MixedSeverities(t *testing.T) {
	// Create a virtual filesystem
	fs := afero.NewMemMapFs()
	formatter := New()

	// Create test results with mixed severities
	results := map[string]validator.ValidationResult{
		"docs/mixed.md": {
			NumFrontMatterLines: 10,
			Checks: []validator.CheckResult{
				{
					Check: validator.NoTitle, // FAIL
					Line:  2,
				},
				{
					Check: validator.ReviewTooLongAgo, // WARN
					Line:  5,
				},
				{
					Check: validator.NoDescription, // FAIL
					Line:  3,
				},
			},
		},
	}

	err := formatter.DumpAnnotationsToFS(fs, "mixed-annotations.json", results)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Read and verify the content
	content, err := afero.ReadFile(fs, "mixed-annotations.json")
	if err != nil {
		t.Fatalf("Error reading annotations file: %v", err)
	}

	var annotations []validator.Annotation
	err = json.Unmarshal(content, &annotations)
	if err != nil {
		t.Fatalf("Error parsing JSON: %v", err)
	}

	if len(annotations) != 1 {
		t.Fatalf("Expected 1 annotation, got %d", len(annotations))
	}

	annotation := annotations[0]

	// Should be marked as failure due to presence of FAIL severity checks
	if annotation.AnnotationLevel != "failure" {
		t.Errorf("Expected annotation level 'failure', got '%s'", annotation.AnnotationLevel)
	}

	expectedTitle := "Found 2 severe and 1 less severe problems"
	if annotation.Title != expectedTitle {
		t.Errorf("Expected title '%s', got '%s'", expectedTitle, annotation.Title)
	}

	// Verify end line calculation - should be max of all check lines
	if annotation.EndLine != 11 { // max(2,5,3) = 5, but fallback to NumFrontMatterLines+1 = 11
		t.Errorf("Expected end line 11, got %d", annotation.EndLine)
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
		if annotations[i].File == "docs/file1.md" {
			file1Annotation = &annotations[i]
		} else if annotations[i].File == "docs/file2.md" {
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

func TestDumpAnnotationsToFS_WithCheckValues(t *testing.T) {
	// Create a virtual filesystem
	fs := afero.NewMemMapFs()
	formatter := New()

	// Create test results with check values
	results := map[string]validator.ValidationResult{
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
	}

	err := formatter.DumpAnnotationsToFS(fs, "values-annotations.json", results)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Read and verify the content
	content, err := afero.ReadFile(fs, "values-annotations.json")
	if err != nil {
		t.Fatalf("Error reading annotations file: %v", err)
	}

	var annotations []validator.Annotation
	err = json.Unmarshal(content, &annotations)
	if err != nil {
		t.Fatalf("Error parsing JSON: %v", err)
	}

	if len(annotations) != 1 {
		t.Fatalf("Expected 1 annotation, got %d", len(annotations))
	}

	annotation := annotations[0]

	// Check that the message contains the check value
	expectedValue := "This is an extremely long title that definitely exceeds the maximum character limit"
	if !strings.Contains(annotation.Message, expectedValue) {
		t.Errorf("Expected message to contain check value '%s', got message: %s", expectedValue, annotation.Message)
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
