package board

type Castle int8

const (
	WKCA = 1 << iota
	WQCA
	BKCA
	BQCA
)

func (c Castle) flip() Castle {
	return (c<<2 | c>>2) & 0xF
}

func (c Castle) String() string {
	var res string

	if c&WKCA != 0 {
		res += "K"
	}

	if c&WQCA != 0 {
		res += "Q"
	}

	if c&BKCA != 0 {
		res += "k"
	}

	if c&BQCA != 0 {
		res += "q"
	}

	if c == 0 {
		res = "-"
	}

	return res
}
