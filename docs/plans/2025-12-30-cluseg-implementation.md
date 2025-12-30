# cluseg Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a minimal TUI that monitors Claude Code usage from ~/.claude/stats-cache.json with configurable warning thresholds.

**Architecture:** Four packages (config, stats, limits, ui) with main.go orchestrating a 30-second refresh loop. Bubbletea handles TUI rendering, Viper handles config/CLI merging.

**Tech Stack:** Go 1.21+, bubbletea (TUI), lipgloss (styling), viper (config), cobra (CLI)

---

## Task 1: Stats Reader - Parse stats-cache.json

**Files:**
- Create: `stats/reader.go`
- Create: `stats/reader_test.go`

**Step 1: Write the failing test**

```go
// stats/reader_test.go
package stats

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadStats_ValidFile(t *testing.T) {
	// Create temp file with sample stats
	tmpDir := t.TempDir()
	statsFile := filepath.Join(tmpDir, "stats-cache.json")
	content := `{
		"version": 1,
		"lastComputedDate": "2025-12-30",
		"dailyModelTokens": [
			{"date": "2025-12-30", "tokensByModel": {"claude-sonnet-4-5": 150000, "claude-opus-4-5": 50000}}
		]
	}`
	if err := os.WriteFile(statsFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	stats, err := ReadStats(statsFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.TodayTokens != 200000 {
		t.Errorf("expected 200000 tokens, got %d", stats.TodayTokens)
	}
}

func TestReadStats_FileNotFound(t *testing.T) {
	_, err := ReadStats("/nonexistent/path")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestReadStats_StaleFile(t *testing.T) {
	tmpDir := t.TempDir()
	statsFile := filepath.Join(tmpDir, "stats-cache.json")
	content := `{"version": 1, "lastComputedDate": "2025-12-29", "dailyModelTokens": []}`
	if err := os.WriteFile(statsFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	// Set mod time to 10 minutes ago
	oldTime := time.Now().Add(-10 * time.Minute)
	os.Chtimes(statsFile, oldTime, oldTime)

	stats, err := ReadStats(statsFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !stats.IsStale {
		t.Error("expected IsStale to be true for old file")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./stats/... -v`
Expected: FAIL - package stats not found

**Step 3: Write minimal implementation**

```go
// stats/reader.go
package stats

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Stats struct {
	TodayTokens int
	LastUpdated time.Time
	IsStale     bool
	Date        string
}

type statsCache struct {
	Version          int              `json:"version"`
	LastComputedDate string           `json:"lastComputedDate"`
	DailyModelTokens []dailyTokens    `json:"dailyModelTokens"`
}

type dailyTokens struct {
	Date          string         `json:"date"`
	TokensByModel map[string]int `json:"tokensByModel"`
}

func ReadStats(path string) (*Stats, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading stats file: %w", err)
	}

	var cache statsCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("parsing stats file: %w", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat stats file: %w", err)
	}

	today := time.Now().Format("2006-01-02")
	var todayTokens int
	for _, dt := range cache.DailyModelTokens {
		if dt.Date == today {
			for _, tokens := range dt.TokensByModel {
				todayTokens += tokens
			}
			break
		}
	}

	isStale := time.Since(info.ModTime()) > 5*time.Minute

	return &Stats{
		TodayTokens: todayTokens,
		LastUpdated: info.ModTime(),
		IsStale:     isStale,
		Date:        cache.LastComputedDate,
	}, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./stats/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add stats/
git commit -m "feat(stats): add reader for stats-cache.json"
```

---

## Task 2: Config - Viper setup with defaults

**Files:**
- Create: `config/config.go`
- Create: `config/config_test.go`

**Step 1: Write the failing test**

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./config/... -v`
Expected: FAIL - package config not found

**Step 3: Write minimal implementation**

```go
// config/config.go
package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Limits     LimitsConfig     `mapstructure:"limits"`
	Thresholds ThresholdsConfig `mapstructure:"thresholds"`
	Display    DisplayConfig    `mapstructure:"display"`
	Source     SourceConfig     `mapstructure:"source"`
	Learned    LearnedConfig    `mapstructure:"learned"`
}

type LimitsConfig struct {
	DailyTokens int    `mapstructure:"daily_tokens"`
	ResetTime   string `mapstructure:"reset_time"`
}

type ThresholdsConfig struct {
	Warning  int `mapstructure:"warning"`
	Critical int `mapstructure:"critical"`
}

type DisplayConfig struct {
	RefreshSeconds int  `mapstructure:"refresh_seconds"`
	Compact        bool `mapstructure:"compact"`
}

type SourceConfig struct {
	StatsFile string `mapstructure:"stats_file"`
}

type LearnedConfig struct {
	LastObservedLimit *int `mapstructure:"last_observed_limit"`
	Confidence        int  `mapstructure:"confidence"`
}

func setDefaults() {
	viper.SetDefault("thresholds.warning", 80)
	viper.SetDefault("thresholds.critical", 95)
	viper.SetDefault("display.refresh_seconds", 30)
	viper.SetDefault("display.compact", false)
	viper.SetDefault("limits.daily_tokens", 0)
	viper.SetDefault("limits.reset_time", "00:00")
	viper.SetDefault("learned.confidence", 0)

	home, _ := os.UserHomeDir()
	viper.SetDefault("source.stats_file", filepath.Join(home, ".claude", "stats-cache.json"))
}

func Load(configPath string) (*Config, error) {
	setDefaults()

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		home, _ := os.UserHomeDir()
		viper.AddConfigPath(filepath.Join(home, ".config", "cluseg"))
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// Ignore file not found - use defaults
	_ = viper.ReadInConfig()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
```

**Step 4: Add dependency and run test**

Run: `go get github.com/spf13/viper && go test ./config/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add config/ go.mod go.sum
git commit -m "feat(config): add viper-based config loading"
```

---

## Task 3: Limits Tracker - Percentage calculation

**Files:**
- Create: `limits/tracker.go`
- Create: `limits/tracker_test.go`

**Step 1: Write the failing test**

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./limits/... -v`
Expected: FAIL - package limits not found

**Step 3: Write minimal implementation**

```go
// limits/tracker.go
package limits

import (
	"github.com/sami/cluseg/config"
)

type WarningLevel int

const (
	LevelNormal WarningLevel = iota
	LevelWarning
	LevelCritical
)

type Usage struct {
	Tokens     int
	Limit      int
	Percentage int
	HasLimit   bool
}

func CalculateUsage(cfg *config.Config, tokens int) Usage {
	limit := cfg.Limits.DailyTokens

	// Use learned limit if manual not set and confidence >= 3
	if limit == 0 && cfg.Learned.LastObservedLimit != nil && cfg.Learned.Confidence >= 3 {
		limit = *cfg.Learned.LastObservedLimit
	}

	usage := Usage{
		Tokens:   tokens,
		Limit:    limit,
		HasLimit: limit > 0,
	}

	if usage.HasLimit {
		usage.Percentage = (tokens * 100) / limit
		if usage.Percentage > 100 {
			usage.Percentage = 100
		}
	}

	return usage
}

func GetWarningLevel(cfg *config.Config, percentage int) WarningLevel {
	if percentage >= cfg.Thresholds.Critical {
		return LevelCritical
	}
	if percentage >= cfg.Thresholds.Warning {
		return LevelWarning
	}
	return LevelNormal
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./limits/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add limits/
git commit -m "feat(limits): add usage calculation and warning levels"
```

---

## Task 4: UI Styles - Lipgloss color definitions

**Files:**
- Create: `ui/styles.go`
- Create: `ui/styles_test.go`

**Step 1: Write the failing test**

```go
// ui/styles_test.go
package ui

import (
	"testing"

	"github.com/sami/cluseg/limits"
)

func TestGetBarStyle_Normal(t *testing.T) {
	style := GetBarStyle(limits.LevelNormal)
	// Just verify it doesn't panic and returns something
	rendered := style.Render("test")
	if rendered == "" {
		t.Error("expected non-empty rendered string")
	}
}

func TestGetBarStyle_Warning(t *testing.T) {
	style := GetBarStyle(limits.LevelWarning)
	rendered := style.Render("test")
	if rendered == "" {
		t.Error("expected non-empty rendered string")
	}
}

func TestGetBarStyle_Critical(t *testing.T) {
	style := GetBarStyle(limits.LevelCritical)
	rendered := style.Render("test")
	if rendered == "" {
		t.Error("expected non-empty rendered string")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./ui/... -v`
Expected: FAIL - package ui not found

**Step 3: Write minimal implementation**

```go
// ui/styles.go
package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/sami/cluseg/limits"
)

var (
	colorNormal   = lipgloss.Color("42")  // Green
	colorWarning  = lipgloss.Color("226") // Yellow
	colorCritical = lipgloss.Color("196") // Red
	colorDim      = lipgloss.Color("240") // Gray
)

func GetBarStyle(level limits.WarningLevel) lipgloss.Style {
	var color lipgloss.Color
	switch level {
	case limits.LevelWarning:
		color = colorWarning
	case limits.LevelCritical:
		color = colorCritical
	default:
		color = colorNormal
	}
	return lipgloss.NewStyle().Foreground(color)
}

func GetTextStyle(level limits.WarningLevel) lipgloss.Style {
	var color lipgloss.Color
	switch level {
	case limits.LevelWarning:
		color = colorWarning
	case limits.LevelCritical:
		color = colorCritical
	default:
		color = lipgloss.Color("255") // White
	}
	return lipgloss.NewStyle().Foreground(color)
}

func DimStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(colorDim)
}
```

**Step 4: Add dependency and run test**

Run: `go get github.com/charmbracelet/lipgloss && go test ./ui/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add ui/ go.mod go.sum
git commit -m "feat(ui): add lipgloss style definitions"
```

---

## Task 5: UI Display - Progress bar rendering

**Files:**
- Modify: `ui/styles.go`
- Create: `ui/display.go`
- Create: `ui/display_test.go`

**Step 1: Write the failing test**

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./ui/... -v`
Expected: FAIL - RenderProgressBar not defined

**Step 3: Write minimal implementation**

```go
// ui/display.go
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/sami/cluseg/limits"
)

const (
	barWidth    = 16
	blockFull   = "█"
	blockEmpty  = "░"
	warningIcon = "⚠"
	criticalIcon = "🔴"
)

func RenderProgressBar(percentage int, width int) string {
	if percentage < 0 {
		percentage = 0
	}
	if percentage > 100 {
		percentage = 100
	}

	filled := (percentage * width) / 100
	empty := width - filled

	return strings.Repeat(blockFull, filled) + strings.Repeat(blockEmpty, empty)
}

func FormatCountdown(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func RenderStatus(percentage int, countdown time.Duration, level limits.WarningLevel, compact bool, stale bool) string {
	countdownStr := FormatCountdown(countdown)

	var indicator string
	switch level {
	case limits.LevelWarning:
		indicator = " " + warningIcon
	case limits.LevelCritical:
		indicator = " " + criticalIcon
	}

	staleStr := ""
	if stale {
		staleStr = " ⏸"
	}

	if compact {
		return fmt.Sprintf("%d%% %s%s%s", percentage, countdownStr, staleStr, indicator)
	}

	bar := RenderProgressBar(percentage, barWidth)
	barStyle := GetBarStyle(level)
	textStyle := GetTextStyle(level)

	return fmt.Sprintf(" %s %s │ resets in %s%s%s",
		barStyle.Render(bar),
		textStyle.Render(fmt.Sprintf("%d%%", percentage)),
		countdownStr,
		staleStr,
		indicator,
	)
}

func RenderNoLimit(tokens int, countdown time.Duration, stale bool) string {
	staleStr := ""
	if stale {
		staleStr = " ⏸"
	}
	return fmt.Sprintf("%dk tokens today │ resets in %s%s │ no limit set",
		tokens/1000,
		FormatCountdown(countdown),
		staleStr,
	)
}

func RenderError(msg string) string {
	return DimStyle().Render(msg)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./ui/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add ui/
git commit -m "feat(ui): add progress bar and status rendering"
```

---

## Task 6: UI Model - Bubbletea integration

**Files:**
- Create: `ui/model.go`
- Create: `ui/model_test.go`

**Step 1: Write the failing test**

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./ui/... -v`
Expected: FAIL - NewModel not defined

**Step 3: Write minimal implementation**

```go
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
	resetTime, _ := time.Parse("15:04", m.cfg.Limits.ResetTime)
	reset := time.Date(now.Year(), now.Month(), now.Day(), resetTime.Hour(), resetTime.Minute(), 0, 0, time.UTC)

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

	return RenderStatus(m.usage.Percentage, m.countdown, m.level, m.cfg.Display.Compact, m.stale)
}
```

**Step 4: Add dependency and run test**

Run: `go get github.com/charmbracelet/bubbletea && go test ./ui/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add ui/ go.mod go.sum
git commit -m "feat(ui): add bubbletea model for TUI"
```

---

## Task 7: CLI - Cobra command setup

**Files:**
- Create: `main.go`

**Step 1: Write minimal main.go**

```go
// main.go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sami/cluseg/config"
	"github.com/sami/cluseg/ui"
)

var (
	cfgFile   string
	warnAt    int
	refresh   int
	compact   bool
	limit     int
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
```

**Step 2: Add dependency and build**

Run: `go get github.com/spf13/cobra && go build .`
Expected: Binary builds successfully

**Step 3: Test run**

Run: `./cluseg --help`
Expected: Shows help with available flags

**Step 4: Commit**

```bash
git add main.go go.mod go.sum
git commit -m "feat: add CLI with cobra, complete TUI monitor"
```

---

## Task 8: Integration Test - End-to-end verification

**Files:**
- Create: `integration_test.go`

**Step 1: Write integration test**

```go
// integration_test.go
package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sami/cluseg/config"
	"github.com/sami/cluseg/limits"
	"github.com/sami/cluseg/stats"
)

func TestIntegration_FullFlow(t *testing.T) {
	// Create temp stats file
	tmpDir := t.TempDir()
	statsFile := filepath.Join(tmpDir, "stats-cache.json")
	content := `{
		"version": 1,
		"lastComputedDate": "2025-12-30",
		"dailyModelTokens": [
			{"date": "2025-12-30", "tokensByModel": {"claude-sonnet-4-5": 400000}}
		]
	}`
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
```

**Step 2: Run all tests**

Run: `go test ./... -v`
Expected: All tests pass

**Step 3: Commit**

```bash
git add integration_test.go
git commit -m "test: add integration test for full flow"
```

---

## Task 9: Final Polish - go mod tidy and verify build

**Step 1: Clean up dependencies**

Run: `go mod tidy`

**Step 2: Run all tests**

Run: `go test ./... -v`
Expected: All pass

**Step 3: Build final binary**

Run: `go build -o cluseg .`

**Step 4: Manual test**

Run: `./cluseg --compact --refresh=5`
Expected: Shows usage status, updates every 5 seconds

**Step 5: Final commit**

```bash
git add go.mod go.sum
git commit -m "chore: tidy dependencies"
```

---

## Summary

| Task | Description | Est. Steps |
|------|-------------|------------|
| 1 | Stats reader | 5 |
| 2 | Config loader | 5 |
| 3 | Limits tracker | 5 |
| 4 | UI styles | 5 |
| 5 | Progress bar rendering | 5 |
| 6 | Bubbletea model | 5 |
| 7 | CLI setup | 4 |
| 8 | Integration test | 3 |
| 9 | Final polish | 5 |

**Total:** 9 tasks, ~42 steps

After completing all tasks, the worktree can be merged to main using `superpowers:finishing-a-development-branch`.
