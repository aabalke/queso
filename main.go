package main

import (
	"fmt"
	"os"
	"strconv"

	"edisa/isa/arm32"
)

func main() {
	args := os.Args[1:]

	var insts []uint32

	for _, s := range args {
		v, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			panic(err)
		}

		insts = append(insts, uint32(v))
	}

	isa := arm32.Arm32{
		Thumb: true,
	}

	asms := isa.Decode(insts)

	for i := range asms {
		fmt.Printf("OP %08X: %s\n", insts[i], asms[i])
	}
}
