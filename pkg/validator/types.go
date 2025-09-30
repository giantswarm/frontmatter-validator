package validator

import (
	"time"
)

// Validation modes
const (
	ValidateAll            = "all"
	ValidateLastReviewDate = "last-reviewed"
)

// Check identifiers
const (
	InvalidDescription    = "INVALID_DESCRIPTION"
	InvalidLastReviewDate = "INVALID_LAST_REVIEW_DATE"
	InvalidOwner          = "INVALID_OWNER"
	LongDescription       = "LONG_DESCRIPTION"
	NoFullStopDescription = "NO_FULL_STOP_DESCRIPTION"
	LongLinkTitle         = "LONG_LINK_TITLE"
	LongTitle             = "LONG_TITLE"
	LongUserQuestion      = "LONG_USER_QUESTION"
	NoDescription         = "NO_DESCRIPTION"
	NoFrontMatter         = "NO_FRONT_MATTER"
	NoLastReviewDate      = "NO_LAST_REVIEW_DATE"
	NoLinkTitle           = "NO_LINK_TITLE"
	NoOwner               = "NO_OWNER"
	NoQuestionMark        = "NO_QUESTION_MARK"
	NoTitle               = "NO_TITLE"
	NoTrailingNewline     = "NO_TRAILING_NEWLINE"
	NoUserQuestions       = "NO_USER_QUESTIONS"
	NoWeight              = "NO_WEIGHT"
	ReviewTooLongAgo      = "REVIEW_TOO_LONG_AGO"
	ShortDescription      = "SHORT_DESCRIPTION"
	ShortTitle            = "SHORT_TITLE"
	UnknownAttribute      = "UNKNOWN_ATTRIBUTE"
	// Runbook checks
	RunbookLayoutNotSet         = "RUNBOOK_LAYOUT_NOT_SET"
	InvalidRunbookVariables     = "INVALID_RUNBOOK_VARIABLES"
	RunbookVariableWithoutName  = "RUNBOOK_VARIABLE_WITHOUT_NAME"
	InvalidRunbookVariableName  = "INVALID_RUNBOOK_VARIABLE_NAME"
	InvalidRunbookVariable      = "INVALID_RUNBOOK_VARIABLE"
	InvalidRunbookDashboards    = "INVALID_RUNBOOK_DASHBOARDS"
	InvalidRunbookDashboard     = "INVALID_RUNBOOK_DASHBOARD"
	InvalidRunbookDashboardLink = "INVALID_RUNBOOK_DASHBOARD_LINK"
	InvalidRunbookKnownIssues   = "INVALID_RUNBOOK_KNOWN_ISSUES"
	InvalidRunbookKnownIssue    = "INVALID_RUNBOOK_KNOWN_ISSUE"
	InvalidRunbookKnownIssueURL = "INVALID_RUNBOOK_KNOWN_ISSUE_URL"
	RunbookAppearsInMenu        = "RUNBOOK_APPEARS_IN_MENU"
)

// Severity levels
const (
	SeverityFail = "FAIL"
	SeverityWarn = "WARN"
)

// Check represents a validation check
type Check struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	HasValue    bool   `json:"has_value,omitempty"`
}

// CheckResult represents the result of a single validation check
type CheckResult struct {
	Check   string      `json:"check"`
	Value   interface{} `json:"value,omitempty"`
	Line    int         `json:"line,omitempty"`
	EndLine int         `json:"end_line,omitempty"`
	Title   string      `json:"title,omitempty"`
	Owner   []string    `json:"owner,omitempty"`
}

// ValidationResult represents the result of validating a single file
type ValidationResult struct {
	NumFrontMatterLines int           `json:"num_front_matter_lines"`
	Checks              []CheckResult `json:"checks"`
}

// RunbookVariable represents a runbook variable
type RunbookVariable struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Default     string `yaml:"default,omitempty"`
}

// RunbookDashboard represents a runbook dashboard
type RunbookDashboard struct {
	Name string `yaml:"name"`
	Link string `yaml:"link"`
}

// RunbookKnownIssue represents a runbook known issue
type RunbookKnownIssue struct {
	URL         string `yaml:"url"`
	Description string `yaml:"description,omitempty"`
}

// Runbook represents the runbook configuration
type Runbook struct {
	Variables   []RunbookVariable   `yaml:"variables,omitempty"`
	Dashboards  []RunbookDashboard  `yaml:"dashboards,omitempty"`
	KnownIssues []RunbookKnownIssue `yaml:"known_issues,omitempty"`
}

// FrontMatter represents the parsed frontmatter structure
type FrontMatter struct {
	Title               string      `yaml:"title"`
	Description         string      `yaml:"description"`
	LinkTitle           string      `yaml:"linkTitle"`
	Owner               []string    `yaml:"owner"`
	LastReviewDate      *time.Time  `yaml:"last_review_date"`
	UserQuestions       []string    `yaml:"user_questions"`
	Weight              *int        `yaml:"weight"`
	Menu                interface{} `yaml:"menu"`
	ExpirationInDays    *int        `yaml:"expiration_in_days"`
	Date                string      `yaml:"date"`
	Aliases             []string    `yaml:"aliases"`
	ChangesCategories   []string    `yaml:"changes_categories"`
	ChangesEntry        interface{} `yaml:"changes_entry"`
	CRD                 interface{} `yaml:"crd"`
	Layout              string      `yaml:"layout"`
	Mermaid             bool        `yaml:"mermaid"`
	Search              interface{} `yaml:"search"`
	SourceRepository    string      `yaml:"source_repository"`
	SourceRepositoryRef string      `yaml:"source_repository_ref"`
	TechnicalName       string      `yaml:"technical_name"`
	TocHide             bool        `yaml:"toc_hide"`
	Runbook             *Runbook    `yaml:"runbook,omitempty"`
}

// JSONOutput represents the JSON output format for issues
type JSONOutput struct {
	Title   string   `json:"title"`
	Message string   `json:"message"`
	Owner   []string `json:"owner"`
}

// Annotation represents a GitHub Actions annotation
type Annotation struct {
	File            string `json:"file"`
	Line            int    `json:"line"`
	EndLine         int    `json:"end_line"`
	Title           string `json:"title"`
	Message         string `json:"message"`
	AnnotationLevel string `json:"annotation_level"`
}
