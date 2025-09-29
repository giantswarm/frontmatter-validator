package validator

import (
	"strings"
)

// ExcludePattern represents a path exclusion pattern
type ExcludePattern struct {
	Path   string   // The path pattern to exclude
	Checks []string // Specific checks to exclude (empty means exclude all checks)
}

// ExcludeConfig manages path-based exclusions
type ExcludeConfig struct {
	patterns []ExcludePattern
}

// NewExcludeConfig creates a new exclude configuration from pattern strings
func NewExcludeConfig(patterns []string) *ExcludeConfig {
	config := &ExcludeConfig{
		patterns: make([]ExcludePattern, 0, len(patterns)),
	}

	for _, pattern := range patterns {
		config.addPattern(pattern)
	}

	return config
}

// addPattern parses and adds a pattern string
func (ec *ExcludeConfig) addPattern(pattern string) {
	if pattern == "" {
		return
	}

	// Split by colon to separate path from checks
	parts := strings.SplitN(pattern, ":", 2)
	path := strings.TrimSpace(parts[0])

	if path == "" {
		return
	}

	excludePattern := ExcludePattern{
		Path:   path,
		Checks: []string{},
	}

	// If there's a second part, parse the check names
	if len(parts) > 1 {
		checksStr := strings.TrimSpace(parts[1])
		if checksStr != "" {
			// Split by comma for multiple checks
			checkNames := strings.Split(checksStr, ",")
			for _, checkName := range checkNames {
				checkName = strings.TrimSpace(checkName)
				if checkName != "" {
					excludePattern.Checks = append(excludePattern.Checks, checkName)
				}
			}
		}
	}

	ec.patterns = append(ec.patterns, excludePattern)
}

// ShouldExclude checks if a file path should be excluded from a specific check
func (ec *ExcludeConfig) ShouldExclude(filePath, checkID string) bool {
	if ec == nil {
		return false
	}

	for _, pattern := range ec.patterns {
		// Check if the file path matches the exclude pattern
		if ec.pathMatches(filePath, pattern.Path) {
			// If no specific checks are listed, exclude from all checks
			if len(pattern.Checks) == 0 {
				return true
			}

			// Check if this specific check should be excluded
			for _, excludeCheck := range pattern.Checks {
				if excludeCheck == checkID {
					return true
				}
			}
		}
	}

	return false
}

// pathMatches checks if a file path matches an exclude pattern
func (ec *ExcludeConfig) pathMatches(filePath, pattern string) bool {
	// Normalize paths by removing leading/trailing slashes and "./"
	filePath = strings.TrimPrefix(filePath, "./")
	filePath = strings.Trim(filePath, "/")
	pattern = strings.TrimPrefix(pattern, "./")
	pattern = strings.Trim(pattern, "/")

	// Exact match
	if filePath == pattern {
		return true
	}

	// Prefix match (pattern is a parent directory)
	if strings.HasPrefix(filePath, pattern+"/") {
		return true
	}

	return false
}

// GetPatterns returns all configured patterns (for debugging/testing)
func (ec *ExcludeConfig) GetPatterns() []ExcludePattern {
	if ec == nil {
		return nil
	}
	return ec.patterns
}
