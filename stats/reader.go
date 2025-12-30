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
