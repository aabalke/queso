package arm32

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
	"r10",
	"fp",
	"ip",
	"sp",
	"lr",
	"pc",
}

type BlockSuffixes [2][4]string

var StandardBlockSuffixes = BlockSuffixes{
	{"da", "ia", "db", "ib"},
	{"da", "ia", "db", "ib"},
}

var StackBlockSuffixes = BlockSuffixes{
	{"ed", "ea", "fd", "ed"},
	{"ea", "fd", "fa", "ed"},
}

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

type Decoder int

type Formatter struct {
	mask, match uint32
}
