package board

import "testing"

func TestColor_Strin(t *testing.T) {
	tests := []struct {
		name     string
		input    Color
		expected byte
	}{
		{
			name:     "White",
			input:    White,
			expected: 'w',
		},
		{
			name:     "Black",
			input:    Black,
			expected: 'b',
		},
		{
			name:     "Both",
			input:    Both,
			expected: '-',
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := test.input.String()

			if output != test.expected {
				t.Errorf("expected: %c\noutput: %c\n", test.expected, output)
			}
		})
	}

}
