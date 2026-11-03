package board

func (b *Board) Mirror() {
	copy := b.Preserve()
	b.Clear()

	for sq := range Square(64) {
		piece := copy.Mailbox[sq]

		if piece == Empty {
			continue
		}

		b.AddPiece(piece.flip(), sq^56)
	}

	b.Side = copy.Side ^ 1

	if copy.EnPassant != NoSquare {
		b.EnPassant = copy.EnPassant ^ 56
	}

	b.Castle = copy.Castle.flip()
	b.Ply = copy.Ply
	b.FiftyMove = copy.FiftyMove

	b.Key = b.GenerateKey()

}
