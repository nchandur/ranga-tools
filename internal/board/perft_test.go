package board

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

func TestPerft_EPDSuite(t *testing.T) {
	epdPath := filepath.Join("testdata", "perftsuite.epd")
	file, err := os.Open(epdPath)
	if err != nil {
		t.Skipf("EPD test file not found at %s; skipping EPD suite", epdPath)
	}
	defer file.Close()

	maxNodes := int64(100000)
	if testing.Short() {
		maxNodes = 20000
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0
	stop := new(atomic.Bool)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lineNum++

		fields := strings.Split(line, ";")
		fen := strings.TrimSpace(fields[0])

		for _, field := range fields[1:] {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}

			parts := strings.Split(field, " ")
			if len(parts) < 2 {
				continue
			}

			depthStr := strings.TrimPrefix(parts[0], "D")
			depth, err := strconv.Atoi(depthStr)
			if err != nil {
				continue
			}

			wantNodes, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
			if err != nil {
				continue
			}

			if wantNodes > maxNodes {
				continue
			}

			t.Run(fen+"_D"+strconv.Itoa(depth), func(t *testing.T) {
				b := &Board{}
				if err := b.ParseFEN(fen); err != nil {
					t.Fatalf("line %d: ParseFEN failed: %v", lineNum, err)
				}

				gotNodes := Perft(b, depth, stop)
				if gotNodes != wantNodes {
					t.Errorf("line %d: depth %d mismatch:\nFEN:  %s\nGot:  %d nodes\nWant: %d nodes",
						lineNum, depth, fen, gotNodes, wantNodes)
				}
			})
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error reading EPD: %v", err)
	}
}
