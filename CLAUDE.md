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
- `main.go` - CLI definition with urfave/cli/v3, subcommands: `exec`, `run`, `hosts`, `tasks`, `check-config`
- `config.go` - TOML configuration loading (hosts, tasks with before/after dependencies)
- `ssh.go` - SSH client with agent, key, and password authentication

## Configuration

Hierarchical loading (local overrides global):
1. `~/.config/gosctl/sctl.toml` - global hosts & tasks
2. `./sctl.toml` - project-specific tasks

**Config flags:**
- `-f, --file` - use different local file (global still loaded)
- `-c, --config` - load only this file (skip hierarchical loading)

```toml
# ~/.config/gosctl/sctl.toml (global hosts)
[hosts.server1]
address = "example.com"
user = "admin"

# ./sctl.toml (project-specific tasks)
[tasks.deploy]
hosts = ["server1"]
before = ["backup"]      # Run before main steps
workdir = "/app"
steps = ["git pull", "systemctl restart app"]
after = ["notify"]       # Run after completion
```

## Dependencies

- `github.com/urfave/cli/v3` - CLI framework
- `github.com/BurntSushi/toml` - Configuration parsing
- `golang.org/x/crypto/ssh` - SSH connections

## Future Features

### Global settings file
A `~/.config/gosctl/settings.toml` for user preferences:
- `icons = true/false` - toggle emoji icons
- `verbose = true/false` - verbose output

### Icon mapping (for future ASCII fallback)
```
ASCII    Emoji   Usage
------   -----   -----
[H]      🖥      Host
[T]      📋      Task / Section header
>        ▶      Step execution
[ok]     ✓      Step completed
[OK]     ✅     Task completed
[err]    ❌     Error
[!]      ⚠      Warning
*        ⚡     Override indicator
->       →      Connection info
```
