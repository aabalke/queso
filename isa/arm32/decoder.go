package arm32

import (
	"math/bits"
)

type Arm32 struct {
	Thumb                 bool
	BigEndian             bool
	ConventionNaming      bool // pc instad of r15
	StackOrientedSuffixes bool // ldm / stm ldmia vs ldmed
	HexImmediateValues    bool // #0xF, #15
}

func (isa *Arm32) Decode(v []uint32) []string {
	if !isa.BigEndian {
		if isa.Thumb {
			for i := range len(v) {
				v[i] = uint32(bits.ReverseBytes16(uint16(v[i])))
			}
		} else {
			for i := range len(v) {
				v[i] = bits.ReverseBytes32(v[i])
			}
		}
	}

	var s []string

	if isa.Thumb {
		for i := range v {
			s = append(s, isa.decodeInstruction16(v[i]))
		}
	} else {
		for i := range v {
			s = append(s, isa.decodeInstruction32(v[i]))
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

	return decoders32[decoder](v)
}

func (isa *Arm32) decodeInstruction16(v uint32) string {
	decoder := INST_THUMB_UNKNOWN
	for i, desc := range format16 {
		if (v & desc.mask) == desc.match {
			decoder = Decoder(i)
			break
		}
	}

	return decoders16[decoder](uint16(v))
}
