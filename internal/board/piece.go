package board

type Piece uint8

const (
	WP Piece = iota
	WN
	WB
	WR
	WQ
	WK
	BP
	BN
	BB
	BR
	BQ
	BK

	Empty
)

// flips color of piece
func (p Piece) flip() Piece {
	if p <= WK {
		return p + 6
	}
	return p - 6
}
