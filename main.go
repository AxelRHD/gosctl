package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

var version = "dev"

func main() {
	app := &cli.Command{
		Name:                  "gosctl",
		Usage:                 "Remote service control over SSH",
		Version:               version,
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
				Action:    runAction,
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

	host, ok := cfg.Hosts[task.Host]
	if !ok {
		return fmt.Errorf("host %q not found in config", task.Host)
	}

	client, err := newSSHClient(host)
	if err != nil {
		return fmt.Errorf("ssh connection failed: %w", err)
	}
	defer client.Close()

	for i, step := range task.Steps {
		cmd := step
		if task.Workdir != "" {
			cmd = fmt.Sprintf("cd %s && %s", task.Workdir, step)
		}
		fmt.Printf("→ [%d/%d] %s\n", i+1, len(task.Steps), step)
		if err := client.Run(cmd); err != nil {
			return fmt.Errorf("step %d failed: %w", i+1, err)
		}
	}

	fmt.Println("✓ Task completed")
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
