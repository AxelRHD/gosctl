package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Hosts map[string]Host `toml:"hosts"`
	Tasks map[string]Task `toml:"tasks"`
}

type Host struct {
	Address  string `toml:"address"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	KeyFile  string `toml:"key_file"`
	Password string `toml:"password"`
}

type Task struct {
	Host    string   `toml:"host"`
	Workdir string   `toml:"workdir"`
	Steps   []string `toml:"steps"`
}

func loadConfig(path string) (*Config, error) {
	if path != "" {
		// Explicit path given - load only this file
		return loadConfigFile(path)
	}

	// Hierarchical loading: global + local
	cfg := &Config{
		Hosts: make(map[string]Host),
		Tasks: make(map[string]Task),
	}

	// 1. Load global config (~/.config/gosctl/sctl.toml)
	home, err := os.UserHomeDir()
	if err == nil {
		globalPath := filepath.Join(home, ".config", "gosctl", "sctl.toml")
		if globalCfg, err := loadConfigFile(globalPath); err == nil {
			mergeConfig(cfg, globalCfg)
		}
	}

	// 2. Load local config (./sctl.toml)
	localPath := "sctl.toml"
	if localCfg, err := loadConfigFile(localPath); err == nil {
		mergeConfig(cfg, localCfg)
	}

	// Check if we have any config at all
	if len(cfg.Hosts) == 0 && len(cfg.Tasks) == 0 {
		return nil, errors.New("no config found (checked ./sctl.toml and ~/.config/gosctl/sctl.toml)")
	}

	applyDefaults(cfg)
	return cfg, nil
}

func loadConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	if cfg.Hosts == nil {
		cfg.Hosts = make(map[string]Host)
	}
	if cfg.Tasks == nil {
		cfg.Tasks = make(map[string]Task)
	}

	applyDefaults(&cfg)
	return &cfg, nil
}

func mergeConfig(base, overlay *Config) {
	// Hosts: overlay overwrites base
	for name, host := range overlay.Hosts {
		base.Hosts[name] = host
	}
	// Tasks: overlay overwrites base
	for name, task := range overlay.Tasks {
		base.Tasks[name] = task
	}
}

func applyDefaults(cfg *Config) {
	for name, host := range cfg.Hosts {
		if host.Port == 0 {
			host.Port = 22
		}
		if host.User == "" {
			host.User = os.Getenv("USER")
		}
		cfg.Hosts[name] = host
	}
}
