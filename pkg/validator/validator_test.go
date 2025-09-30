package validator

import (
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
