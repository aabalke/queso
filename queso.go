package queso

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"
)

type Architecture struct {
	Name            string
	ArmVersion      ArmVersion
	ThumbVersion    ThumbVersion
	EnhancedDsp     EnhancedDspFeature
	Formatters      []Formatter
	ThumbFormatters []Formatter
	LongMultiply    bool
}

var ArchLookup = map[string]Architecture{
	"armv2":     ARMv2,
	"armv3":     ARMv3,
	"armv3m":    ARMv3M,
	"armv4xm":   ARMv4xM,
	"armv4":     ARMv4,
	"armv4txm":  ARMv4TxM,
	"armv4t":    ARMv4T,
	"armv5xm":   ARMv5xM,
	"armv5":     ARMv5,
	"armv5txm":  ARMv5TxM,
	"armv5t":    ARMv5T,
	"armv5texp": ARMv5TExP,
	"armv5te":   ARMv5TE,
}

var (
	ARMv2     = Architecture{"ARMv2", ArmVersion2, ThumbVersionNone, EnhancedDspNone, FormattersArmv2, nil, false}
	ARMv3     = Architecture{"ARMv3", ArmVersion3, ThumbVersionNone, EnhancedDspNone, FormattersArmv3, nil, false}
	ARMv3M    = Architecture{"ARMv3M", ArmVersion3, ThumbVersionNone, EnhancedDspNone, FormattersArmv3, nil, true}
	ARMv4xM   = Architecture{"ARMv4xM", ArmVersion4, ThumbVersionNone, EnhancedDspNone, FormattersArmv4, nil, false}
	ARMv4     = Architecture{"ARMv4", ArmVersion4, ThumbVersionNone, EnhancedDspNone, FormattersArmv4, nil, true}
	ARMv4TxM  = Architecture{"ARMv4TxM", ArmVersion4, ThumbVersion1, EnhancedDspNone, FormattersArmv4, FormattersThumbv1, false}
	ARMv4T    = Architecture{"ARMv4T", ArmVersion4, ThumbVersion1, EnhancedDspNone, FormattersArmv4, FormattersThumbv1, true}
	ARMv5xM   = Architecture{"ARMv5xM", ArmVersion5, ThumbVersionNone, EnhancedDspNone, FormattersArmv5, nil, false}
	ARMv5     = Architecture{"ARMv5", ArmVersion5, ThumbVersionNone, EnhancedDspNone, FormattersArmv5, nil, true}
	ARMv5TxM  = Architecture{"ARMv5TxM", ArmVersion5, ThumbVersion2, EnhancedDspNone, FormattersArmv5, FormattersThumbv2, false}
	ARMv5T    = Architecture{"ARMv5T", ArmVersion5, ThumbVersion2, EnhancedDspNone, FormattersArmv5, FormattersThumbv2, true}
	ARMv5TExP = Architecture{"ARMv5TExP", ArmVersion5, ThumbVersion2, EnhancedDsp5TExP, FormattersArmv5, FormattersThumbv2, true}
	ARMv5TE   = Architecture{"ARMv5TE", ArmVersion5, ThumbVersion2, EnhancedDsp, FormattersArmv5, FormattersThumbv2, true}
)

type ArmVersion int

const (
	ArmVersionNone ArmVersion = iota
	ArmVersion1
	ArmVersion2
	ArmVersion3
	ArmVersion4
	ArmVersion5
)

type ThumbVersion int

const (
	ThumbVersionNone ThumbVersion = iota
	ThumbVersion1
	ThumbVersion2
)

type EnhancedDspFeature int

const (
	EnhancedDspNone EnhancedDspFeature = iota
	EnhancedDsp
	EnhancedDsp5TExP
)

var OrdersLookup = map[string]binary.ByteOrder{
	"little": binary.LittleEndian,
	"big":    binary.BigEndian,
	"native": binary.NativeEndian,
}

// objdump available versions
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
	Arch                               Architecture
	Order                              binary.ByteOrder
	RegisterNaming                     RegisterNaming
	UndefinedOutput                    string
	Thumb, OutputAddress, OutputOpcode bool
	Offset                             int
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
		if isa.Arch.Formatters == nil {
			return nil, fmt.Errorf("arm arch version %s does not support arm instructions", isa.Arch.Name)
		}

		w = 4
		unknown = INST_UNKNOWN
		formatters = isa.Arch.Formatters
		decoderFunc = func(b []byte, isa *Isa, decoder Decoder, op uint32, addr int) []byte {
			return DecodersArm[decoder](b, isa, op, addr)
		}
	} else {

		if isa.Arch.ThumbFormatters == nil {
			return nil, fmt.Errorf("arm arch version %s does not support thumb instructions", isa.Arch.Name)
		}

		w = 2
		unknown = INST_THUMB_UNKNOWN
		formatters = isa.Arch.ThumbFormatters
		decoderFunc = func(b []byte, isa *Isa, decoder Decoder, op uint32, addr int) []byte {
			return DecodersThumb[decoder](b, isa, uint16(op), addr)
		}
	}

	if leftover := len(bs) & (w - 1); leftover != 0 {
		return nil, fmt.Errorf("invalid byte cnt need mod %d, have rem %d", w, leftover)
	}

	// NOTE: start and end are used to only allocate strings once, removes per op allocations
	var (
		cnt     = len(bs) / w
		buf     = make([]byte, 0, cnt*256)
		ends    = make([]uint32, 0, cnt)
		instIdx = 0
	)

	for currByte := 0; currByte < len(bs); {

		doDecoding := true
		addr := isa.Offset + currByte

		var op uint32

		var instWidth int
		if w == 2 {
			instWidth = 2
			op = uint32(isa.Order.Uint16(bs[currByte:]))

			if op&0xF800 == 0xF000 && currByte+4 <= len(bs) {
				op2 := uint32(isa.Order.Uint16(bs[currByte+2:]))

				if op2&0xE800 == 0xE800 {
					var ok bool

					op = (op << 16) | op2

					buf = appendPrefixes(buf, isa, op, addr)
					buf, ok = thumbBlLong(buf, isa, op, addr)
					if ok {
						instWidth = 4
					}

					// skip jump and read next instruction as normal instruction
					doDecoding = false
				}
			}
		} else {
			instWidth = 4
			op = isa.Order.Uint32(bs[currByte:])
		}

		if doDecoding {

			decoder := unknown
			for j, desc := range formatters {

				if unused := desc.mask == 0 && desc.match == 0; unused {
					continue
				}

				if (op & desc.mask) == desc.match {
					decoder = Decoder(j)
					break
				}
			}
			buf = appendPrefixes(buf, isa, op, addr)
			buf = decoderFunc(buf, isa, decoder, op, addr)
		}

		ends = append(ends, uint32(len(buf)))
		currByte += instWidth
		instIdx++
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

func appendPrefixes(b []byte, isa *Isa, op uint32, addr int) []byte {
	if isa.OutputAddress {
		b = append(b, fmt.Sprintf("0x%08X: ", addr)...)
	}

	if isa.OutputOpcode {
		b = append(b, fmt.Sprintf("%08X: ", op)...)
	}

	return b
}
