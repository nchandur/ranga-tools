package uci

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// readLines scans r line by line and delivers trimmed lines on returned chan
func readLines(r io.Reader, stop <-chan struct{}) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		sc := bufio.NewScanner(r)

		if sc.Err() != nil {
			return
		}

		for sc.Scan() {
			select {
			case out <- strings.TrimSpace(sc.Text()):
			case <-stop:
				return
			}
		}
	}()
	return out
}

// consumes lines until one equals token
func waitFor(lines <-chan string, token string, onLine func(string)) error {
	timer := time.NewTimer(StepTimeout)
	defer timer.Stop()

	for {
		select {
		case line, ok := <-lines:
			if !ok {
				return fmt.Errorf("engine exited before sending %q", token)
			}
			if line == token {
				return nil
			}
			if onLine != nil {
				onLine(line)
			}
		case <-timer.C:
			return fmt.Errorf("timed out after %s waiting for %q", StepTimeout, token)
		}
	}
}

func send(w io.Writer, command string) error {
	if _, err := io.WriteString(w, command+"\n"); err != nil {
		return fmt.Errorf("send %q to engine: %w", command, err)
	}
	return nil
}

// shutdown asks the engine to quit, then kills it if it doesn't exit in time
func shutdown(cmd *exec.Cmd, stdin io.WriteCloser, done <-chan error) {
	send(stdin, "quit")
	stdin.Close()

	select {
	case <-done:
	case <-time.After(quitGrace):
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		<-done
	}
}
