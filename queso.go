package queso

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"
)

// TODO: add offset addr option

type Arch struct {
	idx   int
	Name  string
	thumb bool
}

const ( // used for gating features
	ARCH_IDX_ARMv2 = iota
	ARCH_IDX_ARMv3
	ARCH_IDX_ARMv4tm
	ARCH_IDX_ARMv5te
	ARCH_IDX_ARM_ANY
)

var (
	ARCH_ARMv2   = Arch{ARCH_IDX_ARMv2, "armv2", false}
	ARCH_ARMv3   = Arch{ARCH_IDX_ARMv3, "armv3", false}
	ARCH_ARMv4tm = Arch{ARCH_IDX_ARMv4tm, "armv4tm", true}
	ARCH_ARMv5te = Arch{ARCH_IDX_ARMv5te, "armv5te", true}
	ARCH_ARM_ANY = Arch{ARCH_IDX_ARM_ANY, "arm_any", true}
)

var OrdersLookup = map[string]binary.ByteOrder{
	"little": binary.LittleEndian,
	"big":    binary.BigEndian,
	"native": binary.NativeEndian,
}

var ArchLookup = map[string]Arch{
	"armv2":   ARCH_ARMv2,
	"armv3":   ARCH_ARMv3,
	"armv4tm": ARCH_ARMv4tm,
	"armv5te": ARCH_ARMv5te,
	"arm_any": ARCH_ARM_ANY,
}

//"armv1"
//"armv2"
//"armv2a"
//"armv3"
//"armv3m"
//"armv4"
//"armv4t"
//"armv5"
//"armv5t"
//"armv5te"
//"xscale"
//"ep9312"
//"iwmmxt"
//"iwmmxt2"
//"armv5tej"
//"armv6"
//"armv6kz"
//"armv6t2"
//"armv6k"
//"armv7"
//"armv6-m"
//"armv6s-m"
//"armv7e-m"
//"armv8-a"
//"armv8-r"
//"armv8-m.base"
//"armv8-m.main"
//"armv8.1-m.main"
//"arm_any"

type Isa struct {
	Arch                               Arch
	Order                              binary.ByteOrder
	RegisterNaming                     RegisterNaming
	UndefinedOutput                    string
	Thumb, OutputAddress, OutputOpcode bool
}

func (isa *Isa) Decode(r io.Reader) ([]string, error) {
	bs, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("invalid input")
	}

	var (
		w           int
		unknown     Decoder
		formatters  []Formatter
		decoderFunc func(b []byte, isa *Isa, decoder Decoder, op uint32, addr int) []byte
	)

	if !isa.Thumb {
		w = 4
		unknown = INST_UNKNOWN

		switch isa.Arch {
		case ARCH_ARMv2:
			formatters = FORMAT32_ARMv2
		case ARCH_ARMv3:
			formatters = FORMAT32_ARMv3
		case ARCH_ARMv4tm:
			formatters = FORMAT32_ARMv4t
		case ARCH_ARMv5te:
			formatters = FORMAT32_ARMv5te
		default:
			formatters = FORMAT32_ARM
		}

		decoderFunc = func(b []byte, isa *Isa, decoder Decoder, op uint32, addr int) []byte {
			return decoders32[decoder](b, isa, op, addr)
		}
	} else {
		w = 2
		unknown = INST_THUMB_UNKNOWN

		if !isa.Arch.thumb {
			return nil, fmt.Errorf(fmt.Sprintf("arm arch version %s does not support thumb instructions", isa.Arch.Name))
		}

		switch isa.Arch {
		default:
			formatters = FORMAT16_ARCH_ARM
		}

		decoderFunc = func(b []byte, isa *Isa, decoder Decoder, op uint32, addr int) []byte {
			return decoders16[decoder](b, isa, uint16(op), addr)
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

		if isa.OutputAddress {
			buf = append(buf, fmt.Sprintf("0x%08X: ", i*w)...)
		}

		if isa.OutputOpcode {
			if isa.Thumb {
				buf = append(buf, fmt.Sprintf("%04X: ", v)...)
			} else {
				buf = append(buf, fmt.Sprintf("%08X: ", v)...)
			}
		}

		buf = decoderFunc(buf, isa, decoder, v, i*w)
		ends[i] = uint32(len(buf))
	}

	return convertBytesToStringArray(buf, ends), nil
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
