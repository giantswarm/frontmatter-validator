package config

import (
	"fmt"
	"os"
	"strings"

	"go.yaml.in/yaml/v4"

	"github.com/giantswarm/frontmatter-validator/pkg/validator"
)

// Manager handles loading and resolving configuration
type Manager struct {
	config     *Config
	configPath string
}

// NewManager creates a new configuration manager
func NewManager(configPath string) (*Manager, error) {
	manager := &Manager{
		configPath: configPath,
	}

	if err := manager.loadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return manager, nil
}

// loadConfig loads the configuration from the specified file
func (m *Manager) loadConfig() error {
	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// If no config file exists, use default configuration
		m.config = m.getDefaultConfig()
		return nil
	}

	// Read config file
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", m.configPath, err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", m.configPath, err)
	}

	m.config = &config
	return nil
}

// GetEnabledChecksForPath returns the list of enabled checks for a given file path
func (m *Manager) GetEnabledChecksForPath(filePath string) []string {
	if m.config == nil {
		return []string{}
	}

	// Start with default enabled checks
	enabled := make(map[string]bool)
	for _, check := range m.config.DefaultRules.EnabledChecks {
		enabled[check] = true
	}

	// Apply default disabled checks
	for _, check := range m.config.DefaultRules.DisabledChecks {
		enabled[check] = false
	}

	// Apply directory overrides in order
	for _, override := range m.config.DirectoryOverrides {
		if m.pathMatches(filePath, override.Path) {
			// Enable specified checks
			for _, check := range override.EnabledChecks {
				enabled[check] = true
			}
			// Disable specified checks
			for _, check := range override.DisabledChecks {
				enabled[check] = false
			}
		}
	}

	// Convert back to slice
	var result []string
	for check, isEnabled := range enabled {
		if isEnabled {
			result = append(result, check)
		}
	}

	return result
}

// pathMatches checks if a file path matches a glob pattern
func (m *Manager) pathMatches(filePath, pattern string) bool {
	// Normalize paths by removing leading "./"
	filePath = strings.TrimPrefix(filePath, "./")
	pattern = strings.TrimPrefix(pattern, "./")

	// Handle simple glob patterns
	if strings.HasSuffix(pattern, "/**") {
		// Directory prefix match: "src/content/vintage/**"
		prefix := strings.TrimSuffix(pattern, "/**")
		return strings.HasPrefix(filePath, prefix+"/") || filePath == prefix
	}

	if strings.HasSuffix(pattern, "/*") {
		// Single level wildcard: "src/content/*"
		prefix := strings.TrimSuffix(pattern, "/*")
		relativePath := strings.TrimPrefix(filePath, prefix+"/")
		return strings.HasPrefix(filePath, prefix+"/") && !strings.Contains(relativePath, "/")
	}

	// Exact match
	return filePath == pattern
}

// getDefaultConfig returns the default configuration that matches current hardcoded behavior
func (m *Manager) getDefaultConfig() *Config {
	return &Config{
		DefaultRules: RuleSet{
			EnabledChecks: []string{
				validator.NoFrontMatter,
				validator.NoTrailingNewline,
				validator.UnknownAttribute,
				validator.NoTitle,
				validator.LongTitle,
				validator.ShortTitle,
				validator.NoDescription,
				validator.LongDescription,
				validator.ShortDescription,
				validator.NoFullStopDescription,
				validator.InvalidDescription,
				validator.NoLinkTitle,
				validator.LongLinkTitle,
				validator.NoWeight,
				validator.NoOwner,
				validator.InvalidOwner,
				validator.NoLastReviewDate,
				validator.ReviewTooLongAgo,
				validator.InvalidLastReviewDate,
				validator.NoUserQuestions,
				validator.LongUserQuestion,
				validator.NoQuestionMark,
				// INVALID_DIATAXIS_CONTENT_TYPE is safe to enable by default (only fires on a
				// bad value). NO_DIATAXIS_CONTENT_TYPE is opt-in — enable it once pages are tagged.
				validator.InvalidDiataxisContentType,
			},
		},
		DirectoryOverrides: []DirectoryOverride{
			{
				Path: "src/content/reference/platform-api/crd/**",
				DisabledChecks: []string{
					validator.NoDescription,
					validator.LongDescription,
					validator.ShortDescription,
					validator.NoFullStopDescription,
					validator.InvalidDescription,
					validator.NoLinkTitle,
					validator.NoOwner,
					validator.NoUserQuestions,
				},
			},
			{
				Path: "src/content/vintage/use-the-api/management-api/crd/**",
				DisabledChecks: []string{
					validator.NoDescription,
					validator.LongDescription,
					validator.ShortDescription,
					validator.NoFullStopDescription,
					validator.InvalidDescription,
					validator.NoLinkTitle,
					validator.NoOwner,
					validator.NoUserQuestions,
				},
			},
			{
				Path: "src/content/changes/**",
				DisabledChecks: []string{
					validator.NoDescription,
					validator.ShortDescription,
					validator.NoFullStopDescription,
					validator.NoLinkTitle,
					validator.LongLinkTitle,
					validator.NoOwner,
					validator.NoUserQuestions,
				},
			},
			{
				Path: "src/content/vintage/**",
				DisabledChecks: []string{
					validator.NoLastReviewDate,
					validator.ReviewTooLongAgo,
				},
			},
			{
				Path: "src/content/reference/platform-api/cluster-apps/**",
				DisabledChecks: []string{
					validator.NoLastReviewDate,
					validator.ReviewTooLongAgo,
				},
			},
			{
				Path: "src/content/meta/**",
				DisabledChecks: []string{
					validator.NoLastReviewDate,
				},
			},
		},
	}
}

// IsPathIgnored checks if a file path should be completely ignored based on ignore_paths configuration
func (m *Manager) IsPathIgnored(filePath string) bool {
	if m.config == nil {
		return false
	}

	for _, pattern := range m.config.IgnorePaths {
		if m.pathMatches(filePath, pattern) {
			return true
		}
	}

	return false
}

// GetConfig returns the loaded configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// GetConfigPath returns the path to the configuration file
func (m *Manager) GetConfigPath() string {
	return m.configPath
}
