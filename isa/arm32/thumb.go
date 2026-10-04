package arm32

import "fmt"

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

var format16 = [MAX_DECODER16_CNT]Formatter{
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

var decoders16 = [MAX_DECODER16_CNT]func(op uint16) string{
	INST_THUMB_UNKNOWN: func(op uint16) string {
		return "unknown thumb"
	},
	INST_THUMB_BRA: func(op uint16) string {
		// TODO: should negative be uint or int? and format -#0x20 or #0x-20
		return fmt.Sprintf("b #0x%X", (int16((op&0x7FF)<<5)>>4)+4)
	},

	INST_THUMB_SWI: func(op uint16) string {
		return fmt.Sprintf("svc #0x%X", op&0xFF)
	},
	INST_THUMB_BPT: func(op uint16) string {
		return fmt.Sprintf("bkpt #0x%X", op&0xFF)
	},

	INST_THUMB_HI: func(op uint16) string {
		rd := (op & 7) | (((op >> 7) & 1) << 3)
		rs := (op >> 3) & 0xF

		switch inst := (op >> 8) & 3; inst {
		case 0:
			return fmt.Sprintf("add r%d, r%d", rd, rs)
		case 1:
			return fmt.Sprintf("cmp r%d, r%d", rd, rs)
		case 2:
			if op == 0x46C0 {
				return "nop"
			}

			return fmt.Sprintf("mov r%d, r%d", rd, rs)
		case 3:
			mnemonic := "bx"
			if op&0x80 != 0 {
				mnemonic = "blx"
			}

			return fmt.Sprintf("%s r%d", mnemonic, rs)
		default:
			panic("not possible")
		}
	},

	INST_THUMB_CBR: func(op uint16) string {
		// TODO: should labels be abs or relative? +/- vs exact location
		nn := (int(int8(op&0xFF)) << 1) + 4
		return fmt.Sprintf("b%s #0x%X", cond[(op>>8)&0xF], nn)
	},

	INST_THUMB_ALU: func(op uint16) string {
		rd := (op >> 0) & 7
		rs := (op >> 3) & 7
		return fmt.Sprintf("%s r%d, r%d", aluOpsThumb[(op>>6)&0xF], rd, rs)
	},

	INST_THUMB_IMM: func(op uint16) string {
		inst := "subs"

		switch (op >> 11) & 3 {
		case 0:
			inst = "movs"
		case 1:
			inst = "cmp"
		case 2:
			inst = "adds"
		}

		rd := (op >> 8) & 7
		nn := op & 0xFF
		return fmt.Sprintf("%s r%d, #0x%X", inst, rd, nn)
	},

	INST_THUMB_ADD: func(op uint16) string {
		isImm := op&(1<<10) != 0
		sub := op&(1<<9) != 0
		n := (op >> 6) & 7
		rs := (op >> 3) & 7
		rd := (op >> 0) & 7

		inst := "adds"
		if sub {
			inst = "subs"
		}

		switch {
		case isImm && n == 0 && !sub:
			return fmt.Sprintf("mov r%d, r%d", rd, rs)
		case isImm:
			return fmt.Sprintf("%s r%d, r%d, #0x%X", inst, rd, rs, n)
		default:
			return fmt.Sprintf("%s r%d, r%d, r%d", inst, rd, rs, n)
		}
	},

	INST_THUMB_SHI: func(op uint16) string {
		opc := (op >> 11) & 3
		shift := shifts[opc]
		offset := (op >> 6) & 0x1F
		rs := (op >> 3) & 7
		rd := (op >> 0) & 7

		if offset == 0 && opc > 0 {
			offset = 32
		}

		return fmt.Sprintf("%ss r%d, r%d, #0x%X", shift, rd, rs, offset)
	},

	INST_THUMB_PUS: func(op uint16) string {
		pop := op&(1<<11) != 0

		inst := "push"
		if pop {
			inst = "pop"
		}

		rlist := ""

		for i := range 8 {
			if op&(1<<i) != 0 {
				rlist += rawRegs[i]

				if op&0x1FF >= 1<<(i+1) {
					rlist += ", "
				}
			}
		}

		if op&(1<<8) != 0 {
			if pop {
				rlist += "r15"
			} else {
				rlist += "r14"
			}
		}

		return fmt.Sprintf("%s {%s}", inst, rlist)
	},

	INST_THUMB_BLK: func(op uint16) string {
		pop := op&(1<<11) != 0

		rb := (op >> 8) & 7

		inst := "stmia"
		if pop {
			inst = "ldmia"
		}

		rlist := ""

		for i := range 8 {
			if op&(1<<i) != 0 {
				rlist += rawRegs[i]

				if op&0xFF >= 1<<(i+1) {
					rlist += ", "
				}
			}
		}

		return fmt.Sprintf("%s r%d!, {%s}", inst, rb, rlist)
	},

	INST_THUMB_REL: func(op uint16) string {
		nn := (op & 0xFF) << 2
		rd := (op >> 8) & 7

		if pc := op&(1<<11) == 0; pc {
			return fmt.Sprintf("add r%d, r15, #0x%X", rd, (4 + nn))
		}

		return fmt.Sprintf("add r%d, r13, #0x%X", rd, nn)
	},

	INST_THUMB_SP: func(op uint16) string {
		nn := (op & 0x7F) << 2

		inst := "add"
		if op&0x80 != 0 {
			inst = "sub"
		}

		return fmt.Sprintf("%s r13, #0x%X", inst, nn)
	},

	INST_THUMB_PCR: func(op uint16) string {
		nn := (op & 0xFF) << 2
		rd := (op >> 8) & 7

		return fmt.Sprintf("ldr r%d, [r15, #0x%X]", rd, nn)
	},

	INST_THUMB_SDTR: func(op uint16) string {
		ro := (op >> 6) & 7
		rb := (op >> 3) & 7
		rd := (op >> 0) & 7

		inst := ""
		switch (op >> 9) & 7 {
		case 0:
			inst = "str"
		case 1:
			inst = "strh"
		case 2:
			inst = "strb"
		case 3:
			inst = "ldsb"
		case 4:
			inst = "ldr"
		case 5:
			inst = "ldrh"
		case 6:
			inst = "ldrb"
		case 7:
			inst = "ldsh"
		}

		return fmt.Sprintf("%s r%d, [r%d, r%d]", inst, rd, rb, ro)
	},

	INST_THUMB_SDTI: func(op uint16) string {
		nn := (op >> 6) & 0x1F
		rb := (op >> 3) & 7
		rd := (op >> 0) & 7

		inst := ""
		switch (op >> 11) & 3 {
		case 0:
			inst = "str"
			nn *= 4
		case 1:
			inst = "ldr"
			nn *= 4
		case 2:
			inst = "strb"
		case 3:
			inst = "ldrb"
		}

		return fmt.Sprintf("%s r%d, [r%d, #0x%X]", inst, rd, rb, nn)
	},
	INST_THUMB_SDTH: func(op uint16) string {
		nn := ((op >> 6) & 0x1F) * 2
		rb := (op >> 3) & 7
		rd := (op >> 0) & 7

		inst := "strh"
		if op&(1<<11) != 0 {
			inst = "ldrh"
		}

		return fmt.Sprintf("%s r%d, [r%d, #0x%X]", inst, rd, rb, nn)
	},

	INST_THUMB_SPR: func(op uint16) string {
		nn := (op & 0xFF) * 4
		rd := (op >> 8) & 7

		inst := "str"
		if op&(1<<11) != 0 {
			inst = "ldr"
		}

		return fmt.Sprintf("%s r%d, [r13, #0x%X]", inst, rd, nn)
	},
}
