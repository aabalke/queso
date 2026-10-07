package queso

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"
)

type Isa struct {
	Order              binary.ByteOrder
	RegisterNaming     RegisterNaming
	BlockSuffixes      BlockSuffixes
	ImmediateValueBase HexImmediateBase
	UndefinedOutput    string
	Thumb              bool
}

func (isa *Isa) Decode(r io.Reader) []string {
	bs, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}

	w := 4
	unknown := INST_UNKNOWN
	formatters := formatters32
	decoderFunc := func(b []byte, isa *Isa, decoder Decoder, op uint32, addr int) []byte {
		return decoders32[decoder](b, isa, op, addr)
	}
	if isa.Thumb {
		w = 2
		unknown = INST_THUMB_UNKNOWN
		formatters = formatters16
		decoderFunc = func(b []byte, isa *Isa, decoder Decoder, op uint32, addr int) []byte {
			return decoders16[decoder](b, isa, uint16(op), addr)
		}
	}

	if leftover := len(bs) & (w - 1); leftover != 0 {
		panic(fmt.Sprintf("invalid byte cnt need mod %d, have rem %d", w, leftover))
	}

	cnt := len(bs) / w

	// NOTE: start and end are used to only allocate strings once, removes per op allocations
	buf := make([]byte, 0, cnt*256)
	ends := make([]uint32, cnt)
	ops := make([]uint32, cnt)

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
			if (v & desc.mask) == desc.match {
				decoder = Decoder(j)
				break
			}
		}

		buf = decoderFunc(buf, isa, decoder, v, i*w)
		ends[i] = uint32(len(buf))
	}

	s := convertBytesToStringArray(buf, ends)

	//for i := range 128 {
	//	fmt.Printf("%08X %s\n", ops[i], s[i])
	//}

	return s
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
