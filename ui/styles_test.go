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
