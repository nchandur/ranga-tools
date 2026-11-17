package viriformat

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/bits"
	"os"
)

const (
	headerSize = 32
	wordSize   = 4
	bufSize    = 1 << 20
)

var (
	errTruncatedHeader = errors.New("truncated header")
	errTruncatedGame   = errors.New("truncated game")
)

// reads one record into buf
func readGame(r *bufio.Reader, buf []byte) ([]byte, error) {
	buf = buf[:0]

	var head [headerSize]byte
	if _, err := io.ReadFull(r, head[:]); err != nil {
		if err == io.EOF {
			return buf, io.EOF
		}
		if err == io.ErrUnexpectedEOF {
			return buf, errTruncatedHeader
		}
		return buf, err
	}
	buf = append(buf, head[:]...)

	var pair [wordSize]byte
	for {
		if _, err := io.ReadFull(r, pair[:]); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return buf, errTruncatedGame
			}
			return buf, err
		}
		buf = append(buf, pair[:]...)
		if pair == [wordSize]byte{} {
			return buf, nil
		}
	}
}

// verifies the framing of file
func Check(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	r := bufio.NewReaderSize(f, bufSize)
	var buf []byte
	games := 0
	for {
		var err error
		buf, err = readGame(r, buf)
		if err == io.EOF {
			return games, nil
		}
		if err != nil {
			return games, fmt.Errorf("%s: game %d: %w", path, games, err)
		}

		occ := binary.LittleEndian.Uint64(buf[:8])
		if bits.OnesCount64(occ) > 32 {
			return games, fmt.Errorf("%s: game %d: bad occupancy", path, games)
		}
		if wdl := buf[30]; wdl > 2 {
			return games, fmt.Errorf("%s: game %d: bad wdl %d", path, games, wdl)
		}
		games++
	}
}
