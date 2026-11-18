package uci

import (
	"errors"
	"fmt"
	"os/exec"
	"time"
)

var ErrNoName = errors.New("engine did not report an \"id name\"; use --name to provide one")

const (
	stepTimeout   = 5 * time.Second
	searchTimeout = 10 * time.Minute
	quitGrace     = time.Second
)

type Info struct{ Name, Author string }

type Score struct {
	CP     int
	Mate   int
	IsMate bool
}

type Result struct {
	Score    Score
	BestMove string
}

// performs UCI handshake and shuts the engine down again
func Verify(path string) (Info, error) {
	s, err := Start(path, nil)
	if err != nil {
		return Info{}, err
	}
	defer s.Close()
	return s.Info, nil
}

// launches the engine
func Start(path string, options map[string]string) (*Session, error) {
	cmd := exec.Command(path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("open engine stdin: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("open engine stdout: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start engine: %w", err)
	}

	s := &Session{
		cmd:   cmd,
		stdin: stdin,
		lines: make(chan string),
		done:  make(chan error, 1),
		stop:  make(chan struct{}),
	}
	go s.readLines(stdout)
	go func() { s.done <- cmd.Wait() }()

	if err := s.handshake(options); err != nil {
		s.Close()
		return nil, err
	}
	return s, nil
}
