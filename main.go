// Sprout is a terminal chat client for local Ollama models.
package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/user/sprout/agent"
	"github.com/user/sprout/config"
	"github.com/user/sprout/session"
	"github.com/user/sprout/tui"
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

	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List saved sessions",
		RunE:  runLs,
	}
	rootCmd.AddCommand(lsCmd)

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

	sess, err := session.LoadOrCreate(cfg.DataDir, flagSession, cfg.Provider.Model)
	if err != nil {
		return fmt.Errorf("loading session: %w", err)
	}

	loop := agent.NewLoop(cfg)
	session.RestoreMessages(sess, loop.Store())

	m := tui.NewModel(cfg, sess, loop)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}

	session.SyncFromStore(sess, loop.Store())
	if err := session.Save(cfg.DataDir, sess); err != nil {
		return fmt.Errorf("saving session: %w", err)
	}

	fmt.Fprintf(os.Stdout, "\nSession saved: %s\nTo resume: sprout --session %s\n", sess.ID, sess.ID)

	return nil
}

func runLs(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	sessions, err := session.List(cfg.DataDir)
	if err != nil {
		return fmt.Errorf("listing sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No saved sessions.")
		return nil
	}

	for _, s := range sessions {
		msgCount := len(s.Messages)
		age := time.Since(s.UpdatedAt).Truncate(time.Second)
		fmt.Printf("%s  %s ago  %d messages  model: %s\n", s.ID, age, msgCount, s.Model)
	}

	return nil
}
