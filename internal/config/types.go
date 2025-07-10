package config

import (
	"fmt"
	"strings"
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

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s: %s", e.Field, e.Message)
}

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
		return &ValidationError{
			Field:   "operator",
			Message: "multiple comparison operators specified",
		}
	}

	return nil
}

// Validate checks that the configuration is valid
func (c *Config) Validate() error {
	// Check required fields
	if c.MetricCmd == "" {
		return &ValidationError{
			Field:   "metric-cmd",
			Message: "metric command is required",
		}
	}

	// Note: OpUnknown is valid for no-comparison mode

	// Validate base branch (only required for comparison mode)
	if c.Operator != OpUnknown && c.BaseBranch == "" {
		return &ValidationError{
			Field:   "base-branch",
			Message: "base branch cannot be empty when using comparison operator",
		}
	}

	// Check for potentially dangerous commands (warning only)
	if strings.Contains(c.MetricCmd, "rm ") ||
		strings.Contains(c.MetricCmd, "delete") ||
		strings.Contains(c.MetricCmd, "format") {
		// This is just a warning, not an error
		// The progress reporter will handle displaying warnings
	}

	return nil
}