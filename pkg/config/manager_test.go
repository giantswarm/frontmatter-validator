package config

import (
	"os"
	"path/filepath"
	"testing"
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
					EnabledChecks: []string{"NO_TITLE", "NO_DESCRIPTION"},
				},
			},
			filePath: "src/content/docs/example.md",
			want:     []string{"NO_TITLE", "NO_DESCRIPTION"},
		},
		{
			name: "directory override disables check",
			config: &Config{
				DefaultRules: RuleSet{
					EnabledChecks: []string{"NO_TITLE", "NO_DESCRIPTION"},
				},
				DirectoryOverrides: []DirectoryOverride{
					{
						Path:           "src/content/vintage/**",
						DisabledChecks: []string{"NO_DESCRIPTION"},
					},
				},
			},
			filePath: "src/content/vintage/docs/example.md",
			want:     []string{"NO_TITLE"},
		},
		{
			name: "directory override enables additional check",
			config: &Config{
				DefaultRules: RuleSet{
					EnabledChecks: []string{"NO_TITLE"},
				},
				DirectoryOverrides: []DirectoryOverride{
					{
						Path:          "src/content/special/**",
						EnabledChecks: []string{"NO_DESCRIPTION"},
					},
				},
			},
			filePath: "src/content/special/example.md",
			want:     []string{"NO_TITLE", "NO_DESCRIPTION"},
		},
		{
			name: "multiple overrides - most specific wins",
			config: &Config{
				DefaultRules: RuleSet{
					EnabledChecks: []string{"NO_TITLE", "NO_DESCRIPTION", "NO_OWNER"},
				},
				DirectoryOverrides: []DirectoryOverride{
					{
						Path:           "src/content/**",
						DisabledChecks: []string{"NO_DESCRIPTION"},
					},
					{
						Path:          "src/content/special/**",
						EnabledChecks: []string{"NO_DESCRIPTION"},
					},
				},
			},
			filePath: "src/content/special/example.md",
			want:     []string{"NO_TITLE", "NO_OWNER", "NO_DESCRIPTION"},
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
			filePath: "src/content/docs/example.md",
			pattern:  "src/content/docs/example.md",
			want:     true,
		},
		{
			name:     "directory wildcard match",
			filePath: "src/content/vintage/docs/example.md",
			pattern:  "src/content/vintage/**",
			want:     true,
		},
		{
			name:     "directory wildcard no match",
			filePath: "src/content/docs/example.md",
			pattern:  "src/content/vintage/**",
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
			filePath: "src/content/docs/example.md",
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
			pattern:  "src/content/vintage/**",
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

	err := os.WriteFile(configPath, []byte(configContent), 0644)
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
	expected := []string{"NO_TITLE"}

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
