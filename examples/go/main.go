package main

import (
	"bytes"
	"encoding/binary"

	"queso"
)

func main() {
	isa := queso.Isa{
		Arch:           queso.ARMv5TE,
		Order:          binary.LittleEndian,
		RegisterNaming: queso.ConventionNaming,
		OutputAddress:  true,
		OutputOpcode:   true,
	}

	b := []byte{
		0x0E, 0xF0, 0xB0, 0xE1, 0x91, 0x00, 0x02, 0xE1,
	}

	vs, err := isa.Decode(bytes.NewReader(b))
	if err != nil {
		panic(err)
	}

	for _, v := range vs {
		println(v)
	}

	// output:
	// 0x00000000: E1B0F00E: movs pc, lr
	// 0x00000004: E1020091: swp r0, r1, [r2]
}
