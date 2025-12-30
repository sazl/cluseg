// ui/display.go
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/sami/cluseg/limits"
)

const (
	barWidth     = 16
	blockFull    = "█"
	blockEmpty   = "░"
	warningIcon  = "⚠"
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

func FormatTokens(tokens int) string {
	if tokens >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(tokens)/1000000)
	}
	if tokens >= 1000 {
		return fmt.Sprintf("%dk", tokens/1000)
	}
	return fmt.Sprintf("%d", tokens)
}

func RenderStatus(percentage int, tokens int, countdown time.Duration, level limits.WarningLevel, compact bool, stale bool) string {
	countdownStr := FormatCountdown(countdown)
	tokensStr := FormatTokens(tokens)

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
		return fmt.Sprintf("%s (%d%%) %s%s%s", tokensStr, percentage, countdownStr, staleStr, indicator)
	}

	bar := RenderProgressBar(percentage, barWidth)
	barStyle := GetBarStyle(level)
	textStyle := GetTextStyle(level)

	return fmt.Sprintf(" %s %s │ %s │ resets in %s%s%s",
		barStyle.Render(bar),
		textStyle.Render(fmt.Sprintf("%d%%", percentage)),
		tokensStr,
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
