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
	DataDate    string // The actual date of the token data (may differ from today)
}

type statsCache struct {
	Version          int           `json:"version"`
	LastComputedDate string        `json:"lastComputedDate"`
	DailyModelTokens []dailyTokens `json:"dailyModelTokens"`
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
	var tokens int
	var dataDate string

	// First, try to find today's tokens
	for _, dt := range cache.DailyModelTokens {
		if dt.Date == today {
			for _, t := range dt.TokensByModel {
				tokens += t
			}
			dataDate = today
			break
		}
	}

	// If no data for today, use the most recent date available
	if tokens == 0 && len(cache.DailyModelTokens) > 0 {
		// Find the most recent entry (last in the array, sorted by date)
		mostRecent := cache.DailyModelTokens[len(cache.DailyModelTokens)-1]
		for _, t := range mostRecent.TokensByModel {
			tokens += t
		}
		dataDate = mostRecent.Date
	}

	isStale := time.Since(info.ModTime()) > 5*time.Minute

	return &Stats{
		TodayTokens: tokens,
		LastUpdated: info.ModTime(),
		IsStale:     isStale,
		Date:        cache.LastComputedDate,
		DataDate:    dataDate,
	}, nil
}
