package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "ratchet",
	Short: "A CLI tool for [describe purpose]",
	Long: `Ratchet is a CLI application that helps you [describe what it does].
	
This tool can be used both as a standalone CLI and as a GitHub Action.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default action when no subcommand is provided
		_ = cmd.Help()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ratchet",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ratchet v0.1.0")
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the main ratchet process",
	Long:  `Execute the main ratchet functionality with the provided configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if verbose {
			fmt.Println("Running in verbose mode")
			fmt.Printf("Config file: %s\n", viper.ConfigFileUsed())
		}

		// TODO: Implement main functionality
		fmt.Println("Running ratchet...")

		return nil
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ratchet.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Bind flags to viper
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// Add commands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(runCmd)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".ratchet" (without extension)
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".ratchet")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
