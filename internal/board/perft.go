package board

import (
	"fmt"
	"sync/atomic"
)

func Perft(b *Board, depth int, stop *atomic.Bool) int64 {

	if stop.Load() {
		return 0
	}

	if depth == 0 {
		return 1
	}

	var nodes int64 = 0
	var list MoveList

	list.GenerateMoves(b)

	for i := range list.Count {
		move := list.Moves[i]

		copy := b.Preserve()

		if !b.MakeMove(move, false) {
			b.Restore(&copy)
			continue
		}

		count := Perft(b, depth-1, stop)
		nodes += count

		b.Restore(&copy)

		if depth > 2 && stop.Load() {
			return 0
		}

	}

	return nodes
}

func PerftDivide(b *Board, depth int, stop *atomic.Bool) int64 {
	if depth == 0 {
		return 1
	}

	var totalNodes int64 = 0
	var list MoveList
	list.GenerateMoves(b)

	for i := range list.Count {
		move := list.Moves[i]
		copy := b.Preserve()

		if !b.MakeMove(move, false) {
			b.Restore(&copy)
			continue
		}

		nodes := Perft(b, depth-1, stop)
		totalNodes += nodes

		b.Restore(&copy)

		fmt.Println(move, ":", nodes)
	}

	fmt.Println("\nNodes Searched:", totalNodes)
	return totalNodes
}
