package board

// generates bitboard mask representing all potential pawn attack targets for given square and pawn color
func MaskPawnAttacks(square Square, color Color) BitBoard {

	var attack, bb BitBoard

	// set bit for current position
	bb.SetBit(square)

	switch color {
	case White:
		// white pawn capture targets
		if bb>>7&NotAFile != 0 {
			attack |= (bb >> 7)
		}
		if bb>>9&NotHFile != 0 {
			attack |= (bb >> 9)
		}

	case Black:
		// black pawn capture targets
		if bb<<7&NotHFile != 0 {
			attack |= (bb << 7)
		}
		if bb<<9&NotAFile != 0 {
			attack |= (bb << 9)
		}

	}

	return attack
}

// generates bitboard mask representing all legal attack targets for knight
func MaskKnightAttacks(square Square) BitBoard {

	var attack, bb BitBoard

	// set bit current square
	bb.SetBit(square)

	// up 2, left 1
	if bb>>17&NotHFile != 0 {
		attack |= bb >> 17
	}

	// up 2, right 1
	if bb>>15&NotAFile != 0 {
		attack |= bb >> 15
	}

	// up 1, left 2
	if bb>>10&NotGHFile != 0 {
		attack |= bb >> 10
	}

	// up 1, right 2
	if bb>>6&NotABFile != 0 {
		attack |= bb >> 6
	}

	// down 2, right 1
	if bb<<17&NotAFile != 0 {
		attack |= bb << 17
	}

	// down 2, left 1
	if bb<<15&NotHFile != 0 {
		attack |= bb << 15
	}

	// down 1, right 2
	if bb<<10&NotABFile != 0 {
		attack |= bb << 10
	}

	// down 1, left 2
	if bb<<6&NotGHFile != 0 {
		attack |= bb << 6
	}

	return attack

}

// generates a bitboard mask representing all legal attack targets for king
func MaskKingAttacks(square Square) BitBoard {

	var attack, bb BitBoard

	// set bit for current square
	bb.SetBit(square)

	// up
	if bb>>8 != 0 {
		attack |= bb >> 8
	}

	// up and left
	if bb>>9&NotHFile != 0 {
		attack |= bb >> 9
	}

	// up and right
	if bb>>7&NotAFile != 0 {
		attack |= bb >> 7
	}

	// left
	if bb>>1&NotHFile != 0 {
		attack |= bb >> 1
	}

	// down
	if bb<<8 != 0 {
		attack |= bb << 8
	}

	// down and right
	if bb<<9&NotAFile != 0 {
		attack |= bb << 9
	}

	// down and left
	if bb<<7&NotHFile != 0 {
		attack |= bb << 7
	}

	// right
	if bb<<1&NotAFile != 0 {
		attack |= bb << 1
	}

	return attack

}

// generates a bitboard mask representing relevant blocker/attack squares along diagonal rays for bishop
func MaskBishopAttacks(square Square) BitBoard {
	var attack BitBoard

	tr, tf := int(square)/8, int(square)%8

	// up and right
	for r, f := tr+1, tf+1; r <= 6 && f <= 6; r, f = r+1, f+1 {
		attack |= BitBoard(uint64(1) << uint64(FRtoSq(r, f)))
	}

	// down and right
	for r, f := tr-1, tf+1; r >= 1 && f <= 6; r, f = r-1, f+1 {
		attack |= BitBoard(uint64(1) << uint64(FRtoSq(r, f)))
	}

	// down and left
	for r, f := tr-1, tf-1; r >= 1 && f >= 1; r, f = r-1, f-1 {
		attack |= BitBoard(uint64(1) << uint64(FRtoSq(r, f)))
	}

	// up and left
	for r, f := tr+1, tf-1; r <= 6 && f >= 1; r, f = r+1, f-1 {
		attack |= BitBoard(uint64(1) << uint64(FRtoSq(r, f)))
	}

	return attack
}

// generates bitboard mask representing relevant blocker/attack squares orthogonal rays for rook
func MaskRookAttacks(square Square) BitBoard {
	var attack BitBoard

	tr, tf := int(square)/8, int(square)%8

	// up
	for r := tr + 1; r <= 6; r++ {
		attack |= BitBoard(uint64(1) << FRtoSq(r, tf))
	}

	// down
	for r := tr - 1; r >= 1; r-- {
		attack |= BitBoard(uint64(1) << FRtoSq(r, tf))
	}

	// right
	for f := tf + 1; f <= 6; f++ {
		attack |= BitBoard(uint64(1) << FRtoSq(tr, f))
	}

	// left
	for f := tf - 1; f >= 1; f-- {
		attack |= BitBoard(uint64(1) << FRtoSq(tr, f))
	}

	return attack
}

// generates bishop attack target squares on the fly for diagonal rays.
func BishopAttackOTF(square Square, block BitBoard) BitBoard {
	var attack BitBoard

	tr, tf := int(square)/8, int(square)%8

	// up and right
	for r, f := tr+1, tf+1; r <= 7 && f <= 7; r, f = r+1, f+1 {
		attack |= BitBoard(uint64(1) << uint64(FRtoSq(r, f)))
		if BitBoard((uint64(1)<<FRtoSq(r, f)))&block != 0 {
			break
		}
	}

	// down and right
	for r, f := tr-1, tf+1; r >= 0 && f <= 7; r, f = r-1, f+1 {
		attack |= BitBoard(uint64(1) << uint64(FRtoSq(r, f)))
		if BitBoard((uint64(1)<<FRtoSq(r, f)))&block != 0 {
			break
		}
	}

	// down and left
	for r, f := tr-1, tf-1; r >= 0 && f >= 0; r, f = r-1, f-1 {
		attack |= BitBoard(uint64(1) << uint64(FRtoSq(r, f)))
		if BitBoard((uint64(1)<<FRtoSq(r, f)))&block != 0 {
			break
		}

	}

	// up and left
	for r, f := tr+1, tf-1; r <= 7 && f >= 0; r, f = r+1, f-1 {
		attack |= BitBoard(uint64(1) << uint64(FRtoSq(r, f)))
		if BitBoard((uint64(1)<<FRtoSq(r, f)))&block != 0 {
			break
		}
	}

	return attack

}

// generates rook attack target squares on the fly for orthogonal rays
func RookAttackOTF(square Square, block BitBoard) BitBoard {
	var attack BitBoard

	tr, tf := int(square)/8, int(square)%8

	// up
	for r := tr + 1; r <= 7; r++ {
		attack |= BitBoard(uint64(1) << FRtoSq(r, tf))
		if BitBoard((uint64(1)<<FRtoSq(r, tf)))&block != 0 {
			break
		}
	}

	// down
	for r := tr - 1; r >= 0; r-- {
		attack |= BitBoard(uint64(1) << FRtoSq(r, tf))
		if BitBoard((uint64(1)<<FRtoSq(r, tf)))&block != 0 {
			break
		}
	}

	// right
	for f := tf + 1; f <= 7; f++ {
		attack |= BitBoard(uint64(1) << FRtoSq(tr, f))
		if BitBoard((uint64(1)<<FRtoSq(tr, f)))&block != 0 {
			break
		}
	}

	// left
	for f := tf - 1; f >= 0; f-- {
		attack |= BitBoard(uint64(1) << FRtoSq(tr, f))
		if BitBoard((uint64(1)<<FRtoSq(tr, f)))&block != 0 {
			break
		}
	}

	return attack
}

// calculates bishop attack bitboard for given square and occupancy.
func GetBishopAttacks(square Square, occupancy BitBoard) BitBoard {
	occupancy &= BishopMasks[square]
	occupancy *= BishopMagicNumbers[square]
	occupancy >>= 64 - BishopRelevantOccupancy[square]
	return BishopAttacks[square][occupancy]
}

// calculates rook attack bitboard for given square and occupancy
func GetRookAttacks(square Square, occupancy BitBoard) BitBoard {
	occupancy &= RookMasks[square]
	occupancy *= RookMagicNumbers[square]
	occupancy >>= 64 - RookRelevantOccupancy[square]
	return RookAttacks[square][occupancy]
}

// generates specific occupancy/blocker bitboard configuration for a given index
func SetOccupancy(idx, bitsInMask int, attackMask BitBoard) BitBoard {
	var occupancy BitBoard

	for count := range bitsInMask {
		square := attackMask.GetLSB()

		attackMask.PopBit(Square(square))

		if idx&(1<<count) != 0 {
			occupancy |= BitBoard(uint64(1) << uint64(square))
		}

	}
	return occupancy
}

// calculates queen attack bitboard for a given square and occupancy
func GetQueenAttacks(square Square, occupancy BitBoard) BitBoard {

	var result BitBoard

	bishopOccupancy := occupancy
	rookOccupancy := occupancy

	bishopOccupancy &= BishopMasks[square]
	bishopOccupancy *= BishopMagicNumbers[square]
	bishopOccupancy >>= (64 - BishopRelevantOccupancy[square])

	result = BishopAttacks[square][bishopOccupancy]

	rookOccupancy &= RookMasks[square]
	rookOccupancy *= RookMagicNumbers[square]
	rookOccupancy >>= (64 - RookRelevantOccupancy[square])

	result |= RookAttacks[square][rookOccupancy]

	return result
}

