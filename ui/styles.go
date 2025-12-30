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
