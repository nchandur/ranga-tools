package pgn

import (
	"strings"

	"github.com/nchandur/ranga-tools/internal/board"
)

// converts a SAN move string into a valid Move
func ParseSAN(b *board.Board, san string) board.Move {
	san = strings.TrimRight(san, "+#!?")

	list := board.NewMoveList()
	list.GenerateMoves(b)

	if san == "O-O" || san == "0-0" {
		return findCastle(list, b.Side, true)
	}
	if san == "O-O-O" || san == "0-0-0" {
		return findCastle(list, b.Side, false)
	}

	var promo byte
	if i := strings.IndexByte(san, '='); i != -1 {
		promo = san[i+1]
		san = san[:i]
	} else if len(san) >= 2 {
		last := san[len(san)-1]
		if (last == 'Q' || last == 'R' || last == 'B' || last == 'N') &&
			san[0] >= 'a' && san[0] <= 'h' {
			promo = last
			san = san[:len(san)-1]
		}
	}

	pieceLetter := byte('P')
	rest := san
	if len(san) > 0 && san[0] >= 'A' && san[0] <= 'Z' {
		pieceLetter = san[0]
		rest = san[1:]
	}

	rest = strings.Replace(rest, "x", "", 1)

	if len(rest) < 2 {
		return board.NOMOVE
	}
	targetStr := rest[len(rest)-2:]
	disambig := rest[:len(rest)-2]

	if targetStr[0] < 'a' || targetStr[0] > 'h' || targetStr[1] < '1' || targetStr[1] > '8' {
		return board.NOMOVE
	}
	to := board.FRtoSq(int('8'-targetStr[1]), int(targetStr[0]-'a'))

	wantFile, wantRank := -1, -1
	for _, r := range disambig {
		switch {
		case r >= 'a' && r <= 'h':
			wantFile = int(r - 'a')
		case r >= '1' && r <= '8':
			wantRank = int('8' - r)
		}
	}

	wantPiece := pieceFromLetterAndSide(pieceLetter, b.Side)

	for n := range list.Count {
		m := list.Moves[n]

		if m.Target() != to {
			continue
		}

		// FIX: Use board Mailbox instead of non-existent m.Piece()
		if b.Mailbox[m.Source()] != wantPiece {
			continue
		}

		srcFile := int(m.Source() % 8)
		srcRank := int(m.Source() / 8)
		if wantFile != -1 && srcFile != wantFile {
			continue
		}
		if wantRank != -1 && srcRank != wantRank {
			continue
		}

		if promo != 0 {
			// FIX: Pass b.Side to m.Promoted()
			if !promoMatches(m.Promoted(b.Side), promo, b.Side) {
				continue
			}
		} else if m.Promoted(b.Side) != board.Empty {
			continue
		}

		copy := b.Preserve()
		if isLegal := copy.MakeMove(m, false); !isLegal {
			continue
		}

		return m
	}
	return board.NOMOVE
}

// locates the castling move for the side to move
func findCastle(list *board.MoveList, side board.Color, kingside bool) board.Move {
	var home board.Square
	if side == board.White {
		home = board.FRtoSq(7, 4)
	} else {
		home = board.FRtoSq(0, 4)
	}

	srcFile := int(home) % 8

	for n := range list.Count {
		m := list.Moves[n]

		// FIX: Utilize the IsCastle() API exposed in move.go
		if !m.IsCastle() || m.Source() != home {
			continue
		}
		tgtFile := int(m.Target()) % 8
		diff := tgtFile - srcFile

		if kingside && diff == 2 {
			return m
		}
		if !kingside && diff == -2 {
			return m
		}
	}

	return board.NOMOVE
}

func pieceFromLetterAndSide(letter byte, side board.Color) board.Piece {
	white := side == board.White
	switch letter {
	case 'N':
		if white {
			return board.WN
		}
		return board.BN
	case 'B':
		if white {
			return board.WB
		}
		return board.BB
	case 'R':
		if white {
			return board.WR
		}
		return board.BR
	case 'Q':
		if white {
			return board.WQ
		}
		return board.BQ
	case 'K':
		if white {
			return board.WK
		}
		return board.BK
	default:
		if white {
			return board.WP
		}
		return board.BP
	}
}

func promoMatches(p board.Piece, letter byte, side board.Color) bool {
	white := side == board.White
	switch letter {
	case 'Q':
		return p == board.WQ && white || p == board.BQ && !white
	case 'R':
		return p == board.WR && white || p == board.BR && !white
	case 'B':
		return p == board.WB && white || p == board.BB && !white
	case 'N':
		return p == board.WN && white || p == board.BN && !white
	}
	return false
}
