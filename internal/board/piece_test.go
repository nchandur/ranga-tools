package board

import (
	"testing"
)

func TestPiece_Fli(t *testing.T) {
	tests := []struct {
		name     string
		input    Piece
		expected Piece
	}{
		{
			name:     "WP -> BP",
			input:    WP,
			expected: BP,
		},
		{
			name:     "BK -> WK",
			input:    BK,
			expected: WK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := test.input.flip()

			if output != test.expected {
				t.Errorf("expected: %d\noutput: %d", test.expected, output)
			}

		})
	}

}
