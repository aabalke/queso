package queso

import (
	"math/bits"
	"strconv"
)

var aluOpsThumb = [16]string{
	"ands",
	"eors",
	"lsls",
	"lsrs",
	"asrs",
	"adcs",
	"sbcs",
	"rors",
	"tst",
	"negs",
	"cmp",
	"cmn",
	"orrs",
	"muls",
	"bics",
	"mvns",
}

const (
	INST_THUMB_BRA Decoder = iota
	INST_THUMB_SWI
	INST_THUMB_BPT
	INST_THUMB_HI
	INST_THUMB_CBR
	INST_THUMB_REL
	INST_THUMB_SP
	INST_THUMB_PCR
	INST_THUMB_SDTR
	INST_THUMB_SDTI
	INST_THUMB_SDTH
	INST_THUMB_SPR
	INST_THUMB_BLK
	INST_THUMB_PUS
	INST_THUMB_ADD
	INST_THUMB_IMM
	INST_THUMB_SHI
	INST_THUMB_ALU
	INST_THUMB_EXT
	INST_THUMB_REV
	INST_THUMB_CBZ
	INST_THUMB_IT
	INST_THUMB_SET
	INST_THUMB_CPS
	INST_THUMB_UNKNOWN
	MAX_DECODER16_CNT
)

var FORMAT16_ARCH_ARM = []Formatter{
	INST_THUMB_BRA:  {0xF800, 0xE000},
	INST_THUMB_SWI:  {0xFF00, 0xDF00},
	INST_THUMB_BPT:  {0xFF00, 0xBE00},
	INST_THUMB_HI:   {0xFC00, 0x4400},
	INST_THUMB_CBR:  {0xF000, 0xD000},
	INST_THUMB_ALU:  {0xFC00, 0x4000},
	INST_THUMB_IMM:  {0xE000, 0x2000},
	INST_THUMB_ADD:  {0xF800, 0x1800},
	INST_THUMB_SHI:  {0xE000, 0x0000},
	INST_THUMB_PUS:  {0xF600, 0xB400},
	INST_THUMB_BLK:  {0xF000, 0xC000},
	INST_THUMB_REL:  {0xF000, 0xA000},
	INST_THUMB_SP:   {0xFF00, 0xB000},
	INST_THUMB_PCR:  {0xF800, 0x4800},
	INST_THUMB_SDTR: {0xF000, 0x5000},
	INST_THUMB_SDTI: {0xE000, 0x6000},
	INST_THUMB_SDTH: {0xF000, 0x8000},
	INST_THUMB_SPR:  {0xF000, 0x9000},
	INST_THUMB_EXT:  {0xFF00, 0xB200},
	INST_THUMB_REV:  {0xFF00, 0xBA00},
	INST_THUMB_CBZ:  {0xF500, 0xB100},
	INST_THUMB_IT:   {0xFF00, 0xBF00},
	INST_THUMB_SET:  {0xFFF7, 0xB650},
	INST_THUMB_CPS:  {0xFFE8, 0xB660},
}

var decoders16 = []func(b []byte, isa *Isa, op uint16, addr int) []byte{
	INST_THUMB_BRA:     thumbB,
	INST_THUMB_SWI:     thumbSvc,
	INST_THUMB_BPT:     thumbBkpt,
	INST_THUMB_HI:      thumbHi,
	INST_THUMB_CBR:     thumbCondB,
	INST_THUMB_ALU:     thumbAlu,
	INST_THUMB_IMM:     thumbImm,
	INST_THUMB_ADD:     thumbAdd,
	INST_THUMB_SHI:     thumbShift,
	INST_THUMB_PUS:     thumbPushPop,
	INST_THUMB_BLK:     thumbBlk,
	INST_THUMB_REL:     thumbRelative,
	INST_THUMB_SP:      thumbSp,
	INST_THUMB_PCR:     thumbPcRel,
	INST_THUMB_SDTR:    thumbSdtReg,
	INST_THUMB_SDTI:    thumbSdtImm,
	INST_THUMB_SDTH:    thumbSdtHalf,
	INST_THUMB_SPR:     thumbSpRel,
	INST_THUMB_EXT:     thumbExtend,
	INST_THUMB_REV:     thumbReverse,
	INST_THUMB_CBZ:     thumbCbz,
	INST_THUMB_IT:      thumbIt,
	INST_THUMB_SET:     thumbSetEnd,
	INST_THUMB_CPS:     thumbCps,
	INST_THUMB_UNKNOWN: thumbUnknown,
}

func thumbUnknown(b []byte, isa *Isa, op uint16, addr int) []byte {
	b = append(b, isa.UndefinedOutput...)
	return b
}

func thumbB(b []byte, isa *Isa, op uint16, addr int) []byte {
	nn := int(int16((op&0x7FF)<<5)>>4) + 4
	b = append(b, "b "...)
	b = appendAddr(b, addr, nn)
	return b
}

func thumbSvc(b []byte, isa *Isa, op uint16, addr int) []byte {
	b = append(b, "svc "...)
	// TODO: this matches objdump, not sure if should be controllable
	b = strconv.AppendUint(b, uint64(op&0xFF), 10)
	return b
}

func thumbBkpt(b []byte, isa *Isa, op uint16, addr int) []byte {
	b = append(b, "bkpt 0x00"...)
	// TODO: this matches objdump, not sure if should be controllable

	if op&0xFF < 0x10 {
		b = append(b, "0"...)
	}
	b = appendHex(b, uint64(op&0xFF))
	return b
}

func thumbHi(b []byte, isa *Isa, op uint16, addr int) []byte {
	rd := isa.RegisterNaming[(op&7)|(((op>>7)&1)<<3)]
	rs := isa.RegisterNaming[(op>>3)&0xF]

	if op == 0x46C0 {
		b = append(b, "nop"...)
		return b
	}

	inst := (op >> 8) & 3
	if inst == 3 {
		if op&0x80 != 0 {

			if op&7 != 0 {
				b = append(b, isa.UndefinedOutput...)
				return b
			}

			b = append(b, "blx "...)
		} else {
			b = append(b, "bx "...)
		}

		b = append(b, rs...)

		return b
	}

	switch inst {
	case 0:
		b = append(b, "add "...)
	case 1:
		b = append(b, "cmp "...)
	case 2:
		b = append(b, "mov "...)
	}

	b = append(b, rd...)
	b = append(b, ", "...)
	b = append(b, rs...)
	return b
}

func thumbCondB(b []byte, isa *Isa, op uint16, addr int) []byte {
	// TODO: should labels be abs or relative? +/- vs exact location
	b = append(b, "b"...)
	b = appendCond(b, uint32(op>>8)&0xF)
	b = append(b, ' ')
	nn := (int(int8(op&0xFF)) << 1) + 4
	b = appendAddr(b, addr, nn)
	return b
}

func thumbAlu(b []byte, isa *Isa, op uint16, addr int) []byte {
	rd := isa.RegisterNaming[(op>>0)&7]
	rs := isa.RegisterNaming[(op>>3)&7]
	b = append(b, aluOpsThumb[(op>>6)&0xF]...)
	b = append(b, ' ')
	b = append(b, rd...)
	b = append(b, ", "...)
	b = append(b, rs...)
	return b
}

func thumbImm(b []byte, isa *Isa, op uint16, addr int) []byte {
	rd := isa.RegisterNaming[(op>>8)&7]
	nn := int(op & 0xFF)

	switch (op >> 11) & 3 {
	case 0:
		b = append(b, "movs "...)
	case 1:
		b = append(b, "cmp "...)
	case 2:
		b = append(b, "adds "...)
	case 3:
		b = append(b, "subs "...)
	}

	b = append(b, rd...)
	b = append(b, ", "...)
	b = appendImm(b, nn, isa.ImmediateValueBase)
	return b
}

func thumbAdd(b []byte, isa *Isa, op uint16, addr int) []byte {
	isImm := op&(1<<10) != 0
	sub := op&(1<<9) != 0
	n := (op >> 6) & 7
	rs := isa.RegisterNaming[(op>>3)&7]
	rd := isa.RegisterNaming[(op>>0)&7]

	//if isImm && n == 0 && !sub {
	//	b = append(b, "mov "...)
	//	b = append(b, rd...)
	//	b = append(b, ", "...)
	//	b = append(b, rs...)
	//	return b
	//}

	if sub {
		b = append(b, "subs "...)
	} else {
		b = append(b, "adds "...)
	}

	b = append(b, rd...)
	b = append(b, ", "...)
	b = append(b, rs...)
	b = append(b, ", "...)

	if isImm {
		b = appendImm(b, int(n), isa.ImmediateValueBase)
	} else {
		b = append(b, isa.RegisterNaming[n]...)
	}

	return b
}

func thumbShift(b []byte, isa *Isa, op uint16, addr int) []byte {
	opc := (op >> 11) & 3
	offset := (op >> 6) & 0x1F
	rs := isa.RegisterNaming[(op>>3)&7]
	rd := isa.RegisterNaming[(op>>0)&7]

	if offset == 0 {
		if opc == 0 {
			b = append(b, "movs "...)
			b = append(b, rd...)
			b = append(b, ", "...)
			b = append(b, rs...)
			return b
		} else {
			offset = 32
		}
	}

	b = append(b, shifts[opc]...)
	b = append(b, "s "...)
	b = append(b, rd...)
	b = append(b, ", "...)
	b = append(b, rs...)
	b = append(b, ", "...)
	b = appendImm(b, int(offset), isa.ImmediateValueBase)
	return b
}

func thumbPushPop(b []byte, isa *Isa, op uint16, addr int) []byte {
	if pop := op&(1<<11) != 0; pop {
		b = append(b, "pop {"...)
	} else {
		b = append(b, "push {"...)
	}

	for rlist := op & 0x1FF; rlist&0xFF != 0; {
		i := bits.TrailingZeros16(rlist)
		rlist &= rlist - 1
		b = append(b, isa.RegisterNaming[i]...)
		if rlist != 0 {
			b = append(b, ", "...)
		}
	}

	if op&(1<<8) != 0 {
		if pop := op&(1<<11) != 0; pop {
			b = append(b, isa.RegisterNaming[15]...)
		} else {
			b = append(b, isa.RegisterNaming[14]...)
		}
	}

	b = append(b, "}"...)

	return b
}

func thumbBlk(b []byte, isa *Isa, op uint16, addr int) []byte {
	pop := op&(1<<11) != 0
	if pop {
		b = append(b, "ldmia "...)
	} else {
		b = append(b, "stmia "...)
	}

	rb := isa.RegisterNaming[(op>>8)&7]
	b = append(b, rb...)
	if included := op&(1<<((op>>8)&7)) != 0; !included || !pop {
		b = append(b, "!"...)
	}

	b = append(b, ", {"...)

	for rlist := op & 0xFF; rlist != 0; {
		i := bits.TrailingZeros16(rlist)
		rlist &= rlist - 1
		b = append(b, isa.RegisterNaming[i]...)
		if rlist != 0 {
			b = append(b, ", "...)
		}
	}

	b = append(b, "}"...)
	return b
}

func thumbRelative(b []byte, isa *Isa, op uint16, addr int) []byte {
	nn := int((op & 0xFF) << 2)
	rd := isa.RegisterNaming[(op>>8)&7]

	b = append(b, "add "...)
	b = append(b, rd...)
	b = append(b, ", "...)

	if pc := op&(1<<11) == 0; pc {
		b = append(b, isa.RegisterNaming[15]...)
	} else {
		b = append(b, isa.RegisterNaming[13]...)
	}

	b = append(b, ", "...)
	b = appendImm(b, nn, isa.ImmediateValueBase)

	return b
}

func thumbSp(b []byte, isa *Isa, op uint16, addr int) []byte {
	nn := int((op & 0x7F) << 2)

	if op&0x80 != 0 {
		b = append(b, "sub "...)
	} else {
		b = append(b, "add "...)
	}

	b = append(b, isa.RegisterNaming[13]...)
	b = append(b, ", "...)
	b = appendImm(b, nn, isa.ImmediateValueBase)
	return b
}

func thumbSdtReg(b []byte, isa *Isa, op uint16, addr int) []byte {
	ro := isa.RegisterNaming[(op>>6)&7]
	rb := isa.RegisterNaming[(op>>3)&7]
	rd := isa.RegisterNaming[(op>>0)&7]

	switch (op >> 9) & 7 {
	case 0:
		b = append(b, "str "...)
	case 1:
		b = append(b, "strh "...)
	case 2:
		b = append(b, "strb "...)
	case 3:
		b = append(b, "ldrsb "...)
	case 4:
		b = append(b, "ldr "...)
	case 5:
		b = append(b, "ldrh "...)
	case 6:
		b = append(b, "ldrb "...)
	case 7:
		b = append(b, "ldrsh "...)
	}
	b = append(b, rd...)
	b = append(b, ", ["...)
	b = append(b, rb...)
	b = append(b, ", "...)
	b = append(b, ro...)
	b = append(b, "]"...)
	return b
}

func thumbSdtImm(b []byte, isa *Isa, op uint16, addr int) []byte {
	nn := int((op >> 6) & 0x1F)
	rb := isa.RegisterNaming[(op>>3)&7]
	rd := isa.RegisterNaming[(op>>0)&7]

	switch (op >> 11) & 3 {
	case 0:
		b = append(b, "str "...)
		nn *= 4
	case 1:
		b = append(b, "ldr "...)
		nn *= 4
	case 2:
		b = append(b, "strb "...)
	case 3:
		b = append(b, "ldrb "...)
	}

	b = append(b, rd...)
	b = append(b, ", ["...)
	b = append(b, rb...)
	b = append(b, ", "...)
	b = appendImm(b, nn, isa.ImmediateValueBase)
	b = append(b, "]"...)
	return b
}

func thumbSdtHalf(b []byte, isa *Isa, op uint16, addr int) []byte {
	nn := int((op>>6)&0x1F) * 2
	rb := isa.RegisterNaming[(op>>3)&7]
	rd := isa.RegisterNaming[(op>>0)&7]

	if op&(1<<11) != 0 {
		b = append(b, "ldrh "...)
	} else {
		b = append(b, "strh "...)
	}

	b = append(b, rd...)
	b = append(b, ", ["...)
	b = append(b, rb...)
	b = append(b, ", "...)
	b = appendImm(b, nn, isa.ImmediateValueBase)
	b = append(b, "]"...)
	return b
}

func thumbPcRel(b []byte, isa *Isa, op uint16, addr int) []byte {
	nn := int((op & 0xFF) * 4)
	rd := isa.RegisterNaming[(op>>8)&7]

	b = append(b, "ldr "...)
	b = append(b, rd...)
	b = append(b, ", ["...)
	b = append(b, isa.RegisterNaming[15]...)
	b = append(b, ", "...)
	b = appendImm(b, nn, isa.ImmediateValueBase)
	b = append(b, "]"...)
	return b
}

func thumbSpRel(b []byte, isa *Isa, op uint16, addr int) []byte {
	nn := int((op & 0xFF) * 4)
	rd := isa.RegisterNaming[(op>>8)&7]

	if op&(1<<11) != 0 {
		b = append(b, "ldr "...)
	} else {
		b = append(b, "str "...)
	}
	b = append(b, rd...)
	b = append(b, ", ["...)
	b = append(b, isa.RegisterNaming[13]...)
	b = append(b, ", "...)
	b = appendImm(b, nn, isa.ImmediateValueBase)
	b = append(b, "]"...)
	return b
}

func thumbExtend(b []byte, isa *Isa, op uint16, _ int) []byte {
	rm := isa.RegisterNaming[(op>>3)&7]
	rd := isa.RegisterNaming[(op>>0)&7]

	switch inst := (op >> 6) & 3; inst {
	case 0:
		b = append(b, "sxth "...)
	case 1:
		b = append(b, "sxtb "...)
	case 2:
		b = append(b, "uxth "...)
	case 3:
		b = append(b, "uxtb "...)
	}

	b = append(b, rd...)
	b = append(b, ", "...)
	b = append(b, rm...)

	return b
}

func thumbReverse(b []byte, isa *Isa, op uint16, _ int) []byte {
	rm := isa.RegisterNaming[(op>>3)&7]
	rd := isa.RegisterNaming[(op>>0)&7]

	switch inst := (op >> 6) & 3; inst {
	case 0:
		b = append(b, "rev "...)
	case 1:
		b = append(b, "rev16 "...)
	case 2:
		b = append(b, isa.UndefinedOutput...)
		return b
	case 3:
		b = append(b, "revsh "...)
	}

	b = append(b, rd...)
	b = append(b, ", "...)
	b = append(b, rm...)
	return b
}

func thumbCbz(b []byte, isa *Isa, op uint16, addr int) []byte {
	rn := isa.RegisterNaming[(op>>0)&7]

	nn := int(((op >> 3) & 0x1F))

	if op&(1<<9) != 0 {
		nn += 32
	}

	nn *= 2
	nn += 4

	if op&(1<<11) != 0 {
		b = append(b, "cbnz "...)
	} else {
		b = append(b, "cbz "...)
	}

	b = append(b, rn...)
	b = append(b, ", "...)

	//b = appendAddr(b, nn, 0)
	//b = append(b, ", "...)
	//nn = 0
	b = appendAddr(b, addr, nn)
	return b
}

func thumbIt(b []byte, isa *Isa, op uint16, addr int) []byte {
	cond := int(op>>4) & 0xf
	mask := uint(op) & 0xf

	// mask == 0 is a hint instruction (nop/yield/wfe/...), not IT.
	// Guard it here so the trailing-zero loop below can't spin forever.
	if mask == 0 {
		switch cond {
		case 0:
			b = append(b, "nop"...)
		case 1:
			b = append(b, "yield"...)
		case 2:
			b = append(b, "wfe"...)
		case 3:
			b = append(b, "wfi"...)
		case 4:
			b = append(b, "sev"...)
		default:
			b = append(b, isa.UndefinedOutput...)
		}
		return b
	}

	// "t"/"e" flags depend on the condition; rebuild the mask for the
	// "even" condition code (to match objdump output).
	// bit 4 = implied first-condition flag, bit 5 = start of IT block
	itMask := mask | uint((cond&1)<<4) | 0x20

	// Count the instructions in the block (1..4) by finding the lowest set bit.
	ccount := 3
	for mask&1 == 0 {
		ccount--
		mask >>= 1
	}

	mask = itMask & 0x0f
	b = append(b, "it"...)
	for ; ccount > 0; ccount-- {
		if int(mask>>3)&1 == cond&1 {
			b = append(b, 't')
		} else {
			b = append(b, 'e')
		}
		mask = (mask << 1) & 0x0f
	}

	b = append(b, ' ')

	switch cond {
	case 0xE:
		b = append(b, "al"...)
	case 0xF:
		// skip
	default:
		b = appendCond(b, uint32(cond))
	}
	return b
}

func thumbSetEnd(b []byte, isa *Isa, op uint16, addr int) []byte {
	if op&(1<<3) != 0 {
		b = append(b, "setend be"...)
	} else {
		b = append(b, "setend le"...)
	}

	return b
}

func thumbCps(b []byte, isa *Isa, op uint16, addr int) []byte {
	if op&(1<<4) != 0 {
		b = append(b, "cpsid"...)
	} else {
		b = append(b, "cpsie"...)
	}

	if op&7 != 0 {
		b = append(b, " "...)
	}
	if op&(1<<2) != 0 {
		b = append(b, "a"...)
	}
	if op&(1<<1) != 0 {
		b = append(b, "i"...)
	}
	if op&(1<<0) != 0 {
		b = append(b, "f"...)
	}

	return b
}
