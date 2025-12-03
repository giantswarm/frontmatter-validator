package validator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateFile_NoFrontMatter(t *testing.T) {
	v := New()
	content := "# Just a markdown file without frontmatter\n"

	result := v.ValidateFile(content, "test.md")

	if len(result.Checks) != 1 {
		t.Errorf("Expected 1 check result, got %d", len(result.Checks))
	}

	if result.Checks[0].Check != NoFrontMatter {
		t.Errorf("Expected NO_FRONT_MATTER check, got %s", result.Checks[0].Check)
	}
}

func TestValidateFile_NoTrailingNewline(t *testing.T) {
	v := New()
	content := "---\ntitle: Test\n---\nContent without trailing newline"

	result := v.ValidateFile(content, "test.md")

	if len(result.Checks) != 1 {
		t.Errorf("Expected 1 check result, got %d", len(result.Checks))
	}

	if result.Checks[0].Check != NoTrailingNewline {
		t.Errorf("Expected NO_TRAILING_NEWLINE check, got %s", result.Checks[0].Check)
	}
}

func TestValidateFile_ValidFrontMatter(t *testing.T) {
	v := New()
	content := `---
title: Valid Test Page
description: This is a valid description that is long enough and ends with a full stop.
owner:
  - https://github.com/orgs/giantswarm/teams/team-honeybadger
last_review_date: 2024-09-01
user_questions:
  - What is this page about?
  - How do I use this feature?
weight: 100
---

# Test Content

This is a test markdown file with proper frontmatter.
`

	result := v.ValidateFile(content, "test.md")

	// Should have no critical errors, maybe just warnings about review date
	hasFailures := false
	for _, check := range result.Checks {
		checkInfo := GetCheckByID(check.Check)
		if checkInfo != nil && checkInfo.Severity == SeverityFail {
			hasFailures = true
			t.Errorf("Unexpected failure: %s", check.Check)
		}
	}

	if hasFailures {
		t.Error("Valid frontmatter should not produce failure-level checks")
	}
}

func TestValidateFile_InvalidAttributes(t *testing.T) {
	v := New()
	content := `---
title: Test
description: This is a valid description that is long enough and ends with a full stop.
invalid_attribute: should not be allowed
---

# Test Content
`

	result := v.ValidateFile(content, "test.md")

	hasUnknownAttribute := false
	for _, check := range result.Checks {
		if check.Check == UnknownAttribute {
			hasUnknownAttribute = true
			if check.Value != "invalid_attribute" {
				t.Errorf("Expected unknown attribute 'invalid_attribute', got '%v'", check.Value)
			}
		}
	}

	if !hasUnknownAttribute {
		t.Error("Expected UNKNOWN_ATTRIBUTE check for invalid frontmatter attribute")
	}
}

func TestParseFrontMatter(t *testing.T) {
	v := New()
	content := `---
title: Test Page
description: Test description
weight: 100
---

# Content
`

	fm, fmString, numLines, err := v.parseFrontMatter(content)
	if err != nil {
		t.Fatalf("Failed to parse frontmatter: %v", err)
	}

	if fm.Title != "Test Page" {
		t.Errorf("Expected title 'Test Page', got '%s'", fm.Title)
	}

	if fm.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got '%s'", fm.Description)
	}

	if fm.Weight == nil || *fm.Weight != 100 {
		t.Errorf("Expected weight 100, got %v", fm.Weight)
	}

	if !strings.Contains(fmString, "title: Test Page") {
		t.Error("Expected frontmatter string to contain title")
	}

	if numLines <= 0 {
		t.Errorf("Expected positive number of lines, got %d", numLines)
	}
}

func TestValidateFile_DateTimeWithoutTimezone(t *testing.T) {
	v := New()
	// This is the exact content from the failing file
	content := `---
date: 2021-03-12T14:00:00
title: Highlights for the week ending March 12, 2021
changes_categories:
- Highlights
owner:
- https://github.com/orgs/giantswarm/teams/sig-product
---

## Managed apps

[EFK Stack v0.5.0](https://docs.giantswarm.io/changes/managed-apps/efk-stack-app/v0.5.0/) is updated to Elasticsearch and Kibana v7.10.2.
`

	result := v.ValidateFile(content, "test.md")

	// Should NOT have NO_FRONT_MATTER error
	hasNoFrontMatter := false
	for _, check := range result.Checks {
		if check.Check == NoFrontMatter {
			hasNoFrontMatter = true
			t.Errorf("Should not have NO_FRONT_MATTER error when frontmatter exists")
		}
	}

	if hasNoFrontMatter {
		t.Error("File with valid frontmatter should not produce NO_FRONT_MATTER error")
	}

	// Should be able to parse the frontmatter
	if result.NumFrontMatterLines <= 0 {
		t.Error("Should have parsed frontmatter lines")
	}
}

// Runbook validation tests

func TestValidateFile_Runbooks(t *testing.T) {
	tests := []struct {
		name         string
		filename     string
		expectChecks []string // Expected check IDs that should be present
		expectValid  bool     // Whether this should be a valid runbook (no runbook-specific errors)
	}{
		{
			name:        "valid complete runbook",
			filename:    "valid-complete.md",
			expectValid: true,
		},
		{
			name:        "valid minimal runbook",
			filename:    "valid-minimal.md",
			expectValid: true,
		},
		{
			name:        "valid runbook with empty arrays",
			filename:    "valid-empty-arrays.md",
			expectValid: true,
		},
		{
			name:        "valid runbook with only variables",
			filename:    "valid-only-variables.md",
			expectValid: true,
		},
		{
			name:         "runbook layout not set",
			filename:     "layout-not-set.md",
			expectChecks: []string{RunbookLayoutNotSet},
		},
		{
			name:         "runbook appears in menu",
			filename:     "appears-in-menu.md",
			expectChecks: []string{RunbookAppearsInMenu},
		},
		{
			name:         "variable without name",
			filename:     "variable-without-name.md",
			expectChecks: []string{RunbookVariableWithoutName},
		},
		{
			name:         "invalid variable name format",
			filename:     "invalid-variable-name.md",
			expectChecks: []string{InvalidRunbookVariableName},
		},
		{
			name:         "duplicate variable names",
			filename:     "duplicate-variable-names.md",
			expectChecks: []string{InvalidRunbookVariableName},
		},
		{
			name:         "dashboard without name",
			filename:     "dashboard-without-name.md",
			expectChecks: []string{InvalidRunbookDashboard},
		},
		{
			name:         "dashboard without link",
			filename:     "dashboard-without-link.md",
			expectChecks: []string{InvalidRunbookDashboard},
		},
		{
			name:         "dashboard with undefined variable",
			filename:     "dashboard-undefined-variable.md",
			expectChecks: []string{InvalidRunbookDashboardLink},
		},
		{
			name:         "dashboard with invalid URL format",
			filename:     "dashboard-invalid-url.md",
			expectChecks: []string{InvalidRunbookDashboardLink},
		},
		{
			name:         "known issue without URL",
			filename:     "known-issue-without-url.md",
			expectChecks: []string{InvalidRunbookKnownIssue},
		},
		{
			name:         "known issue with invalid URL format",
			filename:     "known-issue-invalid-url.md",
			expectChecks: []string{InvalidRunbookKnownIssueURL},
		},
		{
			name:     "multiple validation errors",
			filename: "multiple-errors.md",
			expectChecks: []string{
				RunbookAppearsInMenu,
				InvalidRunbookVariableName,
				RunbookVariableWithoutName,
				InvalidRunbookDashboardLink,
				InvalidRunbookKnownIssue,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read test file
			content, err := os.ReadFile(filepath.Join("testdata", "runbooks", tt.filename))
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", tt.filename, err)
			}

			// Validate the content
			v := New()
			result := v.ValidateFile(string(content), tt.filename)

			if tt.expectValid {
				// Check that there are no runbook-specific failures
				var runbookErrors []string
				for _, check := range result.Checks {
					if strings.HasPrefix(check.Check, "RUNBOOK_") || strings.HasPrefix(check.Check, "INVALID_RUNBOOK") {
						checkInfo := GetCheckByID(check.Check)
						if checkInfo != nil && checkInfo.Severity == SeverityFail {
							runbookErrors = append(runbookErrors, check.Check)
						}
					}
				}
				if len(runbookErrors) > 0 {
					t.Errorf("Expected valid runbook but got runbook errors: %v", runbookErrors)
				}
			} else {
				// Check that expected checks are present
				foundChecks := make(map[string]bool)
				for _, check := range result.Checks {
					foundChecks[check.Check] = true
				}

				for _, expectedCheck := range tt.expectChecks {
					if !foundChecks[expectedCheck] {
						t.Errorf("Expected check %s but it was not found. Found checks: %v", expectedCheck, getCheckIDs(result.Checks))
					}
				}
			}
		})
	}
}

// Helper function to extract check IDs from results for debugging
func getCheckIDs(checks []CheckResult) []string {
	var ids []string
	for _, check := range checks {
		ids = append(ids, check.Check)
	}
	return ids
}

func TestFlexibleDateParsing(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		expectDate  string
	}{
		{
			name: "quoted date string",
			content: `---
title: Test
description: Test description that is long enough and ends with a full stop.
last_review_date: "2025-01-10"
owner:
  - https://github.com/orgs/giantswarm/teams/team-honeybadger
user_questions:
  - What is this test for?
---
`,
			expectError: false,
			expectDate:  "2025-01-10",
		},
		{
			name: "unquoted date",
			content: `---
title: Test
description: Test description that is long enough and ends with a full stop.
last_review_date: 2025-01-10
owner:
  - https://github.com/orgs/giantswarm/teams/team-honeybadger
user_questions:
  - What is this test for?
---
`,
			expectError: false,
			expectDate:  "2025-01-10",
		},
		{
			name: "RFC3339 timestamp",
			content: `---
title: Test
description: Test description that is long enough and ends with a full stop.
last_review_date: "2025-01-10T00:00:00Z"
owner:
  - https://github.com/orgs/giantswarm/teams/team-honeybadger
user_questions:
  - What is this test for?
---
`,
			expectError: false,
			expectDate:  "2025-01-10",
		},
		{
			name: "ISO 8601 without timezone",
			content: `---
title: Test
description: Test description that is long enough and ends with a full stop.
last_review_date: "2025-01-10T14:30:00"
owner:
  - https://github.com/orgs/giantswarm/teams/team-honeybadger
user_questions:
  - What is this test for?
---
`,
			expectError: false,
			expectDate:  "2025-01-10",
		},
		{
			name: "invalid date format",
			content: `---
title: Test
description: Test description that is long enough and ends with a full stop.
last_review_date: "invalid-date"
owner:
  - https://github.com/orgs/giantswarm/teams/team-honeybadger
user_questions:
  - What is this test for?
---
`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			result := v.ValidateFile(tt.content, "test.md")

			if tt.expectError {
				// Should have NO_FRONT_MATTER error due to parsing failure
				found := false
				for _, check := range result.Checks {
					if check.Check == NoFrontMatter {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected NO_FRONT_MATTER error for invalid date, but didn't find it")
				}
			} else {
				// Should not have NO_FRONT_MATTER error
				for _, check := range result.Checks {
					if check.Check == NoFrontMatter {
						t.Errorf("Unexpected NO_FRONT_MATTER error: %+v", check)
					}
				}
				// Should have parsed front matter successfully
				if result.NumFrontMatterLines == 0 {
					t.Errorf("Expected front matter to be parsed, but NumFrontMatterLines is 0")
				}
			}
		})
	}
}

func TestFlexibleDateValidation(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectChecks   []string
		unexpectChecks []string
	}{
		{
			name: "future date should trigger INVALID_LAST_REVIEW_DATE",
			content: `---
title: Test
description: Test description that is long enough and ends with a full stop.
last_review_date: "2030-01-01"
owner:
  - https://github.com/orgs/giantswarm/teams/team-honeybadger
user_questions:
  - What is this test for?
---
`,
			expectChecks:   []string{InvalidLastReviewDate},
			unexpectChecks: []string{NoFrontMatter, NoLastReviewDate},
		},
		{
			name: "very old date should trigger REVIEW_TOO_LONG_AGO",
			content: `---
title: Test
description: Test description that is long enough and ends with a full stop.
last_review_date: "2020-01-01"
owner:
  - https://github.com/orgs/giantswarm/teams/team-honeybadger
user_questions:
  - What is this test for?
---
`,
			expectChecks:   []string{ReviewTooLongAgo},
			unexpectChecks: []string{NoFrontMatter, NoLastReviewDate, InvalidLastReviewDate},
		},
		{
			name: "recent date should not trigger date-related errors",
			content: `---
title: Test
description: Test description that is long enough and ends with a full stop.
last_review_date: "2025-11-01"
owner:
  - https://github.com/orgs/giantswarm/teams/team-honeybadger
user_questions:
  - What is this test for?
---
`,
			expectChecks:   []string{},
			unexpectChecks: []string{NoFrontMatter, NoLastReviewDate, InvalidLastReviewDate, ReviewTooLongAgo},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			result := v.ValidateFile(tt.content, "test.md")

			checkIDs := getCheckIDs(result.Checks)

			// Check expected checks are present
			for _, expectedCheck := range tt.expectChecks {
				found := false
				for _, checkID := range checkIDs {
					if checkID == expectedCheck {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected check %s not found. Got checks: %v", expectedCheck, checkIDs)
				}
			}

			// Check unexpected checks are not present
			for _, unexpectedCheck := range tt.unexpectChecks {
				for _, checkID := range checkIDs {
					if checkID == unexpectedCheck {
						t.Errorf("Unexpected check %s found. Got checks: %v", unexpectedCheck, checkIDs)
					}
				}
			}
		})
	}
}
