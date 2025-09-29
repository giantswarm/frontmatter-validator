package validator

import (
	"testing"
)

func TestNewExcludeConfig(t *testing.T) {
	patterns := []string{
		"docs/legacy",
		"src/content/old:NO_DESCRIPTION",
		"test/path:CHECK1,CHECK2",
		"",    // empty pattern should be ignored
		"   ", // whitespace only should be ignored
	}

	config := NewExcludeConfig(patterns)

	if len(config.patterns) != 3 {
		t.Errorf("Expected 3 patterns, got %d", len(config.patterns))
	}

	// Test first pattern (exclude all checks)
	if config.patterns[0].Path != "docs/legacy" {
		t.Errorf("Expected path 'docs/legacy', got '%s'", config.patterns[0].Path)
	}
	if len(config.patterns[0].Checks) != 0 {
		t.Errorf("Expected no specific checks for first pattern, got %d", len(config.patterns[0].Checks))
	}

	// Test second pattern (exclude specific check)
	if config.patterns[1].Path != "src/content/old" {
		t.Errorf("Expected path 'src/content/old', got '%s'", config.patterns[1].Path)
	}
	if len(config.patterns[1].Checks) != 1 || config.patterns[1].Checks[0] != "NO_DESCRIPTION" {
		t.Errorf("Expected check 'NO_DESCRIPTION', got %v", config.patterns[1].Checks)
	}

	// Test third pattern (exclude multiple checks)
	if config.patterns[2].Path != "test/path" {
		t.Errorf("Expected path 'test/path', got '%s'", config.patterns[2].Path)
	}
	if len(config.patterns[2].Checks) != 2 {
		t.Errorf("Expected 2 checks, got %d", len(config.patterns[2].Checks))
	}
	expectedChecks := map[string]bool{"CHECK1": true, "CHECK2": true}
	for _, check := range config.patterns[2].Checks {
		if !expectedChecks[check] {
			t.Errorf("Unexpected check '%s'", check)
		}
	}
}

func TestExcludeConfig_ShouldExclude(t *testing.T) {
	patterns := []string{
		"docs/legacy", // exclude all checks
		"src/content/old:NO_DESCRIPTION,NO_TITLE", // exclude specific checks
		"test/specific:NO_OWNER",                  // exclude one specific check
	}

	config := NewExcludeConfig(patterns)

	tests := []struct {
		filePath string
		checkID  string
		expected bool
		desc     string
	}{
		// Test exclude all checks
		{"docs/legacy/file.md", "NO_DESCRIPTION", true, "should exclude all checks for docs/legacy"},
		{"docs/legacy/file.md", "NO_TITLE", true, "should exclude all checks for docs/legacy"},
		{"docs/legacy/subdir/file.md", "NO_OWNER", true, "should exclude all checks for subdirectories"},

		// Test exclude specific checks
		{"src/content/old/file.md", "NO_DESCRIPTION", true, "should exclude NO_DESCRIPTION for src/content/old"},
		{"src/content/old/file.md", "NO_TITLE", true, "should exclude NO_TITLE for src/content/old"},
		{"src/content/old/file.md", "NO_OWNER", false, "should not exclude NO_OWNER for src/content/old"},
		{"src/content/old/sub/file.md", "NO_DESCRIPTION", true, "should exclude for subdirectories"},

		// Test single specific check
		{"test/specific/file.md", "NO_OWNER", true, "should exclude NO_OWNER for test/specific"},
		{"test/specific/file.md", "NO_TITLE", false, "should not exclude NO_TITLE for test/specific"},

		// Test non-matching paths
		{"other/path/file.md", "NO_DESCRIPTION", false, "should not exclude for non-matching paths"},
		{"docs/file.md", "NO_TITLE", false, "should not exclude for parent directories"},
	}

	for _, test := range tests {
		result := config.ShouldExclude(test.filePath, test.checkID)
		if result != test.expected {
			t.Errorf("%s: expected %v, got %v", test.desc, test.expected, result)
		}
	}
}

func TestExcludeConfig_PathMatches(t *testing.T) {
	config := NewExcludeConfig([]string{})

	tests := []struct {
		filePath string
		pattern  string
		expected bool
		desc     string
	}{
		{"docs/file.md", "docs", true, "should match parent directory"},
		{"docs/sub/file.md", "docs", true, "should match nested subdirectory"},
		{"docs/file.md", "docs/file.md", true, "should match exact path"},
		{"other/file.md", "docs", false, "should not match different directory"},
		{"docs-other/file.md", "docs", false, "should not match similar but different directory"},
		{"file.md", "docs", false, "should not match file in root"},

		// Test with leading/trailing slashes
		{"/docs/file.md", "docs", true, "should handle leading slash in file path"},
		{"docs/file.md", "/docs", true, "should handle leading slash in pattern"},
		{"docs/file.md/", "docs/", true, "should handle trailing slashes"},

		// Test with ./ prefix
		{"./README.md", "README.md", true, "should handle ./ prefix in file path"},
		{"README.md", "./README.md", true, "should handle ./ prefix in pattern"},
		{"./docs/file.md", "docs", true, "should handle ./ prefix with subdirectory"},
	}

	for _, test := range tests {
		result := config.pathMatches(test.filePath, test.pattern)
		if result != test.expected {
			t.Errorf("%s: pathMatches('%s', '%s') expected %v, got %v",
				test.desc, test.filePath, test.pattern, test.expected, result)
		}
	}
}

func TestValidatorWithExcludes(t *testing.T) {
	excludePatterns := []string{
		"legacy:NO_DESCRIPTION", // exclude NO_DESCRIPTION for legacy path
		"test/all",              // exclude all checks for test/all path
	}

	v := NewWithExcludes(excludePatterns)

	// Test content that would normally fail NO_DESCRIPTION
	content := `---
title: Test Page
owner:
  - https://github.com/orgs/giantswarm/teams/team-test
user_questions:
  - What is this?
---

# Test Content
`

	// Test with legacy path - should skip NO_DESCRIPTION
	result := v.ValidateFile(content, "legacy/file.md", ValidateAll)
	hasNoDescription := false
	for _, check := range result.Checks {
		if check.Check == NoDescription {
			hasNoDescription = true
		}
	}
	if hasNoDescription {
		t.Error("Should have excluded NO_DESCRIPTION check for legacy path")
	}

	// Test with test/all path - should skip ALL checks
	result = v.ValidateFile(content, "test/all/file.md", ValidateAll)
	if len(result.Checks) > 0 {
		t.Errorf("Should have excluded all checks for test/all path, but got %d checks", len(result.Checks))
	}

	// Test with normal path - should still validate
	result = v.ValidateFile(content, "docs/file.md", ValidateAll)
	hasNoDescription = false
	for _, check := range result.Checks {
		if check.Check == NoDescription {
			hasNoDescription = true
		}
	}
	if !hasNoDescription {
		t.Error("Should have NO_DESCRIPTION check for normal path")
	}
}

func TestValidatorExcludeWithDotSlashPrefix(t *testing.T) {
	excludePatterns := []string{
		"README.md", // exclude all checks for README.md
	}

	v := NewWithExcludes(excludePatterns)

	// Test content that would normally fail
	content := "# Just a header\n"

	// Test with ./README.md path (as it would appear from file walker)
	result := v.ValidateFile(content, "./README.md", ValidateAll)
	if len(result.Checks) > 0 {
		t.Errorf("Should have excluded all checks for ./README.md, but got %d checks: %v",
			len(result.Checks), result.Checks)
	}

	// Test with README.md path (without ./ prefix)
	result = v.ValidateFile(content, "README.md", ValidateAll)
	if len(result.Checks) > 0 {
		t.Errorf("Should have excluded all checks for README.md, but got %d checks: %v",
			len(result.Checks), result.Checks)
	}
}
