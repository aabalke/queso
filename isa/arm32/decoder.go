package arm32

import "math/bits"

type Arm32 struct {
	Thumb                 bool
	BigEndian             bool
	ConventionNaming      bool // pc instad of r15
	StackOrientedSuffixes bool // ldm / stm ldmia vs ldmed
	HexImmediateValues    bool // #0xF, #15
}

func (isa *Arm32) Decode(v []uint32) []string {
	if !isa.BigEndian {
		for i := range len(v) {
			v[i] = bits.ReverseBytes32(v[i])
		}
	}

	var s []string
	for i := range v {
		s = append(s, isa.decodeInstruction(v[i]))
	}

	return s
}

func (isa *Arm32) decodeInstruction(v uint32) string {
	decoder := INST_UNKNOWN
	for i, desc := range format32 {
		if (v & desc.mask) == desc.match {
			decoder = Decoder(i)
			break
		}
	}

	return decoders32[decoder](v)
}
