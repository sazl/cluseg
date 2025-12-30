// limits/tracker_test.go
package limits

import (
	"testing"

	"github.com/sami/cluseg/config"
)

func TestCalculateUsage_WithLimit(t *testing.T) {
	cfg := &config.Config{
		Limits: config.LimitsConfig{DailyTokens: 500000},
	}

	usage := CalculateUsage(cfg, 250000)

	if usage.Percentage != 50 {
		t.Errorf("expected 50%%, got %d%%", usage.Percentage)
	}
	if usage.HasLimit != true {
		t.Error("expected HasLimit to be true")
	}
}

func TestCalculateUsage_NoLimit(t *testing.T) {
	cfg := &config.Config{
		Limits: config.LimitsConfig{DailyTokens: 0},
	}

	usage := CalculateUsage(cfg, 250000)

	if usage.HasLimit != false {
		t.Error("expected HasLimit to be false")
	}
	if usage.Tokens != 250000 {
		t.Errorf("expected 250000 tokens, got %d", usage.Tokens)
	}
}

func TestCalculateUsage_LearnedLimit(t *testing.T) {
	learnedLimit := 400000
	cfg := &config.Config{
		Limits:  config.LimitsConfig{DailyTokens: 0},
		Learned: config.LearnedConfig{LastObservedLimit: &learnedLimit, Confidence: 3},
	}

	usage := CalculateUsage(cfg, 200000)

	if usage.Percentage != 50 {
		t.Errorf("expected 50%%, got %d%%", usage.Percentage)
	}
}

func TestGetWarningLevel(t *testing.T) {
	cfg := &config.Config{
		Thresholds: config.ThresholdsConfig{Warning: 80, Critical: 95},
	}

	tests := []struct {
		pct      int
		expected WarningLevel
	}{
		{50, LevelNormal},
		{80, LevelWarning},
		{95, LevelCritical},
		{100, LevelCritical},
	}

	for _, tt := range tests {
		level := GetWarningLevel(cfg, tt.pct)
		if level != tt.expected {
			t.Errorf("pct=%d: expected %v, got %v", tt.pct, tt.expected, level)
		}
	}
}
