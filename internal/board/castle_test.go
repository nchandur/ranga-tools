package board

import (
	"testing"
)

func TestCastle_Flip(t *testing.T) {
	tests := []struct {
		name     string
		input    Castle
		expected Castle
	}{
		{
			name:     "no castling rights",
			input:    0b0,
			expected: 0b0,
		},
		{
			name:     "white side castle rights",
			input:    0b1100,
			expected: 0b0011,
		},
		{
			name:     "black side castle rights",
			input:    0b0011,
			expected: 0b1100,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := test.input.flip()

			if output != test.expected {
				t.Errorf("expected: %b\noutput: %b\n", test.expected, output)
			}

		})
	}

}

func TestCastle_String(t *testing.T) {
	tests := []struct {
		name     string
		input    Castle
		expected string
	}{
		{
			name:     "all rights",
			input:    0b1111,
			expected: "KQkq",
		},
		{
			name:     "no sides",
			input:    0b0000,
			expected: "-",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := test.input.String()

			if output != test.expected {
				t.Errorf("expected: %s\noutput: %s\n", test.expected, output)
			}

		})
	}
}
