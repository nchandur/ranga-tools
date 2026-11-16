package board

import "testing"

func assertBoardsEqual(t *testing.T, expected, want *Board) {
	t.Helper()

	if expected == nil || want == nil {
		if expected != want {
			t.Fatalf("Board nil mismatch: expected %v, want %v", expected, want)
		}
		return
	}

	for p := range 12 {
		if expected.PieceBitBoards[p] != want.PieceBitBoards[p] {
			t.Fatalf("PieceBitBoards[%d] mismatch:\n expected:  0x%016x\n want: 0x%016x",
				p, expected.PieceBitBoards[p], want.PieceBitBoards[p])
		}
	}

	occupancyNames := [3]string{"White", "Black", "Both"}
	for i := range 3 {
		if expected.Occupancies[i] != want.Occupancies[i] {
			t.Fatalf("Occupancies[%s] mismatch:\n expected:  0x%016x\n want: 0x%016x",
				occupancyNames[i], expected.Occupancies[i], want.Occupancies[i])
		}
	}

	for sq := range 64 {
		if expected.Mailbox[sq] != want.Mailbox[sq] {
			t.Fatalf("Mailbox mismatch at square %d:\n expected:  %v\n want: %v",
				sq, expected.Mailbox[sq], want.Mailbox[sq])
		}
	}

	if expected.Side != want.Side {
		t.Fatalf("Side mismatch: expected %v, want %v", expected.Side, want.Side)
	}

	if expected.EnPassant != want.EnPassant {
		t.Fatalf("EnPassant mismatch: expected %v, want %v", expected.EnPassant, want.EnPassant)
	}

	if expected.Castle != want.Castle {
		t.Fatalf("Castle mismatch: expected %v, want %v", expected.Castle, want.Castle)
	}

	if expected.Ply != want.Ply {
		t.Fatalf("Ply mismatch: expected %d, want %d", expected.Ply, want.Ply)
	}

	if expected.FiftyMove != want.FiftyMove {
		t.Fatalf("FiftyMove mismatch: expected %d, want %d", expected.FiftyMove, want.FiftyMove)
	}

	if expected.Key != want.Key {
		t.Fatalf("Key mismatch: expected 0x%016x, want 0x%016x", expected.Key, want.Key)
	}
}

func TestBoard_Preserve(t *testing.T) {

	original := NewBoard()
	original.ParseFEN("rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 3 2")

	preserved := original.Preserve()
	assertBoardsEqual(t, &preserved, &original)

	original.Side = Black
	original.EnPassant = NoSquare
	original.FiftyMove = 99
	original.Ply = 50
	original.Mailbox[0] = Empty
	original.PieceBitBoards[0] = 0
	original.Occupancies[0] = 0

	if preserved.Side != White {
		t.Errorf("preserved.Side mutated: got %v, want White", preserved.Side)
	}
	if preserved.EnPassant == NoSquare {
		t.Errorf("preserved.EnPassant mutated to NoSquare")
	}
	if preserved.FiftyMove != 3 {
		t.Errorf("preserved.FiftyMove mutated: got %d, want 3", preserved.FiftyMove)
	}
	if preserved.Ply != 2 {
		t.Errorf("preserved.Ply mutated: got %d, want 2", preserved.Ply)
	}
}

func TestBoard_Restore(t *testing.T) {
	current := NewBoard()
	current.ParseFEN(START)

	snapshot := NewBoard()
	snapshot.ParseFEN("rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR b Kq c6 3 2")

	current.Restore(&snapshot)
	assertBoardsEqual(t, &current, &snapshot)

	snapshot.Side = White
	snapshot.EnPassant = NoSquare
	snapshot.FiftyMove = 99
	snapshot.Ply = 100
	snapshot.Mailbox[0] = Empty

	if current.Side != Black {
		t.Errorf("current.Side mutated after snapshot change: got %v, want Black", current.Side)
	}
	if current.EnPassant == NoSquare {
		t.Errorf("current.EnPassant mutated after snapshot change")
	}
	if current.FiftyMove != 3 {
		t.Errorf("current.FiftyMove mutated: got %d, want 3", current.FiftyMove)
	}
	if current.Ply != 3 {
		t.Errorf("current.Ply mutated: got %d, want 3", current.Ply)
	}
}

func TestBoard_Clear(t *testing.T) {
	t.Run("clears a populated board", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 3 2")
		b.Clear()

		for i, bb := range b.PieceBitBoards {
			if bb != 0 {
				t.Errorf("PieceBitBoards[%d] = 0x%016x; want 0", i, bb)
			}
		}
		for i, occ := range b.Occupancies {
			if occ != 0 {
				t.Errorf("Occupancies[%d] = 0x%016x; want 0", i, occ)
			}
		}

		for sq := range 64 {
			if b.Mailbox[sq] != Empty {
				t.Errorf("Mailbox[%d] = %v; want Empty", sq, b.Mailbox[sq])
			}
		}

		if b.Side != Both {
			t.Errorf("Side = %v; want Both", b.Side)
		}
		if b.EnPassant != NoSquare {
			t.Errorf("EnPassant = %v; want NoSquare", b.EnPassant)
		}
		if b.Castle != 0 {
			t.Errorf("Castle = %v; want 0", b.Castle)
		}
		if b.Ply != 0 {
			t.Errorf("Ply = %d; want 0", b.Ply)
		}
		if b.FiftyMove != 0 {
			t.Errorf("FiftyMove = %d; want 0", b.FiftyMove)
		}
		if b.Key != 0 {
			t.Errorf("Key = 0x%016x; want 0", b.Key)
		}
	})
	t.Run("idempotent on empty board", func(t *testing.T) {
		b := NewBoard()
		snapshot := b.Preserve()

		b.Clear()
		assertBoardsEqual(t, &b, &snapshot)
	})
}

func TestBoard_AddPiece(t *testing.T) {
	t.Run("adds white piece correctly", func(t *testing.T) {
		b := NewBoard()

		target := E4
		b.AddPiece(WP, target)

		if b.PieceBitBoards[WP].GetBit(target) == 0 {
			t.Errorf("expected PieceBitBoards[WP] to have bit set at %v", target)
		}

		if b.Occupancies[White].GetBit(target) == 0 {
			t.Errorf("expected Occupancies[White] to have bit set at %v", target)
		}
		if b.Occupancies[Both].GetBit(target) == 0 {
			t.Errorf("expected Occupancies[Both] to have bit set at %v", target)
		}

		if b.Occupancies[Black].GetBit(target) != 0 {
			t.Errorf("expected Occupancies[Black] to NOT have bit set at %v", target)
		}

		if b.Mailbox[target] != WP {
			t.Errorf("expected Mailbox[%v] to be %v, got %v", target, WP, b.Mailbox[target])
		}
	})
	t.Run("adds black piece correctly", func(t *testing.T) {
		b := NewBoard()

		target := E5
		b.AddPiece(BQ, target)

		if b.PieceBitBoards[BQ].GetBit(target) == 0 {
			t.Errorf("expected PieceBitBoards[BQ] to have bit set at %v", target)
		}

		if b.Occupancies[Black].GetBit(target) == 0 {
			t.Errorf("expected Occupancies[Black] to have bit set at %v", target)
		}
		if b.Occupancies[Both].GetBit(target) == 0 {
			t.Errorf("expected Occupancies[Both] to have bit set at %v", target)
		}

		if b.Occupancies[White].GetBit(target) != 0 {
			t.Errorf("expected Occupancies[White] to NOT have bit set at %v", target)
		}

		if b.Mailbox[target] != BQ {
			t.Errorf("expected Mailbox[%v] to be %v, got %v", target, BQ, b.Mailbox[target])
		}
	})
	t.Run("empty piece does nothing", func(t *testing.T) {
		b := NewBoard()
		snapshot := b.Preserve()
		b.AddPiece(Empty, E4)
		assertBoardsEqual(t, &b, &snapshot)
	})
}

func TestBoard_RemovePiece(t *testing.T) {
	t.Run("removes white piece correctly", func(t *testing.T) {
		b := NewBoard()
		target := E4
		b.AddPiece(WP, target)

		b.RemovePiece(target)

		if b.PieceBitBoards[WP].GetBit(target) != 0 {
			t.Errorf("expected PieceBitBoards[WP] to be cleared at %v", target)
		}

		if b.Occupancies[White].GetBit(target) != 0 {
			t.Errorf("expected Occupancies[White] to be cleared at %v", target)
		}
		if b.Occupancies[Both].GetBit(target) != 0 {
			t.Errorf("expected Occupancies[Both] to be cleared at %v", target)
		}

		if b.Mailbox[target] != Empty {
			t.Errorf("expected Mailbox[%v] to be Empty, got %v", target, b.Mailbox[target])
		}
	})
	t.Run("removes black piece correctly", func(t *testing.T) {
		b := NewBoard()
		target := D5
		b.AddPiece(BQ, target)

		b.RemovePiece(target)

		if b.PieceBitBoards[BQ].GetBit(target) != 0 {
			t.Errorf("expected PieceBitBoards[BQ] to be cleared at %v", target)
		}

		if b.Occupancies[Black].GetBit(target) != 0 {
			t.Errorf("expected Occupancies[Black] to be cleared at %v", target)
		}
		if b.Occupancies[Both].GetBit(target) != 0 {
			t.Errorf("expected Occupancies[Both] to be cleared at %v", target)
		}

		if b.Mailbox[target] != Empty {
			t.Errorf("expected Mailbox[%v] to be Empty, got %v", target, b.Mailbox[target])
		}
	})
	t.Run("removing from empty square is a no-op", func(t *testing.T) {
		b := NewBoard()

		b.AddPiece(WN, B1)
		snapshot := b.Preserve()
		b.RemovePiece(E4)

		assertBoardsEqual(t, &b, &snapshot)
	})
	t.Run("removing one piece leaves other pieces intact", func(t *testing.T) {
		b := NewBoard()

		sq1 := E4
		sq2 := E5
		b.AddPiece(WP, sq1)
		b.AddPiece(BP, sq2)

		b.RemovePiece(sq1)

		if b.Mailbox[sq1] != Empty || b.PieceBitBoards[WP].GetBit(sq1) != 0 {
			t.Errorf("piece at %v was not properly removed", sq1)
		}

		if b.Mailbox[sq2] != BP || b.PieceBitBoards[BP].GetBit(sq2) == 0 {
			t.Errorf("piece at %v was unintentionally modified", sq2)
		}
		if b.Occupancies[Black].GetBit(sq2) == 0 || b.Occupancies[Both].GetBit(sq2) == 0 {
			t.Errorf("occupancies for %v were unintentionally cleared", sq2)
		}
	})
}

func TestBoard_IsSquareAttacked(t *testing.T) {
	t.Run("attacks by pawns", func(t *testing.T) {
		b := NewBoard()

		b.AddPiece(WP, E4)
		b.AddPiece(BP, D5)

		if !b.IsSquareAttacked(D5, White) {
			t.Error("expected: true\noutput: false\n")
		}

		if !b.IsSquareAttacked(E4, Black) {
			t.Error("expected: true\noutput: false\n")
		}

		if b.IsSquareAttacked(E5, White) {
			t.Error("expected: false\noutput: true\n")
		}

		if b.IsSquareAttacked(D4, Black) {
			t.Error("expected: true\noutput: false\n")
		}
	})
	t.Run("attacks by knights", func(t *testing.T) {
		b := NewBoard()

		b.AddPiece(WN, B1)
		b.AddPiece(BN, G8)

		if !b.IsSquareAttacked(F6, Black) {
			t.Error("expected: true\noutput: false\n")
		}
		if !b.IsSquareAttacked(H6, Black) {
			t.Error("expected: true\noutput: false\n")
		}
		if b.IsSquareAttacked(G6, Black) {
			t.Error("expected: false\noutput: true\n")
		}

		if !b.IsSquareAttacked(C3, White) {
			t.Error("expected: true\noutput: false\n")
		}
		if !b.IsSquareAttacked(A3, White) {
			t.Error("expected: true\noutput: false\n")
		}
		if b.IsSquareAttacked(B3, White) {
			t.Error("expected: false\noutput: true\n")
		}
	})
	t.Run("attacks by bishops", func(t *testing.T) {
		b := NewBoard()

		b.AddPiece(WB, C1)
		b.AddPiece(BB, C8)

		if !b.IsSquareAttacked(D2, White) {
			t.Error("expected: true\noutput: false\n")
		}
		if !b.IsSquareAttacked(H6, White) {
			t.Error("expected: true\noutput: false\n")
		}
		if !b.IsSquareAttacked(F4, White) {
			t.Error("expected: true\noutput: false\n")
		}

		if !b.IsSquareAttacked(D7, Black) {
			t.Error("expected: true\noutput: false\n")
		}
		if !b.IsSquareAttacked(F5, Black) {
			t.Error("expected: true\noutput: false\n")
		}
		if !b.IsSquareAttacked(H3, Black) {
			t.Error("expected: true\noutput: false")
		}

		if b.IsSquareAttacked(H1, Black) {
			t.Error("expected: false\noutput: true")
		}
		if b.IsSquareAttacked(A1, White) {
			t.Error("expected: false\noutput: true")
		}

	})
	t.Run("attacks by rooks", func(t *testing.T) {
		b := NewBoard()
		b.AddPiece(WR, A1)
		b.AddPiece(BR, H8)

		if !b.IsSquareAttacked(A8, White) {
			t.Error("expected: true\noutput: false\n")
		}
		if !b.IsSquareAttacked(A5, White) {
			t.Error("expected: true\noutput: false\n")
		}
		if !b.IsSquareAttacked(H5, Black) {
			t.Error("expected: true\noutput: false\n")
		}
		if !b.IsSquareAttacked(H1, Black) {
			t.Error("expected: true\noutput: false\n")
		}
		if b.IsSquareAttacked(E4, White) {
			t.Error("expected: false\noutput: true\n")
		}
		if b.IsSquareAttacked(D5, Black) {
			t.Error("expected: false\noutput: true\n")
		}
	})
	t.Run("attacks by queens", func(t *testing.T) {
		b := NewBoard()

		b.AddPiece(WQ, D1)
		b.AddPiece(BQ, D8)

		if !b.IsSquareAttacked(D8, White) {
			t.Error("expected: true\noutput: false\n")
		}
		if !b.IsSquareAttacked(D1, Black) {
			t.Error("expected: true\noutput: false\n")
		}
		if !b.IsSquareAttacked(G4, White) {
			t.Error("expected: true\noutput: false\n")
		}
		if !b.IsSquareAttacked(A5, Black) {
			t.Error("expected: true\noutput: false\n")
		}
		if b.IsSquareAttacked(E4, White) {
			t.Error("expected: false\noutput: true\n")
		}
		if b.IsSquareAttacked(E4, Black) {
			t.Error("expected: false\noutput: true\n")
		}
	})
	t.Run("attacks by kings", func(t *testing.T) {
		b := NewBoard()
		b.AddPiece(WK, E1)
		b.AddPiece(BK, E8)

		if !b.IsSquareAttacked(E2, White) {
			t.Error("expected: true\noutput: false\n")
		}
		if !b.IsSquareAttacked(D2, White) {
			t.Error("expected: true\noutput: false\n")
		}
		if !b.IsSquareAttacked(E7, Black) {
			t.Error("expected: true\noutput: false\n")
		}
		if !b.IsSquareAttacked(D7, Black) {
			t.Error("expected: true\noutput: false\n")
		}
		if b.IsSquareAttacked(E4, White) {
			t.Error("expected: false\noutput: true\n")
		}
		if b.IsSquareAttacked(E4, Black) {
			t.Error("expected: false\noutput: true\n")
		}
	})
}

func TestBoard_MakeMove(t *testing.T) {
	t.Run("quiet move executes when onlyCaptures is false", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN(START)
		move := NewMove(E2, E3, MoveQuiet)
		legal := b.MakeMove(move, false)

		if !legal {
			t.Fatalf("MakeMove rejected legal quiet move")
		}
		if b.Mailbox[E3] != WP || b.PieceBitBoards[WP].GetBit(E3) == 0 {
			t.Errorf("expected white pawn at e3")
		}
		if b.Mailbox[E2] != Empty || b.PieceBitBoards[WP].GetBit(E2) != 0 {
			t.Errorf("expected e2 to be empty")
		}
		if b.Side != Black {
			t.Errorf("Side = %v; want Black", b.Side)
		}
	})
	t.Run("quiet move rejected when onlyCaptures is true", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN(START)
		snapshot := b.Preserve()

		move := NewMove(E2, E3, MoveQuiet)
		legal := b.MakeMove(move, true)

		if legal {
			t.Fatalf("MakeMove executed quiet move with onlyCaptures = true")
		}
		assertBoardsEqual(t, &b, &snapshot)
	})
	t.Run("capture executes and resets fifty-move counter", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 5 3")

		move := NewMove(E4, D5, MoveCapture)
		legal := b.MakeMove(move, true)

		if !legal {
			t.Fatalf("MakeMove rejected capture with onlyCaptures = true")
		}
		if b.Mailbox[D5] != WP {
			t.Errorf("expected white pawn at d5, got %v", b.Mailbox[D5])
		}
		if b.PieceBitBoards[BP].GetBit(D5) != 0 {
			t.Errorf("captured black pawn still set on bitboard")
		}
		if b.FiftyMove != 0 {
			t.Errorf("FiftyMove = %d; want 0", b.FiftyMove)
		}
	})
	t.Run("pawn double push sets en passant square", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN(START)

		move := NewMove(E2, E4, MoveDouble)
		legal := b.MakeMove(move, false)

		if !legal {
			t.Fatalf("MakeMove rejected double push")
		}
		if b.EnPassant != E3 {
			t.Errorf("EnPassant = %v; want e3", b.EnPassant)
		}
	})
	t.Run("en passant capture clears victim pawn from original square", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 3")

		move := NewMove(E5, D6, MoveEnPass)
		legal := b.MakeMove(move, false)

		if !legal {
			t.Fatalf("MakeMove rejected en passant capture")
		}
		if b.Mailbox[D6] != WP || b.PieceBitBoards[WP].GetBit(D6) == 0 {
			t.Errorf("white pawn missing from target square d6")
		}
		if b.Mailbox[D5] != Empty || b.PieceBitBoards[BP].GetBit(D5) != 0 {
			t.Errorf("captured pawn on d5 was not removed")
		}
		if b.EnPassant != NoSquare {
			t.Errorf("EnPassant = %v; want NoSquare", b.EnPassant)
		}
	})
	t.Run("castling relocates king and rook and strips rights", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")

		move := NewMove(E1, G1, MoveCastle)
		legal := b.MakeMove(move, false)

		if !legal {
			t.Fatalf("MakeMove rejected castling")
		}
		if b.Mailbox[G1] != WK || b.Mailbox[E1] != Empty {
			t.Errorf("king placement incorrect after castling")
		}
		if b.Mailbox[F1] != WR || b.Mailbox[H1] != Empty {
			t.Errorf("rook placement incorrect after castling")
		}
		if b.PieceBitBoards[WR].GetBit(F1) == 0 || b.PieceBitBoards[WR].GetBit(H1) != 0 {
			t.Errorf("rook bitboard incorrect after castling")
		}
		if (b.Castle & (WKCA | WQCA)) != 0 {
			t.Errorf("white castling rights not removed: %v", b.Castle)
		}
	})
	t.Run("pawn promotion replaces pawn with target piece", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/4P3/8/8/8/8/8/4K2k w - - 0 1")

		move := NewMove(E7, E8, MovePromoQ)
		legal := b.MakeMove(move, false)

		if !legal {
			t.Fatalf("MakeMove rejected promotion")
		}
		if b.Mailbox[E8] != WQ || b.PieceBitBoards[WQ].GetBit(E8) == 0 {
			t.Errorf("expected promoted Queen at e8")
		}
		if b.PieceBitBoards[WP].GetBit(E8) != 0 {
			t.Errorf("pawn bitboard still set at promotion square e8")
		}
		if b.Mailbox[E7] != Empty {
			t.Errorf("source square e7 was not cleared")
		}
	})
}

func TestBoard_ParseMove(t *testing.T) {
	t.Run("move out of bounds", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN(START)
		moveStr := "a9b9"
		move := b.ParseMove(moveStr)
		if move != NOMOVE {
			t.Errorf("%s should be out of bounds", moveStr)
		}
	})
	t.Run("quiet move", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN(START)

		moveStr := "e2e4"

		move := b.ParseMove(moveStr)
		expected := NewMove(E2, E4, MoveDouble)

		if move != expected {
			t.Errorf("expected: %s\noutput: %s\n", expected, move)
		}

	})
	t.Run("captures", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1")

		moveStr := "f3f6"
		expected := NewMove(F3, F6, MoveCapture)
		move := b.ParseMove(moveStr)

		if move != expected {
			t.Errorf("expected: %s\noutput: %s\n", expected, move)
		}

	})
	t.Run("enpassant captures", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("4k3/4p3/8/3pP3/8/8/8/4K3 w - d6 0 1")

		moveStr := "e5d6"
		expected := NewMove(E5, D6, MoveEnPass)

		move := b.ParseMove(moveStr)

		if move != expected {
			t.Errorf("expected: %s\noutput: %s\n", expected, move)
		}

	})
	t.Run("promotions", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/3P1k2/4p3/8/8/8/8/4K3 w - - 0 1")
		moveStr := "d7d8q"
		expected := NewMove(D7, D8, MovePromoQ)

		move := b.ParseMove(moveStr)

		if move != expected {
			b.Print()
			t.Errorf("expected: %s\noutput: %s\n", expected, move)
		}

	})
	t.Run("invalid move", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN(START)

		moveStr := "a1a8"

		move := b.ParseMove(moveStr)

		if move != NOMOVE {
			t.Errorf("expected: NOMOVE\noutput: %s", move)
		}

	})
}
