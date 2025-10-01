package validator

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Validator handles frontmatter validation
type Validator struct {
	checks        []Check
	validKeys     map[string]bool
	configManager ConfigManager
}

// ConfigManager interface for configuration management
type ConfigManager interface {
	GetEnabledChecksForPath(filePath string) []string
}

// New creates a new Validator instance with default configuration
func New() *Validator {
	// Create a default config manager for backward compatibility
	configManager, _ := createDefaultConfigManager()
	return &Validator{
		checks:        GetChecks(),
		validKeys:     GetValidKeys(),
		configManager: configManager,
	}
}

// NewWithExcludes creates a new Validator instance with default configuration
// Deprecated: Use NewWithConfig instead. Exclude patterns should be handled via configuration files.
func NewWithExcludes(excludePatterns []string) *Validator {
	// Create a default config manager for backward compatibility
	configManager, _ := createDefaultConfigManager()
	return &Validator{
		checks:        GetChecks(),
		validKeys:     GetValidKeys(),
		configManager: configManager,
	}
}

// createDefaultConfigManager creates a config manager with default configuration
func createDefaultConfigManager() (ConfigManager, error) {
	// Import the config package to avoid circular imports
	// We'll create a simple implementation here
	return &defaultConfigManager{}, nil
}

// defaultConfigManager provides default configuration when no config file is used
type defaultConfigManager struct{}

func (dcm *defaultConfigManager) GetEnabledChecksForPath(filePath string) []string {
	// Return all checks enabled by default - this matches the old behavior
	// where all checks were enabled unless specifically ignored
	return []string{
		"NO_FRONT_MATTER",
		"NO_TRAILING_NEWLINE",
		"UNKNOWN_ATTRIBUTE",
		"NO_TITLE",
		"LONG_TITLE",
		"SHORT_TITLE",
		"NO_DESCRIPTION",
		"LONG_DESCRIPTION",
		"SHORT_DESCRIPTION",
		"NO_FULL_STOP_DESCRIPTION",
		"INVALID_DESCRIPTION",
		"NO_LINK_TITLE",
		"LONG_LINK_TITLE",
		"NO_WEIGHT",
		"NO_OWNER",
		"INVALID_OWNER",
		"NO_LAST_REVIEW_DATE",
		"REVIEW_TOO_LONG_AGO",
		"INVALID_LAST_REVIEW_DATE",
		"NO_USER_QUESTIONS",
		"LONG_USER_QUESTION",
		"NO_QUESTION_MARK",
		// Runbook checks
		"RUNBOOK_LAYOUT_NOT_SET",
		"INVALID_RUNBOOK_VARIABLES",
		"RUNBOOK_VARIABLE_WITHOUT_NAME",
		"INVALID_RUNBOOK_VARIABLE_NAME",
		"INVALID_RUNBOOK_VARIABLE",
		"INVALID_RUNBOOK_DASHBOARDS",
		"INVALID_RUNBOOK_DASHBOARD",
		"INVALID_RUNBOOK_DASHBOARD_LINK",
		"INVALID_RUNBOOK_KNOWN_ISSUES",
		"INVALID_RUNBOOK_KNOWN_ISSUE",
		"INVALID_RUNBOOK_KNOWN_ISSUE_URL",
		"RUNBOOK_APPEARS_IN_MENU",
	}
}

// NewWithConfig creates a new Validator instance with a configuration manager
func NewWithConfig(configManager ConfigManager) *Validator {
	return &Validator{
		checks:        GetChecks(),
		validKeys:     GetValidKeys(),
		configManager: configManager,
	}
}

// ValidateFile validates the content of a single markdown file
func (v *Validator) ValidateFile(content, filePath string) ValidationResult {
	result := ValidationResult{
		NumFrontMatterLines: 0,
		Checks:              []CheckResult{},
	}

	// Check for trailing newline
	if !strings.HasSuffix(content, "\n") {
		if !v.shouldSkipCheck(filePath, NoTrailingNewline) {
			result.Checks = append(result.Checks, CheckResult{
				Check: NoTrailingNewline,
				Line:  1 + strings.Count(content, "\n"),
			})
		}
		return result
	}

	// Parse frontmatter
	frontMatter, fmString, numFMLines, err := v.parseFrontMatter(content)
	if err != nil || frontMatter == nil {
		if !v.shouldSkipCheck(filePath, NoFrontMatter) {
			result.Checks = append(result.Checks, CheckResult{
				Check: NoFrontMatter,
				Line:  1,
			})
		}
		return result
	}

	result.NumFrontMatterLines = numFMLines

	// Run validations
	v.validateAll(frontMatter, fmString, filePath, &result)

	return result
}

// parseFrontMatter extracts and parses the frontmatter from content
func (v *Validator) parseFrontMatter(content string) (*FrontMatter, string, int, error) {
	// Find frontmatter boundaries
	re := regexp.MustCompile(`(?m)^---\n`)
	matches := re.FindAllStringIndex(content, -1)
	if len(matches) < 1 {
		return nil, "", 0, nil
	} else if len(matches) < 2 {
		return nil, "", 0, fmt.Errorf("invalid frontmatter format")
	}

	start := matches[0][1] // After first "---\n"
	end := matches[1][0]   // Before second "---"

	fmString := content[start:end]
	numLines := 1 + strings.Count(fmString, "\n")

	var frontMatter FrontMatter
	if err := yaml.Unmarshal([]byte(fmString), &frontMatter); err != nil {
		return nil, "", numLines, err
	}

	return &frontMatter, fmString, numLines, nil
}

// validateAll runs all validation checks
func (v *Validator) validateAll(fm *FrontMatter, fmString, filePath string, result *ValidationResult) {
	// Validate unknown attributes
	v.validateUnknownAttributes(fmString, result)

	// Validate title
	v.validateTitle(fm, result)

	// Validate description
	v.validateDescription(fm, filePath, result)

	// Validate linkTitle
	v.validateLinkTitle(fm, filePath, result)

	// Validate menu and weight
	v.validateMenuAndWeight(fm, result)

	// Validate owner
	v.validateOwner(fm, filePath, result)

	// Validate user questions
	v.validateUserQuestions(fm, filePath, result)

	// Validate last review date
	v.validateLastReviewDate(fm, filePath, result)

	// Validate runbook
	v.validateRunbook(fm, filePath, result)
}

// validateUnknownAttributes checks for unknown frontmatter attributes
func (v *Validator) validateUnknownAttributes(fmString string, result *ValidationResult) {
	// Parse the frontmatter into a generic map to check for unknown keys
	var generic map[string]interface{}
	if err := yaml.Unmarshal([]byte(fmString), &generic); err != nil {
		return
	}

	for key := range generic {
		if !v.validKeys[key] {
			result.Checks = append(result.Checks, CheckResult{
				Check: UnknownAttribute,
				Value: key,
			})
		}
	}
}

// validateTitle validates the title field
func (v *Validator) validateTitle(fm *FrontMatter, result *ValidationResult) {
	if fm.Title == "" {
		result.Checks = append(result.Checks, CheckResult{
			Check: NoTitle,
		})
	} else {
		if len(fm.Title) < 5 {
			result.Checks = append(result.Checks, CheckResult{
				Check: ShortTitle,
				Value: fm.Title,
			})
		}
		if len(fm.Title) > 100 {
			result.Checks = append(result.Checks, CheckResult{
				Check: LongTitle,
				Value: fm.Title,
			})
		}
	}
}

// validateDescription validates the description field
func (v *Validator) validateDescription(fm *FrontMatter, filePath string, result *ValidationResult) {
	if fm.Description == "" {
		if !v.shouldSkipCheck(filePath, NoDescription) {
			result.Checks = append(result.Checks, CheckResult{
				Check: NoDescription,
			})
		}
	} else {
		// Check for line breaks
		if strings.Contains(strings.TrimSpace(fm.Description), "\n") {
			if !v.shouldSkipCheck(filePath, InvalidDescription) {
				result.Checks = append(result.Checks, CheckResult{
					Check: InvalidDescription,
					Value: fm.Description,
				})
			}
		} else {
			if len(fm.Description) < 50 {
				if !v.shouldSkipCheck(filePath, ShortDescription) {
					result.Checks = append(result.Checks, CheckResult{
						Check: ShortDescription,
						Value: fm.Description,
					})
				}
			}
			if len(fm.Description) > 300 {
				if !v.shouldSkipCheck(filePath, LongDescription) {
					result.Checks = append(result.Checks, CheckResult{
						Check: LongDescription,
						Value: fm.Description,
					})
				}
			}
			if !strings.HasSuffix(fm.Description, ".") {
				if !v.shouldSkipCheck(filePath, NoFullStopDescription) {
					result.Checks = append(result.Checks, CheckResult{
						Check: NoFullStopDescription,
						Value: fm.Description,
					})
				}
			}
		}
	}
}

// validateLinkTitle validates the linkTitle field
func (v *Validator) validateLinkTitle(fm *FrontMatter, filePath string, result *ValidationResult) {
	linkTitle := fm.LinkTitle
	if linkTitle == "" {
		linkTitle = fm.Title
	}

	if len(linkTitle) > 40 && !v.shouldSkipCheck(filePath, LongLinkTitle) {
		result.Checks = append(result.Checks, CheckResult{
			Check: LongLinkTitle,
			Value: linkTitle,
		})
	}
}

// validateMenuAndWeight validates menu and weight fields
func (v *Validator) validateMenuAndWeight(fm *FrontMatter, result *ValidationResult) {
	if fm.Menu != nil {
		if fm.LinkTitle == "" && fm.Title == "" {
			result.Checks = append(result.Checks, CheckResult{
				Check: NoLinkTitle,
			})
		}
		if fm.Weight == nil {
			result.Checks = append(result.Checks, CheckResult{
				Check: NoWeight,
			})
		}
	}
}

// validateOwner validates the owner field
func (v *Validator) validateOwner(fm *FrontMatter, filePath string, result *ValidationResult) {
	if len(fm.Owner) == 0 {
		if !v.shouldSkipCheck(filePath, NoOwner) {
			result.Checks = append(result.Checks, CheckResult{
				Check: NoOwner,
			})
		}
	} else {
		for _, owner := range fm.Owner {
			if !strings.HasPrefix(owner, "https://github.com/orgs/giantswarm/teams/") {
				result.Checks = append(result.Checks, CheckResult{
					Check: InvalidOwner,
					Value: fm.Owner,
				})
				break
			}
		}
	}
}

// validateUserQuestions validates the user_questions field
func (v *Validator) validateUserQuestions(fm *FrontMatter, filePath string, result *ValidationResult) {
	if len(fm.UserQuestions) == 0 {
		if !v.shouldSkipCheck(filePath, NoUserQuestions) && !strings.HasSuffix(filePath, "_index.md") {
			result.Checks = append(result.Checks, CheckResult{
				Check: NoUserQuestions,
			})
		}
	} else {
		for _, question := range fm.UserQuestions {
			if len(question) > 100 {
				result.Checks = append(result.Checks, CheckResult{
					Check: LongUserQuestion,
					Value: question,
				})
			}
			if !strings.HasSuffix(question, "?") {
				result.Checks = append(result.Checks, CheckResult{
					Check: NoQuestionMark,
					Value: question,
				})
			}
		}
	}
}

// validateLastReviewDate validates the last_review_date field
func (v *Validator) validateLastReviewDate(fm *FrontMatter, filePath string, result *ValidationResult) {
	if fm.LastReviewDate == nil {
		if !v.shouldSkipCheck(filePath, NoLastReviewDate) {
			result.Checks = append(result.Checks, CheckResult{
				Check: NoLastReviewDate,
			})
		}
	} else {
		today := time.Now()

		// Check if date is in the future
		if fm.LastReviewDate.Time.After(today) {
			result.Checks = append(result.Checks, CheckResult{
				Check: InvalidLastReviewDate,
				Value: fm.LastReviewDate.Time.Format("2006-01-02"),
				Title: fm.Title,
				Owner: fm.Owner,
			})
		} else if !v.shouldSkipCheck(filePath, ReviewTooLongAgo) {
			// Check if review is too long ago
			expiration := 365
			if fm.ExpirationInDays != nil {
				expiration = *fm.ExpirationInDays
			}

			if today.Sub(fm.LastReviewDate.Time) > time.Duration(expiration)*24*time.Hour {
				result.Checks = append(result.Checks, CheckResult{
					Check: ReviewTooLongAgo,
					Value: fm.LastReviewDate.Time.Format("2006-01-02"),
					Title: fm.Title,
					Owner: fm.Owner,
				})
			}
		}
	}
}

// shouldSkipCheck checks if a check should be skipped based on configuration
func (v *Validator) shouldSkipCheck(filePath, checkID string) bool {
	// Use config manager to determine enabled checks
	if v.configManager != nil {
		enabledChecks := v.configManager.GetEnabledChecksForPath(filePath)

		// If check is not in enabled list, skip it
		for _, enabledCheck := range enabledChecks {
			if enabledCheck == checkID {
				return false // Check is enabled
			}
		}
		return true // Check not found in enabled list, skip it
	}

	// If no config manager, don't skip any checks (fallback behavior)
	return false
}

// validateRunbook validates runbook-specific fields
func (v *Validator) validateRunbook(fm *FrontMatter, filePath string, result *ValidationResult) {
	// Check if this is a runbook page
	if fm.Layout == "runbook" {
		// Validate that runbook pages have toc_hide: true
		if !fm.TocHide {
			if !v.shouldSkipCheck(filePath, RunbookAppearsInMenu) {
				result.Checks = append(result.Checks, CheckResult{
					Check: RunbookAppearsInMenu,
				})
			}
		}

		// Validate runbook configuration
		if fm.Runbook == nil {
			// If layout is runbook but no runbook config, that's an issue
			// We'll let the individual field validations catch the specifics
			return
		}

		v.validateRunbookVariables(fm.Runbook, filePath, result)
		v.validateRunbookDashboards(fm.Runbook, filePath, result)
		v.validateRunbookKnownIssues(fm.Runbook, filePath, result)
	} else if fm.Runbook != nil {
		// If runbook config exists but layout is not runbook
		if !v.shouldSkipCheck(filePath, RunbookLayoutNotSet) {
			result.Checks = append(result.Checks, CheckResult{
				Check: RunbookLayoutNotSet,
			})
		}
	}
}

// validateRunbookVariables validates the runbook variables section
func (v *Validator) validateRunbookVariables(runbook *Runbook, filePath string, result *ValidationResult) {
	// Variables are optional, so only validate if they exist
	if len(runbook.Variables) == 0 {
		return
	}

	// Track variable names for uniqueness check
	variableNames := make(map[string]bool)
	variableNameRegex := regexp.MustCompile(`^[A-Z_]+$`)

	for i, variable := range runbook.Variables {
		// Check if variable has a name
		if variable.Name == "" {
			if !v.shouldSkipCheck(filePath, RunbookVariableWithoutName) {
				result.Checks = append(result.Checks, CheckResult{
					Check: RunbookVariableWithoutName,
					Value: fmt.Sprintf("Variable at index %d", i),
				})
			}
			continue
		}

		// Check variable name format (uppercase letters and underscores only)
		if !variableNameRegex.MatchString(variable.Name) {
			if !v.shouldSkipCheck(filePath, InvalidRunbookVariableName) {
				result.Checks = append(result.Checks, CheckResult{
					Check: InvalidRunbookVariableName,
					Value: variable.Name,
				})
			}
		}

		// Check variable name uniqueness
		if variableNames[variable.Name] {
			if !v.shouldSkipCheck(filePath, InvalidRunbookVariableName) {
				result.Checks = append(result.Checks, CheckResult{
					Check: InvalidRunbookVariableName,
					Value: fmt.Sprintf("Duplicate variable name: %s", variable.Name),
				})
			}
		}
		variableNames[variable.Name] = true

		// Validate variable structure (name is required, description and default are optional)
		// This is mostly structural validation - the YAML parsing handles most of this
		// but we can add additional checks if needed
	}
}

// validateRunbookDashboards validates the runbook dashboards section
func (v *Validator) validateRunbookDashboards(runbook *Runbook, filePath string, result *ValidationResult) {
	// Dashboards are optional, so only validate if they exist
	if len(runbook.Dashboards) == 0 {
		return
	}

	// Get variable names for link validation
	variableNames := make(map[string]bool)
	for _, variable := range runbook.Variables {
		if variable.Name != "" {
			variableNames[variable.Name] = true
		}
	}

	for i, dashboard := range runbook.Dashboards {
		// Check if dashboard has name and link
		if dashboard.Name == "" || dashboard.Link == "" {
			if !v.shouldSkipCheck(filePath, InvalidRunbookDashboard) {
				result.Checks = append(result.Checks, CheckResult{
					Check: InvalidRunbookDashboard,
					Value: fmt.Sprintf("Dashboard at index %d missing name or link", i),
				})
			}
			continue
		}

		// Validate dashboard link
		v.validateRunbookDashboardLink(dashboard.Link, variableNames, filePath, result)
	}
}

// validateRunbookDashboardLink validates a dashboard link URL and variable usage
func (v *Validator) validateRunbookDashboardLink(link string, variableNames map[string]bool, filePath string, result *ValidationResult) {
	// Find all variables in the link (format: $VARIABLE_NAME)
	variableRegex := regexp.MustCompile(`\$([A-Z_]+)`)
	matches := variableRegex.FindAllStringSubmatch(link, -1)

	// Check if all variables used in the link are defined
	for _, match := range matches {
		if len(match) > 1 {
			variableName := match[1]
			if !variableNames[variableName] {
				if !v.shouldSkipCheck(filePath, InvalidRunbookDashboardLink) {
					result.Checks = append(result.Checks, CheckResult{
						Check: InvalidRunbookDashboardLink,
						Value: fmt.Sprintf("Undefined variable $%s in link: %s", variableName, link),
					})
				}
			}
		}
	}

	// Replace variables with dummy values for URL validation
	testLink := link
	for variableName := range variableNames {
		testLink = strings.ReplaceAll(testLink, "$"+variableName, "test")
	}

	// Basic URL validation
	if !strings.HasPrefix(testLink, "http://") && !strings.HasPrefix(testLink, "https://") {
		if !v.shouldSkipCheck(filePath, InvalidRunbookDashboardLink) {
			result.Checks = append(result.Checks, CheckResult{
				Check: InvalidRunbookDashboardLink,
				Value: fmt.Sprintf("Invalid URL format: %s", link),
			})
		}
	}
}

// validateRunbookKnownIssues validates the runbook known issues section
func (v *Validator) validateRunbookKnownIssues(runbook *Runbook, filePath string, result *ValidationResult) {
	// Known issues are optional, so only validate if they exist
	// The INVALID_RUNBOOK_KNOWN_ISSUES check would be triggered during YAML parsing
	// if the structure is invalid (e.g., not an array)
	if len(runbook.KnownIssues) == 0 {
		return
	}

	for i, issue := range runbook.KnownIssues {
		// Check if known issue has URL
		if issue.URL == "" {
			if !v.shouldSkipCheck(filePath, InvalidRunbookKnownIssue) {
				result.Checks = append(result.Checks, CheckResult{
					Check: InvalidRunbookKnownIssue,
					Value: fmt.Sprintf("Known issue at index %d missing URL", i),
				})
			}
			continue
		}

		// Validate URL format
		if !strings.HasPrefix(issue.URL, "http://") && !strings.HasPrefix(issue.URL, "https://") {
			if !v.shouldSkipCheck(filePath, InvalidRunbookKnownIssueURL) {
				result.Checks = append(result.Checks, CheckResult{
					Check: InvalidRunbookKnownIssueURL,
					Value: issue.URL,
				})
			}
		}
	}
}
