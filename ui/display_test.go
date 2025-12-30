// ui/display_test.go
package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/sami/cluseg/limits"
)

func TestRenderProgressBar(t *testing.T) {
	bar := RenderProgressBar(50, 16)
	if len([]rune(bar)) != 16 {
		t.Errorf("expected 16 chars, got %d", len([]rune(bar)))
	}
	if !strings.Contains(bar, "█") {
		t.Error("expected filled blocks")
	}
	if !strings.Contains(bar, "░") {
		t.Error("expected empty blocks")
	}
}

func TestRenderProgressBar_Zero(t *testing.T) {
	bar := RenderProgressBar(0, 16)
	filled := strings.Count(bar, "█")
	if filled != 0 {
		t.Errorf("expected 0 filled, got %d", filled)
	}
}

func TestRenderProgressBar_Full(t *testing.T) {
	bar := RenderProgressBar(100, 16)
	empty := strings.Count(bar, "░")
	if empty != 0 {
		t.Errorf("expected 0 empty, got %d", empty)
	}
}

func TestFormatCountdown(t *testing.T) {
	tests := []struct {
		dur      time.Duration
		expected string
	}{
		{4*time.Hour + 32*time.Minute, "4h 32m"},
		{45 * time.Minute, "45m"},
		{1*time.Hour + 5*time.Minute, "1h 5m"},
	}

	for _, tt := range tests {
		result := FormatCountdown(tt.dur)
		if result != tt.expected {
			t.Errorf("dur=%v: expected %q, got %q", tt.dur, tt.expected, result)
		}
	}
}

func TestFormatTokens(t *testing.T) {
	tests := []struct {
		tokens   int
		expected string
	}{
		{500, "500"},
		{1500, "1k"},
		{250000, "250k"},
		{1500000, "1.5M"},
	}
	for _, tt := range tests {
		result := FormatTokens(tt.tokens)
		if result != tt.expected {
			t.Errorf("tokens=%d: expected %q, got %q", tt.tokens, tt.expected, result)
		}
	}
}

func TestRenderStatus_Normal(t *testing.T) {
	status := RenderStatus(67, 335000, time.Hour*4+32*time.Minute, limits.LevelNormal, false, false)
	if !strings.Contains(status, "67%") {
		t.Error("expected percentage in output")
	}
	if !strings.Contains(status, "335k") {
		t.Error("expected token count in output")
	}
	if !strings.Contains(status, "4h 32m") {
		t.Error("expected countdown in output")
	}
}

func TestRenderStatus_Warning(t *testing.T) {
	status := RenderStatus(85, 425000, time.Hour*2, limits.LevelWarning, false, false)
	if !strings.Contains(status, "⚠") {
		t.Error("expected warning indicator")
	}
}

func TestRenderStatus_Critical(t *testing.T) {
	status := RenderStatus(98, 490000, time.Hour, limits.LevelCritical, false, false)
	if !strings.Contains(status, "🔴") {
		t.Error("expected critical indicator")
	}
}

func TestRenderStatus_Compact(t *testing.T) {
	status := RenderStatus(67, 335000, time.Hour*4, limits.LevelNormal, true, false)
	// Compact mode shows tokens and percentage
	if strings.Contains(status, "█") {
		t.Error("compact mode should not have progress bar")
	}
	if !strings.Contains(status, "335k") {
		t.Error("expected token count in compact mode")
	}
	if !strings.Contains(status, "67%") {
		t.Error("expected percentage in compact mode")
	}
}
