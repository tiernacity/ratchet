package config

import (
	"path/filepath"
	"strings"

	"github.com/tiernacity/ratchet/internal/errors"
)

// Config represents the complete application configuration
type Config struct {
	// Core settings
	BaseBranch string `mapstructure:"base-branch"`
	MetricCmd  string `mapstructure:"metric-cmd"`
	PreCmd     string `mapstructure:"pre-cmd"`
	PostCmd    string `mapstructure:"post-cmd"`

	// Comparison operator (exactly one must be set)
	Operator ComparisonOperator `mapstructure:"-"`

	// CLI flags for operators (used for parsing)
	GreaterThan        bool `mapstructure:"greater-than"`
	GreaterThanOrEqual bool `mapstructure:"greater-than-or-equal"`
	Equal              bool `mapstructure:"equal"`
	LessThanOrEqual    bool `mapstructure:"less-than-or-equal"`
	LessThan           bool `mapstructure:"less-than"`

	// Other settings
	ConfigFile string `mapstructure:"-"`
	Verbose    bool   `mapstructure:"verbose"`
}

// ComparisonOperator represents the type of metric comparison to perform
type ComparisonOperator int

const (
	OpUnknown ComparisonOperator = iota
	OpGreaterThan
	OpGreaterThanOrEqual
	OpEqual
	OpLessThanOrEqual
	OpLessThan
)

// String returns the string representation of the operator
func (op ComparisonOperator) String() string {
	switch op {
	case OpGreaterThan:
		return ">"
	case OpGreaterThanOrEqual:
		return ">="
	case OpEqual:
		return "=="
	case OpLessThanOrEqual:
		return "<="
	case OpLessThan:
		return "<"
	default:
		return "unknown"
	}
}

// HumanString returns a human-readable description of the operator
func (op ComparisonOperator) HumanString() string {
	switch op {
	case OpGreaterThan:
		return "greater than"
	case OpGreaterThanOrEqual:
		return "greater than or equal to"
	case OpEqual:
		return "equal to"
	case OpLessThanOrEqual:
		return "less than or equal to"
	case OpLessThan:
		return "less than"
	default:
		return "unknown"
	}
}

// Remove local ValidationError type - use the one from errors package

// Normalize converts boolean operator flags to the Operator enum
func (c *Config) Normalize() error {
	count := 0

	if c.GreaterThan {
		c.Operator = OpGreaterThan
		count++
	}
	if c.GreaterThanOrEqual {
		c.Operator = OpGreaterThanOrEqual
		count++
	}
	if c.Equal {
		c.Operator = OpEqual
		count++
	}
	if c.LessThanOrEqual {
		c.Operator = OpLessThanOrEqual
		count++
	}
	if c.LessThan {
		c.Operator = OpLessThan
		count++
	}

	if count == 0 {
		// No operator specified is valid for no-comparison mode
		c.Operator = OpUnknown
		return nil
	}

	if count > 1 {
		return errors.NewValidationError("operator", "multiple comparison operators specified")
	}

	return nil
}

// Validate checks that the configuration is valid
func (c *Config) Validate() error {
	// Check required fields
	if c.MetricCmd == "" {
		return errors.NewValidationError("metric-cmd", "metric command is required")
	}

	// Note: OpUnknown is valid for no-comparison mode

	// Validate base branch (only required for comparison mode)
	if c.Operator != OpUnknown && c.BaseBranch == "" {
		return errors.NewValidationError("base-branch", "base branch cannot be empty when using comparison operator")
	}

	// Check for potentially dangerous commands (warning only)
	if isDangerousCommand(c.MetricCmd) || 
	   (c.PreCmd != "" && isDangerousCommand(c.PreCmd)) ||
	   (c.PostCmd != "" && isDangerousCommand(c.PostCmd)) {
		// This is just a warning, not an error
		// The progress reporter will handle displaying warnings
	}

	return nil
}

// isDangerousCommand checks if a command contains potentially dangerous operations
// This uses a more robust approach than simple string matching
func isDangerousCommand(cmd string) bool {
	if cmd == "" {
		return false
	}
	
	// Parse the command to get the actual executable name
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return false
	}
	
	// Get the base name of the command (without path)
	executable := filepath.Base(parts[0])
	
	// List of known dangerous commands
	dangerousCommands := map[string]bool{
		"rm":     true,
		"rmdir":  true,
		"del":    true,  // Windows
		"erase":  true,  // Windows
		"format": true,  // Windows/DOS
		"fdisk":  true,
		"mkfs":   true,
		"dd":     true,
		"chmod":  true,
		"chown":  true,
		"sudo":   true,
		"su":     true,
	}
	
	// Check if the main command is dangerous
	if dangerousCommands[executable] {
		return true
	}
	
	// Check for shell operators that could be dangerous
	cmdStr := strings.Join(parts, " ")
	dangerousPatterns := []string{
		"> /",      // Redirecting to root paths
		"rm -",     // rm with flags
		"delete ",  // Generic delete
		"format ",  // Generic format
		">/dev/",   // Writing to device files
	}
	
	for _, pattern := range dangerousPatterns {
		if strings.Contains(cmdStr, pattern) {
			return true
		}
	}
	
	return false
}
