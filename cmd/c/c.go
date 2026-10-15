package main

import "C"
import (
	"bytes"
	"encoding/binary"

	"queso"
)

//export decode
func decode() C.int {
	isa := queso.Isa{
		Arch:           queso.ARCH_ARM_ANY,
		Order:          binary.LittleEndian,
		RegisterNaming: queso.ConventionNaming,
	}

	b := binary.BigEndian.AppendUint32([]byte{}, 0x0eF0B0E1)

	vs, _ := isa.Decode(bytes.NewReader(b))
	for i := range len(vs) {
		println("vs", vs[i])
	}

	return 0
}

func main() {}
