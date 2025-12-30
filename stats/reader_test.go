// stats/reader_test.go
package stats

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadStats_ValidFile(t *testing.T) {
	// Create temp file with sample stats using today's date
	tmpDir := t.TempDir()
	statsFile := filepath.Join(tmpDir, "stats-cache.json")
	today := time.Now().Format("2006-01-02")
	content := fmt.Sprintf(`{
		"version": 1,
		"lastComputedDate": "%s",
		"dailyModelTokens": [
			{"date": "%s", "tokensByModel": {"claude-sonnet-4-5": 150000, "claude-opus-4-5": 50000}}
		]
	}`, today, today)
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

func TestReadStats_FallbackToMostRecentDate(t *testing.T) {
	// When today's date isn't in the file, should use most recent available date
	tmpDir := t.TempDir()
	statsFile := filepath.Join(tmpDir, "stats-cache.json")
	content := `{
		"version": 1,
		"lastComputedDate": "2025-12-29",
		"dailyModelTokens": [
			{"date": "2025-12-28", "tokensByModel": {"claude-sonnet": 100000}},
			{"date": "2025-12-29", "tokensByModel": {"claude-sonnet": 250284, "claude-opus": 239830}}
		]
	}`
	if err := os.WriteFile(statsFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	stats, err := ReadStats(statsFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should get tokens from most recent date (2025-12-29): 250284 + 239830 = 490114
	expectedTokens := 490114
	if stats.TodayTokens != expectedTokens {
		t.Errorf("expected %d tokens from fallback, got %d", expectedTokens, stats.TodayTokens)
	}
	if stats.DataDate != "2025-12-29" {
		t.Errorf("expected DataDate to be 2025-12-29, got %s", stats.DataDate)
	}
}

func TestReadStats_RealFixture(t *testing.T) {
	// Test using the real stats-cache.json fixture
	stats, err := ReadStats("../testdata/stats-cache.json")
	if err != nil {
		t.Fatalf("unexpected error reading fixture: %v", err)
	}

	// The fixture has data for 2025-12-29 with 250284 + 239830 = 490114 tokens
	expectedTokens := 490114
	if stats.TodayTokens != expectedTokens {
		t.Errorf("expected %d tokens from fixture, got %d", expectedTokens, stats.TodayTokens)
	}
	if stats.DataDate != "2025-12-29" {
		t.Errorf("expected DataDate to be 2025-12-29, got %s", stats.DataDate)
	}
	if stats.Date != "2025-12-29" {
		t.Errorf("expected Date (lastComputedDate) to be 2025-12-29, got %s", stats.Date)
	}
}
