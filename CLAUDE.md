# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Description

gosctl is a CLI tool for remote service control over SSH. It executes commands and predefined tasks on remote hosts using TOML configuration files.

## Build & Run

```bash
just build          # Build binary
just build-release  # Build with version
just test           # Run tests
just run -- --help  # Run directly
```

## Architecture

Flat project structure with three main files:
- `main.go` - CLI definition with urfave/cli/v3, subcommands: `exec`, `run`, `hosts`
- `config.go` - TOML configuration loading (hosts, tasks)
- `ssh.go` - SSH client with agent, key, and password authentication

## Configuration

Hierarchical loading (local overrides global):
1. `~/.config/gosctl/sctl.toml` - global hosts & tasks
2. `./sctl.toml` - project-specific tasks

```toml
# ~/.config/gosctl/sctl.toml (global hosts)
[hosts.server1]
address = "example.com"
user = "admin"

# ./sctl.toml (project-specific tasks)
[tasks.deploy]
host = "server1"
steps = ["cd /app", "git pull", "systemctl restart app"]
```

## Dependencies

- `github.com/urfave/cli/v3` - CLI framework
- `github.com/BurntSushi/toml` - Configuration parsing
- `golang.org/x/crypto/ssh` - SSH connections
