package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type ServiceConfig struct {
	Name            string        `yaml:"name"`
	RestartOnFail   bool          `yaml:"restart_on_fail"`
	MaxRestarts     int           `yaml:"max_restarts"`
	RestartCooldown time.Duration `yaml:"restart_cooldown"`
}

type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
	ChatID   string `yaml:"chat_id"`
}

type NotificationConfig struct {
	Enabled  bool           `yaml:"enabled"`
	Telegram TelegramConfig `yaml:"telegram"`
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Listen  string `yaml:"listen"`
}

type Config struct {
	CheckInterval time.Duration      `yaml:"check_interval"`
	LogFile       string             `yaml:"log_file"`
	Notification  NotificationConfig `yaml:"notification"`
	Metrics       MetricsConfig      `yaml:"metrics"`
	Services      []ServiceConfig    `yaml:"services"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = 10 * time.Second
	}

	for i := range cfg.Services {
		if cfg.Services[i].MaxRestarts == 0 {
			cfg.Services[i].MaxRestarts = 3
		}
		if cfg.Services[i].RestartCooldown == 0 {
			cfg.Services[i].RestartCooldown = 30 * time.Second
		}
	}

	return &cfg, nil
}
