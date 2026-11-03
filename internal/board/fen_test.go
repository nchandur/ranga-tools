package board

import (
	"strings"
	"testing"
)

func TestBoard_ParseFEN(t *testing.T) {
	t.Run("valid positions populate board fields correctly", func(t *testing.T) {
		tests := []struct {
			name          string
			fen           string
			wantSide      Color
			wantCastle    Castle
			wantEP        Square
			wantFiftyMove int
			wantPly       int
		}{
			{
				name:          "standard starting position",
				fen:           START,
				wantSide:      White,
				wantCastle:    WKCA | WQCA | BKCA | BQCA,
				wantEP:        NoSquare,
				wantFiftyMove: 0,
				wantPly:       0,
			},
			{
				name:          "black to move with en passant and partial castling",
				fen:           "rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR b Kq c6 3 2",
				wantSide:      Black,
				wantCastle:    WKCA | BQCA,
				wantEP:        C6,
				wantFiftyMove: 3,
				wantPly:       3,
			},
			{
				name:          "endgame without castling and high counters",
				fen:           "8/5k2/8/8/8/8/2K5/8 w - - 45 60",
				wantSide:      White,
				wantCastle:    0,
				wantEP:        NoSquare,
				wantFiftyMove: 45,
				wantPly:       118,
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				b := NewBoard()
				b.ParseFEN(test.fen)

				if b.Side != test.wantSide {
					t.Errorf("Side = %v; want %v", b.Side, test.wantSide)
				}
				if b.Castle != test.wantCastle {
					t.Errorf("Castle = %v; want %v", b.Castle, test.wantCastle)
				}
				if b.EnPassant != test.wantEP {
					t.Errorf("EnPassant = %v; want %v", b.EnPassant, test.wantEP)
				}
				if b.FiftyMove != test.wantFiftyMove {
					t.Errorf("FiftyMove = %d; want %d", b.FiftyMove, test.wantFiftyMove)
				}
				if b.Ply != test.wantPly {
					t.Errorf("Ply = %d; want %d", b.Ply, test.wantPly)
				}
				if b.Key == 0 {
					t.Errorf("expected Key to be generated, got 0")
				}
			})
		}
	})
	t.Run("piece bitboards and mailbox stay synchronized", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("4k3/8/8/8/4P3/8/8/4K3 w - - 0 1")
		expected := map[Square]Piece{
			E1: WK,
			E4: WP,
			E8: BK,
		}

		for sq, piece := range expected {
			if b.Mailbox[sq] != piece {
				t.Errorf("Mailbox[%v] = %v; want %v", sq, b.Mailbox[sq], piece)
			}
			if b.PieceBitBoards[piece].GetBit(sq) == 0 {
				t.Errorf("PieceBitBoards[%v] missing bit at %v", piece, sq)
			}
		}

		for sq := range 64 {
			s := Square(sq)
			piece := b.Mailbox[s]
			if _, ok := expected[s]; !ok && piece != Empty {
				t.Errorf("unexpected piece %v found at %v", piece, s)
			}

			for p := WP; p <= BK; p++ {
				bitSet := b.PieceBitBoards[p].GetBit(s)
				shouldBeSet := piece == p
				if (bitSet == 0 && shouldBeSet) || (bitSet != 0 && !shouldBeSet) {
					t.Errorf("bitboard/mailbox mismatch for piece %v at square %v: bitSet=%v, mailbox=%v",
						p, s, bitSet, piece)
				}
			}
		}
	})
	t.Run("invalid FEN formats return errors", func(t *testing.T) {
		tests := []struct {
			name      string
			fen       string
			errSubstr string
		}{
			{
				name:      "empty string",
				fen:       "",
				errSubstr: "string is empty",
			},
			{
				name:      "fewer than 6 fields",
				fen:       "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -",
				errSubstr: "expected 6 fields",
			},
			{
				name:      "more than 6 fields",
				fen:       "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1 extra",
				errSubstr: "expected 6 fields",
			},
			{
				name:      "invalid active color",
				fen:       "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR x KQkq - 0 1",
				errSubstr: "invalid color value",
			},
			{
				name:      "invalid castling permission character",
				fen:       "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQx - 0 1",
				errSubstr: "invalid castle permissions",
			},
			{
				name:      "invalid empty square digit (> 8)",
				fen:       "rnbqkbnr/pppppppp/9/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
				errSubstr: "invalid empty-square count",
			},
			{
				name:      "invalid empty square digit (0)",
				fen:       "rnbqkbnr/pppppppp/0/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
				errSubstr: "invalid empty-square count",
			},
			{
				name:      "invalid piece character",
				fen:       "rnbqkbnr/pppppppp/8/8/4X3/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
				errSubstr: "invalid piece character",
			},
			{
				name:      "en passant coordinate wrong length",
				fen:       "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq e33 0 1",
				errSubstr: "invalid en passant square",
			},
			{
				name:      "en passant coordinate invalid file",
				fen:       "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq i3 0 1",
				errSubstr: "invalid en passant square",
			},
			{
				name:      "en passant coordinate invalid rank",
				fen:       "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq e9 0 1",
				errSubstr: "invalid en passant square",
			},
			{
				name:      "negative fifty-move counter",
				fen:       "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - -1 1",
				errSubstr: "invalid halfmove clock",
			},
			{
				name:      "zero fullmove number",
				fen:       "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 0",
				errSubstr: "invalid fullmove number",
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				b := NewBoard()
				err := b.ParseFEN(test.fen)
				if !strings.Contains(err.Error(), test.errSubstr) {
					t.Errorf("error %q does not contain expected substring %q", err.Error(), test.errSubstr)
				}
			})
		}
	})
}

func TestBoard_FEN(t *testing.T) {
	t.Run("roundtrip matches original FEN", func(t *testing.T) {
		positions := []struct {
			name string
			fen  string
		}{
			{
				name: "starting position",
				fen:  START,
			},
			{
				name: "kiwipete",
				fen:  "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
			},
			{
				name: "endgame with pawns and kings only",
				fen:  "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
			},
			{
				name: "en passant active with partial castling rights",
				fen:  "rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR b Kq c6 2 2",
			},
			{
				name: "midgame position",
				fen:  "r1bqk2r/pp2bppp/2n1pn2/2pp4/2PP4/2N1PN2/PP2BPPP/R1BQK2R w KQkq - 4 7",
			},
		}

		for _, pos := range positions {
			t.Run(pos.name, func(t *testing.T) {
				b := NewBoard()
				b.ParseFEN(pos.fen)
				got := b.FEN()
				if got != pos.fen {
					t.Errorf("FEN() mismatch:\nwant: %s\ngot:  %s", pos.fen, got)
				}
			})
		}
	})
	t.Run("empty board serialization", func(t *testing.T) {
		b := &Board{}
		b.Clear()
		b.Side = White
		b.EnPassant = NoSquare
		b.Castle = 0
		b.FiftyMove = 0
		b.Ply = 0

		want := "8/8/8/8/8/8/8/8 w - - 0 1"
		got := b.FEN()
		if got != want {
			t.Errorf("FEN() for empty board mismatch:\nwant: %s\ngot:  %s", want, got)
		}
	})
}
