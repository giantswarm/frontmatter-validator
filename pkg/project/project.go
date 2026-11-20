package project

import (
	"fmt"
	"runtime/debug"
)

var (
	version        string
	gitSHA         string
	buildTimestamp string
)

// Version returns the release version, commit and build date.
func Version() string {
	buildInfo, _ := debug.ReadBuildInfo()
	return versionString(buildInfo)
}

func versionString(buildInfo *debug.BuildInfo) string {
	releaseVersion, commit, buildDate := version, gitSHA, buildTimestamp
	if buildInfo != nil {
		if releaseVersion == "" && buildInfo.Main.Version != "(devel)" {
			releaseVersion = buildInfo.Main.Version
		}
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				if commit == "" {
					commit = setting.Value
				}
			case "vcs.time":
				if buildDate == "" {
					buildDate = setting.Value
				}
			}
		}
	}
	if releaseVersion == "" {
		releaseVersion = "dev"
	}
	if commit == "" {
		commit = "unknown"
	}
	if buildDate == "" {
		buildDate = "unknown"
	}
	return fmt.Sprintf("%s (commit: %s, built: %s)", releaseVersion, commit, buildDate)
}
