# Configuration Feature

## Overview
The Configuration module handles all aspects of application configuration, including command-line flags, configuration files, validation, and merging of different configuration sources.

## Configuration Structure

```go
package config

type Config struct {
    // Core settings
    BaseBranch string             `mapstructure:"base-branch"`
    MetricCmd  string             `mapstructure:"metric-cmd"`
    PreCmd     string             `mapstructure:"pre-cmd"`
    PostCmd    string             `mapstructure:"post-cmd"`
    
    // Comparison operator (exactly one must be set)
    Operator   ComparisonOperator `mapstructure:"-"`
    
    // CLI flags for operators (used for parsing)
    GreaterThan         bool `mapstructure:"greater-than"`
    GreaterThanOrEqual  bool `mapstructure:"greater-than-or-equal"`
    Equal               bool `mapstructure:"equal"`
    LessThanOrEqual     bool `mapstructure:"less-than-or-equal"`
    LessThan            bool `mapstructure:"less-than"`
    
    // Other settings
    ConfigFile string `mapstructure:"-"`
    Verbose    bool   `mapstructure:"verbose"`
}

type ComparisonOperator int

const (
    OpUnknown ComparisonOperator = iota
    OpGreaterThan
    OpGreaterThanOrEqual
    OpEqual
    OpLessThanOrEqual
    OpLessThan
)
```

## Operator String Representation

```go
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
```

## Configuration Loading

### Load Order (Priority)
1. Default values
2. Configuration file (if specified)
3. Environment variables
4. Command-line flags (highest priority)

### Implementation
```go
func Load(v *viper.Viper, configFile string) (*Config, error) {
    // Set defaults
    v.SetDefault("base-branch", "main")
    v.SetDefault("verbose", false)
    
    // Load config file if specified
    if configFile != "" {
        v.SetConfigFile(configFile)
        if err := v.ReadInConfig(); err != nil {
            return nil, fmt.Errorf("failed to read config file: %w", err)
        }
    }
    
    // Bind environment variables
    v.SetEnvPrefix("RATCHET")
    v.AutomaticEnv()
    v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
    
    // Unmarshal to struct
    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("failed to parse configuration: %w", err)
    }
    
    // Normalize and validate
    if err := cfg.Normalize(); err != nil {
        return nil, err
    }
    
    if err := cfg.Validate(); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}
```

## Normalization

Convert boolean operator flags to the Operator enum:

```go
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
        return &ValidationError{
            Field:   "operator",
            Message: "no comparison operator specified",
        }
    }
    
    if count > 1 {
        return &ValidationError{
            Field:   "operator",
            Message: "multiple comparison operators specified",
        }
    }
    
    return nil
}
```

## Validation

```go
func (c *Config) Validate() error {
    // Check required fields
    if c.MetricCmd == "" {
        return &ValidationError{
            Field:   "metric-cmd",
            Message: "metric command is required",
        }
    }
    
    // Validate operator
    if c.Operator == OpUnknown {
        return &ValidationError{
            Field:   "operator",
            Message: "comparison operator is required",
        }
    }
    
    // Validate base branch
    if c.BaseBranch == "" {
        return &ValidationError{
            Field:   "base-branch",
            Message: "base branch cannot be empty",
        }
    }
    
    // Warn about potentially dangerous commands
    if strings.Contains(c.MetricCmd, "rm ") || 
       strings.Contains(c.MetricCmd, "delete") {
        // This is just a warning, not an error
        fmt.Fprintf(os.Stderr, "Warning: metric command contains potentially destructive operations\n")
    }
    
    return nil
}
```

## Configuration File Formats

### YAML Example
```yaml
# .ratchet.yml
base-branch: develop
metric-cmd: |
  go test -cover ./... | 
  grep "coverage:" | 
  awk '{print $2}' | 
  sed 's/%//'
pre-cmd: go mod download
post-cmd: go clean -testcache
greater-than: true
verbose: true
```

### JSON Example
```json
{
  "base-branch": "main",
  "metric-cmd": "npm test -- --coverage --json | jq .coveragePercentage",
  "pre-cmd": "npm install",
  "less-than-or-equal": true,
  "verbose": false
}
```

### TOML Example
```toml
base-branch = "main"
metric-cmd = "grep -c TODO src/**/*.js"
less-than = true
verbose = true

[scripts]
pre-cmd = "npm run lint"
post-cmd = "npm run cleanup"
```

## Environment Variables

All configuration options can be set via environment variables with the `RATCHET_` prefix:

```bash
export RATCHET_BASE_BRANCH=develop
export RATCHET_METRIC_CMD="make test-coverage"
export RATCHET_GREATER_THAN=true
export RATCHET_VERBOSE=true
```

## CLI Integration

### Cobra Command Setup
```go
func init() {
    // Bind flags to viper
    rootCmd.PersistentFlags().String("base-branch", "main", "Base branch to compare against")
    viper.BindPFlag("base-branch", rootCmd.PersistentFlags().Lookup("base-branch"))
    
    rootCmd.PersistentFlags().String("metric-cmd", "", "Command that outputs a metric")
    viper.BindPFlag("metric-cmd", rootCmd.PersistentFlags().Lookup("metric-cmd"))
    
    // Operator flags
    rootCmd.PersistentFlags().Bool("gt", false, "Greater than")
    rootCmd.PersistentFlags().Bool("greater-than", false, "Greater than")
    viper.BindPFlag("greater-than", rootCmd.PersistentFlags().Lookup("greater-than"))
    viper.BindPFlag("greater-than", rootCmd.PersistentFlags().Lookup("gt"))
    
    // ... similar for other operators
}
```

## Testing

### Unit Tests
```go
func TestConfig_Normalize(t *testing.T) {
    tests := []struct {
        name    string
        config  Config
        want    ComparisonOperator
        wantErr bool
    }{
        {
            name: "single operator",
            config: Config{
                GreaterThan: true,
            },
            want: OpGreaterThan,
        },
        {
            name: "multiple operators",
            config: Config{
                GreaterThan: true,
                LessThan:    true,
            },
            wantErr: true,
        },
        {
            name:    "no operator",
            config:  Config{},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Normalize()
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.want, tt.config.Operator)
        })
    }
}
```

### Integration Tests
```go
func TestConfig_LoadFromFile(t *testing.T) {
    // Create temporary config file
    configFile := filepath.Join(t.TempDir(), "config.yml")
    content := `
base-branch: feature
metric-cmd: echo 42
greater-than: true
verbose: true
`
    os.WriteFile(configFile, []byte(content), 0644)
    
    // Load config
    v := viper.New()
    cfg, err := Load(v, configFile)
    
    assert.NoError(t, err)
    assert.Equal(t, "feature", cfg.BaseBranch)
    assert.Equal(t, "echo 42", cfg.MetricCmd)
    assert.Equal(t, OpGreaterThan, cfg.Operator)
    assert.True(t, cfg.Verbose)
}
```

## Common Patterns

### Config File Discovery
```go
func findConfigFile() string {
    // Look for config files in order of preference
    locations := []string{
        ".ratchet.yml",
        ".ratchet.yaml",
        ".ratchet.json",
        ".github/ratchet.yml",
        "ratchet.config.yml",
    }
    
    for _, loc := range locations {
        if _, err := os.Stat(loc); err == nil {
            return loc
        }
    }
    
    return ""
}
```

### Preset Configurations
```go
// Support for preset configurations
type Preset struct {
    Name        string
    Description string
    Config      Config
}

var presets = []Preset{
    {
        Name:        "coverage-increase",
        Description: "Ensure test coverage increases",
        Config: Config{
            MetricCmd:   `go test -cover ./... | grep "coverage:" | awk '{print $2}' | sed 's/%//'`,
            GreaterThan: true,
        },
    },
    {
        Name:        "reduce-todos",
        Description: "Ensure TODO comments decrease",
        Config: Config{
            MetricCmd: "grep -c TODO *.go || echo 0",
            LessThan:  true,
        },
    },
}
```

## Error Messages

### Validation Errors
```
validation error: metric-cmd: cannot be empty
validation error: operator: exactly one comparison operator must be specified
validation error: base-branch: branch 'feat/new' does not exist
```

### Config File Errors
```
failed to read config file: open .ratchet.yml: no such file or directory
failed to parse configuration: yaml: line 5: found character that cannot start any token
```

## Future Enhancements

1. **Schema Validation**: JSON Schema for config files
2. **Config Generation**: `ratchet init` to create config file
3. **Multiple Metrics**: Support arrays of metric commands
4. **Conditional Config**: Different settings per branch/environment
5. **Remote Config**: Load configuration from URL
6. **Config Inheritance**: Extend base configurations