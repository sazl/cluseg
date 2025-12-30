// integration_test.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sami/cluseg/config"
	"github.com/sami/cluseg/limits"
	"github.com/sami/cluseg/stats"
)

func TestIntegration_FullFlow(t *testing.T) {
	// Create temp stats file
	tmpDir := t.TempDir()
	statsFile := filepath.Join(tmpDir, "stats-cache.json")

	// Use today's date so ReadStats can find today's tokens
	today := time.Now().Format("2006-01-02")
	content := fmt.Sprintf(`{
		"version": 1,
		"lastComputedDate": "%s",
		"dailyModelTokens": [
			{"date": "%s", "tokensByModel": {"claude-sonnet-4-5": 400000}}
		]
	}`, today, today)
	if err := os.WriteFile(statsFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Create temp config
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `
limits:
  daily_tokens: 500000
thresholds:
  warning: 80
  critical: 95
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Load config
	cfg, err := config.Load(configFile)
	if err != nil {
		t.Fatalf("config load: %v", err)
	}

	// Read stats
	s, err := stats.ReadStats(statsFile)
	if err != nil {
		t.Fatalf("stats read: %v", err)
	}

	// Calculate usage
	usage := limits.CalculateUsage(cfg, s.TodayTokens)
	level := limits.GetWarningLevel(cfg, usage.Percentage)

	// Verify: 400k/500k = 80% = warning level
	if usage.Percentage != 80 {
		t.Errorf("expected 80%%, got %d%%", usage.Percentage)
	}
	if level != limits.LevelWarning {
		t.Errorf("expected warning level, got %v", level)
	}
}
