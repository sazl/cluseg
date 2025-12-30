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
