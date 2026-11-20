package project

import (
	"runtime/debug"
	"testing"
)

func TestVersionString(t *testing.T) {
	const (
		developmentVersion = "dev (commit: unknown, built: unknown)"
		releaseVersion     = "1.2.4"
	)
	originalVersion, originalCommit, originalTimestamp := version, gitSHA, buildTimestamp
	t.Cleanup(func() {
		version, gitSHA, buildTimestamp = originalVersion, originalCommit, originalTimestamp
	})
	buildInfo := &debug.BuildInfo{
		Main: debug.Module{Version: "v1.2.3"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "source-commit"},
			{Key: "vcs.time", Value: "2026-11-19T12:00:00Z"},
		},
	}
	testCases := []struct {
		name           string
		version        string
		commit         string
		buildTimestamp string
		buildInfo      *debug.BuildInfo
		expected       string
	}{
		{
			name:     "missing build info",
			expected: developmentVersion,
		},
		{
			name:      "empty build info",
			buildInfo: &debug.BuildInfo{},
			expected:  developmentVersion,
		},
		{
			name:      "development module",
			buildInfo: &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			expected:  developmentVersion,
		},
		{
			name:      "module and VCS fallback",
			buildInfo: buildInfo,
			expected:  "v1.2.3 (commit: source-commit, built: 2026-11-19T12:00:00Z)",
		},
		{
			name:           "linker values take precedence",
			version:        releaseVersion,
			commit:         "release-commit",
			buildTimestamp: "2026-11-20T12:00:00Z",
			buildInfo:      buildInfo,
			expected:       "1.2.4 (commit: release-commit, built: 2026-11-20T12:00:00Z)",
		},
		{
			name:      "partial linker values",
			version:   releaseVersion,
			buildInfo: buildInfo,
			expected:  "1.2.4 (commit: source-commit, built: 2026-11-19T12:00:00Z)",
		},
		{
			name:           "linker values without build info",
			version:        releaseVersion,
			commit:         "release-commit",
			buildTimestamp: "2026-11-20T12:00:00Z",
			expected:       "1.2.4 (commit: release-commit, built: 2026-11-20T12:00:00Z)",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			version, gitSHA, buildTimestamp = testCase.version, testCase.commit, testCase.buildTimestamp
			if actual := versionString(testCase.buildInfo); actual != testCase.expected {
				t.Fatalf("versionString() = %q, want %q", actual, testCase.expected)
			}
		})
	}
}
