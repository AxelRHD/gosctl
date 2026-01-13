package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

var (
	appVersion = "dev"
	gitVersion = "unknown"
)

func main() {
	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Printf("gosctl %s (git: %s)\n", appVersion, gitVersion)
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
				Usage:   "config file path",
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
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func execAction(ctx context.Context, cmd *cli.Command) error {
	cfg, err := loadConfig(cmd.String("config"))
	if err != nil {
		return err
	}

	hostName := cmd.String("host")
	host, ok := cfg.Hosts[hostName]
	if !ok {
		return fmt.Errorf("host %q not found in config", hostName)
	}

	client, err := newSSHClient(host)
	if err != nil {
		return fmt.Errorf("ssh connection failed: %w", err)
	}
	defer client.Close()

	command := cmd.Args().First()
	if command == "" {
		return fmt.Errorf("no command provided")
	}

	return client.Run(command)
}

func runAction(ctx context.Context, cmd *cli.Command) error {
	cfg, err := loadConfig(cmd.String("config"))
	if err != nil {
		return err
	}

	taskName := cmd.Args().First()
	if taskName == "" {
		return fmt.Errorf("no task name provided")
	}

	task, ok := cfg.Tasks[taskName]
	if !ok {
		return fmt.Errorf("task %q not found in config", taskName)
	}

	// Validate task config
	if err := task.Validate(taskName); err != nil {
		return err
	}

	// Use CLI hosts if provided, otherwise use task config
	hostNames := cmd.StringSlice("host")
	if len(hostNames) == 0 {
		hostNames = task.GetHosts()
	}

	// Run on each host sequentially
	for _, hostName := range hostNames {
		host, ok := cfg.Hosts[hostName]
		if !ok {
			return fmt.Errorf("host %q not found in config", hostName)
		}

		if err := runTaskOnHost(host, hostName, task, len(hostNames) > 1); err != nil {
			return err
		}
	}

	if len(hostNames) > 1 {
		fmt.Printf("✓ Task completed on %d hosts\n", len(hostNames))
	} else {
		fmt.Println("✓ Task completed")
	}
	return nil
}

func runTaskOnHost(host Host, hostName string, task Task, showHostHeader bool) error {
	if showHostHeader {
		fmt.Printf("→ Running on %s...\n", hostName)
	}

	client, err := newSSHClient(host)
	if err != nil {
		return fmt.Errorf("ssh connection to %s failed: %w", hostName, err)
	}
	defer client.Close()

	for i, step := range task.Steps {
		cmd := step
		if task.Workdir != "" {
			cmd = fmt.Sprintf("cd %s && %s", task.Workdir, step)
		}
		if showHostHeader {
			fmt.Printf("  → [%d/%d] %s\n", i+1, len(task.Steps), step)
		} else {
			fmt.Printf("→ [%d/%d] %s\n", i+1, len(task.Steps), step)
		}
		if err := client.Run(cmd); err != nil {
			return fmt.Errorf("step %d on %s failed: %w", i+1, hostName, err)
		}
	}

	if showHostHeader {
		fmt.Printf("  ✓ %s completed\n", hostName)
	}
	return nil
}

func hostsAction(ctx context.Context, cmd *cli.Command) error {
	cfg, err := loadConfig(cmd.String("config"))
	if err != nil {
		return err
	}

	for name, host := range cfg.Hosts {
		fmt.Printf("%s → %s@%s:%d\n", name, host.User, host.Address, host.Port)
	}
	return nil
}
