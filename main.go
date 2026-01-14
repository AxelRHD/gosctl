package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/urfave/cli/v3"
)

var appVersion = "dev"

func main() {
	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Printf("gosctl %s\n", appVersion)
	}

	app := &cli.Command{
		Name:                  "gosctl",
		Usage:                 "Remote service control over SSH",
		Version:               appVersion,
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "load only this file, skip global + local merging",
			},
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "use instead of ./sctl.toml (global config still loaded)",
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "exec",
				Usage:     "Execute a command on a remote host",
				ArgsUsage: "[command]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "host",
						Aliases:  []string{"H"},
						Usage:    "target host",
						Required: true,
					},
				},
				Action: execAction,
			},
			{
				Name:      "run",
				Usage:     "Run a predefined task",
				ArgsUsage: "[task]",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:    "host",
						Aliases: []string{"H"},
						Usage:   "target host (can be specified multiple times)",
					},
				},
				Action: runAction,
			},
			{
				Name:   "hosts",
				Usage:  "List configured hosts",
				Action: hostsAction,
			},
			{
				Name:   "tasks",
				Usage:  "List configured tasks",
				Action: tasksAction,
			},
			{
				Name:   "check-config",
				Usage:  "Validate configuration files",
				Action: checkConfigAction,
			},
			{
				Name:  "init",
				Usage: "Create a sample configuration file",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "local",
						Usage: "create ./sctl.toml instead of global config",
					},
					&cli.BoolFlag{
						Name:  "force",
						Usage: "overwrite existing file",
					},
				},
				Action: initAction,
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func execAction(ctx context.Context, cmd *cli.Command) error {
	cfg, err := loadConfig(cmd.String("config"), cmd.String("file"))
	if err != nil {
		return err
	}

	hostName := cmd.String("host")
	host, ok := cfg.Hosts[hostName]
	if !ok {
		return errorf("host %q not found in config", hostName)
	}

	client, err := newSSHClient(host)
	if err != nil {
		return errorf("ssh connection failed: %w", err)
	}
	defer client.Close()

	command := cmd.Args().First()
	if command == "" {
		return errorf("no command provided")
	}

	return client.Run(command)
}

func runAction(ctx context.Context, cmd *cli.Command) error {
	cfg, err := loadConfig(cmd.String("config"), cmd.String("file"))
	if err != nil {
		return err
	}

	taskName := cmd.Args().First()
	if taskName == "" {
		return errorf("no task name provided")
	}

	task, ok := cfg.Tasks[taskName]
	if !ok {
		return errorf("task %q not found in config", taskName)
	}

	// Validate task config
	if err := task.Validate(taskName); err != nil {
		return errorf("%v", err)
	}

	// Validate task references
	if err := task.ValidateRefs(taskName, cfg.Tasks); err != nil {
		return errorf("%v", err)
	}

	// Use CLI hosts if provided, otherwise use task config
	hostNames := cmd.StringSlice("host")
	if len(hostNames) == 0 {
		hostNames = task.GetHosts()
	}

	// Execute before tasks
	for _, beforeName := range task.Before {
		beforeTask := cfg.Tasks[beforeName]
		if err := executeTask(cfg, beforeName, beforeTask, hostNames); err != nil {
			return err
		}
	}

	// Run main task on each host
	for _, hostName := range hostNames {
		host, ok := cfg.Hosts[hostName]
		if !ok {
			return errorf("host %q not found in config", hostName)
		}

		if err := runTaskOnHost(host, hostName, task, len(hostNames) > 1); err != nil {
			return err
		}
	}

	// Execute after tasks
	for _, afterName := range task.After {
		afterTask := cfg.Tasks[afterName]
		if err := executeTask(cfg, afterName, afterTask, hostNames); err != nil {
			return err
		}
	}

	if len(hostNames) > 1 {
		printSuccess("Task completed on %d hosts", len(hostNames))
	} else {
		printSuccess("Task completed")
	}
	return nil
}

// executeTask runs a referenced task (from before/after) with host mismatch warnings.
func executeTask(cfg *Config, taskName string, task Task, parentHosts []string) error {
	taskHosts := task.GetHosts()

	// Check for host mismatch and warn
	hasOverlap := false
	for _, h := range taskHosts {
		if slices.Contains(parentHosts, h) {
			hasOverlap = true
			break
		}
	}
	if !hasOverlap {
		printWarning("Note: %s runs on different host(s): %s", taskName, strings.Join(taskHosts, ", "))
	}

	printTaskHeader(taskName)

	for _, hostName := range taskHosts {
		host, ok := cfg.Hosts[hostName]
		if !ok {
			return errorf("host %q not found in config", hostName)
		}

		if err := runTaskOnHost(host, hostName, task, len(taskHosts) > 1); err != nil {
			return err
		}
	}

	return nil
}

func runTaskOnHost(host Host, hostName string, task Task, showHostHeader bool) error {
	if showHostHeader {
		printHostHeader(hostName)
	}

	client, err := newSSHClient(host)
	if err != nil {
		return errorf("ssh connection to %s failed: %w", hostName, err)
	}
	defer client.Close()

	for i, step := range task.Steps {
		cmd := step
		if task.Workdir != "" {
			cmd = fmt.Sprintf("cd %s && %s", task.Workdir, step)
		}
		printStep(i+1, len(task.Steps), step, showHostHeader)
		if err := client.Run(cmd); err != nil {
			return errorf("step %d on %s failed: %w", i+1, hostName, err)
		}
	}

	if showHostHeader {
		printStepDone(hostName)
	}
	return nil
}

func hostsAction(ctx context.Context, cmd *cli.Command) error {
	cfg, err := loadConfig(cmd.String("config"), cmd.String("file"))
	if err != nil {
		return err
	}

	for name, host := range cfg.Hosts {
		source := cfg.HostSources[name]
		override := source == "local (overrides global)"
		printHost(name, host.User, host.Address, host.Port, source, override)
	}
	return nil
}

func tasksAction(ctx context.Context, cmd *cli.Command) error {
	cfg, err := loadConfig(cmd.String("config"), cmd.String("file"))
	if err != nil {
		return err
	}

	for name, task := range cfg.Tasks {
		hosts := strings.Join(task.GetHosts(), ", ")
		source := cfg.TaskSources[name]

		var extras []string
		if len(task.Before) > 0 {
			extras = append(extras, fmt.Sprintf("before: %s", strings.Join(task.Before, ", ")))
		}
		if len(task.After) > 0 {
			extras = append(extras, fmt.Sprintf("after: %s", strings.Join(task.After, ", ")))
		}

		info := fmt.Sprintf("hosts: %s, steps: %d", hosts, len(task.Steps))
		if len(extras) > 0 {
			info += ", " + strings.Join(extras, ", ")
		}

		override := source == "local (overrides global)"
		printTask(name, info, source, override)
	}
	return nil
}

func initAction(ctx context.Context, cmd *cli.Command) error {
	local := cmd.Bool("local")
	force := cmd.Bool("force")

	var path string
	var content string

	if local {
		path = "sctl.toml"
		content = localConfigTemplate
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return errorf("could not determine home directory: %w", err)
		}
		dir := filepath.Join(home, ".config", "gosctl")
		path = filepath.Join(dir, "sctl.toml")

		// Create directory if it doesn't exist
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errorf("could not create config directory: %w", err)
		}
		content = globalConfigTemplate
	}

	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		if !force {
			printWarning("File already exists: %s", path)
			fmt.Println("Use --force to overwrite")
			return nil
		}
		printWarning("Overwriting existing file: %s", path)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return errorf("could not write config file: %w", err)
	}

	printSuccess("Created %s", path)
	return nil
}

const globalConfigTemplate = `# gosctl global configuration
# Hosts defined here are available from anywhere on your system.
# Location: ~/.config/gosctl/sctl.toml

# ============================================================================
# HOSTS
# ============================================================================
# Define your remote hosts here. These can be referenced by name in tasks.

[hosts.server1]
address = "server1.example.com"
user = "admin"
# port = 22                    # default: 22
# key_file = "~/.ssh/id_rsa"   # default: ssh-agent

[hosts.server2]
address = "server2.example.com"
user = "admin"

# ============================================================================
# GLOBAL TASKS
# ============================================================================
# Tasks defined here are available from anywhere on your system.
# Useful for common operations you run across projects.

[tasks.system-check]
hosts = ["server1"]
steps = [
    "uptime",
    "df -h",
    "free -m",
]

[tasks.update-system]
hosts = ["server1"]
steps = [
    "sudo apt update",
    "sudo apt upgrade -y",
]
`

const localConfigTemplate = `# gosctl project configuration
# Tasks defined here are specific to this project.
# Location: ./sctl.toml (in your project directory)
#
# Hosts from ~/.config/gosctl/sctl.toml are automatically available.

# ============================================================================
# PROJECT TASKS
# ============================================================================

[tasks.deploy]
hosts = ["server1"]           # Reference hosts from global config
workdir = "/var/www/myapp"
before = ["backup"]           # Run backup task first
steps = [
    "git pull origin main",
    "npm install",
    "npm run build",
    "systemctl restart myapp",
]

[tasks.backup]
hosts = ["server1"]
steps = [
    "tar -czf /backups/myapp-$(date +%Y%m%d).tar.gz /var/www/myapp",
]

[tasks.logs]
hosts = ["server1"]
steps = [
    "journalctl -u myapp -n 50 --no-pager",
]
`

func checkConfigAction(ctx context.Context, cmd *cli.Command) error {
	cfg, err := loadConfig(cmd.String("config"), cmd.String("file"))
	if err != nil {
		return err
	}

	hasErrors := false

	// Check hosts
	printSection("Hosts")
	for name, host := range cfg.Hosts {
		if host.Address == "" {
			printInvalid(name)
			printIssue("missing address")
			hasErrors = true
		} else {
			printValid("%s (%s@%s:%d)", name, host.User, host.Address, host.Port)
		}
	}

	// Check tasks
	fmt.Println()
	printSection("Tasks")
	for name, task := range cfg.Tasks {
		var issues []string

		// Basic validation
		if err := task.Validate(name); err != nil {
			issues = append(issues, err.Error())
		}

		// Check before/after references
		if err := task.ValidateRefs(name, cfg.Tasks); err != nil {
			issues = append(issues, err.Error())
		}

		// Check host references
		for _, hostName := range task.GetHosts() {
			if _, ok := cfg.Hosts[hostName]; !ok {
				issues = append(issues, fmt.Sprintf("host %q not found", hostName))
			}
		}

		if len(issues) > 0 {
			printInvalid(name)
			for _, issue := range issues {
				printIssue(issue)
			}
			hasErrors = true
		} else {
			hosts := strings.Join(task.GetHosts(), ", ")
			printValid("%s (hosts: %s, steps: %d)", name, hosts, len(task.Steps))
		}
	}

	fmt.Println()
	if hasErrors {
		printWarning("Configuration has errors")
		return errorf("configuration validation failed")
	}

	printSuccess("Configuration OK")
	return nil
}
