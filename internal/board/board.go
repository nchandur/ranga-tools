package board

import (
	"fmt"
)

// represents complete chess board state
type Board struct {
	PieceBitBoards [12]BitBoard // bitboards for each piece
	Occupancies    [3]BitBoard  // combined bitboards for white, black and all pieces
	Mailbox        [64]Piece
	Side           Color  // side to move
	EnPassant      Square // en passant square
	Castle                // castle rights
	Ply            int    // current half-move ply
	FiftyMove      int    // halfmove clock for 50-move draw rule
	Key            uint64 // unique hash key for position
}

// creates a new instance of board
func NewBoard() Board {
	var res Board

	res.PieceBitBoards = [12]BitBoard{}
	res.Occupancies = [3]BitBoard{}

	for i := range 64 {
		res.Mailbox[i] = Empty
	}

	res.Side = Both
	res.EnPassant = NoSquare
	res.Castle = 0
	res.Ply = 0
	res.FiftyMove = 0
	res.Key = 0

	return res
}

// deep copies board state
func (b *Board) Preserve() Board {
	var res Board

	res.PieceBitBoards = b.PieceBitBoards
	res.Occupancies = b.Occupancies
	res.Mailbox = b.Mailbox
	res.Side = b.Side
	res.EnPassant = b.EnPassant
	res.Castle = b.Castle
	res.Ply = b.Ply
	res.FiftyMove = b.FiftyMove
	res.Key = b.Key

	return res
}

// restores board state from copy
func (b *Board) Restore(copy *Board) {
	b.PieceBitBoards = copy.PieceBitBoards
	b.Occupancies = copy.Occupancies
	b.Mailbox = copy.Mailbox
	b.Side = copy.Side
	b.EnPassant = copy.EnPassant
	b.Castle = copy.Castle
	b.Ply = copy.Ply
	b.FiftyMove = copy.FiftyMove
	b.Key = copy.Key
}

func (b *Board) Clear() {
	b.PieceBitBoards = [12]BitBoard{}
	b.Occupancies = [3]BitBoard{}

	for i := range 64 {
		b.Mailbox[i] = Empty
	}
	b.Side = Both
	b.EnPassant = NoSquare
	b.Castle = 0
	b.Ply = 0
	b.FiftyMove = 0
	b.Key = 0
}

// add piece to the board
func (b *Board) AddPiece(piece Piece, sq Square) {
	if piece == Empty {
		return
	}

	b.PieceBitBoards[piece].SetBit(sq)

	if piece <= WK {
		b.Occupancies[White].SetBit(sq)
	} else if piece > WK && piece <= BK {
		b.Occupancies[Black].SetBit(sq)
	}

	b.Occupancies[Both].SetBit(sq)
	b.Mailbox[sq] = piece

}

// remove piece from the board
func (b *Board) RemovePiece(sq Square) {
	piece := b.Mailbox[sq]
	if piece == Empty {
		return
	}

	b.PieceBitBoards[piece].PopBit(sq)

	if piece <= WK {
		b.Occupancies[White].PopBit(sq)
	} else if piece > WK && piece <= BK {
		b.Occupancies[Black].PopBit(sq)
	}

	b.Occupancies[Both].PopBit(sq)
	b.Mailbox[sq] = Empty
}

// determines whether given square is under attack by any enemy piece
func (b *Board) IsSquareAttacked(square Square, color Color) bool {

	// checks pawn attacks
	if (color == White) && (PawnAttacks[Black][square]&b.PieceBitBoards[WP]) != 0 {
		return true
	}

	if (color == Black) && (PawnAttacks[White][square]&b.PieceBitBoards[BP]) != 0 {
		return true
	}

	// check knight attacks
	var enemyKnights BitBoard
	switch color {
	case White:
		enemyKnights = b.PieceBitBoards[WN]
	case Black:
		enemyKnights = b.PieceBitBoards[BN]
	}

	if (KnightAttacks[square] & enemyKnights) != 0 {
		return true
	}

	// check rook attacks
	var enemyRooks BitBoard
	switch color {
	case White:
		enemyRooks = b.PieceBitBoards[WR]
	case Black:
		enemyRooks = b.PieceBitBoards[BR]
	}

	if (GetRookAttacks(square, b.Occupancies[Both]) & enemyRooks) != 0 {
		return true
	}

	// check bishop attacks
	var enemyBishops BitBoard
	switch color {
	case White:
		enemyBishops = b.PieceBitBoards[WB]
	case Black:
		enemyBishops = b.PieceBitBoards[BB]
	}

	if (GetBishopAttacks(square, b.Occupancies[Both]) & enemyBishops) != 0 {
		return true
	}

	// check king attacks
	var enemyKing BitBoard
	switch color {
	case White:
		enemyKing = b.PieceBitBoards[WK]
	case Black:
		enemyKing = b.PieceBitBoards[BK]
	}

	if (KingAttacks[square] & enemyKing) != 0 {
		return true
	}

	// check queen attacks
	var enemyQueens BitBoard
	switch color {
	case White:
		enemyQueens = b.PieceBitBoards[WQ]
	case Black:
		enemyQueens = b.PieceBitBoards[BQ]
	}

	if (GetQueenAttacks(square, b.Occupancies[Both]) & enemyQueens) != 0 {
		return true
	}

	return false
}

// attempts to make a move on board and updates internal board state
func (b *Board) MakeMove(move Move, onlyCaptures bool) bool {
	// skips quiet moves during capture-only searches
	if onlyCaptures && !move.IsCapture() {
		return false
	}

	source, target := move.Source(), move.Target()
	piece := b.Mailbox[source]

	// preserves current board state
	stateCopy := b.Preserve()

	// updates 50-move rule counter (reset on pawn move or captures)
	b.FiftyMove++

	if move.IsCapture() || piece == WP || piece == BP {
		b.FiftyMove = 0
	}

	// remove piece on board and updates Zobrist hash key
	b.RemovePiece(source)
	b.Key ^= PieceKeys[piece][source]

	// removes target piece from opponent bitboard during regular captures
	if move.IsCapture() && !move.IsEnpass() {
		captured := b.Mailbox[target]
		b.Key ^= PieceKeys[captured][target]
		b.RemovePiece(target)
	}

	// add piece on board and updates hash key
	b.AddPiece(piece, target)
	b.Key ^= PieceKeys[piece][target]

	// handles pawn promotion (replaces pawn with promoted piece)
	if move.IsPromotion() {
		b.RemovePiece(target)
		b.Key ^= PieceKeys[piece][target]

		promoted := move.Promoted(b.Side)
		b.AddPiece(promoted, target)
		b.Key ^= PieceKeys[promoted][target]
	}

	// handle en passant capture (removes captured pawn behind target square)
	if move.IsEnpass() {
		if b.Side == White {
			b.RemovePiece(target + 8)
			b.Key ^= PieceKeys[BP][target+8]
		} else {
			b.RemovePiece(target - 8)
			b.Key ^= PieceKeys[WP][target-8]
		}
	}

	// clears previous en passant target square key
	if b.EnPassant != NoSquare {
		b.Key ^= EnpassantKeys[b.EnPassant]
	}
	b.EnPassant = NoSquare

	// sets en passant target square on double pawn push
	if move.IsDouble() {
		if b.Side == White {
			b.EnPassant = target + 8
			b.Key ^= EnpassantKeys[target+8]
		} else {
			b.EnPassant = target - 8
			b.Key ^= EnpassantKeys[target-8]
		}
	}

	// moves castling rook into position
	if move.IsCastle() {
		switch target {
		case G1:
			b.RemovePiece(H1)
			b.AddPiece(WR, F1)
			b.Key ^= PieceKeys[WR][H1] ^ PieceKeys[WR][F1]
		case C1:
			b.RemovePiece(A1)
			b.AddPiece(WR, D1)
			b.Key ^= PieceKeys[WR][A1] ^ PieceKeys[WR][D1]
		case G8:
			b.RemovePiece(H8)
			b.AddPiece(BR, F8)
			b.Key ^= PieceKeys[BR][H8] ^ PieceKeys[BR][F8]
		case C8:
			b.RemovePiece(A8)
			b.AddPiece(BR, D8)
			b.Key ^= PieceKeys[BR][A8] ^ PieceKeys[BR][D8]
		}
	}

	// updates castling rights permissions & Zobrist keys
	b.Key ^= CastleKeys[b.Castle]
	b.Castle &= CastleRights[source]
	b.Castle &= CastleRights[target]
	b.Key ^= CastleKeys[b.Castle]

	// updates overall board occupancy bitboard
	b.Occupancies[Both] = b.Occupancies[White] | b.Occupancies[Black]

	// switches side to move
	b.Side ^= 1
	b.Key ^= SideKey

	// verifies king safety
	var inCheck bool
	if b.Side == White {
		inCheck = b.IsSquareAttacked(Square(b.PieceBitBoards[BK].GetLSB()), b.Side)
	} else {
		inCheck = b.IsSquareAttacked(Square(b.PieceBitBoards[WK].GetLSB()), b.Side)
	}

	// revert illegal moves that leave king in check
	if inCheck {
		b.Restore(&stateCopy)
		return false
	}

	return true
}

// converts a UCI move string into a valid move
func (b *Board) ParseMove(move string) Move {

	if move[1] > '8' || move[1] < '1' {
		return NOMOVE
	}

	if move[3] > '8' || move[3] < '1' {
		return NOMOVE
	}

	if move[0] > 'h' || move[0] < 'a' {
		return NOMOVE
	}

	if move[2] > 'h' || move[2] < 'a' {
		return NOMOVE
	}

	from := FRtoSq(int('8'-move[1]), int(move[0]-'a'))
	to := FRtoSq(int('8'-move[3]), int(move[2]-'a'))

	list := NewMoveList()

	list.GenerateMoves(b)

	for n := range list.Count {
		m := list.Moves[n]

		if m.Source() == from && m.Target() == to {
			if m.IsPromotion() {
				if len(move) > 4 && move[4] == m.PromotionChar() {
					return m
				}
				continue
			}
			return m
		}

	}
	return NOMOVE
}

func (b *Board) Print() {
	fmt.Println("     a   b   c   d   e   f   g   h")
	fmt.Println("   +---+---+---+---+---+---+---+---+")

	for rank := 0; rank <= 7; rank++ {
		fmt.Printf(" %d |", 8-rank)

		for file := 0; file <= 7; file++ {
			sq := FRtoSq(rank, file)
			piece := b.Mailbox[sq]
			fmt.Printf(" %c |", PieceChar[piece])
		}

		fmt.Printf(" %d\n", 8-rank)
		fmt.Println("   +---+---+---+---+---+---+---+---+")
	}

	fmt.Println("     a   b   c   d   e   f   g   h")

	fmt.Printf("\nSide to Move : %c\n", b.Side.String())
	fmt.Printf("En Passant   : %s\n", b.EnPassant)
	fmt.Printf("Castling     : %s\n", b.Castle)
	fmt.Printf("Zobrist      : 0x%x\n", b.Key)
}
