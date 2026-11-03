package board

import (
	"math/rand/v2"
)

// precomputes attack lookup tables for leaper pieces
func initLeaperAttacks() {
	for square := range 64 {

		sq := Square(square)
		// precompute pawn attacks
		PawnAttacks[White][square] = MaskPawnAttacks(sq, White)
		PawnAttacks[Black][square] = MaskPawnAttacks(sq, Black)

		// precompute knight attacks
		KnightAttacks[square] = MaskKnightAttacks(sq)

		// precompute king attacks
		KingAttacks[square] = MaskKingAttacks(sq)
	}

}

// precomputes attack lookup tables for slider pieces
func initSliderAttacks(isBishop bool) {

	for sq := range 64 {
		BishopMasks[sq] = MaskBishopAttacks(Square(sq))
		RookMasks[sq] = MaskRookAttacks(Square(sq))

		attackMask := BitBoard(0)

		// set mask and fetch relevant occupancy mask for target piece
		if isBishop {
			attackMask = BishopMasks[sq]
		} else {
			attackMask = RookMasks[sq]
		}

		relevantBitCount := attackMask.CountBits()
		occupancyIndices := (1 << relevantBitCount)

		// populate magic lookup table for all blocker permutations
		for idx := range occupancyIndices {
			if isBishop {
				occupancy := SetOccupancy(idx, relevantBitCount, attackMask)
				magicIdx := (occupancy * BishopMagicNumbers[sq]) >> (64 - BishopRelevantOccupancy[sq])
				BishopAttacks[sq][magicIdx] = BishopAttackOTF(Square(sq), occupancy)
			} else {
				occupancy := SetOccupancy(idx, relevantBitCount, attackMask)
				magicIdx := (occupancy * RookMagicNumbers[sq]) >> (64 - RookRelevantOccupancy[sq])
				RookAttacks[sq][magicIdx] = RookAttackOTF(Square(sq), occupancy)
			}
		}

	}

}

// populates the Zobrist hashing tables with 64-bit pseudo-random numbers
func initHashKeys() {

	pcg := rand.NewPCG(SEED, 0)
	r := rand.New(pcg)

	// piece keys
	for pce := WP; pce <= BK; pce++ {
		for sq := range 64 {
			PieceKeys[pce][sq] = r.Uint64()
		}
	}

	// castle keys
	for pce := range 16 {
		CastleKeys[pce] = r.Uint64()
	}

	// enpassant keys
	for sq := range 64 {
		EnpassantKeys[sq] = r.Uint64()
	}

	// side keys
	SideKey = r.Uint64()

}

func init() {
	initLeaperAttacks()
	initSliderAttacks(true)
	initSliderAttacks(false)
	initHashKeys()
}
