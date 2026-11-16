package board

import "testing"

func TestFRtoSq(t *testing.T) {
	tests := []struct {
		name     string
		file     int
		rank     int
		expected Square
	}{
		{
			name:     "E4",
			file:     4,
			rank:     4,
			expected: E4,
		},
		{
			name:     "A1",
			file:     0,
			rank:     7,
			expected: A1,
		},
		{
			name:     "A8",
			file:     0,
			rank:     0,
			expected: A8,
		},
		{
			name:     "H1",
			file:     7,
			rank:     7,
			expected: H1,
		},
		{
			name:     "H8",
			file:     7,
			rank:     0,
			expected: H8,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sq := FRtoSq(test.rank, test.file)

			if sq != test.expected {
				t.Errorf("\nexpected: %s\noutput:   %s\n", test.expected, sq)
			}

		})
	}

}

func TestSqToFR(t *testing.T) {
	tests := []struct {
		name string
		sq   Square
		file int
		rank int
	}{
		{
			name: "E4",
			sq:   E4,
			file: 4,
			rank: 4,
		},
		{
			name: "A1",
			sq:   A1,
			file: 0,
			rank: 7,
		},
		{
			name: "A8",
			sq:   A8,
			file: 0,
			rank: 0,
		},
		{
			name: "H1",
			sq:   H1,
			file: 7,
			rank: 7,
		},
		{
			name: "H8",
			sq:   H8,
			file: 7,
			rank: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r, f := SqToFR(test.sq)

			if r != test.rank {
				t.Errorf("\nrank expected: %d\nrank output:   %d\n", test.rank, r)
			}

			if f != test.file {
				t.Errorf("\nfile expected: %d\nfile output:   %d\n", test.file, f)
			}

		})
	}

}
