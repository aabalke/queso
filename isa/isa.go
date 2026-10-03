package isa

type Isa interface {
	Decode(insts []uint32) string
}
