// config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Thresholds.Warning != 80 {
		t.Errorf("expected warning threshold 80, got %d", cfg.Thresholds.Warning)
	}
	if cfg.Thresholds.Critical != 95 {
		t.Errorf("expected critical threshold 95, got %d", cfg.Thresholds.Critical)
	}
	if cfg.Display.RefreshSeconds != 30 {
		t.Errorf("expected refresh 30s, got %d", cfg.Display.RefreshSeconds)
	}
}

func TestLoad_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	content := `
thresholds:
  warning: 70
  critical: 90
display:
  refresh_seconds: 10
  compact: true
limits:
  daily_tokens: 300000
`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Thresholds.Warning != 70 {
		t.Errorf("expected warning 70, got %d", cfg.Thresholds.Warning)
	}
	if cfg.Display.Compact != true {
		t.Error("expected compact mode enabled")
	}
	if cfg.Limits.DailyTokens != 300000 {
		t.Errorf("expected daily tokens 300000, got %d", cfg.Limits.DailyTokens)
	}
}
