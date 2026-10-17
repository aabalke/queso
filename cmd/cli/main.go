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

func main() {
	var (
		versionPtr    = flag.Bool("version", false, "get version")
		profilePtr    = flag.Bool("profile", false, "run profiler")
		addrPtr       = flag.Bool("addr", false, "output address with disassembly")
		opcodePtr     = flag.Bool("opcode", false, "output opcode with disassembly")
		thumbPtr      = flag.Bool("thumb", false, "force thumb")
		orderPtr      = flag.String("order", "little", "byte order / endianness (\"little\", \"big\", or \"native\")")
		regsPtr       = flag.String("regs", "convention", "register names, \"convention\" (lr, pc), \"general\" (r14, r15) or custom comma separated values \"r0,r1,r2,specialr4,r5...\"")
		archPtr       = flag.String("arch", "ARMv5TE", "architecture version (\"armv2\", \"armv3\", \"armv4tm\", \"armv5te\")")
		inputPtr      = flag.String("input", "", "input binary file path")
		outputPtr     = flag.String("output", "", "output path")
		undefinedPtr  = flag.String("undefined", "unknown instruction", "output on undefined instruction")
		addrOffsetPtr = flag.Int("offset", 0, "address offset")
	)

	// TODO: print help if no args

	flag.Parse()

	if *versionPtr {
		fmt.Printf("queso disassembler version %s", version)
		return
	}

	order, ok := queso.OrdersLookup[strings.ToLower(*orderPtr)]
	if !ok {
		fmt.Printf("order/endianness %s is not valid, using little\n", *orderPtr)
		order = binary.LittleEndian
	}

	arch, ok := queso.ArchLookup[strings.ToLower(*archPtr)]
	if !ok {
		fmt.Printf("arch version %s is not valid, using arm_any\n", *archPtr)
		arch = queso.ARMv5TE
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
		Offset:          *addrOffsetPtr,
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
	var err error
	if *inputPtr != "" {
		f, err = os.Open(*inputPtr)
		if err != nil {
			panic("could not open input file path")
		}

	} else {
		f, err = queso.NewInstReaderFromStrings(flag.Args())
		if err != nil {
			panic(err)
		}
	}

	var asms []string
	if *profilePtr {
		StartProfiler()
		asms, err = isa.Decode(f)
		if err != nil {
			asms = []string{err.Error()}
		}
		EndProfiler()
	} else {
		asms, err = isa.Decode(f)
		if err != nil {
			asms = []string{err.Error()}
		}
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
