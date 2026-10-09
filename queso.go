package queso

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"
)

type Arch string

const (
	// ARCH_ARM     Arch = "arm"
	ARCH_ARMv2 Arch = "armv2"
	// ARCH_ARMv2a  Arch = "armv2a"
	// ARCH_ARMv3   Arch = "armv3"
	// ARCH_ARMv3m  Arch = "armv3m"
	// ARCH_ARMv4   Arch = "armv4"
	ARCH_ARMv4t Arch = "armv4t"
	// ARCH_ARMv5   Arch = "armv5"
	// ARCH_ARMv5t  Arch = "armv5t"
	ARCH_ARMv5te Arch = "armv5te"
	// "xscale"
	// "ep9312"
	// "iwmmxt"
	// "iwmmxt2"
	// "armv5tej"
	// "armv6"
	// "armv6kz"
	// "armv6t2"
	// "armv6k"
	// "armv7"
	// "armv6-m"
	// "armv6s-m"
	// "armv7e-m"
	// "armv8-a"
	// "armv8-r"
	// "armv8-m.base"
	// "armv8-m.main"
	// "armv8.1-m.main"
	ARCH_ARM_ANY Arch = "arm_any"
)

type Isa struct {
	Arch               Arch
	Order              binary.ByteOrder
	RegisterNaming     RegisterNaming
	ImmediateValueBase HexImmediateBase
	UndefinedOutput    string
	Thumb              bool
}

func (isa *Isa) Decode(r io.Reader) []string {
	bs, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}

	var (
		w           int
		unknown     Decoder
		formatters  []Formatter
		decoderFunc func(b []byte, isa *Isa, decoder Decoder, op uint32, addr int) []byte
	)

	if isa.Thumb {
		w = 2
		unknown = INST_THUMB_UNKNOWN

		switch isa.Arch {
		default:
			formatters = FORMAT16_ARCH_ARM
		}

		decoderFunc = func(b []byte, isa *Isa, decoder Decoder, op uint32, addr int) []byte {
			return decoders16[decoder](b, isa, uint16(op), addr)
		}
	} else {
		w = 4
		unknown = INST_UNKNOWN

		switch isa.Arch {
		case ARCH_ARMv2:
			formatters = FORMAT32_ARMv2
		case ARCH_ARMv4t:
			formatters = FORMAT32_ARMv4t
		case ARCH_ARMv5te:
			formatters = FORMAT32_ARMv5te
		default:
			formatters = FORMAT16_ARCH_ARM
		}

		decoderFunc = func(b []byte, isa *Isa, decoder Decoder, op uint32, addr int) []byte {
			return decoders32[decoder](b, isa, op, addr)
		}
	}

	if leftover := len(bs) & (w - 1); leftover != 0 {
		panic(fmt.Sprintf("invalid byte cnt need mod %d, have rem %d", w, leftover))
	}

	// NOTE: start and end are used to only allocate strings once, removes per op allocations
	var (
		cnt  = len(bs) / w
		buf  = make([]byte, 0, cnt*256)
		ends = make([]uint32, cnt)
		ops  = make([]uint32, cnt)
	)

	for i := range cnt {
		var v uint32
		if w == 4 {
			v = isa.Order.Uint32(bs[i*w:])
		} else {
			v = uint32(isa.Order.Uint16(bs[i*w:]))
		}
		ops[i] = v

		decoder := unknown
		for j, desc := range formatters {

			if desc.mask == 0 && desc.match == 0 {
				continue
			}

			if (v & desc.mask) == desc.match {
				decoder = Decoder(j)
				break
			}
		}

		buf = decoderFunc(buf, isa, decoder, v, i*w)
		ends[i] = uint32(len(buf))
	}

	return convertBytesToStringArray(buf, ends)
}

func convertBytesToStringArray(buf []byte, ends []uint32) []string {
	allStrings := unsafe.String(unsafe.SliceData(buf), len(buf))

	// NOTE: do not append to buf after unsafe.String

	s := make([]string, len(ends))
	start := uint32(0)
	for i, end := range ends {
		s[i] = allStrings[start:end]
		start = end
	}

	return s
}
