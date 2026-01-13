package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

type SSHClient struct {
	client    *ssh.Client
	host      Host
	agentConn net.Conn
}

func newSSHClient(host Host) (*SSHClient, error) {
	authMethods, agentConn := buildAuthMethods(host)
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication methods available")
	}

	hostKeyCallback, err := buildHostKeyCallback()
	if err != nil {
		if agentConn != nil {
			agentConn.Close()
		}
		return nil, fmt.Errorf("failed to load known_hosts: %w (add host with: ssh-keyscan -H %s >> ~/.ssh/known_hosts)", err, host.Address)
	}

	config := &ssh.ClientConfig{
		User:            host.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host.Address, host.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		if agentConn != nil {
			agentConn.Close()
		}
		return nil, err
	}

	return &SSHClient{client: client, host: host, agentConn: agentConn}, nil
}

func buildAuthMethods(host Host) ([]ssh.AuthMethod, net.Conn) {
	var methods []ssh.AuthMethod
	var agentConn net.Conn

	// 1. Try SSH agent first
	if auth, conn := sshAgentAuth(); auth != nil {
		methods = append(methods, auth)
		agentConn = conn
	}

	// 2. Try specific key file if configured
	if host.KeyFile != "" {
		if keyAuth := publicKeyAuth(host.KeyFile); keyAuth != nil {
			methods = append(methods, keyAuth)
		}
	}

	// 3. Try default key locations
	home, _ := os.UserHomeDir()
	defaultKeys := []string{
		filepath.Join(home, ".ssh", "id_ed25519"),
		filepath.Join(home, ".ssh", "id_rsa"),
		filepath.Join(home, ".ssh", "id_ecdsa"),
	}
	for _, keyPath := range defaultKeys {
		if keyAuth := publicKeyAuth(keyPath); keyAuth != nil {
			methods = append(methods, keyAuth)
		}
	}

	// 4. Password as fallback
	if host.Password != "" {
		methods = append(methods, ssh.Password(host.Password))
	}

	return methods, agentConn
}

func sshAgentAuth() (ssh.AuthMethod, net.Conn) {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return nil, nil
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, nil
	}

	agentClient := agent.NewClient(conn)
	return ssh.PublicKeysCallback(agentClient.Signers), conn
}

func publicKeyAuth(keyPath string) ssh.AuthMethod {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		// Try with empty passphrase - if key is encrypted, this will fail
		// Could add interactive passphrase prompt here later
		return nil
	}

	return ssh.PublicKeys(signer)
}

func buildHostKeyCallback() (ssh.HostKeyCallback, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	knownHostsPath := filepath.Join(home, ".ssh", "known_hosts")
	return knownhosts.New(knownHostsPath)
}

func (c *SSHClient) Run(command string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	return session.Run(command)
}

func (c *SSHClient) Close() error {
	if c.agentConn != nil {
		c.agentConn.Close()
	}
	return c.client.Close()
}
