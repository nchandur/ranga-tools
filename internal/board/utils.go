package board

// returns square for given rank and file
func FRtoSq(rank, file int) Square {
	return Square((rank * 8) + file)
}

// returns rank, file for given square
func SqToFR(square Square) (int, int) {
	tr, tf := int(square)/8, int(square)%8
	return tr, tf
}
