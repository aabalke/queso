package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"queso"
)

const version = "0.0.1"

var orders = map[string]binary.ByteOrder{
	"little": binary.LittleEndian,
	"big":    binary.BigEndian,
	"native": binary.NativeEndian,
}

var arch = map[string]queso.Arch{
	"armv2":   queso.ARCH_ARMv2,
	"armv3":   queso.ARCH_ARMv3,
	"armv4t":  queso.ARCH_ARMv4t,
	"armv5te": queso.ARCH_ARMv5te,
}

func main() {
	var (
		versionPtr   = flag.Bool("version", false, "get version")
		profilePtr   = flag.Bool("profile", false, "run profiler")
		addrPtr      = flag.Bool("addr", false, "output address with disassembly")
		opcodePtr    = flag.Bool("opcode", false, "output opcode with disassembly")
		thumbPtr     = flag.Bool("thumb", false, "force thumb")
		orderPtr     = flag.String("order", "little", "byte order / endianness (\"little\", \"big\", or \"native\")")
		regsPtr      = flag.String("regs", "convention", "register names, \"convention\" (lr, pc), \"general\" (r14, r15) or custom comma separated values \"r0,r1,r2,specialr4,r5...\"")
		archPtr      = flag.String("arch", "arm_any", "architecture version (\"arm_any\", \"armv2\", \"armv3\", \"armv4t\", \"armv5te\")")
		inputPtr     = flag.String("input", "", "input binary file path")
		outputPtr    = flag.String("output", "", "output path")
		undefinedPtr = flag.String("undefined", "unknown instruction", "output on undefined instruction")
	)

	flag.Parse()

	if *versionPtr {
		fmt.Printf("queso disassembler version %s", version)
		return
	}

	order, ok := orders[strings.ToLower(*orderPtr)]
	if !ok {
		fmt.Printf("order/endianness %s is not valid, using little\n", *orderPtr)
		order = binary.LittleEndian
	}

	arch, ok := arch[strings.ToLower(*archPtr)]
	if !ok {
		fmt.Printf("arch version %s is not valid, using arm_any\n", *archPtr)
		arch = queso.ARCH_ARM_ANY
	}

	var registers queso.RegisterNaming
	switch *regsPtr {
	case "convention":
		registers = queso.ConventionNaming
	case "general":
		registers = queso.GeneralNaming
	default:
		registers = queso.ConventionNaming
		copy(registers[:], strings.Split(*regsPtr, ","))
	}

	isa := queso.Isa{
		Order:           order,
		Thumb:           *thumbPtr,
		Arch:            arch,
		RegisterNaming:  registers,
		OutputAddress:   *addrPtr,
		OutputOpcode:    *opcodePtr,
		UndefinedOutput: *undefinedPtr,
	}

	var out *os.File

	if *outputPtr != "" {
		var err error
		out, err = os.OpenFile(*outputPtr, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0o644)
		if err != nil {
			panic(err)
		}
	}

	var f io.Reader
	if *inputPtr != "" {
		var err error
		f, err = os.Open(*inputPtr)
		if err != nil {
			panic("could not open input file path")
		}

	} else {
		f = NewArgsUint32(flag.Args(), *thumbPtr)
	}

	var asms []string
	if *profilePtr {
		StartProfiler()
		asms = isa.Decode(f)
		EndProfiler()
	} else {
		asms = isa.Decode(f)
	}

	var s []byte
	for _, asm := range asms {
		s = append(s, asm...)
		s = append(s, "\n"...)
	}

	if *outputPtr != "" {
		if _, err := out.Write(s); err != nil {
			panic(err)
		}
	} else {
		fmt.Print(string(s))
	}
}
