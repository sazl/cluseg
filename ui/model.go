// ui/model.go
package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sami/cluseg/config"
	"github.com/sami/cluseg/limits"
	"github.com/sami/cluseg/stats"
)

type TickMsg time.Time

type Model struct {
	cfg       *config.Config
	statsPath string
	usage     limits.Usage
	level     limits.WarningLevel
	countdown time.Duration
	stale     bool
	err       error
	quitting  bool
}

func NewModel(cfg *config.Config, statsPath string) Model {
	return Model{
		cfg:       cfg,
		statsPath: statsPath,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.refresh(),
		m.tick(),
	)
}

func (m Model) tick() tea.Cmd {
	return tea.Tick(time.Duration(m.cfg.Display.RefreshSeconds)*time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m Model) refresh() tea.Cmd {
	return func() tea.Msg {
		return TickMsg(time.Now())
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		}
		switch msg.String() {
		case "q":
			m.quitting = true
			return m, tea.Quit
		}

	case TickMsg:
		s, err := stats.ReadStats(m.statsPath)
		if err != nil {
			m.err = err
			return m, m.tick()
		}

		m.err = nil
		m.usage = limits.CalculateUsage(m.cfg, s.TodayTokens)
		m.level = limits.GetWarningLevel(m.cfg, m.usage.Percentage)
		m.stale = s.IsStale
		m.countdown = m.calculateCountdown()

		return m, m.tick()

	case tea.WindowSizeMsg:
		// Handle resize if needed
		return m, nil
	}

	return m, nil
}

func (m Model) calculateCountdown() time.Duration {
	now := time.Now()
	resetTime, err := time.Parse("15:04", m.cfg.Limits.ResetTime)
	if err != nil {
		// Default to midnight if parse fails
		resetTime, _ = time.Parse("15:04", "00:00")
	}
	reset := time.Date(now.Year(), now.Month(), now.Day(), resetTime.Hour(), resetTime.Minute(), 0, 0, now.Location())

	if now.After(reset) {
		reset = reset.Add(24 * time.Hour)
	}

	return reset.Sub(now)
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if m.err != nil {
		return RenderError(m.err.Error())
	}

	if !m.usage.HasLimit {
		return RenderNoLimit(m.usage.Tokens, m.countdown, m.stale)
	}

	return RenderStatus(m.usage.Percentage, m.usage.Tokens, m.countdown, m.level, m.cfg.Display.Compact, m.stale)
}
