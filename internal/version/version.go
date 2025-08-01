// Package version provides version information for the ratchet application.
// It supports both compile-time version injection via ldflags and
// runtime version detection through Go's debug.BuildInfo.
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// These variables are populated at build time using ldflags.
// For example: go build -ldflags "-X github.com/tiernacity/ratchet/internal/version.version=v1.0.0"
var (
	// version is the semantic version (e.g., v1.0.0)
	version = "dev"
	// commit is the git commit SHA
	commit = "none"
	// date is the build date in RFC3339 format
	date = "unknown"
)

// Info contains detailed version information
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
	Compiler  string `json:"compiler"`
	Platform  string `json:"platform"`
}

// GetVersion returns a formatted version string suitable for display.
// In production builds, it shows the version, commit, and build date.
// In development builds, it attempts to use Go module information.
func GetVersion() string {
	if version != "dev" {
		// Production build with injected version
		parts := []string{version}
		if commit != "none" && len(commit) >= 7 {
			parts = append(parts, commit[:7])
		}
		if date != "unknown" {
			parts = append(parts, "built "+date)
		}
		return strings.Join(parts, " ")
	}

	// Development build - try to get version from build info
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return fmt.Sprintf("%s (dev)", info.Main.Version)
	}

	return "dev"
}

// GetInfo returns detailed version information
func GetInfo() Info {
	info := Info{
		Version:   version,
		Commit:    commit,
		BuildDate: date,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	// If we're in dev mode, try to enhance with build info
	if version == "dev" {
		if buildInfo, ok := debug.ReadBuildInfo(); ok {
			// Look for version control information
			for _, setting := range buildInfo.Settings {
				switch setting.Key {
				case "vcs.revision":
					if info.Commit == "none" && setting.Value != "" {
						info.Commit = setting.Value
					}
				case "vcs.time":
					if info.BuildDate == "unknown" && setting.Value != "" {
						info.BuildDate = setting.Value
					}
				}
			}

			// Use module version if available
			if buildInfo.Main.Version != "" && buildInfo.Main.Version != "(devel)" {
				info.Version = buildInfo.Main.Version
			}
		}
	}

	return info
}

// IsDevBuild returns true if this is a development build
func IsDevBuild() bool {
	return version == "dev"
}