package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/giantswarm/frontmatter-validator/pkg/validator"
)

// Shared fixtures for the tests in this file.
const (
	testDocsExamplePath = "src/content/docs/example.md"
	testVintageGlob     = "src/content/vintage/**"
	testReadmePath      = "README.md"
)

func TestManager_GetEnabledChecksForPath(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		filePath string
		want     []string
	}{
		{
			name: "default rules only",
			config: &Config{
				DefaultRules: RuleSet{
					EnabledChecks: []string{validator.NoTitle, validator.NoDescription},
				},
			},
			filePath: testDocsExamplePath,
			want:     []string{validator.NoTitle, validator.NoDescription},
		},
		{
			name: "directory override disables check",
			config: &Config{
				DefaultRules: RuleSet{
					EnabledChecks: []string{validator.NoTitle, validator.NoDescription},
				},
				DirectoryOverrides: []DirectoryOverride{
					{
						Path:           testVintageGlob,
						DisabledChecks: []string{validator.NoDescription},
					},
				},
			},
			filePath: "src/content/vintage/docs/example.md",
			want:     []string{validator.NoTitle},
		},
		{
			name: "directory override enables additional check",
			config: &Config{
				DefaultRules: RuleSet{
					EnabledChecks: []string{validator.NoTitle},
				},
				DirectoryOverrides: []DirectoryOverride{
					{
						Path:          "src/content/special/**",
						EnabledChecks: []string{validator.NoDescription},
					},
				},
			},
			filePath: "src/content/special/example.md",
			want:     []string{validator.NoTitle, validator.NoDescription},
		},
		{
			name: "multiple overrides - most specific wins",
			config: &Config{
				DefaultRules: RuleSet{
					EnabledChecks: []string{validator.NoTitle, validator.NoDescription, validator.NoOwner},
				},
				DirectoryOverrides: []DirectoryOverride{
					{
						Path:           "src/content/**",
						DisabledChecks: []string{validator.NoDescription},
					},
					{
						Path:          "src/content/special/**",
						EnabledChecks: []string{validator.NoDescription},
					},
				},
			},
			filePath: "src/content/special/example.md",
			want:     []string{validator.NoTitle, validator.NoOwner, validator.NoDescription},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{config: tt.config}
			got := m.GetEnabledChecksForPath(tt.filePath)

			// Convert to map for easier comparison
			gotMap := make(map[string]bool)
			for _, check := range got {
				gotMap[check] = true
			}

			wantMap := make(map[string]bool)
			for _, check := range tt.want {
				wantMap[check] = true
			}

			if len(gotMap) != len(wantMap) {
				t.Errorf("GetEnabledChecksForPath() got %v checks, want %v checks", len(gotMap), len(wantMap))
				t.Errorf("Got: %v", got)
				t.Errorf("Want: %v", tt.want)
				return
			}

			for check := range wantMap {
				if !gotMap[check] {
					t.Errorf("GetEnabledChecksForPath() missing check %v", check)
				}
			}

			for check := range gotMap {
				if !wantMap[check] {
					t.Errorf("GetEnabledChecksForPath() unexpected check %v", check)
				}
			}
		})
	}
}

func TestManager_pathMatches(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		pattern  string
		want     bool
	}{
		{
			name:     "exact match",
			filePath: testDocsExamplePath,
			pattern:  testDocsExamplePath,
			want:     true,
		},
		{
			name:     "directory wildcard match",
			filePath: "src/content/vintage/docs/example.md",
			pattern:  testVintageGlob,
			want:     true,
		},
		{
			name:     "directory wildcard no match",
			filePath: testDocsExamplePath,
			pattern:  testVintageGlob,
			want:     false,
		},
		{
			name:     "single level wildcard match",
			filePath: "src/content/example.md",
			pattern:  "src/content/*",
			want:     true,
		},
		{
			name:     "single level wildcard no match - too deep",
			filePath: testDocsExamplePath,
			pattern:  "src/content/*",
			want:     false,
		},
		{
			name:     "normalize leading dot-slash",
			filePath: "./src/content/docs/example.md",
			pattern:  "src/content/docs/**",
			want:     true,
		},
		{
			name:     "directory exact match",
			filePath: "src/content/vintage",
			pattern:  testVintageGlob,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{}
			got := m.pathMatches(tt.filePath, tt.pattern)
			if got != tt.want {
				t.Errorf("pathMatches(%q, %q) = %v, want %v", tt.filePath, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestNewManager_WithConfigFile(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	configContent := `default_rules:
  enabled_checks:
    - "NO_TITLE"
    - "NO_DESCRIPTION"
directory_overrides:
  - path: "test/**"
    disabled_checks:
      - "NO_DESCRIPTION"
`

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loading the config
	manager, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Test that config was loaded correctly
	checks := manager.GetEnabledChecksForPath("test/example.md")
	expected := []string{validator.NoTitle}

	if len(checks) != len(expected) {
		t.Errorf("Expected %d checks, got %d", len(expected), len(checks))
	}

	checkMap := make(map[string]bool)
	for _, check := range checks {
		checkMap[check] = true
	}

	for _, expectedCheck := range expected {
		if !checkMap[expectedCheck] {
			t.Errorf("Expected check %s not found in result", expectedCheck)
		}
	}
}

func TestManager_IsPathIgnored(t *testing.T) {
	tests := []struct {
		name        string
		ignorePaths []string
		filePath    string
		want        bool
	}{
		{
			name:        "exact match ignored",
			ignorePaths: []string{testReadmePath},
			filePath:    testReadmePath,
			want:        true,
		},
		{
			name:        "exact match not ignored",
			ignorePaths: []string{testReadmePath},
			filePath:    testDocsExamplePath,
			want:        false,
		},
		{
			name:        "glob pattern ignored",
			ignorePaths: []string{"vendor/**"},
			filePath:    "vendor/some/lib/file.md",
			want:        true,
		},
		{
			name:        "glob pattern not ignored",
			ignorePaths: []string{"vendor/**"},
			filePath:    testDocsExamplePath,
			want:        false,
		},
		{
			name:        "multiple patterns",
			ignorePaths: []string{testReadmePath, "CONTRIBUTING.md", ".claude/**"},
			filePath:    ".claude/skills/test/SKILL.md",
			want:        true,
		},
		{
			name:        "empty ignore paths",
			ignorePaths: nil,
			filePath:    testReadmePath,
			want:        false,
		},
		{
			name:        "leading dot-slash normalized",
			ignorePaths: []string{testReadmePath},
			filePath:    "./README.md",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{
				config: &Config{
					IgnorePaths: tt.ignorePaths,
				},
			}
			got := m.IsPathIgnored(tt.filePath)
			if got != tt.want {
				t.Errorf("IsPathIgnored(%q) = %v, want %v", tt.filePath, got, tt.want)
			}
		})
	}
}

func TestNewManager_NoConfigFile(t *testing.T) {
	// Test with non-existent config file
	manager, err := NewManager("/non/existent/config.yaml")
	if err != nil {
		t.Fatalf("NewManager() should not error when config file doesn't exist, got: %v", err)
	}

	// Should use default config
	config := manager.GetConfig()
	if config == nil {
		t.Fatal("Expected default config, got nil")
	}

	if len(config.DefaultRules.EnabledChecks) == 0 {
		t.Error("Expected default config to have enabled checks")
	}
}
