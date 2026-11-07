package uci

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// represents running UCI engine process and its communication pipes
type Client struct {
	Path    string
	process *exec.Cmd
	stdin   io.WriteCloser
	stdout  *bufio.Scanner
}

// verifies binary, spawns process, and performs UCI handshake
func Load(path string) (*Client, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("engine file not found: %s", path)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start engine process: %w", err)
	}

	client := &Client{
		Path:    path,
		process: cmd,
		stdin:   stdin,
		stdout:  bufio.NewScanner(stdout),
	}

	if err := client.handshake(); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}

// handshake sequentially verifies UCI protocol compliance
func (c *Client) handshake() error {
	if _, err := fmt.Fprintln(c.stdin, "uci"); err != nil {
		return fmt.Errorf("failed to write 'uci': %w", err)
	}

	uciok := false
	for c.stdout.Scan() {
		line := strings.TrimSpace(c.stdout.Text())
		if line == "uciok" {
			uciok = true
			break
		}
	}
	if !uciok {
		return fmt.Errorf("engine did not respond with 'uciok'")
	}

	if _, err := fmt.Fprintln(c.stdin, "isready"); err != nil {
		return fmt.Errorf("failed to write 'isready': %w", err)
	}

	readyok := false
	for c.stdout.Scan() {
		line := strings.TrimSpace(c.stdout.Text())
		if line == "readyok" {
			readyok = true
			break
		}
	}
	if !readyok {
		return fmt.Errorf("engine did not respond with 'readyok'")
	}

	return nil
}

// gracefully terminates engine process
func (c *Client) Close() {
	if c.stdin != nil {
		fmt.Fprintln(c.stdin, "quit")
		c.stdin.Close()
	}
	if c.process != nil {
		c.process.Wait()
	}
}
