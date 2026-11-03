package board

import (
	"testing"
)

func TestBitBoard_GetBit(t *testing.T) {
	tests := []struct {
		name     string
		bb       BitBoard
		sq       Square
		expected BitBoard
	}{
		{
			name:     "empty bitboard",
			bb:       0x0,
			sq:       E4,
			expected: 0x0,
		},
		{
			name:     "piece on E4",
			bb:       0x1000000000,
			sq:       E4,
			expected: 0x1000000000,
		},
		{
			name:     "no piece on A1",
			bb:       0x1000000000,
			sq:       A1,
			expected: 0x0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := test.bb.GetBit(test.sq)

			if output != test.expected {
				t.Errorf("\nexpected: %064b\noutput:   %064b", test.expected, output)
			}

		})
	}
}

func TestBitBoard_SetBit(t *testing.T) {
	tests := []struct {
		name     string
		initial  BitBoard
		sq       Square
		expected BitBoard
	}{
		{
			name:     "set E4 on empty board",
			initial:  0x0,
			sq:       E4,
			expected: 0x1000000000,
		},
		{
			name:     "set E4 on non-empty board",
			initial:  0x8100000000000081,
			sq:       E4,
			expected: 0x8100001000000081,
		},
		{
			name:     "set E4 when already set",
			initial:  0x1000000000,
			sq:       E4,
			expected: 0x1000000000,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.initial.SetBit(test.sq)

			if test.initial != test.expected {
				t.Errorf("\nexpected: %064b\noutput:   %064b", test.expected, test.initial)
			}

		})
	}

}

func TestBitBoard_PopBit(t *testing.T) {
	tests := []struct {
		name     string
		initial  BitBoard
		sq       Square
		expected BitBoard
	}{
		{
			name:     "pop bit on empty",
			initial:  0x0,
			sq:       E4,
			expected: 0x0,
		},
		{
			name:     "pop bit on E4",
			initial:  0x1000000000,
			sq:       E4,
			expected: 0x0,
		},
		{
			name:     "pop bit on invalid square",
			initial:  0x1000000000,
			sq:       Square(65),
			expected: 0x1000000000,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.initial.PopBit(test.sq)

			if test.initial != test.expected {
				t.Errorf("\nexpected: %064b\noutput:   %064b", test.expected, test.initial)
			}

		})
	}
}

func TestBitBoard_CountBits(t *testing.T) {
	tests := []struct {
		name     string
		bb       BitBoard
		expected int
	}{
		{
			name:     "empty board",
			bb:       0x0,
			expected: 0,
		},
		{
			name:     "one bit set",
			bb:       0x1000000000,
			expected: 1,
		},
		{
			name:     "starting position",
			bb:       0xffff00000000ffff,
			expected: 32,
		},
		{
			name:     "kiwipete",
			bb:       0x91ffa41218737d91,
			expected: 32,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := test.bb.CountBits()

			if output != test.expected {
				t.Errorf("\nexpected: %d\noutput:  %d\n", test.expected, output)
			}

		})
	}

}

func TestBitBoard_GetLSB(t *testing.T) {
	tests := []struct {
		name     string
		bb       BitBoard
		expected int
	}{
		{
			name:     "empty board",
			bb:       0x0,
			expected: -1,
		},
		{
			name:     "bit set on E4",
			bb:       0x1000000000,
			expected: 36,
		},
		{
			name:     "starting position",
			bb:       0xffff00000000ffff,
			expected: 0,
		},
		{
			name: "kiwipete",
			bb: 0x91ffa41218737d91,
			expected: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := test.bb.GetLSB()

			if output != test.expected {
				t.Errorf("\nexpected: %d\noutput:  %d\n", test.expected, output)
			}

		})
	}

}
