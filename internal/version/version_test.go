package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name            string
		version         string
		commit          string
		date            string
		expectedPattern string
	}{
		{
			name:            "dev build",
			version:         "dev",
			commit:          "none",
			date:            "unknown",
			expectedPattern: "dev",
		},
		{
			name:            "production build full info",
			version:         "v1.0.0",
			commit:          "1234567890abcdef",
			date:            "2024-01-01T00:00:00Z",
			expectedPattern: "v1.0.0 1234567 built 2024-01-01T00:00:00Z",
		},
		{
			name:            "production build version only",
			version:         "v2.0.0",
			commit:          "none",
			date:            "unknown",
			expectedPattern: "v2.0.0",
		},
		{
			name:            "production build with short commit",
			version:         "v1.2.3",
			commit:          "abc",
			date:            "unknown",
			expectedPattern: "v1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origVersion := version
			origCommit := commit
			origDate := date

			// Set test values
			version = tt.version
			commit = tt.commit
			date = tt.date

			// Test
			result := GetVersion()
			assert.Equal(t, tt.expectedPattern, result)

			// Restore original values
			version = origVersion
			commit = origCommit
			date = origDate
		})
	}
}

func TestGetInfo(t *testing.T) {
	// Save original values
	origVersion := version
	origCommit := commit
	origDate := date

	// Test production build
	version = "v1.0.0"
	commit = "abcdef1234567890"
	date = "2024-01-01T00:00:00Z"

	info := GetInfo()
	assert.Equal(t, "v1.0.0", info.Version)
	assert.Equal(t, "abcdef1234567890", info.Commit)
	assert.Equal(t, "2024-01-01T00:00:00Z", info.BuildDate)
	assert.Contains(t, info.GoVersion, "go")
	assert.NotEmpty(t, info.Compiler)
	assert.Contains(t, info.Platform, "/")

	// Restore original values
	version = origVersion
	commit = origCommit
	date = origDate
}

func TestIsDevBuild(t *testing.T) {
	// Save original value
	origVersion := version

	// Test dev build
	version = "dev"
	assert.True(t, IsDevBuild())

	// Test production build
	version = "v1.0.0"
	assert.False(t, IsDevBuild())

	// Restore original value
	version = origVersion
}
