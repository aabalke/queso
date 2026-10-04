package arm32

import (
	"encoding/binary"
	"io"
)

type Arm32 struct {
	Order              binary.ByteOrder
	RegisterNaming     RegisterNaming
	BlockSuffixes      BlockSuffixes
	Thumb              bool
	HexImmediateValues bool // #0xF, #15
}

func (isa *Arm32) Decode(r io.Reader) []string {
	var s []string

	bs, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}

	w := 4
	if isa.Thumb {
		w = 2
	}

	for i := 0; i < len(bs); i += w {

		if i+w > len(bs) {
			break
		}

		if !isa.Thumb {
			v := isa.Order.Uint32(bs[i:])
			s = append(s, isa.decodeInstruction32(v))
		} else {
			v := uint32(isa.Order.Uint16(bs[i:]))
			s = append(s, isa.decodeInstruction16(v))
		}
	}

	return s
}

func (isa *Arm32) decodeInstruction32(v uint32) string {
	decoder := INST_UNKNOWN
	for i, desc := range format32 {
		if (v & desc.mask) == desc.match {
			decoder = Decoder(i)
			break
		}
	}

	return decoders32[decoder](isa, v)
}

func (isa *Arm32) decodeInstruction16(v uint32) string {
	decoder := INST_THUMB_UNKNOWN
	for i, desc := range format16 {
		if (v & desc.mask) == desc.match {
			decoder = Decoder(i)
			break
		}
	}

	return decoders16[decoder](isa, uint16(v))
}
