package board

// generates pawn moves
func (m *MoveList) generatePawnMoves(b *Board, capturesOnly bool) {

	switch b.Side {
	case White:
		{
			bitboard := b.PieceBitBoards[WP]

			for bitboard != 0 {
				source := Square(bitboard.GetLSB())
				target := source - 8

				// quiet single/double pawn push
				if !capturesOnly && target >= A8 && b.Occupancies[Both].GetBit(target) == 0 {
					// quiet promotion
					if source >= A7 && source <= H7 {
						m.AddMove(NewMove(source, target, MovePromoQ))
						m.AddMove(NewMove(source, target, MovePromoR))
						m.AddMove(NewMove(source, target, MovePromoB))
						m.AddMove(NewMove(source, target, MovePromoN))
					} else {
						m.AddMove(NewMove(source, target, MoveQuiet))
						// double push from rank 2
						if (source >= A2 && source <= H2) && b.Occupancies[Both].GetBit(target-8) == 0 {
							m.AddMove(NewMove(source, target-8, MoveDouble))
						}
					}
				}

				// standard pawn captures
				attacks := PawnAttacks[b.Side][source] & b.Occupancies[Black]

				for attacks != 0 {
					target = Square(attacks.GetLSB())
					if source >= A7 && source <= H7 {
						m.AddMove(NewMove(source, target, MovePromoCaptureQ))
						m.AddMove(NewMove(source, target, MovePromoCaptureR))
						m.AddMove(NewMove(source, target, MovePromoCaptureB))
						m.AddMove(NewMove(source, target, MovePromoCaptureN))
					} else {
						m.AddMove(NewMove(source, target, MoveCapture))
					}
					attacks.PopBit(target)
				}

				// en-passant captures
				if b.EnPassant != NoSquare {
					enpassantAttacks := PawnAttacks[b.Side][source] & (BitBoard(uint64(1) << uint64(b.EnPassant)))
					if enpassantAttacks != 0 {
						targetEnpass := Square(enpassantAttacks.GetLSB())
						m.AddMove(NewMove(source, targetEnpass, MoveEnPass))
					}
				}

				bitboard.PopBit(source)
			}
		}
	case Black:
		{
			bitboard := b.PieceBitBoards[BP]

			for bitboard != 0 {
				source := Square(bitboard.GetLSB())
				target := source + 8

				// quiet single/double pawn push
				if !capturesOnly && target <= H1 && b.Occupancies[Both].GetBit(target) == 0 {
					// quiet promotion
					if source >= A2 && source <= H2 {
						m.AddMove(NewMove(source, target, MovePromoQ))
						m.AddMove(NewMove(source, target, MovePromoR))
						m.AddMove(NewMove(source, target, MovePromoB))
						m.AddMove(NewMove(source, target, MovePromoN))
					} else {
						m.AddMove(NewMove(source, target, MoveQuiet))
						// double push from rank 7
						if (source >= A7 && source <= H7) && b.Occupancies[Both].GetBit(target+8) == 0 {
							m.AddMove(NewMove(source, target+8, MoveDouble))
						}
					}
				}

				// standard pawn captures
				attacks := PawnAttacks[b.Side][source] & b.Occupancies[White]

				for attacks != 0 {
					target = Square(attacks.GetLSB())
					if source >= A2 && source <= H2 {
						m.AddMove(NewMove(source, target, MovePromoCaptureQ))
						m.AddMove(NewMove(source, target, MovePromoCaptureR))
						m.AddMove(NewMove(source, target, MovePromoCaptureB))
						m.AddMove(NewMove(source, target, MovePromoCaptureN))
					} else {
						m.AddMove(NewMove(source, target, MoveCapture))
					}
					attacks.PopBit(target)
				}

				// en-passant captures
				if b.EnPassant != NoSquare {
					enpassantAttacks := PawnAttacks[b.Side][source] & (BitBoard(uint64(1) << uint64(b.EnPassant)))
					if enpassantAttacks != 0 {
						targetEnpass := Square(enpassantAttacks.GetLSB())
						m.AddMove(NewMove(source, targetEnpass, MoveEnPass))
					}
				}

				bitboard.PopBit(source)
			}

		}
	}
}

// generate knight moves
func (m *MoveList) generateKnightMoves(b *Board, capturesOnly bool) {
	var pce Piece
	var friendly BitBoard
	var enemy BitBoard

	// set side-specific piece type and occupancy bitboards
	if b.Side == White {
		pce = WN
		friendly = b.Occupancies[White]
		enemy = b.Occupancies[Black]
	} else {
		pce = BN
		friendly = b.Occupancies[Black]
		enemy = b.Occupancies[White]
	}

	bitboard := b.PieceBitBoards[pce]

	// for each knight on board
	for bitboard != 0 {
		source := Square(bitboard.GetLSB())

		// excludes friendly blockers
		attacks := KnightAttacks[source] & ^friendly
		if capturesOnly {
			attacks &= enemy
		}

		for attacks != 0 {
			target := Square(attacks.GetLSB())

			if enemy.GetBit(target) == 0 {
				// quiet move
				m.AddMove(NewMove(source, target, MoveQuiet))
			} else {
				// capture move
				m.AddMove(NewMove(source, target, MoveCapture))
			}

			attacks.PopBit(target)
		}

		bitboard.PopBit(source)
	}
}

// generate king moves
func (m *MoveList) generateKingMoves(b *Board, capturesOnly bool) {
	var pce Piece
	var friendly BitBoard
	var enemy BitBoard

	// set side-specific piece type and occupancy bitboards
	if b.Side == White {
		pce = WK
		friendly = b.Occupancies[White]
		enemy = b.Occupancies[Black]
	} else {
		pce = BK
		friendly = b.Occupancies[Black]
		enemy = b.Occupancies[White]
	}

	bitboard := b.PieceBitBoards[pce]

	if bitboard != 0 {
		source := Square(bitboard.GetLSB())

		// excludes friendly blockers
		attacks := KingAttacks[source] & ^friendly
		if capturesOnly {
			attacks &= enemy
		}

		for attacks != 0 {
			target := Square(attacks.GetLSB())

			if enemy.GetBit(target) == 0 {
				// quiet move
				m.AddMove(NewMove(source, target, MoveQuiet))
			} else {
				// capture move
				m.AddMove(NewMove(source, target, MoveCapture))
			}

			attacks.PopBit(target)
		}

		bitboard.PopBit(source)
	}

	// white castle
	if pce == WK && !capturesOnly {
		// king-side
		if b.Castle&WKCA != 0 {
			if b.Occupancies[Both].GetBit(F1) == 0 && b.Occupancies[Both].GetBit(G1) == 0 {
				if !b.IsSquareAttacked(E1, Black) && !b.IsSquareAttacked(F1, Black) {
					m.AddMove(NewMove(E1, G1, MoveCastle))
				}
			}
		}
		// queen-side
		if b.Castle&WQCA != 0 {
			if b.Occupancies[Both].GetBit(D1) == 0 && b.Occupancies[Both].GetBit(C1) == 0 && b.Occupancies[Both].GetBit(B1) == 0 {
				if !b.IsSquareAttacked(E1, Black) && !b.IsSquareAttacked(D1, Black) {
					m.AddMove(NewMove(E1, C1, MoveCastle))
				}
			}
		}
	}

	if pce == BK && !capturesOnly {
		// king-side
		if b.Castle&BKCA != 0 {
			if b.Occupancies[Both].GetBit(F8) == 0 && b.Occupancies[Both].GetBit(G8) == 0 {
				if !b.IsSquareAttacked(E8, White) && !b.IsSquareAttacked(F8, White) {
					m.AddMove(NewMove(E8, G8, MoveCastle))
				}
			}
		}
		// queen-side
		if b.Castle&BQCA != 0 {
			if b.Occupancies[Both].GetBit(D8) == 0 && b.Occupancies[Both].GetBit(C8) == 0 && b.Occupancies[Both].GetBit(B8) == 0 {
				if !b.IsSquareAttacked(E8, White) && !b.IsSquareAttacked(D8, White) {
					m.AddMove(NewMove(E8, C8, MoveCastle))
				}
			}
		}
	}

}

// generates bishop moves
func (m *MoveList) generateBishopMoves(b *Board, capturesOnly bool) {
	var pce Piece
	var friendly BitBoard
	var enemy BitBoard

	// set side-specific piece type and occupancy bitboards
	if b.Side == White {
		pce = WB
		friendly = b.Occupancies[White]
		enemy = b.Occupancies[Black]
	} else {
		pce = BB
		friendly = b.Occupancies[Black]
		enemy = b.Occupancies[White]
	}

	bitboard := b.PieceBitBoards[pce]

	// for each bishop on board
	for bitboard != 0 {
		source := Square(bitboard.GetLSB())

		// excludes friendly blockers
		attacks := GetBishopAttacks(source, b.Occupancies[Both]) & ^friendly
		if capturesOnly {
			attacks &= enemy
		}

		for attacks != 0 {
			target := Square(attacks.GetLSB())

			if enemy.GetBit(target) == 0 {
				// quiet move
				m.AddMove(NewMove(source, target, MoveQuiet))
			} else {
				// capture move
				m.AddMove(NewMove(source, target, MoveCapture))
			}

			attacks.PopBit(target)
		}

		bitboard.PopBit(source)
	}

}

// generates pseudo-legal rook moves
func (m *MoveList) generateRookMoves(b *Board, capturesOnly bool) {
	var pce Piece
	var friendly BitBoard
	var enemy BitBoard

	// set side-specific piece type and occupancy bitboards
	if b.Side == White {
		pce = WR
		friendly = b.Occupancies[White]
		enemy = b.Occupancies[Black]
	} else {
		pce = BR
		friendly = b.Occupancies[Black]
		enemy = b.Occupancies[White]
	}

	bitboard := b.PieceBitBoards[pce]

	// for each rook on board
	for bitboard != 0 {
		source := Square(bitboard.GetLSB())

		// excludes friendly blockers
		attacks := GetRookAttacks(source, b.Occupancies[Both]) & ^friendly
		if capturesOnly {
			attacks &= enemy
		}

		for attacks != 0 {
			target := Square(attacks.GetLSB())

			if enemy.GetBit(target) == 0 {
				// quiet move
				m.AddMove(NewMove(source, target, MoveQuiet))
			} else {
				// capture move
				m.AddMove(NewMove(source, target, MoveCapture))
			}

			attacks.PopBit(target)
		}

		bitboard.PopBit(source)
	}

}

// generates pseudo-legal queen moves
func (m *MoveList) generateQueenMoves(b *Board, capturesOnly bool) {
	var pce Piece
	var friendly BitBoard
	var enemy BitBoard

	// set side-specific piece type and occupancy bitboards
	if b.Side == White {
		pce = WQ
		friendly = b.Occupancies[White]
		enemy = b.Occupancies[Black]
	} else {
		pce = BQ
		friendly = b.Occupancies[Black]
		enemy = b.Occupancies[White]
	}

	bitboard := b.PieceBitBoards[pce]

	// for each queen on board
	for bitboard != 0 {
		source := Square(bitboard.GetLSB())

		// excludes friendly blockers
		attacks := GetQueenAttacks(source, b.Occupancies[Both]) & ^friendly
		if capturesOnly {
			attacks &= enemy
		}

		for attacks != 0 {
			target := Square(attacks.GetLSB())

			if enemy.GetBit(target) == 0 {
				// quiet moves
				m.AddMove(NewMove(source, target, MoveQuiet))
			} else {
				// capture moves
				m.AddMove(NewMove(source, target, MoveCapture))
			}

			attacks.PopBit(target)
		}

		bitboard.PopBit(source)
	}

}

// generate pseudo-legal moves for all pieces
func (m *MoveList) generate(b *Board, capturesOnly bool) {
	m.generatePawnMoves(b, capturesOnly)
	m.generateKnightMoves(b, capturesOnly)
	m.generateKingMoves(b, capturesOnly)
	m.generateBishopMoves(b, capturesOnly)
	m.generateRookMoves(b, capturesOnly)
	m.generateQueenMoves(b, capturesOnly)
}

// generate all pseudo-legal moves
func (m *MoveList) GenerateMoves(b *Board) {
	m.generate(b, false)
}

// generate only capture moves
func (m *MoveList) GenerateCaptures(b *Board) {
	m.generate(b, true)
}
