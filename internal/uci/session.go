package uci

import (
	"fmt"
	"io"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// bounds a single fixed-depth search
const SearchTimeout = 10 * time.Minute

// an engine evaluation from the point of view of the side to move
type Score struct {
	CP     int
	Mate   int
	IsMate bool
}

// the outcome of one search.
type Result struct {
	Score    Score
	Depth    int
	BestMove string
	PV       []string
}

// running engine that can be searched repeatedly
type Session struct {
	Info Info

	cmd   *exec.Cmd
	stdin io.WriteCloser
	lines <-chan string
	done  chan error
	stop  chan struct{}
	once  sync.Once
}

// launches the engine, performs the UCI handshake, applies options and waits until the engine reports it is ready
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
		stop:  make(chan struct{}),
		done:  make(chan error, 1),
	}
	s.lines = readLines(stdout, s.stop)
	go func() { s.done <- cmd.Wait() }()

	if err := s.handshake(options); err != nil {
		s.Close()
		return nil, err
	}
	return s, nil
}

func (s *Session) handshake(options map[string]string) error {
	if err := send(s.stdin, "uci"); err != nil {
		return err
	}
	err := waitFor(s.lines, "uciok", func(line string) {
		switch {
		case strings.HasPrefix(line, "id name "):
			s.Info.Name = strings.TrimSpace(strings.TrimPrefix(line, "id name "))
		case strings.HasPrefix(line, "id author "):
			s.Info.Author = strings.TrimSpace(strings.TrimPrefix(line, "id author "))
		}
	})
	if err != nil {
		return err
	}

	names := make([]string, 0, len(options))
	for k := range options {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		if err := send(s.stdin, fmt.Sprintf("setoption name %s value %s", k, options[k])); err != nil {
			return err
		}
	}
	return s.sync()
}

// sends isready and waits for readyok.
func (s *Session) sync() error {
	if err := send(s.stdin, "isready"); err != nil {
		return err
	}
	return waitFor(s.lines, "readyok", nil)
}

// tells the engine new game starts
func (s *Session) NewGame() error {
	if err := send(s.stdin, "ucinewgame"); err != nil {
		return err
	}
	return s.sync()
}

// analyzes the position reached from startFEN after playing moves to fixed depth
func (s *Session) Search(startFEN string, moves []string, depth int) (Result, error) {
	pos := "position fen " + startFEN
	if len(moves) > 0 {
		pos += " moves " + strings.Join(moves, " ")
	}
	if err := send(s.stdin, pos); err != nil {
		return Result{}, err
	}
	if err := send(s.stdin, "go depth "+strconv.Itoa(depth)); err != nil {
		return Result{}, err
	}

	var res Result
	haveScore := false
	timer := time.NewTimer(SearchTimeout)
	defer timer.Stop()

	for {
		select {
		case line, ok := <-s.lines:
			if !ok {
				return Result{}, fmt.Errorf("engine exited during search")
			}
			switch {
			case strings.HasPrefix(line, "info "):
				if parseInfo(line, &res) {
					haveScore = true
				}
			case strings.HasPrefix(line, "bestmove "):
				f := strings.Fields(line)
				if len(f) < 2 || f[1] == "(none)" || f[1] == "0000" {
					return Result{}, fmt.Errorf("engine has no move in position %q", pos)
				}
				res.BestMove = f[1]
				if !haveScore {
					return Result{}, fmt.Errorf("engine sent no score for %q", pos)
				}
				return res, nil
			}
		case <-timer.C:
			_ = send(s.stdin, "stop")
			return Result{}, fmt.Errorf("search timed out after %s", SearchTimeout)
		}
	}
}

// shuts the engine down
func (s *Session) Close() {
	s.once.Do(func() {
		close(s.stop)
		shutdown(s.cmd, s.stdin, s.done)
	})
}

// reports whether it carried usable score
func parseInfo(line string, res *Result) bool {
	f := strings.Fields(line)
	if len(f) < 2 || f[1] == "string" {
		return false
	}

	var (
		score    Score
		hasScore bool
		depth    int
		pv       []string
	)
	for i := 1; i < len(f); i++ {
		switch f[i] {
		case "depth":
			if i+1 < len(f) {
				depth, _ = strconv.Atoi(f[i+1])
			}
		case "multipv":
			if i+1 < len(f) && f[i+1] != "1" {
				return false
			}
		case "lowerbound", "upperbound":
			return false
		case "score":
			if i+2 >= len(f) {
				return false
			}
			v, err := strconv.Atoi(f[i+2])
			if err != nil {
				return false
			}
			switch f[i+1] {
			case "cp":
				score, hasScore = Score{CP: v}, true
			case "mate":
				score, hasScore = Score{Mate: v, IsMate: true}, true
			}
		case "pv":
			pv = f[i+1:]
			i = len(f) // pv is always last
		}
	}

	if !hasScore {
		return false
	}
	res.Score, res.PV = score, pv
	if depth > 0 {
		res.Depth = depth
	}
	return true
}
