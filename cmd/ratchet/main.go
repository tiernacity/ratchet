package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tiernacity/ratchet/internal/config"
	"github.com/tiernacity/ratchet/internal/ratchet"
)

var (
	// Comparison flags
	lessThan     string
	lessEqual    string
	equalTo      string
	greaterEqual string
	greaterThan  string

	// Setup/teardown
	pre  string
	post string

	// Config
	configFile string
	configStr  string

	// Other flags
	verbose bool
	version = "0.1.0"
)

var rootCmd = &cobra.Command{
	Use:           "ratchet [flags] <metric command>",
	Short:         "A software ratchet tool that ensures metrics only improve",
	Long:          `Ratchet compares a metric output between your current branch/HEAD and a base branch and applies the test that you specify`,
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runRatchet,
}

func runRatchet(cmd *cobra.Command, args []string) error {
	if cmd.Flags().Changed("version") {
		fmt.Printf("ratchet v%s\n", version)
		return nil
	}

	// Validate that only one comparison operator is provided via CLI
	cliComparisons := 0
	if lessThan != "" {
		cliComparisons++
	}
	if lessEqual != "" {
		cliComparisons++
	}
	if equalTo != "" {
		cliComparisons++
	}
	if greaterEqual != "" {
		cliComparisons++
	}
	if greaterThan != "" {
		cliComparisons++
	}

	if cliComparisons > 1 {
		return fmt.Errorf("only one comparison operator can be specified")
	}

	// Determine metric source
	var metric string
	if len(args) > 0 {
		metric = args[0]
	}

	// Load configuration
	var cfg *config.Config
	var err error

	if configStr != "" {
		// Load from config string (YAML or JSON)
		cfg, err = config.LoadFromConfigString(configStr)
		if err != nil {
			return err
		}
	} else if configFile != "" {
		// Load from specified file
		cfg, err = config.LoadFromFile(configFile)
		if err != nil {
			return err
		}
	} else {
		// Try to load default config
		cfg, err = config.LoadDefault()
		if err != nil {
			return err
		}
	}

	// Merge with command-line flags (flags take precedence)
	cfg.MergeWithFlags(metric, pre, post, lessThan, lessEqual, equalTo, greaterEqual, greaterThan, verbose)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return err
	}

	// Get comparison info from config
	compType, baseRef := cfg.GetComparisonInfo()

	// Map comparison type string to enum
	var comparisonType ratchet.ComparisonType
	switch compType {
	case "lt":
		comparisonType = ratchet.LessThan
	case "le":
		comparisonType = ratchet.LessEqual
	case "eq":
		comparisonType = ratchet.Equal
	case "ge":
		comparisonType = ratchet.GreaterEqual
	case "gt":
		comparisonType = ratchet.GreaterThan
	default:
		comparisonType = ratchet.NoComparison
	}

	opts := ratchet.Options{
		Metric:         cfg.Metric,
		BaseRef:        baseRef,
		ComparisonType: comparisonType,
		Pre:            cfg.Pre,
		Post:           cfg.Post,
		Verbose:        cfg.Verbose,
	}

	return ratchet.Run(opts)
}

func init() {
	// Comparison flags
	rootCmd.Flags().StringVar(&lessThan, "less-than", "", "test that HEAD metric < base branch metric")
	rootCmd.Flags().StringVar(&lessThan, "lt", "", "test that HEAD metric < base branch metric")
	rootCmd.Flags().StringVar(&lessEqual, "less-equal", "", "test that HEAD metric <= base branch metric")
	rootCmd.Flags().StringVar(&lessEqual, "le", "", "test that HEAD metric <= base branch metric")
	rootCmd.Flags().StringVar(&equalTo, "equal-to", "", "test that HEAD metric == base branch metric")
	rootCmd.Flags().StringVar(&equalTo, "eq", "", "test that HEAD metric == base branch metric")
	if err := rootCmd.Flags().MarkHidden("eq"); err != nil {
		// Log the error but don't fail - this is not critical
		fmt.Fprintf(os.Stderr, "Warning: failed to mark 'eq' flag as hidden: %v\n", err)
	}
	rootCmd.Flags().StringVar(&greaterEqual, "greater-equal", "", "test that HEAD metric >= base branch metric")
	rootCmd.Flags().StringVar(&greaterEqual, "ge", "", "test that HEAD metric >= base branch metric")
	rootCmd.Flags().StringVar(&greaterThan, "greater-than", "", "test that HEAD metric > base branch metric")
	rootCmd.Flags().StringVar(&greaterThan, "gt", "", "test that HEAD metric > base branch metric")

	// Setup/teardown flags
	rootCmd.Flags().StringVar(&pre, "pre", "", "command to run before metric command")
	rootCmd.Flags().StringVar(&post, "post", "", "command to run after metric command")

	// Config flags
	rootCmd.Flags().StringVar(&configFile, "config-file", "", "path to config file (YAML or JSON)")
	rootCmd.Flags().StringVar(&configStr, "config", "", "config string (YAML or JSON)")

	// Other flags
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "show detailed output including both values")
	rootCmd.Flags().Bool("version", false, "show version information")

	// Custom usage template to group comparison operators
	rootCmd.SetUsageTemplate(`Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Comparison operators (choose one):
      --less-than, --lt <base>       test that HEAD metric < base branch metric
      --less-equal, --le <base>      test that HEAD metric <= base branch metric
      --equal-to, --eq <base>        test that HEAD metric == base branch metric
      --greater-equal, --ge <base>   test that HEAD metric >= base branch metric
      --greater-than, --gt <base>    test that HEAD metric > base branch metric

Other flags:
  -h, --help                   help for ratchet
      --pre <command>          Command to run before metric command
      --post <command>         Command to run after metric command
      --config-file string     Path to config file (YAML or JSON)
      --config string          Config string (YAML or JSON)
  -v, --verbose                Show detailed output including both values
      --version                Show version information{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		// For config validation errors, show usage
		if err.Error() == "a metric command is required" || err.Error() == "only one comparison operator can be specified" {
			fmt.Fprintln(os.Stderr, "Error:", err)
			if err := rootCmd.Usage(); err != nil {
				// If we can't print usage, just continue with exit
				fmt.Fprintf(os.Stderr, "Warning: failed to print usage: %v\n", err)
			}
			os.Exit(2)
		}

		// Exit code 1 for metric test failures
		if err.Error() == "metric test failed" {
			os.Exit(1)
		}

		// Exit code 2 for other errors (no usage)
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(2)
	}
}
