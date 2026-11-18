package pgn

import (
	"strings"

	"github.com/nchandur/ranga-tools/internal/board"
)

var pieces = map[byte][2]board.Piece{
	'P': {board.WP, board.BP}, 'N': {board.WN, board.BN}, 'B': {board.WB, board.BB},
	'R': {board.WR, board.BR}, 'Q': {board.WQ, board.BQ}, 'K': {board.WK, board.BK},
}

func pieceFor(letter byte, side board.Color) board.Piece {
	p, ok := pieces[letter]
	if !ok {
		return board.Empty
	}
	if side == board.White {
		return p[0]
	}
	return p[1]
}

// converts a SAN move into the legal move
func ParseSAN(b *board.Board, san string) board.Move {
	san = strings.TrimRight(san, "+#!?")
	if san == "" {
		return board.NOMOVE
	}

	list := board.NewMoveList()
	list.GenerateMoves(b)

	switch san {
	case "O-O", "0-0":
		return findCastle(list, b.Side, 6)
	case "O-O-O", "0-0-0":
		return findCastle(list, b.Side, 2)
	}

	var promo byte
	if s, p, ok := strings.Cut(san, "="); ok {
		san = s
		if p != "" {
			promo = p[0]
		}
	} else if n := len(san); n >= 2 && san[0] >= 'a' && san[0] <= 'h' && strings.IndexByte("QRBN", san[n-1]) >= 0 {
		promo, san = san[n-1], san[:n-1]
	}

	piece := byte('P')
	if san != "" && san[0] >= 'A' && san[0] <= 'Z' {
		piece, san = san[0], san[1:]
	}
	san = strings.Replace(san, "x", "", 1)
	if len(san) < 2 {
		return board.NOMOVE
	}
	target, disambig := san[len(san)-2:], san[:len(san)-2]
	if target[0] < 'a' || target[0] > 'h' || target[1] < '1' || target[1] > '8' {
		return board.NOMOVE
	}
	to := board.FRtoSq(int('8'-target[1]), int(target[0]-'a'))

	wantFile, wantRank := -1, -1
	for _, r := range disambig {
		switch {
		case r >= 'a' && r <= 'h':
			wantFile = int(r - 'a')
		case r >= '1' && r <= '8':
			wantRank = int('8' - r)
		}
	}

	wantPiece := pieceFor(piece, b.Side)
	wantPromo := board.Empty
	if promo != 0 {
		wantPromo = pieceFor(promo, b.Side)
	}

	for n := range list.Count {
		m := list.Moves[n]
		src := int(m.Source())
		if m.Target() != to || b.Mailbox[m.Source()] != wantPiece || m.Promoted(b.Side) != wantPromo {
			continue
		}
		if (wantFile >= 0 && src%8 != wantFile) || (wantRank >= 0 && src/8 != wantRank) {
			continue
		}
		c := b.Preserve() // generated moves are pseudo-legal: test on a copy
		if c.MakeMove(m, false) {
			return m
		}
	}
	return board.NOMOVE
}

// returns the castling move
func findCastle(list *board.MoveList, side board.Color, targetFile int) board.Move {
	home := board.FRtoSq(0, 4)
	if side == board.White {
		home = board.FRtoSq(7, 4)
	}
	for n := range list.Count {
		m := list.Moves[n]
		if m.IsCastle() && m.Source() == home && int(m.Target())%8 == targetFile {
			return m
		}
	}
	return board.NOMOVE
}
