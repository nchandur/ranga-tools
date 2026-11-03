package board

import (
	"fmt"
	"strings"
)

// represents generated legal or pseudo-legal moves
type MoveList struct {
	Moves [MAX_MOVES]Move
	Count int
}

// initializes empty move list
func NewMoveList() *MoveList {
	return &MoveList{
		Moves: [MAX_MOVES]Move{},
		Count: 0,
	}
}

// appends move to move list and increments the move counter
func (m *MoveList) AddMove(move Move) {
	m.Moves[m.Count] = move
	m.Count++
}

func (m MoveList) String() string {
	if m.Count == 0 {
		return "\nNo moves in list\n"
	}

	var b strings.Builder
	const columns = 4

	fmt.Fprintln(&b)

	for idx := 0; idx < m.Count; idx++ {
		moveStr := fmt.Sprintf("#%-2d %-5s", idx+1, m.Moves[idx])
		fmt.Fprintf(&b, "  %-10s", moveStr)

		if (idx+1)%columns == 0 && idx != m.Count-1 {
			fmt.Fprintln(&b)
		}
	}

	fmt.Fprintf(&b, "\n\n  Total: %-3d\n", m.Count)

	return b.String()
}
