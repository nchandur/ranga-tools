package viriformat

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/rand/v2"
	"os"
)

type ShuffleStats struct {
	Games int
	Bytes int64
}

type span struct {
	off int64
	n   int32
}

// rewrites input to output with its games in random order
func Shuffle(input, output string, seed int64, expectGames int) (ShuffleStats, error) {
	var st ShuffleStats

	if _, err := os.Stat(output); err == nil {
		return st, fmt.Errorf("refusing to overwrite %s", output)
	}

	in, err := os.Open(input)
	if err != nil {
		return st, err
	}
	defer in.Close()

	fi, err := in.Stat()
	if err != nil {
		return st, err
	}
	size := fi.Size()
	if size%wordSize != 0 {
		return st, errors.New("file size is not a multiple of 4: not the layout I assumed")
	}

	spans, err := scanSpans(in)
	if err != nil {
		return st, fmt.Errorf("%s: %w", input, err)
	}
	st = ShuffleStats{Games: len(spans), Bytes: size}

	if expectGames >= 0 && len(spans) != expectGames {
		return st, fmt.Errorf("expected %d games, got %d: layout assumption is wrong", expectGames, len(spans))
	}

	out, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return st, fmt.Errorf("refusing to overwrite %s", output)
		}
		return st, err
	}

	if err := writeShuffled(in, out, spans, seed); err != nil {
		out.Close()
		os.Remove(output)
		return st, err
	}
	if err := out.Close(); err != nil {
		return st, err
	}

	if oi, err := os.Stat(output); err != nil || oi.Size() != size {
		return st, errors.New("output size mismatch")
	}
	return st, nil
}

// records byte offset and length of every game.
func scanSpans(f *os.File) ([]span, error) {
	r := bufio.NewReaderSize(f, bufSize)
	var (
		spans []span
		off   int64
		buf   = make([]byte, 0, 4096)
	)
	for {
		var err error
		buf, err = readGame(r, buf)
		if err == io.EOF {
			return spans, nil
		}
		if err != nil {
			return nil, fmt.Errorf("game starting at byte %d: %w", off, err)
		}
		spans = append(spans, span{off: off, n: int32(len(buf))})
		off += int64(len(buf))
	}
}

func writeShuffled(in *os.File, out *os.File, spans []span, seed int64) error {
	perm := rand.New(rand.NewPCG(uint64(seed), 0)).Perm(len(spans))

	w := bufio.NewWriterSize(out, bufSize)
	buf := make([]byte, 0, 4096)

	for _, j := range perm {
		s := spans[j]
		if cap(buf) < int(s.n) {
			buf = make([]byte, s.n)
		}
		b := buf[:s.n]

		n, err := in.ReadAt(b, s.off)
		if n < len(b) {
			if err == nil || err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return fmt.Errorf("read game at byte %d: %w", s.off, err)
		}
		if _, err := w.Write(b); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}
	return w.Flush()
}
