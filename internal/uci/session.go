package uci

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Session struct {
	Info Info

	cmd   *exec.Cmd
	stdin io.WriteCloser
	lines chan string
	done  chan error
	stop  chan struct{}
}

func (s *Session) handshake(options map[string]string) error {
	if err := s.send("uci"); err != nil {
		return err
	}
	err := s.waitFor("uciok", func(line string) {
		if v, ok := strings.CutPrefix(line, "id name "); ok {
			s.Info.Name = strings.TrimSpace(v)
		} else if v, ok := strings.CutPrefix(line, "id author "); ok {
			s.Info.Author = strings.TrimSpace(v)
		}
	})
	if err != nil {
		return err
	}
	for k, v := range options {
		if err := s.send("setoption name " + k + " value " + v); err != nil {
			return err
		}
	}
	return s.sync()
}

// tells engine a new game starts and waits until it is ready
func (s *Session) NewGame() error {
	if err := s.send("ucinewgame"); err != nil {
		return err
	}
	return s.sync()
}

// analyzes the position after playing moves
func (s *Session) Search(startFEN string, moves []string, depth int) (Result, error) {
	pos := "position fen " + startFEN
	if len(moves) > 0 {
		pos += " moves " + strings.Join(moves, " ")
	}
	if err := s.send(pos); err != nil {
		return Result{}, err
	}
	if err := s.send("go depth " + strconv.Itoa(depth)); err != nil {
		return Result{}, err
	}

	var (
		res       Result
		haveScore bool
		timer     = time.NewTimer(searchTimeout)
	)
	defer timer.Stop()
	for {
		select {
		case line, ok := <-s.lines:
			if !ok {
				return Result{}, errors.New("engine exited during search")
			}
			if sc, ok := parseScore(line); ok {
				res.Score, haveScore = sc, true
			} else if v, ok := strings.CutPrefix(line, "bestmove "); ok {
				res.BestMove, _, _ = strings.Cut(v, " ")
				if res.BestMove == "" || res.BestMove == "(none)" || res.BestMove == "0000" {
					return Result{}, fmt.Errorf("engine has no move in position %q", pos)
				}
				if !haveScore {
					return Result{}, fmt.Errorf("engine sent no score for %q", pos)
				}
				return res, nil
			}
		case <-timer.C:
			s.send("stop")
			return Result{}, fmt.Errorf("search timed out after %s", searchTimeout)
		}
	}
}

// asks the engine to quit, then kills it if it doesn't exit in time
func (s *Session) Close() {
	close(s.stop)
	s.send("quit")
	s.stdin.Close()
	select {
	case <-s.done:
	case <-time.After(quitGrace):
		s.cmd.Process.Kill()
		<-s.done
	}
}

func (s *Session) send(command string) error {
	if _, err := io.WriteString(s.stdin, command+"\n"); err != nil {
		return fmt.Errorf("send %q to engine: %w", command, err)
	}
	return nil
}

func (s *Session) sync() error {
	if err := s.send("isready"); err != nil {
		return err
	}
	return s.waitFor("readyok", nil)
}

// consumes lines until one equals token
func (s *Session) waitFor(token string, onLine func(string)) error {
	timer := time.NewTimer(stepTimeout)
	defer timer.Stop()
	for {
		select {
		case line, ok := <-s.lines:
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
			return fmt.Errorf("timed out after %s waiting for %q", stepTimeout, token)
		}
	}
}

// forwards trimmed stdout lines
func (s *Session) readLines(r io.Reader) {
	defer close(s.lines)
	sc := bufio.NewScanner(r)

	if sc.Err() != nil {
		return
	}

	for sc.Scan() {
		select {
		case s.lines <- strings.TrimSpace(sc.Text()):
		case <-s.stop:
			return
		}
	}
}

// extracts the score from a main-line "info" line
func parseScore(line string) (Score, bool) {
	f := strings.Fields(line)
	if len(f) < 2 || f[0] != "info" || f[1] == "string" {
		return Score{}, false
	}
	var (
		sc Score
		ok bool
	)
loop:
	for i := 1; i < len(f); i++ {
		switch f[i] {
		case "multipv":
			if i+1 < len(f) && f[i+1] != "1" {
				return Score{}, false
			}
		case "lowerbound", "upperbound":
			return Score{}, false
		case "score":
			if i+2 >= len(f) {
				return Score{}, false
			}
			v, err := strconv.Atoi(f[i+2])
			if err != nil {
				return Score{}, false
			}
			switch f[i+1] {
			case "cp":
				sc, ok = Score{CP: v}, true
			case "mate":
				sc, ok = Score{Mate: v, IsMate: true}, true
			}
		case "pv":
			break loop
		}
	}
	return sc, ok
}
