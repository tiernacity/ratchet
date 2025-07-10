package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
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
)

var (
	cfgFile string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "ratchet [flags] <metric command>",
	Short: "A software ratchet that enforces continuous improvement",
	Long: `Ratchet is a CLI tool that ensures metrics move in the right direction.
It compares command output metrics between Git branches to enforce improvements.

Examples:
  ratchet --gt main "grep -c TODO *.go"          # TODO count must decrease
  ratchet --le origin/main "npm test | grep failed"  # Failures must not increase
  ratchet "echo 42"                              # Just output the metric`,
	Args: cobra.ExactArgs(1),
	RunE: runRatchet,
}

// Comparison operator flags
var (
	gtBranch string
	geBranch string
	eqBranch string
	leBranch string
	ltBranch string
	preCmd   string
	postCmd  string
)

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config-file", "", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Comparison operators (mutually exclusive)
	rootCmd.Flags().StringVar(&gtBranch, "gt", "", "test that HEAD metric > base branch metric")
	rootCmd.Flags().StringVar(&gtBranch, "greater-than", "", "test that HEAD metric > base branch metric")
	rootCmd.Flags().StringVar(&geBranch, "ge", "", "test that HEAD metric >= base branch metric")
	rootCmd.Flags().StringVar(&geBranch, "greater-equal", "", "test that HEAD metric >= base branch metric")
	rootCmd.Flags().StringVar(&eqBranch, "eq", "", "test that HEAD metric == base branch metric")
	rootCmd.Flags().StringVar(&eqBranch, "equal-to", "", "test that HEAD metric == base branch metric")
	rootCmd.Flags().StringVar(&leBranch, "le", "", "test that HEAD metric <= base branch metric")
	rootCmd.Flags().StringVar(&leBranch, "less-equal", "", "test that HEAD metric <= base branch metric")
	rootCmd.Flags().StringVar(&ltBranch, "lt", "", "test that HEAD metric < base branch metric")
	rootCmd.Flags().StringVar(&ltBranch, "less-than", "", "test that HEAD metric < base branch metric")

	// Other flags
	rootCmd.Flags().StringVar(&preCmd, "pre", "", "command to run before metric command")
	rootCmd.Flags().StringVar(&postCmd, "post", "", "command to run after metric command")

	// Note: viper binding handled manually in buildConfig()
}

func runRatchet(cmd *cobra.Command, args []string) error {
	// Create context with cancellation for interrupt handling
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Build configuration from flags and config file
	cfg, err := buildConfig(args[0])
	if err != nil {
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
	return orch.Run(ctx, cfg)
}

func buildConfig(metricCmd string) (*config.Config, error) {
	// Initialize viper
	v := viper.New()

	// Load config file if specified
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", cfgFile, err)
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
			return nil, fmt.Errorf("multiple comparison operators specified")
		}
		greaterThanOrEqual = true
		baseBranch = geBranch
	}
	if eqBranch != "" {
		if baseBranch != "" {
			return nil, fmt.Errorf("multiple comparison operators specified")
		}
		equal = true
		baseBranch = eqBranch
	}
	if leBranch != "" {
		if baseBranch != "" {
			return nil, fmt.Errorf("multiple comparison operators specified")
		}
		lessThanOrEqual = true
		baseBranch = leBranch
	}
	if ltBranch != "" {
		if baseBranch != "" {
			return nil, fmt.Errorf("multiple comparison operators specified")
		}
		lessThan = true
		baseBranch = ltBranch
	}

	// Set values from flags
	v.Set("metric-cmd", metricCmd)
	v.Set("verbose", verbose)
	v.Set("greater-than", greaterThan)
	v.Set("greater-than-or-equal", greaterThanOrEqual)
	v.Set("equal", equal)
	v.Set("less-than-or-equal", lessThanOrEqual)
	v.Set("less-than", lessThan)
	v.Set("base-branch", baseBranch)
	v.Set("pre-cmd", preCmd)
	v.Set("post-cmd", postCmd)

	// Parse into config struct
	var cfg config.Config
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

// getTempDir returns the appropriate temporary directory, preferring RUNNER_TEMP in GitHub Actions
func getTempDir() string {
	if runnerTemp := os.Getenv("RUNNER_TEMP"); runnerTemp != "" {
		return runnerTemp
	}
	return os.TempDir()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		// Print error to stderr
		fmt.Fprintln(os.Stderr, err)

		// Exit with appropriate code
		os.Exit(errors.GetExitCode(err))
	}
}
