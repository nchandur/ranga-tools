package board

// zobrist hashing
var PieceKeys [12][64]uint64
var CastleKeys [16]uint64
var EnpassantKeys [64]uint64
var SideKey uint64

// computes 64-bit hash key representing board state
func (b *Board) GenerateKey() uint64 {

	var res uint64

	var temp BitBoard

	// hash pieces
	for pce := WP; pce <= BK; pce++ {
		temp = b.PieceBitBoards[pce]

		for temp != 0 {
			sq := temp.GetLSB()
			res ^= PieceKeys[pce][sq]
			temp.PopBit(Square(sq))
		}
	}

	// hash enpassant square
	if b.EnPassant != NoSquare {
		res ^= EnpassantKeys[b.EnPassant]
	}

	// hash castling
	res ^= CastleKeys[b.Castle]

	// hash side
	res ^= SideKey

	return res
}
