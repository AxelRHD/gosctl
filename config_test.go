package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	configContent := `
[hosts.server1]
address = "example.com"
port = 2222
user = "admin"
key_file = "/path/to/key"

[hosts.server2]
address = "other.com"

[tasks.deploy]
host = "server1"
steps = ["echo hello", "echo world"]
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}

	// Test hosts
	if len(cfg.Hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d", len(cfg.Hosts))
	}

	server1 := cfg.Hosts["server1"]
	if server1.Address != "example.com" {
		t.Errorf("expected address example.com, got %s", server1.Address)
	}
	if server1.Port != 2222 {
		t.Errorf("expected port 2222, got %d", server1.Port)
	}
	if server1.User != "admin" {
		t.Errorf("expected user admin, got %s", server1.User)
	}
	if server1.KeyFile != "/path/to/key" {
		t.Errorf("expected key_file /path/to/key, got %s", server1.KeyFile)
	}

	// Test defaults
	server2 := cfg.Hosts["server2"]
	if server2.Port != 22 {
		t.Errorf("expected default port 22, got %d", server2.Port)
	}
	if server2.User == "" {
		t.Error("expected default user from $USER, got empty")
	}

	// Test tasks
	if len(cfg.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(cfg.Tasks))
	}

	deploy := cfg.Tasks["deploy"]
	if deploy.Host != "server1" {
		t.Errorf("expected task host server1, got %s", deploy.Host)
	}
	if len(deploy.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(deploy.Steps))
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := loadConfig("/nonexistent/path/config.toml")
	if err == nil {
		t.Error("expected error for missing config file")
	}
}

func TestLoadConfigInvalidTOML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	if err := os.WriteFile(configPath, []byte("invalid toml [[["), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := loadConfig(configPath)
	if err == nil {
		t.Error("expected error for invalid TOML")
	}
}
