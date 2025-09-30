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

// Runbook validation tests

func TestValidateFile_ValidRunbook(t *testing.T) {
	v := New()
	content := `---
title: Test Runbook
description: This is a valid runbook description that is long enough and ends with a full stop.
layout: runbook
toc_hide: true
owner:
  - https://github.com/orgs/giantswarm/teams/team-honeybadger
last_review_date: 2024-09-01
runbook:
  variables:
    - name: INSTALLATION
      description: Installation name
      default: golem
    - name: CLUSTER_ID
      description: Cluster identifier
  dashboards:
    - name: Cilium performance
      link: https://grafana-$INSTALLATION.teleport.giantswarm.io/d/cilium-performance
  known_issues:
    - url: https://github.com/giantswarm/giantswarm/issues/28493
      description: Version 1.13 performance is not great in general
    - url: https://github.com/giantswarm/giantswarm/issues/29998
---

# Test Runbook Content
`

	result := v.ValidateFile(content, "test.md")

	// Should have no runbook-specific failures
	hasRunbookFailures := false
	for _, check := range result.Checks {
		if strings.HasPrefix(check.Check, "RUNBOOK_") || strings.HasPrefix(check.Check, "INVALID_RUNBOOK") {
			checkInfo := GetCheckByID(check.Check)
			if checkInfo != nil && checkInfo.Severity == SeverityFail {
				hasRunbookFailures = true
				t.Errorf("Unexpected runbook failure: %s", check.Check)
			}
		}
	}

	if hasRunbookFailures {
		t.Error("Valid runbook should not produce runbook-specific failure checks")
	}
}

func TestValidateFile_RunbookLayoutNotSet(t *testing.T) {
	v := New()
	content := `---
title: Test Runbook
description: This is a valid runbook description that is long enough and ends with a full stop.
owner:
  - https://github.com/orgs/giantswarm/teams/team-honeybadger
runbook:
  variables:
    - name: INSTALLATION
      description: Installation name
  dashboards:
    - name: Test Dashboard
      link: https://example.com
  known_issues:
    - url: https://github.com/giantswarm/giantswarm/issues/1
---

# Test Content
`

	result := v.ValidateFile(content, "test.md")

	hasLayoutCheck := false
	for _, check := range result.Checks {
		if check.Check == RunbookLayoutNotSet {
			hasLayoutCheck = true
		}
	}

	if !hasLayoutCheck {
		t.Error("Expected RUNBOOK_LAYOUT_NOT_SET check when runbook config exists without layout: runbook")
	}
}

func TestValidateFile_RunbookAppearsInMenu(t *testing.T) {
	v := New()
	content := `---
title: Test Runbook
description: This is a valid runbook description that is long enough and ends with a full stop.
layout: runbook
owner:
  - https://github.com/orgs/giantswarm/teams/team-honeybadger
runbook:
  variables:
    - name: INSTALLATION
      description: Installation name
  dashboards:
    - name: Test Dashboard
      link: https://example.com
  known_issues:
    - url: https://github.com/giantswarm/giantswarm/issues/1
---

# Test Content
`

	result := v.ValidateFile(content, "test.md")

	hasMenuCheck := false
	for _, check := range result.Checks {
		if check.Check == RunbookAppearsInMenu {
			hasMenuCheck = true
		}
	}

	if !hasMenuCheck {
		t.Error("Expected RUNBOOK_APPEARS_IN_MENU check when toc_hide is not set to true")
	}
}

func TestValidateFile_InvalidRunbookVariables(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectedErr string
	}{
		{
			name: "empty variables array",
			content: `---
title: Test Runbook
layout: runbook
toc_hide: true
runbook:
  variables: []
  dashboards:
    - name: Test Dashboard
      link: https://example.com
  known_issues:
    - url: https://github.com/giantswarm/giantswarm/issues/1
---`,
			expectedErr: InvalidRunbookVariables,
		},
		{
			name: "variable without name",
			content: `---
title: Test Runbook
layout: runbook
toc_hide: true
runbook:
  variables:
    - description: Missing name
  dashboards:
    - name: Test Dashboard
      link: https://example.com
  known_issues:
    - url: https://github.com/giantswarm/giantswarm/issues/1
---`,
			expectedErr: RunbookVariableWithoutName,
		},
		{
			name: "invalid variable name format",
			content: `---
title: Test Runbook
layout: runbook
toc_hide: true
runbook:
  variables:
    - name: invalid-name
      description: Invalid name format
  dashboards:
    - name: Test Dashboard
      link: https://example.com
  known_issues:
    - url: https://github.com/giantswarm/giantswarm/issues/1
---`,
			expectedErr: InvalidRunbookVariableName,
		},
		{
			name: "duplicate variable names",
			content: `---
title: Test Runbook
layout: runbook
toc_hide: true
runbook:
  variables:
    - name: INSTALLATION
      description: First installation
    - name: INSTALLATION
      description: Duplicate installation
  dashboards:
    - name: Test Dashboard
      link: https://example.com
  known_issues:
    - url: https://github.com/giantswarm/giantswarm/issues/1
---`,
			expectedErr: InvalidRunbookVariableName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			result := v.ValidateFile(tt.content+"\n# Test Content\n", "test.md")

			hasExpectedCheck := false
			for _, check := range result.Checks {
				if check.Check == tt.expectedErr {
					hasExpectedCheck = true
					break
				}
			}

			if !hasExpectedCheck {
				t.Errorf("Expected %s check for %s", tt.expectedErr, tt.name)
			}
		})
	}
}

func TestValidateFile_InvalidRunbookDashboards(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectedErr string
	}{
		{
			name: "empty dashboards array",
			content: `---
title: Test Runbook
layout: runbook
toc_hide: true
runbook:
  variables:
    - name: INSTALLATION
      description: Installation name
  dashboards: []
  known_issues:
    - url: https://github.com/giantswarm/giantswarm/issues/1
---`,
			expectedErr: InvalidRunbookDashboards,
		},
		{
			name: "dashboard without name",
			content: `---
title: Test Runbook
layout: runbook
toc_hide: true
runbook:
  variables:
    - name: INSTALLATION
      description: Installation name
  dashboards:
    - link: https://example.com
  known_issues:
    - url: https://github.com/giantswarm/giantswarm/issues/1
---`,
			expectedErr: InvalidRunbookDashboard,
		},
		{
			name: "dashboard without link",
			content: `---
title: Test Runbook
layout: runbook
toc_hide: true
runbook:
  variables:
    - name: INSTALLATION
      description: Installation name
  dashboards:
    - name: Test Dashboard
  known_issues:
    - url: https://github.com/giantswarm/giantswarm/issues/1
---`,
			expectedErr: InvalidRunbookDashboard,
		},
		{
			name: "dashboard with undefined variable",
			content: `---
title: Test Runbook
layout: runbook
toc_hide: true
runbook:
  variables:
    - name: INSTALLATION
      description: Installation name
  dashboards:
    - name: Test Dashboard
      link: https://grafana-$UNDEFINED_VAR.example.com
  known_issues:
    - url: https://github.com/giantswarm/giantswarm/issues/1
---`,
			expectedErr: InvalidRunbookDashboardLink,
		},
		{
			name: "dashboard with invalid URL format",
			content: `---
title: Test Runbook
layout: runbook
toc_hide: true
runbook:
  variables:
    - name: INSTALLATION
      description: Installation name
  dashboards:
    - name: Test Dashboard
      link: not-a-valid-url
  known_issues:
    - url: https://github.com/giantswarm/giantswarm/issues/1
---`,
			expectedErr: InvalidRunbookDashboardLink,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			result := v.ValidateFile(tt.content+"\n# Test Content\n", "test.md")

			hasExpectedCheck := false
			for _, check := range result.Checks {
				if check.Check == tt.expectedErr {
					hasExpectedCheck = true
					break
				}
			}

			if !hasExpectedCheck {
				t.Errorf("Expected %s check for %s", tt.expectedErr, tt.name)
			}
		})
	}
}

func TestValidateFile_InvalidRunbookKnownIssues(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectedErr string
	}{
		{
			name: "empty known issues array",
			content: `---
title: Test Runbook
layout: runbook
toc_hide: true
runbook:
  variables:
    - name: INSTALLATION
      description: Installation name
  dashboards:
    - name: Test Dashboard
      link: https://example.com
  known_issues: []
---`,
			expectedErr: InvalidRunbookKnownIssues,
		},
		{
			name: "known issue without URL",
			content: `---
title: Test Runbook
layout: runbook
toc_hide: true
runbook:
  variables:
    - name: INSTALLATION
      description: Installation name
  dashboards:
    - name: Test Dashboard
      link: https://example.com
  known_issues:
    - description: Missing URL
---`,
			expectedErr: InvalidRunbookKnownIssue,
		},
		{
			name: "known issue with invalid URL format",
			content: `---
title: Test Runbook
layout: runbook
toc_hide: true
runbook:
  variables:
    - name: INSTALLATION
      description: Installation name
  dashboards:
    - name: Test Dashboard
      link: https://example.com
  known_issues:
    - url: not-a-valid-url
      description: Invalid URL format
---`,
			expectedErr: InvalidRunbookKnownIssueURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			result := v.ValidateFile(tt.content+"\n# Test Content\n", "test.md")

			hasExpectedCheck := false
			for _, check := range result.Checks {
				if check.Check == tt.expectedErr {
					hasExpectedCheck = true
					break
				}
			}

			if !hasExpectedCheck {
				t.Errorf("Expected %s check for %s", tt.expectedErr, tt.name)
			}
		})
	}
}
