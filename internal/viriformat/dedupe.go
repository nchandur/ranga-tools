package viriformat

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

type DedupeStats struct {
	Kept, Dropped, Conflicts int
}

// copies every game from inputs to output, skipping any game whose start position and move sequence were already seen
func Dedupe(inputs []string, output string) (DedupeStats, error) {
	var st DedupeStats
	if len(inputs) == 0 {
		return st, fmt.Errorf("no input files")
	}

	outInfo, outErr := os.Stat(output)
	for _, p := range inputs {
		ii, err := os.Stat(p)
		if err != nil {
			return st, err
		}
		if outErr == nil && os.SameFile(outInfo, ii) {
			return st, fmt.Errorf("output %s is also an input; it would be truncated", output)
		}
	}

	out, err := os.Create(output)
	if err != nil {
		return st, err
	}
	w := bufio.NewWriterSize(out, bufSize)

	seen := make(map[[16]byte]byte)
	var raw, keyBuf []byte

	process := func(path string) error {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		r := bufio.NewReaderSize(f, bufSize)

		for {
			var err error
			raw, err = readGame(r, raw)
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}

			keyBuf = append(keyBuf[:0], raw[:28]...)
			for i := headerSize; i < len(raw)-wordSize; i += wordSize {
				keyBuf = append(keyBuf, raw[i], raw[i+1])
			}
			sum := sha256.Sum256(keyBuf)
			var key [16]byte
			copy(key[:], sum[:16])
			wdl := raw[30]

			if prev, dup := seen[key]; dup {
				st.Dropped++
				if prev != wdl {
					st.Conflicts++
				}
				continue
			}
			seen[key] = wdl

			if _, err := w.Write(raw); err != nil {
				return fmt.Errorf("write %s: %w", output, err)
			}
			st.Kept++
		}
	}

	for _, p := range inputs {
		if err := process(p); err != nil {
			out.Close()
			return st, err
		}
	}
	if err := w.Flush(); err != nil {
		out.Close()
		return st, fmt.Errorf("write %s: %w", output, err)
	}
	return st, out.Close()
}
