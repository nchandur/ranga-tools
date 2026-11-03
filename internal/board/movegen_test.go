package board

import "testing"

func containsMove(ml *MoveList, m Move) bool {
	for i := range ml.Count {
		if ml.Moves[i] == m {
			return true
		}
	}
	return false
}

func TestMoveList_GeneratePawnMoves(t *testing.T) {
	t.Run("white pawn pushes and double pushes", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/8/8/8/8/3P4/4P3/8 w - - 0 1")

		var list MoveList
		list.generatePawnMoves(&b, false)

		wantMoves := []Move{
			NewMove(E2, E3, MoveQuiet),
			NewMove(E2, E4, MoveDouble),
			NewMove(D3, D4, MoveQuiet),
		}

		if list.Count != len(wantMoves) {
			t.Fatalf("Move count = %d; want %d", list.Count, len(wantMoves))
		}
		for _, m := range wantMoves {
			if !containsMove(&list, m) {
				t.Errorf("missing expected pawn push: %v", m)
			}
		}
	})
	t.Run("black pawn pushes and double pushes", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/3p4/4p3/8/8/8/8/8 b - - 0 1")

		var list MoveList
		list.generatePawnMoves(&b, false)

		wantMoves := []Move{
			NewMove(D7, D6, MoveQuiet),
			NewMove(D7, D5, MoveDouble),
			NewMove(E6, E5, MoveQuiet),
		}

		if list.Count != len(wantMoves) {
			t.Fatalf("Move count = %d; want %d", list.Count, len(wantMoves))
		}
		for _, m := range wantMoves {
			if !containsMove(&list, m) {
				t.Errorf("missing expected black pawn push: %v", m)
			}
		}
	})
	t.Run("blocked pawns generate no quiet pushes", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/8/8/8/3r4/4p3/3P1P2/8 w - - 0 1")

		var list MoveList
		list.generatePawnMoves(&b, false)

		if containsMove(&list, NewMove(E2, E3, MoveQuiet)) {
			t.Errorf("blocked pawn at e2 generated push to e3")
		}
		if containsMove(&list, NewMove(E2, E4, MoveDouble)) {
			t.Errorf("blocked pawn at e2 generated double push to e4")
		}
	})
	t.Run("white pawn promotions (quiet and capture)", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("3r4/4P3/8/8/8/8/8/8 w - - 0 1")

		var list MoveList
		list.generatePawnMoves(&b, false)

		// 4 quiet promotions to e8 + 4 capture promotions to d8 = 8 moves
		promoPieces := []Piece{WQ, WR, WB, WN}
		for _, piece := range promoPieces {
			quietPromo := NewMove(E7, E8, MovePromoQ)
			capturePromo := NewMove(E7, D8, MovePromoCaptureQ)

			if !containsMove(&list, quietPromo) {
				t.Errorf("missing quiet promotion to %v", piece)
			}
			if !containsMove(&list, capturePromo) {
				t.Errorf("missing capture promotion to %v", piece)
			}
		}
		if list.Count != 8 {
			t.Errorf("expected 8 promotion moves, got %d", list.Count)
		}
	})
	t.Run("black pawn promotions (quiet and capture)", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/8/8/8/8/8/3p4/4N3 b - - 0 1")

		var list MoveList
		list.generatePawnMoves(&b, false)

		promoPieces := []Piece{BQ, BR, BB, BN}
		for _, piece := range promoPieces {
			quietPromo := NewMove(D2, D1, MovePromoQ)
			capturePromo := NewMove(D2, E1, MovePromoCaptureQ)

			if !containsMove(&list, quietPromo) {
				t.Errorf("missing quiet promotion to %v", piece)
			}
			if !containsMove(&list, capturePromo) {
				t.Errorf("missing capture promotion to %v", piece)
			}
		}
		if list.Count != 8 {
			t.Errorf("expected 8 promotion moves, got %d", list.Count)
		}
	})
	t.Run("en passant capture generation", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/8/8/2P1P3/8/8/8/8 w - d6 0 1")

		var list MoveList
		list.generatePawnMoves(&b, false)

		epMoveRight := NewMove(C5, D6, MoveEnPass)
		epMoveLeft := NewMove(E5, D6, MoveEnPass)

		if !containsMove(&list, epMoveRight) {
			t.Errorf("missing en passant move c5xd6")
		}
		if !containsMove(&list, epMoveLeft) {
			t.Errorf("missing en passant move e5xd6")
		}
	})
	t.Run("capturesOnly ignores quiet moves", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("3r4/4P3/8/3p4/4P3/8/4P3/8 w - - 0 1")

		var list MoveList
		list.generatePawnMoves(&b, true)

		if containsMove(&list, NewMove(E2, E3, MoveQuiet)) {
			t.Errorf("quiet push included when capturesOnly = true")
		}
		if containsMove(&list, NewMove(E7, E8, MoveQuiet)) {
			t.Errorf("quiet promotion included when capturesOnly = true")
		}

		if !containsMove(&list, NewMove(E4, D5, MoveCapture)) {
			t.Errorf("expected capture e4xd5 missing")
		}

		if !containsMove(&list, NewMove(E7, D8, MovePromoCaptureQ)) {
			t.Errorf("expected capture promotion e7xd8 missing")
		}
	})
}

func TestMoveList_GenerateKnightMoves(t *testing.T) {
	t.Run("white knight in center with friendly blockers and captures", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/8/2P1p3/8/3N4/8/8/8 w - - 0 1")

		var list MoveList
		list.generateKnightMoves(&b, false)

		wantMoves := []Move{
			NewMove(D4, E6, MoveCapture),
			NewMove(D4, B3, MoveQuiet),
			NewMove(D4, B5, MoveQuiet),
			NewMove(D4, C2, MoveQuiet),
			NewMove(D4, E2, MoveQuiet),
			NewMove(D4, F3, MoveQuiet),
			NewMove(D4, F5, MoveQuiet),
		}

		if list.Count != len(wantMoves) {
			t.Fatalf("Move count = %d; want %d", list.Count, len(wantMoves))
		}

		blockedMove := NewMove(D4, C6, MoveQuiet)
		if containsMove(&list, blockedMove) {
			t.Errorf("knight generated illegal move onto friendly blocker square c6")
		}

		for _, m := range wantMoves {
			if !containsMove(&list, m) {
				t.Errorf("missing expected knight move: %v", m)
			}
		}
	})
	t.Run("black knight quiet moves and captures", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/3p1P2/8/4n3/8/8/8/8 b - - 0 1")

		var list MoveList
		list.generateKnightMoves(&b, false)

		captureMove := NewMove(E5, F7, MoveCapture)
		if !containsMove(&list, captureMove) {
			t.Errorf("missing black knight capture on f7: %v", captureMove)
		}

		blockedMove := NewMove(E5, D7, MoveQuiet)
		if containsMove(&list, blockedMove) {
			t.Errorf("black knight generated move onto friendly pawn at d7")
		}
	})
	t.Run("corner knight jumps are clamped to board", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/8/8/8/8/8/8/N7 w - - 0 1")

		var list MoveList
		list.generateKnightMoves(&b, false)

		wantMoves := []Move{
			NewMove(A1, B3, MoveQuiet),
			NewMove(A1, C2, MoveQuiet),
		}

		if list.Count != len(wantMoves) {
			t.Fatalf("expected 2 corner moves, got %d", list.Count)
		}
		for _, m := range wantMoves {
			if !containsMove(&list, m) {
				t.Errorf("missing expected corner jump: %v", m)
			}
		}
	})
	t.Run("capturesOnly ignores quiet knight moves", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/8/2p1p3/8/3N4/8/8/8 w - - 0 1")

		var list MoveList
		list.generateKnightMoves(&b, true)

		if list.Count != 2 {
			t.Fatalf("expected exactly 2 captures, got %d", list.Count)
		}

		capture1 := NewMove(D4, C6, MoveCapture)
		capture2 := NewMove(D4, E6, MoveCapture)

		if !containsMove(&list, capture1) || !containsMove(&list, capture2) {
			t.Errorf("missing expected knight captures")
		}

		quietMove := NewMove(D4, F5, MoveQuiet)
		if containsMove(&list, quietMove) {
			t.Errorf("quiet move generated when capturesOnly is true: %v", quietMove)
		}
	})
}

func TestMoveList_GenerateKingMoves(t *testing.T) {
	t.Run("white king basic moves, captures, and friendly blockers", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/8/8/3pP3/4K3/8/8/8 w - - 0 1")

		var list MoveList
		list.generateKingMoves(&b, false)

		captureMove := NewMove(E4, D5, MoveCapture)
		if !containsMove(&list, captureMove) {
			t.Errorf("missing expected king capture on d5: %v", captureMove)
		}

		blockedMove := NewMove(E4, E5, MoveQuiet)
		if containsMove(&list, blockedMove) {
			t.Errorf("king generated illegal move onto friendly blocker square e5")
		}

		if list.Count != 7 {
			t.Errorf("Move count = %d; want 7", list.Count)
		}
	})
	t.Run("black king moves and captures", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/8/8/4k3/3P1p2/8/8/8 b - - 0 1")

		var list MoveList
		list.generateKingMoves(&b, false)

		captureMove := NewMove(E5, D4, MoveCapture)
		if !containsMove(&list, captureMove) {
			t.Errorf("missing black king capture on d4: %v", captureMove)
		}

		blockedMove := NewMove(E5, F4, MoveQuiet)
		if containsMove(&list, blockedMove) {
			t.Errorf("black king generated move onto friendly pawn at f4")
		}
	})
	t.Run("white castling generated when legal", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/8/8/8/8/8/8/R3K2R w KQkq - 0 1")

		var list MoveList
		list.generateKingMoves(&b, false)

		kingsideCastle := NewMove(E1, G1, MoveCastle)
		queensideCastle := NewMove(E1, C1, MoveCastle)

		if !containsMove(&list, kingsideCastle) {
			t.Errorf("missing legal white kingside castle (e1-g1)")
		}
		if !containsMove(&list, queensideCastle) {
			t.Errorf("missing legal white queenside castle (e1-c1)")
		}
	})
	t.Run("white castling blocked by intervening pieces", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/8/8/8/8/8/8/RB2KNNR w KQkq - 0 1")

		var list MoveList
		list.generateKingMoves(&b, false)

		kingsideCastle := NewMove(E1, G1, MoveCastle)
		queensideCastle := NewMove(E1, C1, MoveCastle)

		if containsMove(&list, kingsideCastle) {
			t.Errorf("white kingside castle generated despite f1 blocker")
		}
		if containsMove(&list, queensideCastle) {
			t.Errorf("white queenside castle generated despite b1 blocker")
		}
	})
	t.Run("white castling blocked when in check or crossing check", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("4r3/8/8/8/8/8/8/R3K2R w KQkq - 0 1")

		var list MoveList
		list.generateKingMoves(&b, false)

		kingsideCastle := NewMove(E1, G1, MoveCastle)
		queensideCastle := NewMove(E1, C1, MoveCastle)

		if containsMove(&list, kingsideCastle) || containsMove(&list, queensideCastle) {
			t.Errorf("castling generated while king is in check at e1")
		}

		b2 := NewBoard()
		b2.ParseFEN("5r2/8/8/8/8/8/8/R3K2R w KQkq - 0 1")

		var list2 MoveList
		list2.generateKingMoves(&b2, false)

		if containsMove(&list2, kingsideCastle) {
			t.Errorf("kingside castling generated while transit square f1 is attacked")
		}
	})
	t.Run("black castling generated and blocked appropriately", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("r3k2r/8/8/8/8/8/8/8 b kq - 0 1")

		var list MoveList
		list.generateKingMoves(&b, false)

		kingsideCastle := NewMove(E8, G8, MoveCastle)
		queensideCastle := NewMove(E8, C8, MoveCastle)

		if !containsMove(&list, kingsideCastle) {
			t.Errorf("missing legal black kingside castle (e8-g8)")
		}
		if !containsMove(&list, queensideCastle) {
			t.Errorf("missing legal black queenside castle (e8-c8)")
		}

		b2 := NewBoard()
		b2.ParseFEN("r3k2r/8/8/8/8/8/8/3R4 b kq - 0 1")

		var list2 MoveList
		list2.generateKingMoves(&b2, false)

		if containsMove(&list2, queensideCastle) {
			t.Errorf("black queenside castling generated while transit square d8 is attacked")
		}
	})
	t.Run("capturesOnly skips quiet moves and castling", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/8/8/8/8/8/3p4/R3K2R w KQkq - 0 1")

		var list MoveList
		list.generateKingMoves(&b, true)

		if list.Count != 1 {
			t.Fatalf("expected exactly 1 capture, got %d", list.Count)
		}

		captureMove := NewMove(E1, D2, MoveCapture)
		if !containsMove(&list, captureMove) {
			t.Errorf("missing king capture on d2")
		}

		kingsideCastle := NewMove(E1, G1, MoveCastle)
		if containsMove(&list, kingsideCastle) {
			t.Errorf("castling generated when capturesOnly is true")
		}
	})
}

func TestMoveList_GenerateBishopMoves(t *testing.T) {
	t.Run("white bishop ray sliding, blockers, and captures", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/8/5p2/2P5/3B4/8/8/8 w - - 0 1")

		var list MoveList
		list.generateBishopMoves(&b, false)

		captureMove := NewMove(D4, F6, MoveCapture)
		if !containsMove(&list, captureMove) {
			t.Errorf("missing bishop capture on f6: %v", captureMove)
		}

		if containsMove(&list, NewMove(D4, G7, MoveQuiet)) ||
			containsMove(&list, NewMove(D4, H8, MoveQuiet)) {
			t.Errorf("bishop ray leaped through enemy blocker on f6")
		}

		if containsMove(&list, NewMove(D4, C5, MoveQuiet)) {
			t.Errorf("bishop generated move onto friendly blocker on c5")
		}
		if containsMove(&list, NewMove(D4, B6, MoveQuiet)) {
			t.Errorf("bishop ray leaped through friendly blocker on c5")
		}

		wantQuietMoves := []Move{
			NewMove(D4, E5, MoveQuiet),
			NewMove(D4, C3, MoveQuiet),
			NewMove(D4, B2, MoveQuiet),
			NewMove(D4, A1, MoveQuiet),
			NewMove(D4, E3, MoveQuiet),
			NewMove(D4, F2, MoveQuiet),
			NewMove(D4, G1, MoveQuiet),
		}

		for _, m := range wantQuietMoves {
			if !containsMove(&list, m) {
				t.Errorf("missing expected quiet bishop move: %v", m)
			}
		}

		if list.Count != 8 {
			t.Errorf("Move count = %d; want 8", list.Count)
		}
	})
	t.Run("black bishop generates quiet moves and captures", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/2P3b1/8/4b3/8/8/8/8 b - - 0 1")

		var list MoveList
		list.generateBishopMoves(&b, false)

		captureMove := NewMove(E5, C7, MoveCapture)
		if !containsMove(&list, captureMove) {
			t.Errorf("missing black bishop capture on c7: %v", captureMove)
		}

		friendlyBlocker := NewMove(E5, G7, MoveQuiet)
		if containsMove(&list, friendlyBlocker) {
			t.Errorf("black bishop generated move onto friendly pawn at g7")
		}
	})
	t.Run("capturesOnly ignores all quiet diagonals", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/8/1p3p2/8/3B4/8/8/8 w - - 0 1")

		var list MoveList
		list.generateBishopMoves(&b, true)

		if list.Count != 2 {
			t.Fatalf("expected exactly 2 captures, got %d", list.Count)
		}

		cap1 := NewMove(D4, B6, MoveCapture)
		cap2 := NewMove(D4, F6, MoveCapture)

		if !containsMove(&list, cap1) || !containsMove(&list, cap2) {
			t.Errorf("missing expected bishop captures when capturesOnly = true")
		}

		quietMove := NewMove(D4, C3, MoveQuiet)
		if containsMove(&list, quietMove) {
			t.Errorf("quiet move generated when capturesOnly is true: %v", quietMove)
		}
	})
}

func TestMoveList_GenerateRookMoves(t *testing.T) {
	t.Run("white rook orthogonal sliding, friendly blockers, and captures", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/8/3p4/8/3R4/3P4/8/8 w - - 0 1")

		var list MoveList
		list.generateRookMoves(&b, false)

		captureMove := NewMove(D4, D6, MoveCapture)
		if !containsMove(&list, captureMove) {
			t.Errorf("missing rook capture on d6: %v", captureMove)
		}

		if containsMove(&list, NewMove(D4, D7, MoveQuiet)) ||
			containsMove(&list, NewMove(D4, D8, MoveQuiet)) {
			t.Errorf("rook ray leaped through enemy blocker on d6")
		}

		if containsMove(&list, NewMove(D4, D3, MoveQuiet)) {
			t.Errorf("rook generated move onto friendly blocker on d3")
		}
		if containsMove(&list, NewMove(D4, D2, MoveQuiet)) ||
			containsMove(&list, NewMove(D4, D1, MoveQuiet)) {
			t.Errorf("rook ray leaped through friendly blocker on d3")
		}
		wantQuietMoves := []Move{
			NewMove(D4, D5, MoveQuiet),
			NewMove(D4, C4, MoveQuiet),
			NewMove(D4, B4, MoveQuiet),
			NewMove(D4, A4, MoveQuiet),
			NewMove(D4, E4, MoveQuiet),
			NewMove(D4, F4, MoveQuiet),
			NewMove(D4, G4, MoveQuiet),
			NewMove(D4, H4, MoveQuiet),
		}

		for _, m := range wantQuietMoves {
			if !containsMove(&list, m) {
				t.Errorf("missing expected quiet rook move: %v", m)
			}
		}

		if list.Count != 9 {
			t.Errorf("Move count = %d; want 9", list.Count)
		}
	})
	t.Run("black rook generates quiet moves and captures", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/4p3/8/4r3/8/8/4P3/8 b - - 0 1")

		var list MoveList
		list.generateRookMoves(&b, false)

		captureMove := NewMove(E5, E2, MoveCapture)
		if !containsMove(&list, captureMove) {
			t.Errorf("missing black rook capture on e2: %v", captureMove)
		}

		friendlyBlocker := NewMove(E5, E7, MoveQuiet)
		if containsMove(&list, friendlyBlocker) {
			t.Errorf("black rook generated move onto friendly pawn at e7")
		}
	})
	t.Run("capturesOnly ignores all quiet orthogonal squares", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/3p4/8/8/3R2p1/8/8/8 w - - 0 1")

		var list MoveList
		list.generateRookMoves(&b, true)

		if list.Count != 2 {
			t.Fatalf("expected exactly 2 captures, got %d", list.Count)
		}

		cap1 := NewMove(D4, D7, MoveCapture)
		cap2 := NewMove(D4, G4, MoveCapture)

		if !containsMove(&list, cap1) || !containsMove(&list, cap2) {
			t.Errorf("missing expected rook captures when capturesOnly = true")
		}

		// Ensure no quiet move leaked through
		quietMove := NewMove(D4, D5, MoveQuiet)
		if containsMove(&list, quietMove) {
			t.Errorf("quiet move generated when capturesOnly is true: %v", quietMove)
		}
	})
}

func TestMoveList_GenerateQueenMoves(t *testing.T) {
	t.Run("white queen 8-direction sliding with blockers and captures", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/8/5p2/8/3Q4/3P4/8/8 w - - 0 1")

		var list MoveList
		list.generateQueenMoves(&b, false)

		captureMove := NewMove(D4, F6, MoveCapture)
		if !containsMove(&list, captureMove) {
			t.Errorf("missing queen capture on f6: %v", captureMove)
		}

		if containsMove(&list, NewMove(D4, G7, MoveQuiet)) ||
			containsMove(&list, NewMove(D4, H8, MoveQuiet)) {
			t.Errorf("queen ray leaped through enemy blocker on f6")
		}

		if containsMove(&list, NewMove(D4, D3, MoveQuiet)) {
			t.Errorf("queen generated move onto friendly blocker on d3")
		}
		if containsMove(&list, NewMove(D4, D2, MoveQuiet)) ||
			containsMove(&list, NewMove(D4, D1, MoveQuiet)) {
			t.Errorf("queen ray leaped through friendly blocker on d3")
		}

		wantQuietMoves := []Move{
			NewMove(D4, E5, MoveQuiet),
			NewMove(D4, D5, MoveQuiet),
			NewMove(D4, C4, MoveQuiet),
			NewMove(D4, E4, MoveQuiet),
			NewMove(D4, C5, MoveQuiet),
			NewMove(D4, C3, MoveQuiet),
			NewMove(D4, E3, MoveQuiet),
		}

		for _, m := range wantQuietMoves {
			if !containsMove(&list, m) {
				t.Errorf("missing expected quiet queen move: %v", m)
			}
		}

		if list.Count != 22 {
			t.Errorf("Move count = %d; want 22", list.Count)
		}
	})
	t.Run("black queen generates quiet moves and captures", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/2P1p3/8/4q3/8/8/4P3/8 b - - 0 1")

		var list MoveList
		list.generateQueenMoves(&b, false)

		orthoCap := NewMove(E5, E2, MoveCapture)
		diagCap := NewMove(E5, C7, MoveCapture)

		if !containsMove(&list, orthoCap) {
			t.Errorf("missing orthogonal capture on e2: %v", orthoCap)
		}
		if !containsMove(&list, diagCap) {
			t.Errorf("missing diagonal capture on c7: %v", diagCap)
		}

		friendlyBlocker := NewMove(E5, E7, MoveQuiet)
		if containsMove(&list, friendlyBlocker) {
			t.Errorf("black queen generated move onto friendly pawn at e7")
		}
	})
	t.Run("capturesOnly ignores all quiet rays", func(t *testing.T) {
		b := NewBoard()
		b.ParseFEN("8/3p2p1/8/8/3Q4/8/8/8 w - - 0 1")

		var list MoveList
		list.generateQueenMoves(&b, true)

		if list.Count != 2 {
			t.Fatalf("expected exactly 2 captures, got %d", list.Count)
		}

		cap1 := NewMove(D4, D7, MoveCapture)
		cap2 := NewMove(D4, G7, MoveCapture)

		if !containsMove(&list, cap1) || !containsMove(&list, cap2) {
			t.Errorf("missing expected queen captures when capturesOnly = true")
		}

		quietMove := NewMove(D4, D5, MoveQuiet)
		if containsMove(&list, quietMove) {
			t.Errorf("quiet move generated when capturesOnly is true: %v", quietMove)
		}
	})
}
