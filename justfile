app_name := "gosctl"
bin_dir := "bin"
bin_file := bin_dir / app_name
git_version := `printf "%s%s" "$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')" "$([ -z "$(git status --porcelain 2>/dev/null)" ] || echo "-dirty")"`

[private]
default:
    @just --list --unsorted

# ============================================================
# Development
# ============================================================

# Run program directly
[group('dev')]
run *args:
    @go run . {{args}}

# Format code
[group('dev')]
fmt:
    @go fmt ./...

# Static analysis
[group('dev')]
vet:
    @go vet ./...

# Run tests
[group('dev')]
test:
    @go test -v ./...

# ============================================================
# Build
# ============================================================

# Build binary
[group('build')]
build:
    @mkdir -p {{bin_dir}}
    @go build -o {{bin_file}} .

# Build release binary with version
[group('build')]
build-release version="0.1.0":
    @mkdir -p {{bin_dir}}
    @go build -ldflags "-X main.version=v{{version}}" -o {{bin_file}} .

# Build binaries for all platforms
[group('build')]
build-all version="0.1.0":
    @mkdir -p {{bin_dir}}
    @GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=v{{version}}" -o {{bin_dir}}/{{app_name}}-linux-amd64 .
    @GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=v{{version}}" -o {{bin_dir}}/{{app_name}}-linux-arm64 .
    @GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=v{{version}}" -o {{bin_dir}}/{{app_name}}-darwin-amd64 .
    @GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=v{{version}}" -o {{bin_dir}}/{{app_name}}-darwin-arm64 .

# ============================================================
# Install
# ============================================================

# Install locally (go install)
[group('install')]
install:
    @go install .

# Install with version
[group('install')]
install-release version="0.1.0":
    @go install -ldflags "-X main.version=v{{version}}" .

# ============================================================
# Deploy
# ============================================================

# Deploy binary and completions locally
[group('deploy')]
deploy shell="fish": build (deploy-bin) (deploy-completion shell)

# Copy binary to ~/.local/bin
[group('deploy')]
deploy-bin:
    @mkdir -p ~/.local/bin
    @cp {{bin_file}} ~/.local/bin/{{app_name}}
    @echo "Installed {{app_name}} to ~/.local/bin/"

# Generate and install shell completions
[group('deploy')]
deploy-completion shell="fish":
    #!/usr/bin/env sh
    case "{{shell}}" in
        fish)
            mkdir -p ~/.config/fish/completions
            {{bin_file}} completion fish > ~/.config/fish/completions/{{app_name}}.fish
            echo "Installed fish completions"
            ;;
        bash)
            mkdir -p ~/.local/share/bash-completion/completions
            {{bin_file}} completion bash > ~/.local/share/bash-completion/completions/{{app_name}}
            echo "Installed bash completions"
            ;;
        zsh)
            mkdir -p ~/.local/share/zsh/site-functions
            {{bin_file}} completion zsh > ~/.local/share/zsh/site-functions/_{{app_name}}
            echo "Installed zsh completions"
            ;;
        *)
            echo "Unknown shell: {{shell}}. Available: fish, bash, zsh"
            exit 1
            ;;
    esac

# ============================================================
# Clean
# ============================================================

# Remove build artifacts
[group('clean')]
clean:
    @rm -rf {{bin_dir}}
