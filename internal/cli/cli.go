// Package cli provides the command-line interface implementation for ratchet.
// This package is separated from main to enable testing with testscript.
package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tiernacity/ratchet/internal/config"
	"github.com/tiernacity/ratchet/internal/errors"
	"github.com/tiernacity/ratchet/internal/executor"
	"github.com/tiernacity/ratchet/internal/git"
	"github.com/tiernacity/ratchet/internal/orchestrator"
	"github.com/tiernacity/ratchet/internal/parser"
	"github.com/tiernacity/ratchet/internal/progress"
	"github.com/tiernacity/ratchet/internal/version"
)

var (
	cfgFile string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "ratchet [flags] <metric command>",
	Short: "Implement a software ratchet to enforce continuous improvement",
	Long: `Ratchet is a CLI tool to test that a metric has changed as you require, in a git repo.
Supply a command that outputs your metric, and ratchet will compare the metric against a base branch.`,

	Args: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			return nil // Skip arg validation for version
		}
		// Either positional arg or --metric flag, but not both
		if len(args) > 1 {
			return errors.NewValidationError("args", "too many arguments provided")
		}
		if len(args) == 1 && metricCmd != "" {
			return errors.NewValidationError("metric", "metric specified both as argument and flag")
		}
		if len(args) == 0 && metricCmd == "" {
			// Will be checked later in buildConfig to allow config file to provide metric
			return nil
		}
		return nil
	},
	RunE: runRatchet,
}

// Comparison operator flags
var (
	gtBranch    string
	geBranch    string
	eqBranch    string
	leBranch    string
	ltBranch    string
	preCmd      string
	postCmd     string
	configStr   string
	showVersion bool
	metricCmd   string
)

func init() {
	// Set custom help template to match README format
	rootCmd.SetHelpTemplate(`{{.Short}}

Usage:
  {{.UseLine}}

Comparison operators (choose one):
      --less-than, --lt <base>       test that HEAD metric < base branch metric
      --less-equal, --le <base>      test that HEAD metric <= base branch metric
      --equal-to, --eq <base>        test that HEAD metric == base branch metric
      --greater-equal, --ge <base>   test that HEAD metric >= base branch metric
      --greater-than, --gt <base>    test that HEAD metric > base branch metric

Other flags:
  -h, --help                   help for ratchet
      --metric <command>       Metric command (alternative to positional argument)
      --pre <command>          Command to run before metric command
      --post <command>         Command to run after metric command
      --config-file string     Path to config file (YAML or JSON)
      --config string          Config string (YAML or JSON)
  -v, --verbose                Show detailed output including both values
      --version                Show version information
`)

	// Register flags (order doesn't matter now since we have custom template)
	rootCmd.Flags().StringVar(&ltBranch, "less-than", "", "")
	rootCmd.Flags().StringVar(&ltBranch, "lt", "", "")
	rootCmd.Flags().StringVar(&leBranch, "less-equal", "", "")
	rootCmd.Flags().StringVar(&leBranch, "le", "", "")
	rootCmd.Flags().StringVar(&eqBranch, "equal-to", "", "")
	rootCmd.Flags().StringVar(&eqBranch, "eq", "", "")
	rootCmd.Flags().StringVar(&geBranch, "greater-equal", "", "")
	rootCmd.Flags().StringVar(&geBranch, "ge", "", "")
	rootCmd.Flags().StringVar(&gtBranch, "greater-than", "", "")
	rootCmd.Flags().StringVar(&gtBranch, "gt", "", "")

	// Other flags
	rootCmd.Flags().StringVar(&metricCmd, "metric", "", "")
	rootCmd.Flags().StringVar(&preCmd, "pre", "", "")
	rootCmd.Flags().StringVar(&postCmd, "post", "", "")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config-file", "", "")
	rootCmd.Flags().StringVar(&configStr, "config", "", "")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "")
	rootCmd.Flags().BoolVar(&showVersion, "version", false, "")

	// Note: viper binding handled manually in buildConfig()
}

func runRatchet(cmd *cobra.Command, args []string) error {
	// Handle version flag
	if showVersion {
		fmt.Println(version.GetVersion())
		return nil
	}

	// Create context with cancellation for interrupt handling
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Build configuration from flags and config file
	var metric string
	if len(args) > 0 {
		metric = args[0]
	}
	cfg, err := buildConfig(metric)
	if err != nil {
		if errors.ShouldSuppressHelp(err) {
			cmd.SilenceUsage = true
		}
		return err
	}

	// Create dependencies
	gitOps := git.New()
	cmdExecutor := executor.New()
	progressReporter := progress.NewConsole()
	metricParser := parser.New()

	// Create orchestrator
	tempDir := getTempDir()
	orch := orchestrator.New(gitOps, cmdExecutor, metricParser, progressReporter, tempDir)

	// Run the orchestrator
	err = orch.Run(ctx, cfg)
	if err != nil && errors.ShouldSuppressHelp(err) {
		cmd.SilenceUsage = true
	}
	return err
}

func buildConfig(metricArg string) (*config.Config, error) {
	// Initialize viper
	v := viper.New()

	// Load config file if specified
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, errors.NewValidationError("config-file", fmt.Sprintf("failed to read config file %s: %v", cfgFile, err))
		}
	}

	// Load config string if specified
	if configStr != "" {
		// Try to detect format based on content
		configStr = strings.TrimSpace(configStr)
		if strings.HasPrefix(configStr, "{") && strings.HasSuffix(configStr, "}") {
			v.SetConfigType("json")
		} else {
			v.SetConfigType("yaml")
		}
		if err := v.ReadConfig(strings.NewReader(configStr)); err != nil {
			return nil, errors.NewValidationError("config", fmt.Sprintf("failed to parse config string: %v", err))
		}
	}

	// Determine which operator was used and set corresponding flags
	var baseBranch string
	var greaterThan, greaterThanOrEqual, equal, lessThanOrEqual, lessThan bool

	if gtBranch != "" {
		greaterThan = true
		baseBranch = gtBranch
	}
	if geBranch != "" {
		if baseBranch != "" {
			return nil, errors.NewValidationError("operators", "multiple comparison operators specified")
		}
		greaterThanOrEqual = true
		baseBranch = geBranch
	}
	if eqBranch != "" {
		if baseBranch != "" {
			return nil, errors.NewValidationError("operators", "multiple comparison operators specified")
		}
		equal = true
		baseBranch = eqBranch
	}
	if leBranch != "" {
		if baseBranch != "" {
			return nil, errors.NewValidationError("operators", "multiple comparison operators specified")
		}
		lessThanOrEqual = true
		baseBranch = leBranch
	}
	if ltBranch != "" {
		if baseBranch != "" {
			return nil, errors.NewValidationError("operators", "multiple comparison operators specified")
		}
		lessThan = true
		baseBranch = ltBranch
	}

	// Determine metric command from various sources
	// Priority: positional arg > --metric flag > config
	finalMetric := ""
	if metricArg != "" {
		finalMetric = metricArg
	} else if metricCmd != "" {
		finalMetric = metricCmd
	} else if v.IsSet("metric") {
		finalMetric = v.GetString("metric")
	}

	// Set values from flags
	v.Set("metric", finalMetric)
	v.Set("verbose", verbose)
	v.Set("greater-than", greaterThan)
	v.Set("greater-than-or-equal", greaterThanOrEqual)
	v.Set("equal", equal)
	v.Set("less-than-or-equal", lessThanOrEqual)
	v.Set("less-than", lessThan)
	v.Set("base-branch", baseBranch)
	v.Set("pre", preCmd)
	v.Set("post", postCmd)

	// Parse into config struct
	var cfg config.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, errors.NewValidationError("config", fmt.Sprintf("failed to parse configuration: %v", err))
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

// getTempDir returns the appropriate temporary directory, preferring RUNNER_TEMP in GitHub Actions
func getTempDir() string {
	if runnerTemp := os.Getenv("RUNNER_TEMP"); runnerTemp != "" {
		return runnerTemp
	}
	return os.TempDir()
}

// Main runs the CLI and returns the exit code (for testing)
func Main() int {
	if err := rootCmd.Execute(); err != nil {
		return errors.GetExitCode(err)
	}
	return 0
}
