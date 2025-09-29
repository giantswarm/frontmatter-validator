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
	excludeConfig *ExcludeConfig
}

// New creates a new Validator instance
func New() *Validator {
	return &Validator{
		checks:        GetChecks(),
		validKeys:     GetValidKeys(),
		excludeConfig: nil,
	}
}

// NewWithExcludes creates a new Validator instance with exclude patterns
func NewWithExcludes(excludePatterns []string) *Validator {
	return &Validator{
		checks:        GetChecks(),
		validKeys:     GetValidKeys(),
		excludeConfig: NewExcludeConfig(excludePatterns),
	}
}

// ValidateFile validates the content of a single markdown file
func (v *Validator) ValidateFile(content, filePath, validationMode string) ValidationResult {
	result := ValidationResult{
		NumFrontMatterLines: 0,
		Checks:              []CheckResult{},
	}

	// Check for trailing newline
	if !strings.HasSuffix(content, "\n") {
		if !v.shouldSkipCheck(filePath, NoTrailingNewline) {
			result.Checks = append(result.Checks, CheckResult{
				Check: NoTrailingNewline,
				Line:  strings.Count(content, "\n"),
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

	// Run validations based on mode
	if validationMode == ValidateAll {
		v.validateAll(frontMatter, fmString, filePath, &result)
	} else if validationMode == ValidateLastReviewDate {
		v.validateLastReviewDate(frontMatter, filePath, &result)
	}

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
	numLines := strings.Count(fmString, "\n")

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
		if fm.LastReviewDate.After(today) {
			result.Checks = append(result.Checks, CheckResult{
				Check: InvalidLastReviewDate,
				Value: fm.LastReviewDate.Format("2006-01-02"),
				Title: fm.Title,
				Owner: fm.Owner,
			})
		} else if !v.shouldSkipCheck(filePath, ReviewTooLongAgo) {
			// Check if review is too long ago
			expiration := 365
			if fm.ExpirationInDays != nil {
				expiration = *fm.ExpirationInDays
			}

			if today.Sub(*fm.LastReviewDate) > time.Duration(expiration)*24*time.Hour {
				result.Checks = append(result.Checks, CheckResult{
					Check: ReviewTooLongAgo,
					Value: fm.LastReviewDate.Format("2006-01-02"),
					Title: fm.Title,
					Owner: fm.Owner,
				})
			}
		}
	}
}

// isIgnoredPath checks if a file path should be ignored for a specific check
func (v *Validator) isIgnoredPath(filePath, checkID string) bool {
	check := GetCheckByID(checkID)
	if check == nil {
		return false
	}

	for _, ignorePath := range check.IgnorePaths {
		if strings.HasPrefix(filePath, ignorePath) {
			return true
		}
	}

	return false
}

// shouldSkipCheck checks if a check should be skipped due to exclusion patterns
func (v *Validator) shouldSkipCheck(filePath, checkID string) bool {
	// Check built-in ignore paths
	if v.isIgnoredPath(filePath, checkID) {
		return true
	}

	// Check user-defined exclusion patterns
	if v.excludeConfig != nil && v.excludeConfig.ShouldExclude(filePath, checkID) {
		return true
	}

	return false
}
