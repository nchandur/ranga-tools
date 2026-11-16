package uci

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	StepTimeout = 5 * time.Second // how long we wait for each handshake step
	quitGrace   = time.Second     // quitGrace is how long we let the engine exit on its own
)

// holds the identity an engine reports during the UCI handshake.
type Info struct {
	Name   string
	Author string
}

// launches the executable at path and performs UCI handshake
func Verify(path string) (Info, error) {
	cmd := exec.Command(path)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return Info{}, fmt.Errorf("open engine stdin: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Info{}, fmt.Errorf("open engine stdout: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return Info{}, fmt.Errorf("start engine: %w", err)
	}

	stop := make(chan struct{})
	lines := readLines(stdout, stop)

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	defer func() {
		close(stop)
		shutdown(cmd, stdin, done)
	}()

	var info Info

	if err := send(stdin, "uci"); err != nil {
		return Info{}, err
	}
	err = waitFor(lines, "uciok", func(line string) {
		switch {
		case strings.HasPrefix(line, "id name "):
			info.Name = strings.TrimSpace(strings.TrimPrefix(line, "id name "))
		case strings.HasPrefix(line, "id author "):
			info.Author = strings.TrimSpace(strings.TrimPrefix(line, "id author "))
		}
	})
	if err != nil {
		return Info{}, err
	}

	if err := send(stdin, "isready"); err != nil {
		return Info{}, err
	}
	if err := waitFor(lines, "readyok", nil); err != nil {
		return Info{}, err
	}

	return info, nil
}
