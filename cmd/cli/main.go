package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"runtime/pprof"
	"time"

	"queso"
)

func main() {
	var f io.Reader
	path := os.Args[1]

	f, err := os.Open(path)
	if err != nil {
		f = NewArgsUint32()
	}

	w := 4

	isa := queso.Isa{
		Order:              binary.LittleEndian,
		Thumb:              w == 2,
		RegisterNaming:     queso.ConventionNaming,
		ImmediateValueBase: queso.HEX_BASE,
	}

	if profile {
		StartProfiler()
	}

	asms := isa.Decode(f)

	if profile {
		EndProfiler()
	}

	for i := range asms[:16] {
		fmt.Printf("%08X: %s\n", i*w, asms[i])
	}
}

var (
	profile = true
	t       time.Time
	f       *os.File
)

func StartProfiler() {
	var err error
	f, err = os.Create("profile.pprof")
	if err != nil {
		panic(err)
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		panic(err)
	}

	fmt.Printf("Starting Profiler...\n")
	t = time.Now()
}

func EndProfiler() {
	pprof.StopCPUProfile()
	if err := f.Close(); err != nil {
		panic(err)
	}

	fmt.Printf("Closing Profiler: Duration: %.2f\n", time.Since(t).Seconds())
}
