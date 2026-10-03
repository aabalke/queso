package arm32

import (
	"fmt"
)

var shifts = [4]string{"lsl", "lsr", "asr", "ror"}

var cond = [16]string{
	"eq",
	"ne",
	"hs",
	"lo",
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

var rawReg = [16]string{
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

var convReg = [16]string{
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
	"fp",
	"ip",
	"sp",
	"lr",
	"pc",
}

var aluOps = [16]string{
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

type Decoder int

const (
	INST_SVC Decoder = iota
	INST_BKP
	INST_BRA
	INST_BRX
	INST_UND
	INST_MUL
	INST_HWM
	INST_CLZ
	INST_QAL
	INST_SWP
	INST_BLK
	INST_HAL
	INST_SDT
	INST_ALU
	INST_ALU_SHIFT_REG
	INST_ALU_SHIFT_IMM

	INST_UNKNOWN
	MAX_DECODER32_CNT
)

type Formatter struct {
	mask, match uint32
}

var format32 = [MAX_DECODER32_CNT]Formatter{
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
}

var decoders32 = [MAX_DECODER32_CNT]func(op uint32) string{
	INST_UNKNOWN: func(op uint32) string {
		return "unknown"
	},
	INST_SVC: func(op uint32) string {
		return fmt.Sprintf("svc%s #0x%X", cond[op>>28], op&0xFF_FFFF)
	},
	INST_BKP: func(op uint32) string {
		return fmt.Sprintf("bkpt #0x%X", (op&0xF)|(((op>>8)&0xFFF)<<4))
	},
	INST_UND: func(op uint32) string {
		return "undefined"
	},
	INST_BRA: func(op uint32) string {
		nn := uint32((int32(op)<<8)>>6) + 8

		switch {
		case op>>28 == 0xF:
			nn += ((op >> 24) & 1) << 1
			return fmt.Sprintf("blx #0x%X", nn)

		case op&(1<<24) != 0:
			return fmt.Sprintf("bl%s #0x%X", cond[op>>28], nn)
		default:
			return fmt.Sprintf("b%s #0x%X", cond[op>>28], nn)
		}
	},
	INST_BRX: func(op uint32) string {
		rn := op & 0xF

		switch inst := (op >> 4) & 0xF; inst {
		case 1:
			return fmt.Sprintf("bx%s r%d", cond[op>>28], rn)
		case 2:
			return fmt.Sprintf("bxj%s r%d", cond[op>>28], rn)
		case 3:
			return fmt.Sprintf("blx%s r%d", cond[op>>28], rn)
		default:
			panic("unknown branch exchange instruction")
		}
	},

	INST_MUL: func(op uint32) string {
		rd := (op >> 16) & 0xF
		rn := (op >> 12) & 0xF
		rs := (op >> 8) & 0xF
		rm := (op >> 0) & 0xF

		s := ""
		if op&(1<<20) != 0 {
			s = "s"
		}

		switch inst := (op >> 21) & 7; inst {
		case 0:
			return fmt.Sprintf("mul%s%s r%d, r%d, r%d", s, cond[op>>28], rd, rm, rs)
		case 1:
			return fmt.Sprintf("mla%s%s r%d, r%d, r%d, r%d", s, cond[op>>28], rd, rm, rs, rn)
		case 2:

			// TODO: only armv6+

			if op&(1<<20) != 0 {
				return "undefined"
			}

			return fmt.Sprintf("umaal%s r%d, r%d, r%d, r%d", cond[op>>28], rn, rd, rm, rs)
		case 4:
			return fmt.Sprintf("umull%s%s r%d, r%d, r%d, r%d", s, cond[op>>28], rn, rd, rm, rs)
		case 5:
			return fmt.Sprintf("umlal%s%s r%d, r%d, r%d, r%d", s, cond[op>>28], rn, rd, rm, rs)
		case 6:
			return fmt.Sprintf("smull%s%s r%d, r%d, r%d, r%d", s, cond[op>>28], rn, rd, rm, rs)
		case 7:
			return fmt.Sprintf("smlal%s%s r%d, r%d, r%d, r%d", s, cond[op>>28], rn, rd, rm, rs)
		default:
			panic("unknown muliplty instruction")
		}
	},

	INST_HWM: func(op uint32) string {
		rd := (op >> 16) & 0xF
		rn := (op >> 12) & 0xF
		rs := (op >> 8) & 0xF
		rm := (op >> 0) & 0xF

		y := "b"
		if op&(1<<6) != 0 {
			y = "t"
		}

		x := "b"
		if op&(1<<5) != 0 {
			x = "t"
		}

		switch inst := (op >> 21) & 7; inst {
		case 0:
			return fmt.Sprintf("smla%s%s%s r%d, r%d, r%d, r%d", x, y, cond[op>>28], rd, rm, rs, rn)
		case 1:

			if op&(1<<5) != 0 {
				return fmt.Sprintf("smlaw%s%s r%d, r%d, r%d, r%d", y, cond[op>>28], rd, rm, rs, rn)
			}

			return fmt.Sprintf("smulw%s%s r%d, r%d, r%d", y, cond[op>>28], rd, rm, rs)

		case 2:
			return fmt.Sprintf("smlal%s%s%s r%d, r%d, r%d, r%d", x, y, cond[op>>28], rn, rd, rm, rs)
		case 3:
			return fmt.Sprintf("smul%s%s%s r%d, r%d, r%d, r%d", x, y, cond[op>>28], rn, rd, rm, rs)
		default:
			panic("unknown multiply instruction")
		}
	},

	INST_CLZ: func(op uint32) string {
		return fmt.Sprintf("clz%s r%d, r%d", cond[op>>28], (op>>12)&0xF, op&0xF)
	},

	INST_QAL: func(op uint32) string {
		rn := (op >> 16) & 0xF
		rd := (op >> 12) & 0xF
		rm := (op >> 0) & 0xF

		switch inst := (op >> 20) & 0xF; inst {
		case 0:
			return fmt.Sprintf("qadd%s r%d, r%d, r%d", cond[op>>28], rd, rm, rn)
		case 1:
			return fmt.Sprintf("qsub%s r%d, r%d, r%d", cond[op>>28], rd, rm, rn)
		case 2:
			return fmt.Sprintf("qdadd%s r%d, r%d, r%d", cond[op>>28], rd, rm, rn)
		case 3:
			return fmt.Sprintf("qdsub%s r%d, r%d, r%d", cond[op>>28], rd, rm, rn)
		default:
			panic("unknown qalu instruction")
		}
	},

	INST_SWP: func(op uint32) string {
		rn := (op >> 16) & 0xF
		rd := (op >> 12) & 0xF
		rm := (op >> 0) & 0xF

		b := ""
		if op&(1<<22) != 0 {
			b = "b"
		}

		return fmt.Sprintf("swp%s%s r%d, r%d, [r%d]", b, cond[op>>28], rd, rm, rn)
	},

	INST_BLK: func(op uint32) string {
		amode := ""
		switch amodeBits := (op >> 23) & 3; amodeBits {
		case 0:
			amode = "da"
		case 1:
			amode = "ia"
		case 2:
			amode = "db"
		case 3:
			amode = "ib"
		}

		// TODO: alt mnemonics

		s := ""
		if op&(1<<22) != 0 {
			s = "^"
		}

		w := ""
		if op&(1<<21) != 0 {
			w = "!"
		}

		rlist := ""

		for i := range 16 {
			if op&(1<<i) != 0 {
				rlist += rawReg[i]

				if op&0xFFFF >= 1<<(i+1) {
					rlist += ", "
				}
			}
		}

		rn := (op >> 16) & 0xF

		if op&(1<<20) != 0 {
			return fmt.Sprintf("ldm%s%s r%x%s, {%s}%s", amode, cond[op>>28], rn, w, rlist, s)
		}

		return fmt.Sprintf("stm%s%s r%x%s, {%s}%s", amode, cond[op>>28], rn, w, rlist, s)
	},

	INST_HAL: func(op uint32) string {
		rn := (op >> 16) & 0xF
		rd := (op >> 12) & 0xF
		pre := (op>>24)&1 != 0
		load := (op>>20)&1 != 0
		inst := (op >> 5) & 3

		w := ""
		if wb := (op>>21)&1 != 0 || !pre; wb {
			w = "!"
		}

		sign := ""
		if op&(1<<23) == 0 {
			sign = "-"
		}

		offset := ""
		if imm := (op>>22)&1 != 0; imm {

			ofs := ((op & 0xF) | ((op >> 4) & 0xF0))
			if ofs != 0 {
				offset = fmt.Sprintf(", #%s0x%X", sign, ofs)
			}

		} else {
			offset = fmt.Sprintf(", %sr%d", sign, op&0xF)
		}

		mnemonic := ""
		if !load {
			switch inst {
			case 0:
				panic("invalid half !load")
			case 1:
				mnemonic = "strh"
			case 2:
				mnemonic = "ldrd"
			case 3:
				mnemonic = "strd"
			}
		} else {
			switch inst {
			case 0:
				panic("invalid half load")
			case 1:
				mnemonic = "ldrh"
			case 2:
				mnemonic = "ldrsb"
			case 3:
				mnemonic = "ldrsh"
			}
		}

		return fmt.Sprintf("%s%s r%d, [r%d%s]%s", mnemonic, cond[op>>28], rd, rn, offset, w)
	},

	INST_SDT: func(op uint32) string {
		reg := (op>>25)&1 != 0
		pre := (op>>24)&1 != 0
		byte := (op>>22)&1 != 0
		load := (op>>20)&1 != 0
		rn := (op >> 16) & 0xF
		rd := (op >> 12) & 0xF

		pld := op>>28 == 0xF

		t := ""
		if !pre && (op>>21)&1 != 0 {
			t = "t"
		}

		w := ""
		if wb := (op>>21)&1 != 0 && pre; wb {
			w = "!"
		}

		sign := ""
		if op&(1<<23) == 0 {
			sign = "-"
		}

		offset := ""
		if !reg {
			ofs := op & 0xFFF
			if ofs != 0 {
				offset = fmt.Sprintf("#%s0x%X", sign, ofs)
			}
		} else {

			is := (op >> 7) & 0x1F
			sType := (op >> 5) & 3
			rm := op & 0xF

			if sType == 0 && is == 0 {
				offset = fmt.Sprintf("r%d", rm)
			} else {
				shift := shifts[sType]

				offset = fmt.Sprintf("%sr%d, %s #0x%X", sign, rm, shift, is)
			}
		}

		addr := fmt.Sprintf("[r%d", rn)
		if pre {
			addr += fmt.Sprintf(", %s]%s", offset, w)
		} else {
			addr += fmt.Sprintf("], %s", offset)
		}

		if pld {
			return "pld"
		}

		mnemonic := ""
		if load {
			mnemonic += "ldr"
		} else {
			mnemonic += "str"
		}

		if byte {
			mnemonic += "b"
		}

		return fmt.Sprintf("%s%s%s r%d, %s", mnemonic, t, cond[op>>28], rd, addr)
	},

	INST_ALU: func(op uint32) string {
		return "alu"
	},

	INST_ALU_SHIFT_REG: func(op uint32) string {
		inst := (op >> 21) & 0xF
		sType := (op >> 5) & 3
		rn := (op >> 16) & 0xF
		rd := (op >> 12) & 0xF
		rs := (op >> 8) & 0xF
		rm := op & 0xF

		s := ""
		if op&(1<<20) != 0 && (inst < 8 || inst >= 0xC) {
			s = "s"
		}

		operands := ""

		switch {
		case inst < 8 || inst == 0xC || inst == 0xE:
			operands = fmt.Sprintf("r%d, r%d, r%d, %s r%d", rd, rn, rm, shifts[sType], rs)
		case inst < 0xC:
			operands = fmt.Sprintf("r%d, r%d, %s r%d", rn, rm, shifts[sType], rs)
		case inst == 0xD:
			return fmt.Sprintf("%s%s%s r%d, r%d, r%d", shifts[sType], s, cond[op>>28], rd, rm, rs)
		default:
			operands = fmt.Sprintf("r%d, r%d, %s r%d", rd, rm, shifts[sType], rs)
		}

		return fmt.Sprintf("%s%s%s %s", aluOps[inst], s, cond[op>>28], operands)
	},

	INST_ALU_SHIFT_IMM: func(op uint32) string {
		// TODO: NOP (MOV r0, r0)
		// TODO: shifts with mov are lsl, lsr, asr, and ror (rrx) separately

		inst := (op >> 21) & 0xF
		sType := (op >> 5) & 3
		rn := (op >> 16) & 0xF
		rd := (op >> 12) & 0xF
		is := (op >> 7) & 0x1F
		rm := op & 0xF

		s := ""
		if op&(1<<20) != 0 && inst < 8 && inst >= 0xC {
			s = "s"
		}

		op2 := ""

		shift := shifts[sType]
		if is == 0 {
			switch sType {
			case 0:
			case 1, 2:
				op2 = fmt.Sprintf("%s #0x%X", shift, 32)
			case 3:
				op2 = "rrx #0x1"
			}
		} else {
			op2 = fmt.Sprintf("%s #0x%X", shift, is)
		}

		operands := ""

		switch {
		case inst < 8 || inst == 0xC || inst == 0xE:
			operands = fmt.Sprintf("r%d, r%d, r%d, %s", rd, rn, rm, op2)
		case inst < 0xC:
			operands = fmt.Sprintf("r%d, r%d, %s", rn, rm, op2)
		default:
			operands = fmt.Sprintf("r%d, r%d, %s", rd, rm, op2)
		}

		return fmt.Sprintf("%s%s%s %s", aluOps[inst], s, cond[op>>28], operands)
	},
}
