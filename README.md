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

Create `~/.config/gosctl/sctl.toml` for global hosts, or `./sctl.toml` for project-specific ones:

```toml
[hosts.web1]
address = "web1.example.com"
user = "deploy"

[hosts.web2]
address = "web2.example.com"
user = "deploy"

[hosts.db]
address = "db.example.com"
user = "admin"
port = 2222
```

### 2. Run ad-hoc commands

```bash
gosctl exec -H web1 "uptime"
gosctl exec -H db "systemctl status postgresql"
```

### 3. Define project tasks

Create `sctl.toml` in your project directory:

```toml
[tasks.deploy]
hosts = ["web1", "web2"]
workdir = "/var/www/myapp"
steps = [
    "git pull origin main",
    "npm install --production",
    "systemctl --user restart myapp",
]

[tasks.logs]
host = "web1"
steps = ["journalctl --user -u myapp -f"]
```

### 4. Run tasks

```bash
gosctl run deploy
# → Running on web1...
#   → [1/3] git pull origin main
#   → [2/3] npm install --production
#   → [3/3] systemctl --user restart myapp
#   ✓ web1 completed
# → Running on web2...
#   → [1/3] git pull origin main
#   → [2/3] npm install --production
#   → [3/3] systemctl --user restart myapp
#   ✓ web2 completed
# ✓ Task completed on 2 hosts
```

## Configuration

gosctl loads configuration hierarchically:

| Order | File | Purpose |
|-------|------|---------|
| 1 | `~/.config/gosctl/sctl.toml` | Global hosts and tasks |
| 2 | `./sctl.toml` | Project-specific hosts and tasks |

**Merge behavior:** Both files can define hosts and tasks. Local definitions override global ones with the same name. This allows you to define shared hosts globally and project-specific tasks (or host overrides) locally.

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
host = "myserver"          # Single host
workdir = "/app"           # Optional: working directory for all steps
steps = [                  # Required: commands to execute
    "echo 'Hello'",
    "date",
]

[tasks.deploy-all]
hosts = ["web1", "web2"]   # Multiple hosts (runs sequentially)
workdir = "/var/www/app"
steps = ["git pull", "systemctl restart app"]
```

> **Note:** Use either `host` or `hosts`, not both.

## Commands

| Command | Description |
|---------|-------------|
| `gosctl exec -H <host> "<cmd>"` | Execute a single command on a host |
| `gosctl run <task>` | Run a predefined task |
| `gosctl run <task> -H host1 -H host2` | Run task on specific hosts (overrides config) |
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

Or use `just deploy-completion` (default: fish) or `just deploy-completion bash`.

## License

[MIT](LICENSE)
