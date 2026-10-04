package main

import (
	"encoding/binary"
	"fmt"

	"edisa/isa/arm32"
)

func main() {
	args := NewArgsUint32()

	isa := arm32.Arm32{
		Order:          binary.LittleEndian,
		Thumb:          false,
		RegisterNaming: arm32.GeneralNaming,
		BlockSuffixes:  arm32.StackBlockSuffixes,
	}

	asms := isa.Decode(args)

	for i := range asms {
		fmt.Printf("%s\n", asms[i])
	}
}
