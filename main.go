package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/user/sprout/config"
)

var (
	flagSession  string
	flagModel    string
	flagEndpoint string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "sprout",
		Short: "CLI chat for local Ollama models",
		Long:  "A minimal CLI chat application for local Ollama models with streaming, syntax highlighting, and session persistence.",
		RunE:  run,
	}

	rootCmd.Flags().StringVar(&flagSession, "session", "", "resume a specific session by ID")
	rootCmd.Flags().StringVar(&flagModel, "model", "", "override config model")
	rootCmd.Flags().StringVar(&flagEndpoint, "endpoint", "", "override config endpoint")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if flagModel != "" {
		cfg.Provider.Model = flagModel
	}
	if flagEndpoint != "" {
		cfg.Provider.BaseURL = flagEndpoint
	}

	if err := config.EnsureDataDirs(cfg); err != nil {
		return fmt.Errorf("ensuring data directories: %w", err)
	}

	fmt.Printf("Sprout started — model: %s, endpoint: %s\n", cfg.Provider.Model, cfg.Provider.BaseURL)
	if flagSession != "" {
		fmt.Printf("Resuming session: %s\n", flagSession)
	}

	return nil
}
