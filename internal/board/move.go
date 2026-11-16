package board

import (
	"fmt"
)

//       binary move bits (16 bits)                     hexidecimal constants

// 0000 0000 0011 1111    source square                 0x003f
// 0000 1111 1100 0000    target square                 0x0fc0
// 1111 0000 0000 0000    move type                     0xf000

//	type   meaning
//	0      quiet move
//	1      double pawn push
//	2      castle
//	4      capture
//	5      en passant capture
//	8-11   promotion to knight, bishop, rook, queen
//	12-15  promotion with capture to knight, bishop, rook, queen

type Move uint16

// 4-bit type stored in the top bits of a Move
type MoveType uint16

const (
	SourceMask    = 0x3f
	TargetMask    = 0xfc0
	TypeMask      = 0xf000
	CaptureFlag   = 0x4000 // set for every capture type, including en passant and promotion captures
	PromotionFlag = 0x8000 // set for every promotion type

	typeShift = 12
)

const (
	MoveQuiet   MoveType = 0
	MoveDouble  MoveType = 1
	MoveCastle  MoveType = 2
	MoveCapture MoveType = 4
	MoveEnPass  MoveType = 5

	MovePromoN MoveType = 8
	MovePromoB MoveType = 9
	MovePromoR MoveType = 10
	MovePromoQ MoveType = 11

	MovePromoCaptureN MoveType = 12
	MovePromoCaptureB MoveType = 13
	MovePromoCaptureR MoveType = 14
	MovePromoCaptureQ MoveType = 15
)

const NOMOVE = Move(0)
const MAX_MOVES = 512

func NewMove(source, target Square, moveType MoveType) Move {
	return Move(uint16(source) | uint16(target)<<6 | uint16(moveType)<<typeShift)
}

// returns square from where the piece was moved
func (m Move) Source() Square {
	return Square(m & SourceMask)
}

// returns square to which the piece is moved
func (m Move) Target() Square {
	return Square((m & TargetMask) >> 6)
}

// returns the 4-bit move type
func (m Move) Type() MoveType {
	return MoveType(m >> typeShift)
}

// returns if move is a promotion
func (m Move) IsPromotion() bool {
	return (m & PromotionFlag) != 0
}

// returns the piece to which pawn promoted for the given side
func (m Move) Promoted(side Color) Piece {
	if !m.IsPromotion() {
		return Empty
	}

	p := WN + Piece((m>>typeShift)&3)
	if side == Black {
		p += 6
	}

	return p
}

// returns promotion letter
func (m Move) PromotionChar() byte {
	return "nbrq"[(m>>typeShift)&3]
}

// returns if move was a capture
func (m Move) IsCapture() bool {
	return (m & CaptureFlag) != 0
}

// returns if move was a double pawn push
func (m Move) IsDouble() bool {
	return m.Type() == MoveDouble
}

// returns if move was an enpassant
func (m Move) IsEnpass() bool {
	return m.Type() == MoveEnPass
}

// returns if move was a castle
func (m Move) IsCastle() bool {
	return m.Type() == MoveCastle
}

// print move
func (m Move) String() string {

	if m == 0 {
		return ""
	}

	res := fmt.Sprintf("%s%s", m.Source(), m.Target())

	if m.IsPromotion() {
		res += string(m.PromotionChar())
	}

	return res
}
