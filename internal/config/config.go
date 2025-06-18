package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration for ratchet
type Config struct {
	Metric  string `yaml:"metric" json:"metric"`
	Pre     string `yaml:"pre" json:"pre"`
	Post    string `yaml:"post" json:"post"`
	LT      string `yaml:"lt" json:"lt"`
	LE      string `yaml:"le" json:"le"`
	EQ      string `yaml:"eq" json:"eq"`
	GE      string `yaml:"ge" json:"ge"`
	GT      string `yaml:"gt" json:"gt"`
	Verbose bool   `yaml:"verbose" json:"verbose"`
}

// LoadFromFile loads configuration from a YAML or JSON file
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s", path)
	}

	// Determine format based on file extension
	ext := strings.ToLower(filepath.Ext(path))
	var cfg *Config

	switch ext {
	case ".json":
		cfg, err = LoadFromJSONString(string(data))
	case ".yaml", ".yml":
		cfg, err = LoadFromString(string(data))
	default:
		// For files without clear extension (like .ratchet), try YAML first, then JSON
		cfg, err = LoadFromString(string(data))
		if err != nil {
			// If YAML parsing fails, try JSON
			cfg, err = LoadFromJSONString(string(data))
			if err != nil {
				return nil, fmt.Errorf("failed to parse config file %s as either YAML or JSON", path)
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse config file %s", path)
	}

	return cfg, nil
}

// LoadFromString loads configuration from a YAML string
func LoadFromString(yamlStr string) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal([]byte(yamlStr), &cfg); err != nil {
		return nil, fmt.Errorf("invalid config was supplied:\n%s", yamlStr)
	}

	return &cfg, nil
}

// LoadFromJSONString loads configuration from a JSON string
func LoadFromJSONString(jsonStr string) (*Config, error) {
	var cfg Config
	if err := json.Unmarshal([]byte(jsonStr), &cfg); err != nil {
		return nil, fmt.Errorf("invalid JSON config was supplied:\n%s", jsonStr)
	}

	return &cfg, nil
}

// LoadFromConfigString loads configuration from a YAML or JSON string (auto-detects format)
func LoadFromConfigString(configStr string) (*Config, error) {
	// Try to detect JSON format (starts with { and ends with })
	trimmed := strings.TrimSpace(configStr)
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		return LoadFromJSONString(configStr)
	}

	// Try YAML first, then JSON as fallback
	cfg, err := LoadFromString(configStr)
	if err != nil {
		// If YAML parsing fails, try JSON
		cfg, err = LoadFromJSONString(configStr)
		if err != nil {
			return nil, fmt.Errorf("invalid config string (tried both YAML and JSON):\n%s", configStr)
		}
	}

	return cfg, nil
}

// Validate ensures the configuration is valid
func (c *Config) Validate() error {
	if c.Metric == "" {
		return fmt.Errorf("a metric command is required")
	}

	// Count comparison operators
	count := 0
	if c.LT != "" {
		count++
	}
	if c.LE != "" {
		count++
	}
	if c.EQ != "" {
		count++
	}
	if c.GE != "" {
		count++
	}
	if c.GT != "" {
		count++
	}

	if count > 1 {
		return fmt.Errorf("only one comparison operator can be specified")
	}

	return nil
}

// MergeWithFlags merges config with command-line flags, with flags taking precedence
func (c *Config) MergeWithFlags(metric string, pre string, post string, lt string, le string, equalTo string, ge string, gt string, verbose bool) {
	// Metric from args takes precedence
	if metric != "" {
		c.Metric = metric
	}

	// Flags take precedence over config file
	if pre != "" {
		c.Pre = pre
	}
	if post != "" {
		c.Post = post
	}

	// Check if any CLI comparison operator is provided
	cliHasComparison := lt != "" || le != "" || equalTo != "" || ge != "" || gt != ""

	// If CLI has a comparison operator, clear all config comparison operators first,
	// then set the CLI one. This allows CLI to override config even with different operators.
	if cliHasComparison {
		c.LT = ""
		c.LE = ""
		c.EQ = ""
		c.GE = ""
		c.GT = ""

		// Now set the CLI comparison operator
		if lt != "" {
			c.LT = lt
		}
		if le != "" {
			c.LE = le
		}
		if equalTo != "" {
			c.EQ = equalTo
		}
		if ge != "" {
			c.GE = ge
		}
		if gt != "" {
			c.GT = gt
		}
	}

	if verbose {
		c.Verbose = true
	}
}

// GetComparisonInfo returns the comparison type and base reference
func (c *Config) GetComparisonInfo() (compType string, baseRef string) {
	if c.LT != "" {
		return "lt", c.LT
	}
	if c.LE != "" {
		return "le", c.LE
	}
	if c.EQ != "" {
		return "eq", c.EQ
	}
	if c.GE != "" {
		return "ge", c.GE
	}
	if c.GT != "" {
		return "gt", c.GT
	}
	return "", ""
}

// LoadDefault attempts to load config from default locations
func LoadDefault() (*Config, error) {
	// Look for .ratchet in current directory
	if _, err := os.Stat(".ratchet"); err == nil {
		return LoadFromFile(".ratchet")
	}

	// No default config found
	return &Config{}, nil
}
