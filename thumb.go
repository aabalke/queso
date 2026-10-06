package queso

import (
	"math/bits"
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
	INST_THUMB_UNKNOWN
	MAX_DECODER16_CNT
)

var formatters16 = []Formatter{
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
}

var decoders16 = []func(isa *Isa, op uint16, b []byte) []byte{
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
	INST_THUMB_UNKNOWN: thumbUnknown,
}

func thumbUnknown(isa *Isa, op uint16, b []byte) []byte {
	b = append(b, "unknown thumb"...)
	return b
}

func thumbB(isa *Isa, op uint16, b []byte) []byte {
	nn := int(int16((op&0x7FF)<<5)>>4) + 4
	b = append(b, "b "...)
	b = appendImm(b, nn, isa.ImmediateValueBase)
	return b
}

func thumbSvc(isa *Isa, op uint16, b []byte) []byte {
	b = append(b, "svc "...)
	b = appendImm(b, int(op&0xFF), isa.ImmediateValueBase)
	return b
}

func thumbBkpt(isa *Isa, op uint16, b []byte) []byte {
	b = append(b, "bkpt "...)
	b = appendImm(b, int(op&0xFF), isa.ImmediateValueBase)
	return b
}

func thumbHi(isa *Isa, op uint16, b []byte) []byte {
	rd := isa.RegisterNaming[(op&7)|(((op>>7)&1)<<3)]
	rs := isa.RegisterNaming[(op>>3)&0xF]

	if op == 0x46C0 {
		b = append(b, "nop"...)
		return b
	}

	inst := (op >> 8) & 3
	if inst == 3 {
		if op&0x80 != 0 {
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

func thumbCondB(isa *Isa, op uint16, b []byte) []byte {
	// TODO: should labels be abs or relative? +/- vs exact location
	b = append(b, "b"...)
	b = appendCond(b, uint32(op>>8)&0xF)
	b = append(b, ' ')
	nn := (int(int8(op&0xFF)) << 1) + 4
	b = appendImm(b, int(nn), isa.ImmediateValueBase)
	return b
}

func thumbAlu(isa *Isa, op uint16, b []byte) []byte {
	rd := isa.RegisterNaming[(op>>0)&7]
	rs := isa.RegisterNaming[(op>>3)&7]
	b = append(b, aluOpsThumb[(op>>6)&0xF]...)
	b = append(b, ' ')
	b = append(b, rd...)
	b = append(b, ", "...)
	b = append(b, rs...)
	return b
}

func thumbImm(isa *Isa, op uint16, b []byte) []byte {
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

func thumbAdd(isa *Isa, op uint16, b []byte) []byte {
	isImm := op&(1<<10) != 0
	sub := op&(1<<9) != 0
	n := (op >> 6) & 7
	rs := isa.RegisterNaming[(op>>3)&7]
	rd := isa.RegisterNaming[(op>>0)&7]

	if isImm && n == 0 && !sub {
		b = append(b, "mov "...)
		b = append(b, rd...)
		b = append(b, ", "...)
		b = append(b, rs...)
		return b
	}

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

func thumbShift(isa *Isa, op uint16, b []byte) []byte {
	opc := (op >> 11) & 3
	offset := (op >> 6) & 0x1F
	rs := isa.RegisterNaming[(op>>3)&7]
	rd := isa.RegisterNaming[(op>>0)&7]

	if offset == 0 && opc > 0 {
		offset = 32
	}

	b = append(b, shifts[opc]...)
	b = append(b, " "...)
	b = append(b, rd...)
	b = append(b, ", "...)
	b = append(b, rs...)
	b = append(b, ", "...)
	b = appendImm(b, int(offset), isa.ImmediateValueBase)
	return b
}

func thumbPushPop(isa *Isa, op uint16, b []byte) []byte {
	if pop := op&(1<<11) != 0; pop {
		b = append(b, "push "...)
	} else {
		b = append(b, "pop "...)
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

func thumbBlk(isa *Isa, op uint16, b []byte) []byte {
	if pop := op&(1<<11) != 0; pop {
		b = append(b, "ldmia "...)
	} else {
		b = append(b, "stmia "...)
	}

	rb := isa.RegisterNaming[(op>>8)&7]
	b = append(b, rb...)
	b = append(b, "!, {"...)

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

func thumbRelative(isa *Isa, op uint16, b []byte) []byte {
	nn := int((op & 0xFF) << 2)
	rd := isa.RegisterNaming[(op>>8)&7]

	b = append(b, "add "...)
	b = append(b, rd...)
	b = append(b, ", "...)

	if pc := op&(1<<11) == 0; pc {
		nn += 4
		b = append(b, isa.RegisterNaming[15]...)
	} else {
		b = append(b, isa.RegisterNaming[13]...)
	}

	b = append(b, ", "...)
	b = appendImm(b, nn, isa.ImmediateValueBase)

	return b
}

func thumbSp(isa *Isa, op uint16, b []byte) []byte {
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

func thumbSdtReg(isa *Isa, op uint16, b []byte) []byte {
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
		b = append(b, "ldsb "...)
	case 4:
		b = append(b, "ldr "...)
	case 5:
		b = append(b, "ldrh "...)
	case 6:
		b = append(b, "ldrb "...)
	case 7:
		b = append(b, "ldsh "...)
	}
	b = append(b, rd...)
	b = append(b, ", ["...)
	b = append(b, rb...)
	b = append(b, ", "...)
	b = append(b, ro...)
	b = append(b, "]"...)
	return b
}

func thumbSdtImm(isa *Isa, op uint16, b []byte) []byte {
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

func thumbSdtHalf(isa *Isa, op uint16, b []byte) []byte {
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

func thumbPcRel(isa *Isa, op uint16, b []byte) []byte {
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

func thumbSpRel(isa *Isa, op uint16, b []byte) []byte {
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
