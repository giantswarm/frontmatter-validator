package validator

// DocsHost is the base URL for documentation links
const (
	DocsHost = "https://github.com/giantswarm/docs/blob/main/"
)

// GetChecks returns all validation checks in logical order
func GetChecks() []Check {
	return []Check{
		// Prerequisites
		{
			ID:          NoFrontMatter,
			Description: "No front matter found in the beginning of the page",
			Severity:    SeverityFail,
		},
		{
			ID:          NoTrailingNewline,
			Description: "There must be a newline character at the end of the page to ensure proper parsing",
			Severity:    SeverityFail,
		},
		{
			ID:          UnknownAttribute,
			Description: "There is an unknown front matter attribute in this page",
			Severity:    SeverityFail,
			HasValue:    true,
		},
		// Standard attributes
		{
			ID:          NoTitle,
			Description: "The page should have a title",
			Severity:    SeverityFail,
		},
		{
			ID:          LongTitle,
			Description: "The title should be less than 100 characters",
			Severity:    SeverityFail,
			HasValue:    true,
		},
		{
			ID:          ShortTitle,
			Description: "The title should be longer than 5 characters",
			Severity:    SeverityFail,
			HasValue:    true,
		},
		{
			ID:          NoDescription,
			Description: "Each page should have a description",
			Severity:    SeverityFail,
		},
		{
			ID:          LongDescription,
			Description: "The description should be less than 300 characters",
			Severity:    SeverityFail,
			HasValue:    true,
		},
		{
			ID:          NoFullStopDescription,
			Description: "The description should end with a full stop",
			Severity:    SeverityFail,
			HasValue:    true,
		},
		{
			ID:          ShortDescription,
			Description: "The description should be longer than 50 characters",
			Severity:    SeverityFail,
			HasValue:    true,
		},
		{
			ID:          InvalidDescription,
			Description: "Description must be a simple string without any markup or line breaks",
			Severity:    SeverityFail,
		},
		{
			ID:          NoLinkTitle,
			Description: "The page should have a linkTitle, which appears in menus and list pages. If not given, title will be used and should be no longer than 40 characters.",
			Severity:    SeverityWarn,
		},
		{
			ID:          LongLinkTitle,
			Description: "The linkTitle (used in menu and list pages; title is used if linkTitle is not given) should be less than 40 characters",
			Severity:    SeverityFail,
		},
		{
			ID:          NoWeight,
			Description: "The page should have a weight attribute, to control the sort order",
			Severity:    SeverityWarn,
		},
		// Custom attributes
		{
			ID:          NoOwner,
			Description: "The page should have an owner assigned",
			Severity:    SeverityFail,
		},
		{
			ID:          InvalidOwner,
			Description: "The owner field values must start with a Github teams URL",
			Severity:    SeverityFail,
			HasValue:    true,
		},
		{
			ID:          NoLastReviewDate,
			Description: "The page should have a last_review_date",
			Severity:    SeverityWarn,
		},
		{
			ID:          ReviewTooLongAgo,
			Description: "The last review date is too long ago",
			Severity:    SeverityWarn,
			HasValue:    true,
		},
		{
			ID:          InvalidLastReviewDate,
			Description: "The last_review_date should be in format YYYY-MM-DD and not in the future",
			Severity:    SeverityFail,
			HasValue:    true,
		},
		{
			ID:          NoUserQuestions,
			Description: "The page should have user_questions assigned",
			Severity:    SeverityFail,
		},
		{
			ID:          LongUserQuestion,
			Description: "Each user question should be no longer than 100 characters",
			Severity:    SeverityFail,
			HasValue:    true,
		},
		{
			ID:          NoQuestionMark,
			Description: "Questions should end with a question mark",
			Severity:    SeverityFail,
			HasValue:    true,
		},
		// Runbook checks
		{
			ID:          RunbookLayoutNotSet,
			Description: "Runbook pages must have layout: runbook",
			Severity:    SeverityFail,
		},
		{
			ID:          InvalidRunbookVariables,
			Description: "Runbook variables must be an array and not empty",
			Severity:    SeverityFail,
		},
		{
			ID:          RunbookVariableWithoutName,
			Description: "Each runbook variable must have a name specified",
			Severity:    SeverityFail,
			HasValue:    true,
		},
		{
			ID:          InvalidRunbookVariableName,
			Description: "Variable names must use only uppercase letters and underscores, and be unique",
			Severity:    SeverityFail,
			HasValue:    true,
		},
		{
			ID:          InvalidRunbookVariable,
			Description: "Each variable must be a valid object with name field and optional description and default fields",
			Severity:    SeverityFail,
			HasValue:    true,
		},
		{
			ID:          InvalidRunbookDashboards,
			Description: "Runbook dashboards must be an array and not empty",
			Severity:    SeverityFail,
		},
		{
			ID:          InvalidRunbookDashboard,
			Description: "Each runbook dashboard must have name and link specified and non-empty",
			Severity:    SeverityFail,
			HasValue:    true,
		},
		{
			ID:          InvalidRunbookDashboardLink,
			Description: "Dashboard link must be a valid URL with properly defined variables",
			Severity:    SeverityFail,
			HasValue:    true,
		},
		{
			ID:          InvalidRunbookKnownIssues,
			Description: "Runbook known issues must be an array and not empty",
			Severity:    SeverityFail,
		},
		{
			ID:          InvalidRunbookKnownIssue,
			Description: "Each known issue must have url defined and may have optional description field",
			Severity:    SeverityFail,
			HasValue:    true,
		},
		{
			ID:          InvalidRunbookKnownIssueURL,
			Description: "Known issue URL must be a valid URL",
			Severity:    SeverityFail,
			HasValue:    true,
		},
		{
			ID:          RunbookAppearsInMenu,
			Description: "Runbook pages must have toc_hide: true to prevent appearing in menus",
			Severity:    SeverityFail,
		},
	}
}

// GetValidKeys returns the set of valid frontmatter keys
func GetValidKeys() map[string]bool {
	return map[string]bool{
		"aliases":               true,
		"changes_categories":    true,
		"changes_entry":         true,
		"classification":        true,
		"crd":                   true,
		"date":                  true,
		"description":           true,
		"expiration_in_days":    true,
		"last_review_date":      true,
		"layout":                true,
		"linkTitle":             true,
		"menu":                  true,
		"mermaid":               true,
		"owner":                 true,
		"runbook":               true,
		"search":                true,
		"source_repository":     true,
		"source_repository_ref": true,
		"technical_name":        true,
		"title":                 true,
		"toc_hide":              true,
		"user_questions":        true,
		"weight":                true,
	}
}

// GetCheckByID returns a check by its ID
func GetCheckByID(id string) *Check {
	checks := GetChecks()
	for _, check := range checks {
		if check.ID == id {
			return &check
		}
	}
	return nil
}
