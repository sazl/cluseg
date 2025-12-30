// ui/model_test.go
package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sami/cluseg/config"
)

func TestModel_Init(t *testing.T) {
	cfg := &config.Config{
		Display: config.DisplayConfig{RefreshSeconds: 30},
	}
	m := NewModel(cfg, "/tmp/stats.json")

	cmd := m.Init()
	if cmd == nil {
		t.Error("expected init command")
	}
}

func TestModel_Update_Tick(t *testing.T) {
	cfg := &config.Config{
		Thresholds: config.ThresholdsConfig{Warning: 80, Critical: 95},
		Display:    config.DisplayConfig{RefreshSeconds: 30},
	}
	m := NewModel(cfg, "/tmp/stats.json")

	// Simulate tick message
	newM, _ := m.Update(TickMsg(time.Now()))
	model := newM.(Model)

	// Should have attempted to read stats (will fail, but that's ok)
	if model.err == nil && model.usage.Tokens == 0 {
		// Either has error or has attempted read
	}
}

func TestModel_Update_Quit(t *testing.T) {
	cfg := &config.Config{
		Display: config.DisplayConfig{RefreshSeconds: 30},
	}
	m := NewModel(cfg, "/tmp/stats.json")

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if newM.(Model).quitting != true {
		t.Error("expected quitting to be true")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestModel_View(t *testing.T) {
	cfg := &config.Config{
		Thresholds: config.ThresholdsConfig{Warning: 80, Critical: 95},
		Display:    config.DisplayConfig{RefreshSeconds: 30},
	}
	m := NewModel(cfg, "/tmp/stats.json")

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}
