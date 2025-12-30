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
