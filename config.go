package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Hosts map[string]Host `toml:"hosts"`
	Tasks map[string]Task `toml:"tasks"`

	// Source tracking (not from TOML)
	HostSources map[string]string `toml:"-"`
	TaskSources map[string]string `toml:"-"`
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
	Hosts   []string `toml:"hosts"`
	Workdir string   `toml:"workdir"`
	Before  []string `toml:"before"`
	Steps   []string `toml:"steps"`
	After   []string `toml:"after"`
}

// GetHosts returns the target hosts for this task.
func (t Task) GetHosts() []string {
	if len(t.Hosts) > 0 {
		return t.Hosts
	}
	if t.Host != "" {
		return []string{t.Host}
	}
	return nil
}

// Validate checks the task configuration for errors.
func (t Task) Validate(name string) error {
	if t.Host != "" && len(t.Hosts) > 0 {
		return fmt.Errorf("task %q: use either 'host' or 'hosts', not both", name)
	}
	if t.Host == "" && len(t.Hosts) == 0 {
		return fmt.Errorf("task %q: missing 'host' or 'hosts'", name)
	}
	if len(t.Steps) == 0 {
		return fmt.Errorf("task %q: missing 'steps'", name)
	}
	return nil
}

// ValidateRefs checks that all before/after task references exist.
func (t Task) ValidateRefs(name string, tasks map[string]Task) error {
	for _, ref := range t.Before {
		if _, ok := tasks[ref]; !ok {
			return fmt.Errorf("task %q: before task %q not found", name, ref)
		}
	}
	for _, ref := range t.After {
		if _, ok := tasks[ref]; !ok {
			return fmt.Errorf("task %q: after task %q not found", name, ref)
		}
	}
	return nil
}

func loadConfig(configPath, filePath string) (*Config, error) {
	if configPath != "" {
		// --config: load only this file, skip hierarchical loading
		cfg, err := loadConfigFile(configPath)
		if err != nil {
			return nil, err
		}
		// Mark all as from this file
		cfg.HostSources = make(map[string]string)
		cfg.TaskSources = make(map[string]string)
		for name := range cfg.Hosts {
			cfg.HostSources[name] = configPath
		}
		for name := range cfg.Tasks {
			cfg.TaskSources[name] = configPath
		}
		return cfg, nil
	}

	// Hierarchical loading: global + local
	cfg := &Config{
		Hosts:       make(map[string]Host),
		Tasks:       make(map[string]Task),
		HostSources: make(map[string]string),
		TaskSources: make(map[string]string),
	}

	// 1. Load global config (~/.config/gosctl/sctl.toml)
	home, err := os.UserHomeDir()
	if err == nil {
		globalPath := filepath.Join(home, ".config", "gosctl", "sctl.toml")
		if globalCfg, err := loadConfigFile(globalPath); err == nil {
			mergeConfigWithSource(cfg, globalCfg, "global")
		}
	}

	// 2. Load local config (--file or ./sctl.toml)
	localPath := "sctl.toml"
	if filePath != "" {
		localPath = filePath
	}
	if localCfg, err := loadConfigFile(localPath); err == nil {
		mergeConfigWithSource(cfg, localCfg, "local")
	} else if filePath != "" {
		// --file was explicit, so error if not found
		return nil, fmt.Errorf("config file not found: %s", filePath)
	}

	// Check if we have any config at all
	if len(cfg.Hosts) == 0 && len(cfg.Tasks) == 0 {
		return nil, fmt.Errorf("no config found (checked ./sctl.toml and ~/.config/gosctl/sctl.toml)")
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

func mergeConfigWithSource(base, overlay *Config, source string) {
	// Hosts: overlay overwrites base, track if overwritten
	for name, host := range overlay.Hosts {
		if _, exists := base.Hosts[name]; exists {
			base.HostSources[name] = "local (overrides global)"
		} else {
			base.HostSources[name] = source
		}
		base.Hosts[name] = host
	}
	// Tasks: overlay overwrites base, track if overwritten
	for name, task := range overlay.Tasks {
		if _, exists := base.Tasks[name]; exists {
			base.TaskSources[name] = "local (overrides global)"
		} else {
			base.TaskSources[name] = source
		}
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
