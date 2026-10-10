package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"time"
)

var (
	t time.Time
	f *os.File
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
