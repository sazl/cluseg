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

func TestRenderStatus_Normal(t *testing.T) {
	status := RenderStatus(67, time.Hour*4+32*time.Minute, limits.LevelNormal, false, false)
	if !strings.Contains(status, "67%") {
		t.Error("expected percentage in output")
	}
	if !strings.Contains(status, "4h 32m") {
		t.Error("expected countdown in output")
	}
}

func TestRenderStatus_Warning(t *testing.T) {
	status := RenderStatus(85, time.Hour*2, limits.LevelWarning, false, false)
	if !strings.Contains(status, "⚠") {
		t.Error("expected warning indicator")
	}
}

func TestRenderStatus_Critical(t *testing.T) {
	status := RenderStatus(98, time.Hour, limits.LevelCritical, false, false)
	if !strings.Contains(status, "🔴") {
		t.Error("expected critical indicator")
	}
}

func TestRenderStatus_Compact(t *testing.T) {
	status := RenderStatus(67, time.Hour*4, limits.LevelNormal, true, false)
	// Compact mode: just "67% 4h 0m"
	if strings.Contains(status, "█") {
		t.Error("compact mode should not have progress bar")
	}
}
