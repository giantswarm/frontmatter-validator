package config

// Config represents the complete configuration for frontmatter validation
type Config struct {
	DefaultRules       RuleSet             `yaml:"default_rules"`
	DirectoryOverrides []DirectoryOverride `yaml:"directory_overrides"`
}

// RuleSet defines which validation checks are enabled or disabled
type RuleSet struct {
	EnabledChecks  []string `yaml:"enabled_checks"`
	DisabledChecks []string `yaml:"disabled_checks,omitempty"`
}

// DirectoryOverride allows overriding rules for specific directory patterns
type DirectoryOverride struct {
	Path           string   `yaml:"path"`                      // Glob pattern like "src/content/vintage/**"
	EnabledChecks  []string `yaml:"enabled_checks,omitempty"`  // Additional checks to enable for this path
	DisabledChecks []string `yaml:"disabled_checks,omitempty"` // Checks to disable for this path
}
