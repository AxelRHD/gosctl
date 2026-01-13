<p align="center">
  <img src="gosctl-logo.png" alt="gosctl logo" width="440">
</p>

<p align="center"><strong>Remote service control over SSH</strong><br>Execute commands and predefined tasks on remote hosts with a simple, declarative configuration.</p>

## Features

- 🔐 **Multiple auth methods** — SSH agent, key files, or password
- 📁 **Hierarchical config** — Global hosts + project-specific tasks
- 🚀 **Task automation** — Define multi-step deployment workflows
- 🐚 **Shell completions** — Fish, Bash, and Zsh supported
- ⚡ **Zero dependencies** — Single binary, no runtime required

## Installation

### From source

```bash
go install github.com/axelrhd/gosctl@latest
```

### With just

```bash
git clone https://github.com/axelrhd/gosctl.git
cd gosctl
just deploy  # Builds and installs to ~/.local/bin with shell completions
```

## Quick Start

### 1. Define your hosts

Create `~/.config/gosctl/sctl.toml`:

```toml
[hosts.web]
address = "web.example.com"
user = "deploy"

[hosts.db]
address = "db.example.com"
user = "admin"
port = 2222
```

### 2. Run ad-hoc commands

```bash
gosctl exec -H web "uptime"
gosctl exec -H db "systemctl status postgresql"
```

### 3. Define project tasks

Create `sctl.toml` in your project directory:

```toml
[tasks.deploy]
host = "web"
workdir = "/var/www/myapp"
steps = [
    "git pull origin main",
    "npm install --production",
    "systemctl --user restart myapp",
]

[tasks.logs]
host = "web"
steps = ["journalctl --user -u myapp -f"]
```

### 4. Run tasks

```bash
gosctl run deploy
# → [1/3] git pull origin main
# → [2/3] npm install --production
# → [3/3] systemctl --user restart myapp
# ✓ Task completed
```

## Configuration

gosctl loads configuration hierarchically:

| File | Purpose |
|------|---------|
| `~/.config/gosctl/sctl.toml` | Global hosts and tasks |
| `./sctl.toml` | Project-specific tasks (overrides global) |

### Host options

```toml
[hosts.myserver]
address = "example.com"    # Required
user = "deploy"            # Default: $USER
port = 22                  # Default: 22
key_file = "~/.ssh/id_ed25519"  # Optional, uses SSH agent by default
password = "secret"        # Optional, not recommended
```

### Task options

```toml
[tasks.mytask]
host = "myserver"          # Required: host name from config
workdir = "/app"           # Optional: working directory for all steps
steps = [                  # Required: commands to execute
    "echo 'Hello'",
    "date",
]
```

## Commands

| Command | Description |
|---------|-------------|
| `gosctl exec -H <host> "<cmd>"` | Execute a single command on a host |
| `gosctl run <task>` | Run a predefined task |
| `gosctl hosts` | List all configured hosts |
| `gosctl completion <shell>` | Generate shell completions |

## Shell Completions

```bash
# Fish (recommended)
gosctl completion fish > ~/.config/fish/completions/gosctl.fish

# Bash
gosctl completion bash > ~/.local/share/bash-completion/completions/gosctl

# Zsh
gosctl completion zsh > ~/.local/share/zsh/site-functions/_gosctl
```

## License

MIT
