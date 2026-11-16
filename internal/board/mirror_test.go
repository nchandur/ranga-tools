package board

import "testing"

func TestBoard_Mirror(t *testing.T) {
	t.Run("double mirror restores exact original state", func(t *testing.T) {
		positions := []string{
			START,
			"rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 3 2",
			"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
			"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		}

		for _, fen := range positions {
			t.Run(fen, func(t *testing.T) {
				b := NewBoard()
				b.ParseFEN(fen)
				original := b.Preserve()
				b.Mirror()
				b.Mirror()

				assertBoardsEqual(t, &b, &original)
			})
		}
	})
	t.Run("mirrors piece placements and colors correctly", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("4k3/8/8/8/4P3/8/8/4K3 w - - 0 1")
		b.Mirror()

		if b.Mailbox[E5] != BP || b.PieceBitBoards[BP].GetBit(E5) == 0 {
			t.Errorf("expected Black Pawn at %v, got %v", E5, b.Mailbox[E5])
		}

		if b.Mailbox[E8] != BK || b.PieceBitBoards[BK].GetBit(E8) == 0 {
			t.Errorf("expected Black King at %v, got %v", E8, b.Mailbox[E8])
		}

		if b.Mailbox[E1] != WK || b.PieceBitBoards[WK].GetBit(E1) == 0 {
			t.Errorf("expected White King at %v, got %v", E1, b.Mailbox[E1])
		}

		if b.Side != Black {
			t.Errorf("Side = %v; want Black", b.Side)
		}
	})
	t.Run("mirrors castling, en passant, and preserves counters", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w Kq c6 5 12")

		b.Mirror()
		if b.Side != Black {
			t.Errorf("Side = %v; want Black", b.Side)
		}

		wantCastle := Castle(BKCA | WQCA)
		if b.Castle != wantCastle {
			t.Errorf("Castle = %v; want %v", b.Castle, wantCastle)
		}

		wantEP := C6 ^ 56
		if b.EnPassant != wantEP {
			t.Errorf("EnPassant = %v; want %v", b.EnPassant, wantEP)
		}

		if b.FiftyMove != 5 {
			t.Errorf("FiftyMove = %d; want 5", b.FiftyMove)
		}
		if b.Ply != 22 {
			t.Errorf("Ply = %d; want 22", b.Ply)
		}

		if b.Key == 0 {
			t.Errorf("expected non-zero Key generated")
		}
	})
}
