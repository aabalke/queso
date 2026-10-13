package queso

import (
	"math/bits"
)

type HexImmediateBase int

const (
	DEC_BASE HexImmediateBase = 10
	HEX_BASE HexImmediateBase = 16
)

type RegisterNaming [16]string

var GeneralNaming = RegisterNaming{
	"r0",
	"r1",
	"r2",
	"r3",
	"r4",
	"r5",
	"r6",
	"r7",
	"r8",
	"r9",
	"r10",
	"r11",
	"r12",
	"r13",
	"r14",
	"r15",
}

var ConventionNaming = RegisterNaming{
	"r0",
	"r1",
	"r2",
	"r3",
	"r4",
	"r5",
	"r6",
	"r7",
	"r8",
	"r9",
	"sl",
	"fp",
	"ip",
	"sp",
	"lr",
	"pc",
}

type Decoder int

type Formatter struct {
	mask, match uint32
}

func appendHex(dst []byte, v uint32) []byte {
	const digits = "0123456789abcdef"
	const base = 16
	var tmp [16]byte // max 64bit value would be 16bytes in ascii hex
	i := len(tmp)
	shift := uint(bits.TrailingZeros(uint(base)))
	m := uint(base) - 1 // == 1<<shift - 1
	for v >= base {
		i--
		tmp[i] = digits[uint(v)&m]
		v >>= shift
	}
	// v < base
	i--
	tmp[i] = digits[uint(v)]

	dst = append(dst, tmp[i:]...)
	return dst
}

var cond = [16]string{
	"eq",
	"ne",
	"cs",
	"cc",
	"mi",
	"pl",
	"vs",
	"vc",
	"hi",
	"ls",
	"ge",
	"lt",
	"gt",
	"le",
	"",
	"nv",
}

func appendCond(b []byte, c uint32) []byte {
	if c != 0xE {
		b = append(b, cond[c]...)
	}
	return b
}

func appendAddr(b []byte, addr, absOffset, relOffset int) []byte {
	b = append(b, "0x"...)
	b = appendHex(b, uint32(addr+absOffset+relOffset))
	return b
}
