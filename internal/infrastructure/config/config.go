package configinfra

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Binance  BinanceConfig `yaml:"binance"`
	Symbols  []string      `yaml:"symbols"`
	Interval string        `yaml:"interval"`
	Storage  StorageConfig `yaml:"storage"`
}

type BinanceConfig struct {
	WSURL          string        `yaml:"ws_url"`
	ReconnectDelay time.Duration `yaml:"reconnect_delay"`
}

type StorageConfig struct {
	BTCUSDT string `yaml:"btcusdt"`
	ETHUSDT string `yaml:"ethusdt"`
}

func Load(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) Validate() error {
	if cfg.Binance.WSURL == "" {
		return fmt.Errorf("binance.ws_url is required")
	}
	if !strings.HasSuffix(cfg.Binance.WSURL, "/ws") {
		return fmt.Errorf("binance.ws_url must end with /ws")
	}
	if cfg.Binance.ReconnectDelay <= 0 {
		return fmt.Errorf("binance.reconnect_delay must be positive")
	}
	if len(cfg.Symbols) != 2 || cfg.Symbols[0] != "BTCUSDT" || cfg.Symbols[1] != "ETHUSDT" {
		return fmt.Errorf("symbols must be exactly [BTCUSDT, ETHUSDT]")
	}
	if cfg.Interval != "1m" {
		return fmt.Errorf("interval must be 1m")
	}
	if cfg.Storage.BTCUSDT == "" || cfg.Storage.ETHUSDT == "" {
		return fmt.Errorf("storage paths must be set")
	}
	if filepath.Clean(cfg.Storage.BTCUSDT) != filepath.Clean("data/btcusdt.json") {
		return fmt.Errorf("storage.btcusdt must be data/btcusdt.json")
	}
	if filepath.Clean(cfg.Storage.ETHUSDT) != filepath.Clean("data/ethusdt.json") {
		return fmt.Errorf("storage.ethusdt must be data/ethusdt.json")
	}
	return nil
}
