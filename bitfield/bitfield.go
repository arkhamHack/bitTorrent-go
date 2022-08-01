package bitfield

type Bitfield []byte

func (bf Bitfield) SetPiece(index int) {
	bf[index/8] |= 1 << (7 - (index % 8))
}

func (bf Bitfield) HasPiece(index int) bool {
	return bf[index/8]>>(7-(index%8))&1 != 0
}
