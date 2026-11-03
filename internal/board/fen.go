package board

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const START = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

// parses a FEN (Forsyth-Edwards Notation) string and populates board
func (b *Board) ParseFEN(fen string) error {
	if len(fen) == 0 {
		return fmt.Errorf("failed to parse FEN: string is empty")
	}

	// checks if FEN strings contain exactly 6 space-separated fields
	fields := strings.Fields(fen)
	if len(fields) != 6 {
		return fmt.Errorf("failed to parse FEN: expected 6 fields, got %d", len(fields))
	}

	b.Clear()

	rank := 0
	file := 0

	// parses piece placement
	for _, char := range fields[0] {
		if char == '/' {
			rank++
			file = 0
			continue
		}

		if unicode.IsDigit(char) {
			emptySquares := int(char - '0')
			if emptySquares < 1 || emptySquares > 8 {
				return fmt.Errorf("failed to parse FEN: invalid empty-square count '%c'", char)
			}
			file += emptySquares
		} else {
			idx := strings.IndexRune(PieceChar, char)
			if idx < 0 {
				return fmt.Errorf("failed to parse FEN: invalid piece character '%c'", char)
			}
			piece := Piece(idx)

			if rank < 0 || rank > 7 || file < 0 || file > 7 {
				return fmt.Errorf("failed to parse FEN: piece placement out of bounds at rank %d file %d", rank, file)
			}

			sq := FRtoSq(rank, file)
			if piece != Empty && sq != NoSquare {
				b.AddPiece(piece, sq)
			}
			file++
		}
	}

	// parses current side to move
	if fields[1] != "w" && fields[1] != "b" {
		return fmt.Errorf("failed to parse FEN: invalid color value %s", fields[1])
	}

	if fields[1] == "w" {
		b.Side = White
	} else {
		b.Side = Black
	}

	// parses castle rights
	if fields[2] != "-" {
		for _, char := range fields[2] {
			switch char {
			case 'K':
				b.Castle |= WKCA
			case 'Q':
				b.Castle |= WQCA
			case 'k':
				b.Castle |= BKCA
			case 'q':
				b.Castle |= BQCA
			default:
				return fmt.Errorf("failed to parse FEN: invalid castle permissions %s", fields[2])
			}
		}
	}

	// parses en passant square
	if fields[3] == "-" {
		b.EnPassant = NoSquare
	} else {
		if len(fields[3]) != 2 {
			return fmt.Errorf("failed to parse FEN: invalid en passant square %s", fields[3])
		}

		fileChar := fields[3][0]
		rankChar := fields[3][1]

		if fileChar < 'a' || fileChar > 'h' || rankChar < '1' || rankChar > '8' {
			return fmt.Errorf("failed to parse FEN: invalid en passant square %s", fields[3])
		}

		f := int(fileChar - 'a')
		r := int('8' - rankChar)
		sq := FRtoSq(r, f)

		if sq == NoSquare {
			return fmt.Errorf("failed to parse FEN: invalid en passant square %s", fields[3])
		}

		b.EnPassant = sq
	}

	var err error

	// parses fifty-move counter
	b.FiftyMove, err = strconv.Atoi(fields[4])
	if err != nil {
		return fmt.Errorf("failed to parse FEN: %v", err)
	}
	if b.FiftyMove < 0 {
		return fmt.Errorf("failed to parse FEN: invalid halfmove clock %d", b.FiftyMove)
	}

	fullMove, err := strconv.Atoi(fields[5])
	if err != nil {
		return fmt.Errorf("failed to parse FEN: %v", err)
	}

	if fullMove < 1 {
		return fmt.Errorf("failed to parse FEN: invalid fullmove number %d", fullMove)
	}

	// parses ply count
	b.Ply = (fullMove - 1) * 2
	if b.Side == Black {
		b.Ply++
	}

	// initializes position hash key
	b.Key = b.GenerateKey()

	return nil
}

// returns FEN string from board state
func (b *Board) FEN() string {
	var builder strings.Builder

	fenRanks := make([]string, 0)

	for rank := range 8 {
		var rankBuilder strings.Builder
		empty := 0

		for file := range 8 {
			sq := FRtoSq(rank, file)
			piece := b.Mailbox[sq]

			if piece == Empty {
				empty++
			} else {
				if empty > 0 {
					fmt.Fprintf(&rankBuilder, "%d", empty)
					empty = 0
				}
				fmt.Fprintf(&rankBuilder, "%c", PieceChar[piece])
			}
		}

		if empty > 0 {
			fmt.Fprintf(&rankBuilder, "%d", empty)
		}
		fenRanks = append(fenRanks, rankBuilder.String())
	}

	fmt.Fprint(&builder, strings.Join(fenRanks, "/"))
	fmt.Fprintf(&builder, " %c %s %s %d %d", b.Side.String(), b.Castle.String(), b.EnPassant.String(), b.FiftyMove, b.Ply/2+1)

	return builder.String()
}
