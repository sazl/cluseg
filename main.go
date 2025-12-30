// main.go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/sami/cluseg/config"
	"github.com/sami/cluseg/ui"
)

var (
	cfgFile string
	warnAt  int
	refresh int
	compact bool
	limit   int
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "cluseg",
		Short: "Claude Usage Monitor - track your Claude Code usage",
		RunE:  runMonitor,
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default ~/.config/cluseg/config.yaml)")
	rootCmd.Flags().IntVar(&warnAt, "warn-at", 0, "warning threshold percentage (overrides config)")
	rootCmd.Flags().IntVar(&refresh, "refresh", 0, "refresh interval in seconds (overrides config)")
	rootCmd.Flags().BoolVar(&compact, "compact", false, "ultra-minimal display mode")
	rootCmd.Flags().IntVar(&limit, "limit", 0, "daily token limit (overrides config)")

	hitLimitCmd := &cobra.Command{
		Use:   "hit-limit",
		Short: "Record that you hit a rate limit (for learning)",
		RunE:  runHitLimit,
	}
	rootCmd.AddCommand(hitLimitCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runMonitor(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Apply CLI overrides
	if warnAt > 0 {
		cfg.Thresholds.Warning = warnAt
	}
	if refresh > 0 {
		cfg.Display.RefreshSeconds = refresh
	}
	if compact {
		cfg.Display.Compact = true
	}
	if limit > 0 {
		cfg.Limits.DailyTokens = limit
	}

	statsPath := cfg.Source.StatsFile
	if statsPath == "" {
		home, _ := os.UserHomeDir()
		statsPath = home + "/.claude/stats-cache.json"
	}
	// Expand ~ in path
	if len(statsPath) > 0 && statsPath[0] == '~' {
		home, _ := os.UserHomeDir()
		statsPath = home + statsPath[1:]
	}

	model := ui.NewModel(cfg, statsPath)
	p := tea.NewProgram(model, tea.WithAltScreen())

	_, err = p.Run()
	return err
}

func runHitLimit(cmd *cobra.Command, args []string) error {
	// TODO: Implement limit learning
	fmt.Println("Rate limit recorded. This will improve future predictions.")
	return nil
}
