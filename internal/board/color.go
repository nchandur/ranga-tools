package board

type Color uint8

const (
	White Color = iota
	Black
	Both
)

func (c Color) String() byte {
	switch c {
	case White:
		return 'w'
	case Black:
		return 'b'
	default:
		return '-'
	}
}
