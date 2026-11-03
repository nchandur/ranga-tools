package board

import (
	"fmt"
	"math/bits"
	"strings"
)

type BitBoard uint64

// checks if bit corresponding to given square is set
func (b *BitBoard) GetBit(square Square) BitBoard {
	return (*b) & BitBoard(uint64(1)<<uint64(square))
}

// sets bit corresponding to specified square to 1
func (b *BitBoard) SetBit(square Square) {
	(*b) |= BitBoard(uint64(1) << square)
}

// clears bit corresponding to specified square if currently set
func (b *BitBoard) PopBit(square Square) {
	*b &^= BitBoard(uint64(1) << square)
}

// returns the number of set bits in bitboard
func (b BitBoard) CountBits() int {
	return bits.OnesCount64(uint64(b))
}

// returns the index of Least Significant Bit set to 1 in bitboard
func (b BitBoard) GetLSB() int {

	if b == 0 {
		return -1
	}

	return bits.TrailingZeros64(uint64(b))
}

func (b BitBoard) String() string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "\n")

	for rank := range 8 {
		for file := range 8 {
			square := Square((rank * 8) + file)

			bit := b.GetBit(square)

			if bit != 0 {
				fmt.Fprintf(&builder, " %d ", 1)
			} else {

				fmt.Fprintf(&builder, " %d ", 0)
			}

		}
		fmt.Fprintf(&builder, "\n")
	}
	fmt.Fprintf(&builder, "Value: 0x%x\n\n", uint64(b))
	return builder.String()
}
