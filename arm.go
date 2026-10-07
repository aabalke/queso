package queso

import (
	"math/bits"
	"strconv"
)

var shifts = [4]string{"lsl", "lsr", "asr", "ror"}

var aluOpsArm = [16]string{
	"and",
	"eor",
	"sub",
	"rsb",
	"add",
	"adc",
	"sbc",
	"rsc",
	"tst",
	"teq",
	"cmp",
	"cmn",
	"orr",
	"mov",
	"bic",
	"mvn",
}

const (
	INST_SVC Decoder = iota
	INST_EXT
	INST_BKP
	INST_COP
	INST_BRA
	INST_BRX
	INST_SAT
	INST_SAT16
	INST_USA
	INST_PACK
	INST_REV
	INST_SDT
	INST_BLK
	INST_MUL
	INST_HAL
	INST_UND
	INST_MRS
	INST_MSR
	INST_SWP
	INST_HWM
	INST_ALU
	INST_ALU_SHIFT_REG
	INST_ALU_SHIFT_IMM
	INST_CLZ
	INST_QAL
	INST_CDP
	INST_UNKNOWN
	MAX_DECODER32_CNT
)

var formatters32 = []Formatter{
	INST_SVC:           {0x0F00_0000, 0x0F00_0000},
	INST_BKP:           {0xFFF0_00F0, 0xE120_0070},
	INST_BRA:           {0x0E00_0000, 0x0A00_0000},
	INST_BRX:           {0x0FFF_FF00, 0x012F_FF00},
	INST_UND:           {0x0E00_0010, 0x0600_0010},
	INST_MUL:           {0x0F00_00F0, 0x0000_0090},
	INST_HWM:           {0x0F90_0090, 0x0100_0080},
	INST_CLZ:           {0x0FFF_0F10, 0x016F_0F10},
	INST_QAL:           {0x0100_0FF0, 0x0100_0050},
	INST_SWP:           {0x0FB0_0FF0, 0x0100_0090},
	INST_BLK:           {0x0E00_0000, 0x0800_0000},
	INST_HAL:           {0x0E00_0090, 0x0000_0090},
	INST_SDT:           {0x0C00_0000, 0x0400_0000},
	INST_ALU:           {0x0E00_0000, 0x0200_0000},
	INST_ALU_SHIFT_REG: {0x0E00_0090, 0x0000_0010},
	INST_ALU_SHIFT_IMM: {0x0E00_0010, 0x0000_0000},
	INST_COP:           {0x0F00_0010, 0x0E00_0010},
	INST_MRS:           {0x0FBF_0FFF, 0x010F_0000},
	INST_MSR:           {0x0DB0_F000, 0x0120_F000},
	INST_CDP:           {0x0F00_0010, 0x0E00_0000},
	INST_EXT:           {0x0F80_03F0, 0x0680_0070},
	INST_SAT:           {0x0FA0_0030, 0x06A0_0010},
	INST_SAT16:         {0x0FA0_0FF0, 0x06A0_0F30},
	INST_USA:           {0x0FF0_00F0, 0x0780_0010},
	INST_PACK:          {0x0FF0_0000, 0x0680_0000},
	INST_REV:           {0x0FBF_0F70, 0x06BF_0F30},
}

var decoders32 = [MAX_DECODER32_CNT]func(b []byte, isa *Isa, op uint32, addr int) []byte{
	INST_UNKNOWN:       unknown,
	INST_SVC:           svc,
	INST_BKP:           bkpt,
	INST_UND:           und,
	INST_BRA:           br,
	INST_BRX:           bx,
	INST_MUL:           mul,
	INST_HWM:           hwm,
	INST_CLZ:           clz,
	INST_QAL:           qalu,
	INST_SWP:           swp,
	INST_BLK:           blk,
	INST_HAL:           half,
	INST_SDT:           sdt,
	INST_ALU_SHIFT_REG: aluShiftReg,
	INST_ALU_SHIFT_IMM: aluShiftImm,
	INST_ALU:           alu,
	INST_COP:           cop,
	INST_MRS:           mrs,
	INST_MSR:           msr,
	INST_CDP:           cdp,
	INST_EXT:           ext,
	INST_SAT:           sat,
	INST_SAT16:         sat16,
	INST_USA:           usad,
	INST_PACK:          pack,
	INST_REV:           rev,
}

func unknown(b []byte, isa *Isa, op uint32, addr int) []byte {
	b = append(b, isa.UndefinedOutput...)
	return b
}

func svc(b []byte, isa *Isa, op uint32, addr int) []byte {
	b = append(b, "svc"...)
	b = appendCond(b, op>>28)
	b = append(b, " 0x00000000"...)
	c := strconv.AppendUint([]byte{}, uint64(op&0xFFFFFF), 16)
	b = append(b[:len(b)-len(c)], c[:]...)

	//b = appendHex(b, uint64(uint32(op&0xFFFFFF)))
	//b = strconv.AppendUint(b, uint64(op&0xFFFFFF), 10)
	//b = appendImm(b, int(op&0xFF_FFFF), isa.ImmediateValueBase)
	return b
}

func bkpt(b []byte, isa *Isa, op uint32, addr int) []byte {
	b = append(b, "bkpt "...)
	b = appendImm(b, int((op&0xF)|(((op>>8)&0xFFF)<<4)), isa.ImmediateValueBase)
	return b
}

func und(b []byte, isa *Isa, op uint32, addr int) []byte {
	b = append(b, "undefined"...)
	return b
}

func br(b []byte, isa *Isa, op uint32, addr int) []byte {
	nn := int((int32(op)<<8)>>6) + 8

	switch {
	case op>>28 == 0xF:
		nn += int((op>>24)&1) * 2
		b = append(b, "blx "...)
	case op&(1<<24) != 0:
		b = append(b, "bl"...)
		b = appendCond(b, op>>28)
		b = append(b, ' ')
	default:
		b = append(b, "b"...)
		b = appendCond(b, op>>28)
		b = append(b, ' ')
	}

	b = appendAddr(b, addr, nn)
	return b
}

func bx(b []byte, isa *Isa, op uint32, addr int) []byte {
	rn := isa.RegisterNaming[op&0xF]

	switch inst := (op >> 4) & 0xF; inst {
	case 1:
		b = append(b, "bx"...)
	case 2:
		b = append(b, "bxj"...)
	case 3:
		b = append(b, "blx"...)
	default:
		b = append(b, "unknown branch exchange instruction"...)
		return b
	}

	b = appendCond(b, op>>28)
	b = append(b, ' ')
	b = append(b, rn...)

	return b
}

var mulOps = []string{
	"mul",
	"mla",
	"umaal",
	"",
	"umull",
	"umlal",
	"smull",
	"smlal",
}

func mul(b []byte, isa *Isa, op uint32, addr int) []byte {
	rd := isa.RegisterNaming[(op>>16)&0xF]
	rn := isa.RegisterNaming[(op>>12)&0xF]
	rs := isa.RegisterNaming[(op>>8)&0xF]
	rm := isa.RegisterNaming[(op>>0)&0xF]

	inst := (op >> 21) & 7

	// TODO: only armv6+
	if inst == 2 && op&(1<<20) != 0 {
		b = append(b, "undefined"...)
		return b
	}

	b = append(b, mulOps[inst]...)
	b = appendCond(b, op>>28)

	if inst != 2 && op&(1<<20) != 0 {
		b = append(b, 's')
	}

	b = append(b, ' ')

	switch inst {
	case 0, 1:
		b = append(b, rd...)
		b = append(b, ", "...)
		b = append(b, rm...)
		b = append(b, ", "...)
		b = append(b, rs...)

		if inst == 1 {
			b = append(b, ", "...)
			b = append(b, rn...)
		}
	default:
		b = append(b, rn...)
		b = append(b, ", "...)
		b = append(b, rd...)
		b = append(b, ", "...)
		b = append(b, rm...)
		b = append(b, ", "...)
		b = append(b, rs...)
	}

	return b
}

var hwmOps = []string{
	"smla",
	"",
	"smlal",
	"smul",
}

func hwm(b []byte, isa *Isa, op uint32, addr int) []byte {
	rd := isa.RegisterNaming[(op>>16)&0xF]
	rn := isa.RegisterNaming[(op>>12)&0xF]
	rs := isa.RegisterNaming[(op>>8)&0xF]
	rm := isa.RegisterNaming[(op>>0)&0xF]

	inst := (op >> 21) & 7

	if inst >= 4 {
		b = append(b, "unknown multiply instruction"...)
		return b
	}

	if inst == 1 {
		if op&(1<<5) != 0 {
			b = append(b, "smulw"...)
		} else {
			b = append(b, "smlaw"...)
		}
	} else {
		b = append(b, hwmOps[inst]...)

		if op&(1<<5) != 0 {
			b = append(b, 't')
		} else {
			b = append(b, 'b')
		}
	}

	if op&(1<<6) != 0 {
		b = append(b, 't')
	} else {
		b = append(b, 'b')
	}

	b = appendCond(b, op>>28)
	b = append(b, ' ')

	if inst == 2 {
		b = append(b, rn...)
		b = append(b, ", "...)
	}
	b = append(b, rd...)
	b = append(b, ", "...)
	b = append(b, rm...)
	b = append(b, ", "...)
	b = append(b, rs...)

	if inst == 0 || (inst == 1 && op&(1<<5) == 0) {
		b = append(b, ", "...)
		b = append(b, rn...)
	}

	return b
}

var qaluOps = []string{"qadd", "qsub", "qdadd", "qdsub"}

func qalu(b []byte, isa *Isa, op uint32, addr int) []byte {
	rn := isa.RegisterNaming[(op>>16)&0xF]
	rd := isa.RegisterNaming[(op>>12)&0xF]
	rm := isa.RegisterNaming[(op>>0)&0xF]
	inst := (op >> 20) & 0xF

	if inst >= 4 {
		b = append(b, "unknown qalu"...)
		return b
	}

	b = append(b, qaluOps[inst]...)
	b = appendCond(b, op>>28)
	b = append(b, " "...)
	b = append(b, rd...)
	b = append(b, ", "...)
	b = append(b, rm...)
	b = append(b, ", "...)
	b = append(b, rn...)

	return b
}

func clz(b []byte, isa *Isa, op uint32, addr int) []byte {
	rd := isa.RegisterNaming[(op>>12)&0xF]
	rm := isa.RegisterNaming[op&0xF]

	b = append(b, "clz"...)
	b = appendCond(b, op>>28)
	b = append(b, " "...)
	b = append(b, rd...)
	b = append(b, ", "...)
	b = append(b, rm...)

	return b
}

func swp(b []byte, isa *Isa, op uint32, addr int) []byte {
	rn := isa.RegisterNaming[(op>>16)&0xF]
	rd := isa.RegisterNaming[(op>>12)&0xF]
	rm := isa.RegisterNaming[(op>>0)&0xF]

	b = append(b, "swp"...)

	if op&(1<<22) != 0 {
		b = append(b, "b"...)
	}
	b = appendCond(b, op>>28)
	b = append(b, " "...)
	b = append(b, rd...)
	b = append(b, ", "...)
	b = append(b, rm...)
	b = append(b, ", ["...)
	b = append(b, rn...)
	b = append(b, "]"...)

	return b
}

func blk(b []byte, isa *Isa, op uint32, addr int) []byte {
	if op&(1<<20) != 0 {
		b = append(b, "ldm"...)
	} else {
		b = append(b, "stm"...)
	}

	amode := isa.BlockSuffixes[(op>>20)&1][(op>>23)&3]
	b = append(b, amode...)
	b = appendCond(b, op>>28)
	b = append(b, ' ')

	rn := isa.RegisterNaming[(op>>16)&0xF]
	b = append(b, rn...)

	if op&(1<<21) != 0 {
		b = append(b, '!')
	}

	b = append(b, ", {"...)

	for rlist := op & 0xFFFF; rlist != 0; {
		i := bits.TrailingZeros32(rlist)
		rlist &= rlist - 1
		b = append(b, isa.RegisterNaming[i]...)
		if rlist != 0 {
			b = append(b, ", "...)
		}
	}

	b = append(b, "}"...)

	if op&(1<<22) != 0 {
		b = append(b, '^')
	}

	return b
}

func half(b []byte, isa *Isa, op uint32, addr int) []byte {
	rn := isa.RegisterNaming[(op>>16)&0xF]
	rd := isa.RegisterNaming[(op>>12)&0xF]
	pre := (op>>24)&1 != 0
	inst := (op >> 5) & 3

	if load := (op>>20)&1 != 0; !load {
		switch inst {
		case 0:
			b = append(b, "unknown half !load"...)
			return b
		case 1:
			b = append(b, "strh"...)
		case 2:
			b = append(b, "ldrd"...)
		case 3:
			b = append(b, "strd"...)
		}
	} else {
		switch inst {
		case 0:
			b = append(b, "unknown half load"...)
			return b
		case 1:
			b = append(b, "ldrh"...)
		case 2:
			b = append(b, "ldrsb"...)
		case 3:
			b = append(b, "ldrsh"...)
		}
	}

	offset := make([]byte, 0, 8)
	if imm := (op>>22)&1 != 0; imm {
		if ofs := int((op & 0xF) | ((op >> 4) & 0xF0)); ofs != 0 {
			if op&(1<<23) == 0 {
				ofs = -ofs
			}

			offset = append(offset, ", "...)
			offset = appendImm(offset, ofs, isa.ImmediateValueBase)
		}
	} else {
		offset = append(offset, ", "...)

		if op&(1<<23) == 0 {
			offset = append(offset, '-')
		}

		offset = append(offset, isa.RegisterNaming[op&0xF]...)
	}

	b = appendCond(b, op>>28)
	b = append(b, ' ')
	b = append(b, rd...)

	b = append(b, ", "...)

	if pre {
		b = append(b, '[')
		b = append(b, rn...)
		b = append(b, offset...)
		b = append(b, ']')
		if wb := (op>>21)&1 != 0 && pre; wb {
			b = append(b, '!')
		}
	} else {
		b = append(b, '[')
		b = append(b, rn...)
		b = append(b, ']')
		b = append(b, offset...)
	}

	return b
}

func sdt(b []byte, isa *Isa, op uint32, addr int) []byte {
	pre := (op>>24)&1 != 0
	rn := isa.RegisterNaming[(op>>16)&0xF]
	rd := isa.RegisterNaming[(op>>12)&0xF]
	pld := op>>28 == 0xF
	if pld {
		// TODO: confirm and correct pld
		b = append(b, "pld"...)
		return b
	}

	offset := make([]byte, 0, 32)
	if reg := (op>>25)&1 != 0; !reg {
		if ofs := int(op & 0xFFF); ofs != 0 {

			if op&(1<<23) == 0 {
				ofs = -ofs
			}

			offset = append(offset, ", "...)
			offset = appendImm(offset, ofs, isa.ImmediateValueBase)
		}
	} else {

		is := int((op >> 7) & 0x1F)
		sType := (op >> 5) & 3
		rm := isa.RegisterNaming[op&0xF]

		if sType == 0 && is == 0 {
			offset = append(offset, ", "...)
			offset = append(offset, rm...)
		} else {
			offset = append(offset, ", "...)

			if op&(1<<23) == 0 {
				offset = append(offset, '-')
			}

			offset = append(offset, rm...)
			offset = append(offset, ", "...)
			offset = append(offset, shifts[sType]...)
			offset = append(offset, ' ')
			offset = appendImm(offset, is, isa.ImmediateValueBase)
		}
	}

	if load := (op>>20)&1 != 0; load {
		b = append(b, "ldr"...)
	} else {
		b = append(b, "str"...)
	}

	if byte := (op>>22)&1 != 0; byte {
		b = append(b, 'b')
	}

	if !pre && (op>>21)&1 != 0 {
		b = append(b, 't')
	}

	b = appendCond(b, op>>28)
	b = append(b, ' ')
	b = append(b, rd...)
	b = append(b, ", "...)

	if pre {
		b = append(b, '[')
		b = append(b, rn...)
		b = append(b, offset...)
		b = append(b, ']')
		if wb := (op>>21)&1 != 0 && pre; wb {
			b = append(b, '!')
		}
	} else {
		b = append(b, '[')
		b = append(b, rn...)
		b = append(b, ']')
		b = append(b, offset...)
	}

	return b
}

func aluShiftReg(b []byte, isa *Isa, op uint32, addr int) []byte {
	rm := isa.RegisterNaming[op&0xF]
	rs := isa.RegisterNaming[(op>>8)&0xF]
	sType := (op >> 5) & 3
	shift := shifts[sType]

	b = appendAluPre(b, isa, op, addr)

	b = append(b, rm...)
	b = append(b, ", "...)
	b = append(b, shift...)
	b = append(b, " "...)
	b = append(b, rs...)
	return b
}

func aluShiftImm(b []byte, isa *Isa, op uint32, addr int) []byte {
	rm := isa.RegisterNaming[op&0xF]
	sType := (op >> 5) & 3
	is := (op >> 7) & 0x1F

	b = appendAluPre(b, isa, op, addr)

	b = append(b, rm...)
	b = append(b, ", "...)

	shift := shifts[sType]
	if is == 0 {
		switch sType {
		case 0:
		case 1, 2:
			b = append(b, shift...)
			b = append(b, " #32"...)
		case 3:
			b = append(b, "rrx #1"...)
		}
	} else {
		b = append(b, shift...)
		b = append(b, " #"...)
		b = strconv.AppendUint(b, uint64(is), 10)
	}

	return b
	//// TODO: NOP (MOV r0, r0)
	//// TODO: shifts with mov are lsl, lsr, asr, and ror (rrx) separately
}

func appendAluPre(b []byte, isa *Isa, op uint32, addr int) []byte {
	inst := (op >> 21) & 0xF
	logical := inst >= 8 && inst < 0xC
	b = append(b, aluOpsArm[inst]...)
	b = appendCond(b, op>>28)
	if op&(1<<20) != 0 && !logical {
		b = append(b, "s"...)
	}

	b = append(b, " "...)

	rd := (op >> 12) & 0xF
	rn := (op >> 16) & 0xF

	if !logical {
		b = append(b, isa.RegisterNaming[rd]...)
		b = append(b, ", "...)
	}

	if inst != 0xD && inst != 0xF {
		b = append(b, isa.RegisterNaming[rn]...)
		b = append(b, ", "...)
	}

	return b
}

func alu(b []byte, isa *Isa, op uint32, addr int) []byte {
	b = appendAluPre(b, isa, op, addr)
	b = append(b, "#"...)

	// TODO: what decides imm full vs nn, ror rot?
	ror := (int((op>>8)&0xF) * 2)
	if useRor := false; useRor {

		b = strconv.AppendUint(b, uint64(op&0xFF), 10)

		if ror != 0 {
			b = append(b, ", "...)
			b = strconv.AppendUint(b, uint64(ror), 10)
		}

	} else {
		nn := bits.RotateLeft32(op&0xFF, -ror)
		b = strconv.AppendUint(b, uint64(nn), 10)

	}

	return b
}

func cop(b []byte, isa *Isa, op uint32, addr int) []byte {
	var (
		opc = (op >> 21) & 7
		cn  = (op >> 16) & 0xF
		rd  = isa.RegisterNaming[(op>>12)&0xF]
		pn  = (op >> 8) & 0xF
		cp  = (op >> 5) & 7
		cm  = (op >> 0) & 0xF
	)

	if m2 := op>>28 == 0xF; m2 {
		if op&(1<<20) != 0 {
			b = append(b, "mrc2 "...)
		} else {
			b = append(b, "mcr2 "...)
		}
	} else {
		if op&(1<<20) != 0 {
			b = append(b, "mrc "...)
			b = appendCond(b, op>>28)
		} else {
			b = append(b, "mcr "...)
			b = appendCond(b, op>>28)
		}
	}

	b = strconv.AppendUint(b, uint64(pn), 10)
	b = append(b, ", "...)
	b = strconv.AppendUint(b, uint64(opc), 10)
	b = append(b, ", "...)
	b = append(b, rd...)
	b = append(b, ", cr"...)
	b = strconv.AppendUint(b, uint64(cn), 10)
	b = append(b, ", cr"...)
	b = strconv.AppendUint(b, uint64(cm), 10)

	if cp != 0 {
		b = append(b, ", {"...)
		b = strconv.AppendUint(b, uint64(cp), 10)
		b = append(b, "}"...)
	}

	return b
}

var psrs = []string{
	"cpsr",
	"spsr",
}

func mrs(b []byte, isa *Isa, op uint32, addr int) []byte {
	rd := isa.RegisterNaming[(op>>12)&0xF]
	b = append(b, "mrs"...)
	b = appendCond(b, op>>28)
	b = append(b, " "...)
	b = append(b, rd...)
	b = append(b, ", "...)
	b = append(b, psrs[(op>>22)&1]...)
	return b
}

func msr(b []byte, isa *Isa, op uint32, addr int) []byte {
	b = append(b, "msr"...)
	b = appendCond(b, op>>28)
	b = append(b, " "...)
	b = append(b, psrs[(op>>22)&1]...)
	b = append(b, ", "...)

	if imm := (op>>25)&1 != 0; imm {
		v := ((op >> 8) & 0xF) << 1
		v = bits.RotateLeft32(op&0xFF, int(v))
		b = appendImm(b, int(v), isa.ImmediateValueBase)
	} else {
		b = append(b, isa.RegisterNaming[op&0xF]...)
	}
	return b
}

func cdp(b []byte, isa *Isa, op uint32, addr int) []byte {
	if cond := op >> 28; cond == 0xF {
		b = append(b, "cdp2"...)
	} else {
		b = append(b, "cdp"...)
		b = appendCond(b, op>>28)
	}

	b = append(b, " "...)

	rm := (op >> 0) & 0xF
	opc2 := (op >> 5) & 7
	coproc := (op >> 8) & 0xF
	rd := (op >> 12) & 0xF
	rn := (op >> 16) & 0xF
	opc1 := (op >> 20) & 0xF

	b = strconv.AppendUint(b, uint64(coproc), 10)
	b = append(b, ", "...)
	b = strconv.AppendUint(b, uint64(opc1), 10)
	b = append(b, ", cr"...)
	b = strconv.AppendUint(b, uint64(rd), 10)
	b = append(b, ", cr"...)
	b = strconv.AppendUint(b, uint64(rn), 10)
	b = append(b, ", cr"...)
	b = strconv.AppendUint(b, uint64(rm), 10)
	b = append(b, ", {"...)
	b = strconv.AppendUint(b, uint64(opc2), 10)
	b = append(b, "}"...)

	return b
}

func ext(b []byte, isa *Isa, op uint32, addr int) []byte {
	rn := (op >> 16) & 0xF
	rd := (op >> 12) & 0xF
	rm := (op >> 0) & 0xF
	ror := (op >> 10) & 3
	opc := (op >> 20) & 7

	if rn == 0xF {
		switch opc {
		case 0:
			b = append(b, "sxtb16"...)
		case 2:
			b = append(b, "sxtb"...)
		case 3:
			b = append(b, "sxth"...)
		case 4:
			b = append(b, "uxtb16"...)
		case 6:
			b = append(b, "uxtb"...)
		case 7:
			b = append(b, "uxth"...)
		}
	} else {
		switch opc {
		case 0:
			b = append(b, "sxtab16"...)
		case 2:
			b = append(b, "sxtab"...)
		case 3:
			b = append(b, "sxtah"...)
		case 4:
			b = append(b, "uxtab16"...)
		case 6:
			b = append(b, "uxtab"...)
		case 7:
			b = append(b, "uxtah"...)
		}
	}

	b = appendCond(b, op>>28)
	b = append(b, " "...)
	b = append(b, isa.RegisterNaming[rd]...)
	b = append(b, ", "...)

	if rn != 0xF {
		b = append(b, isa.RegisterNaming[rn]...)
		b = append(b, ", "...)
	}

	b = append(b, isa.RegisterNaming[rm]...)

	if ror == 0 {
		return b
	}

	b = append(b, ", ror #"...)

	switch ror {
	case 1:
		b = append(b, "8"...)
	case 2:
		b = append(b, "16"...)
	case 3:
		b = append(b, "24"...)
	}

	return b
}

func sat(b []byte, isa *Isa, op uint32, addr int) []byte {
	immediate := (op >> 16) & 0x1F
	if op&(1<<22) != 0 {
		b = append(b, "usat"...)
	} else {
		immediate++
		b = append(b, "ssat"...)
	}

	b = appendCond(b, op>>28)
	b = append(b, " "...)
	b = append(b, isa.RegisterNaming[(op>>12)&0xF]...)
	b = append(b, ", #"...)
	b = strconv.AppendUint(b, uint64(immediate), 10)
	b = append(b, ", "...)
	b = append(b, isa.RegisterNaming[op&0xF]...)

	if (op>>6)&0x3F == 0 {
		return b
	}

	b = append(b, ", "...)
	imm := (op >> 7) & 0x1F

	if op&(1<<6) != 0 {
		b = append(b, "asr #"...)
		b = strconv.AppendUint(b, uint64(imm), 10)
	} else {
		b = append(b, "lsl #"...)
		b = strconv.AppendUint(b, uint64(imm), 10)
	}

	return b
}

func sat16(b []byte, isa *Isa, op uint32, addr int) []byte {
	immediate := (op >> 16) & 0xF
	if op&(1<<22) != 0 {
		b = append(b, "usat16"...)
	} else {
		immediate++
		b = append(b, "ssat16"...)
	}

	b = appendCond(b, op>>28)
	b = append(b, " "...)
	b = append(b, isa.RegisterNaming[(op>>12)&0xF]...)
	b = append(b, ", #"...)
	b = strconv.AppendUint(b, uint64(immediate), 10)
	b = append(b, ", "...)
	b = append(b, isa.RegisterNaming[op&0xF]...)
	return b
}

func usad(b []byte, isa *Isa, op uint32, addr int) []byte {
	rd := (op >> 16) & 0xF
	rn := (op >> 12) & 0xF
	rs := (op >> 8) & 0xF
	rm := (op >> 0) & 0xF

	if rn == 0xF {
		b = append(b, "usad8"...)
	} else {
		b = append(b, "usada8"...)
	}

	b = appendCond(b, op>>28)
	b = append(b, " "...)
	b = append(b, isa.RegisterNaming[rd]...)
	b = append(b, ", "...)
	b = append(b, isa.RegisterNaming[rm]...)
	b = append(b, ", "...)
	b = append(b, isa.RegisterNaming[rs]...)

	if rn != 0xF {
		b = append(b, ", "...)
		b = append(b, isa.RegisterNaming[rn]...)
	}

	return b
}

func pack(b []byte, isa *Isa, op uint32, addr int) []byte {
	rn := (op >> 16) & 0xF
	rd := (op >> 12) & 0xF
	rm := (op >> 0) & 0xF

	switch (op >> 4) & 7 {
	case 0b001:
		b = append(b, "pkhbt"...)
	case 0b101:
		b = append(b, "pkhtb"...)
	}

	b = appendCond(b, op>>28)
	b = append(b, " "...)
	b = append(b, isa.RegisterNaming[rd]...)
	b = append(b, ", "...)
	b = append(b, isa.RegisterNaming[rn]...)
	b = append(b, ", "...)
	b = append(b, isa.RegisterNaming[rm]...)

	imm := (op >> 7) & 0x1F

	if imm != 0 {
		b = append(b, ", "...)
		switch (op >> 4) & 7 {
		case 0b001:
			b = append(b, "lsl #"...)
		case 0b101:
			b = append(b, "asr #"...)
		}

		b = strconv.AppendUint(b, uint64(imm), 10)
	}

	return b
}

func rev(b []byte, isa *Isa, op uint32, addr int) []byte {
	opc := (((op >> 22) & 1) << 1) | ((op >> 7) & 1)

	rd := (op >> 12) & 0xF
	rm := (op >> 0) & 0xF

	switch opc {
	case 0:
		b = append(b, "rev"...)
	case 1:
		b = append(b, "rev16"...)
	case 2:
		b = append(b, "revsh"...)
	}

	b = appendCond(b, op>>28)

	b = append(b, " "...)
	b = append(b, isa.RegisterNaming[rd]...)
	b = append(b, ", "...)
	b = append(b, isa.RegisterNaming[rm]...)

	return b
}
