package queso

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// NOTE: tests uses https://gitlab.arm.com/tooling/gnu-toolchains-for-arm/-/packages
// objdump and as to disassemble and assembler instructions
// a possible source is https://git.sr.ht/~sourceware/binutils-gdb/tree/master/item/opcodes/arm-dis.c
// objdump is not perfect but is a great 99% test comparison
// thumb16 instructions can be all tested exhaustively
// arm32 and thumb-2 (32 bit) have to be randomly tested

// NOTE: Hlt instruction: "Defined in V8 but is in undefined encoding space for earlier
// architectures.  However earlier architectures are required to treat
// this instuction as a semihosting trap as well.  Hence while not explicitly
// defined as such, it is in fact correct to define the instruction for all
// architectures." - binutils gas. Since this projects goal is to accurately disassemble
// emulated, and decomped arm binaries, hlt will not be included.

func newIsa(arch Architecture, thumb bool) *Isa {
	return &Isa{
		Arch:           arch,
		Order:          binary.LittleEndian,
		Thumb:          thumb,
		RegisterNaming: ConventionNaming,
	}
}

func norm(s string) string {
	s = strings.ToLower(s)
	return strings.Join(strings.Fields(strings.ReplaceAll(s, "\t", " ")), " ")
}

func decode(t testing.TB, isa *Isa, b []byte) []string {
	t.Helper()

	vs, err := isa.Decode(strings.NewReader(string(b)))
	if err != nil {
		panic(err)
	}

	return vs
}

const refCount = 2_000_000

var (
	objdumpBin = envOr("OBJDUMP", "arm-none-eabi-objdump")
	asBin      = envOr("AS", "arm-none-eabi-as")
	lineRe     = regexp.MustCompile(`^\s*([0-9a-f]+):\t([0-9a-f ]+?)\s*\t(.*)$`)
	commentRe  = regexp.MustCompile(`\s*[;@].*$`)
	labelRe    = regexp.MustCompile(`\s*<[^>]*>`)
	narrowRe   = regexp.MustCompile(`^([a-z0-9]+)\.n\b`) // removes .n thumb qualifier
)

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func normRef(s string) string {
	s = commentRe.ReplaceAllString(s, "")
	s = labelRe.ReplaceAllString(s, "")
	s = narrowRe.ReplaceAllString(s, "$1")
	return norm(s)
}

func runObjdump(t *testing.T, data []byte, arch Architecture, thumb bool) map[int]string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "in.bin")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	args := []string{"-D", "-z", "-b", "binary", "-marm", "--endian=little", "-m", string(arch.Name)}
	if thumb {
		args = append(args, "-M", "force-thumb")
	}
	args = append(args, path)
	out, err := exec.Command(objdumpBin, args...).Output()
	if err != nil {
		t.Skipf("objdump unavailable/failed: %v", err)
	}

	res := make(map[int]string, len(data)/2)
	sc := bufio.NewScanner(bytes.NewReader(out))

	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		if m := lineRe.FindStringSubmatch(sc.Text()); m != nil {
			addr, _ := strconv.ParseInt(m[1], 16, 64)
			res[int(addr)] = m[3]
		}
	}

	markUnassemblable(t, res, arch, thumb)

	return res
}

const notAssemblable = "not assemblable"

var asErrRe = regexp.MustCompile(`:(\d+): Error:`)

func isUndefinedText(s string) bool {
	s = strings.ToLower(s)
	return strings.HasPrefix(s, "udf") ||
		strings.Contains(s, "undefined") ||
		strings.Contains(s, "unpredictable") ||
		strings.HasPrefix(s, ".word") ||
		strings.HasPrefix(s, ".short") ||
		strings.Contains(s, notAssemblable) ||
		strings.HasPrefix(s, "hlt")
}

// 2nd pass against assembler to confirm it is not assemblable (cleans up objdump disassembly)
func markUnassemblable(t *testing.T, res map[int]string, arch Architecture, thumb bool) {
	t.Helper()

	addrs := make([]int, 0, len(res))
	for a := range res {
		addrs = append(addrs, a)
	}
	sort.Ints(addrs)

	var sb strings.Builder
	sb.WriteString(".syntax unified\n")
	if thumb {
		sb.WriteString(".thumb\n")
	} else {
		sb.WriteString(".arm\n")
	}
	const header = 2 // instruction i lives on line header+1+i

	for _, a := range addrs {
		txt := res[a]
		if i := strings.IndexAny(txt, ";@"); i >= 0 { // strip comments
			txt = txt[:i]
		}
		txt = strings.TrimSpace(txt)
		if isUndefinedText(txt) {
			txt = ""
		}
		sb.WriteString(txt)
		sb.WriteByte('\n')
	}

	dir := t.TempDir()
	src := filepath.Join(dir, "t.s")
	if err := os.WriteFile(src, []byte(sb.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	// gas exits non-zero when anything fails; that's expected here.
	out, _ := exec.Command(asBin, "-march="+string(arch.Name), src,
		"-o", filepath.Join(dir, "t.o")).CombinedOutput()

	bad := 0
	for _, m := range asErrRe.FindAllStringSubmatch(string(out), -1) {
		line, _ := strconv.Atoi(m[1])
		idx := line - header - 1
		if idx < 0 || idx >= len(addrs) {
			continue
		}
		res[addrs[idx]] = notAssemblable
		bad++
	}
	t.Logf("%s: %d/%d lines not assemblable", arch.Name, bad, len(addrs))
}

type mismatch struct {
	op        uint32
	got, want string
}

// isThumbWide reports whether hw1 starts a 32-bit Thumb instruction
// (the BL/BLX prefix on v4T/v5T, any 32-bit encoding on Thumb-2).
func isThumbWide(hw1 uint16) bool {
	switch hw1 >> 11 {
	case 0b11101, 0b11110, 0b11111:
		return true
	}
	return false
}

// encode serialises ops and returns the byte offset of each op.
// Thumb: ops <= 0xFFFF are 16-bit, otherwise hw1<<16|hw2.
func encode(t *testing.T, ops []uint32, thumb bool) (data []byte, offs []int) {
	t.Helper()
	offs = make([]int, len(ops))

	if !thumb {
		data = make([]byte, 0, len(ops)*4)
		for i, op := range ops {
			offs[i] = i * 4
			data = binary.LittleEndian.AppendUint32(data, op)
		}
		return
	}

	data = make([]byte, 0, len(ops)*4)
	for i, op := range ops {
		offs[i] = len(data)
		if op>>16 == 0 {
			hw := uint16(op)
			if isThumbWide(hw) {
				t.Fatalf("op %d: %04X is a 32-bit prefix, give it as hw1<<16|hw2", i, hw)
			}
			data = binary.LittleEndian.AppendUint16(data, hw)
			continue
		}
		hw1, hw2 := uint16(op>>16), uint16(op)
		if !isThumbWide(hw1) {
			t.Fatalf("op %d: %08X has a 32-bit layout but %04X is not a wide prefix", i, op, hw1)
		}
		// halfwords are stored in order: hw1 first, each little-endian
		data = binary.LittleEndian.AppendUint16(data, hw1)
		data = binary.LittleEndian.AppendUint16(data, hw2)
	}
	return
}

func compare(t *testing.T, arch Architecture, thumb bool, ops []uint32) {
	data, offs := encode(t, ops, thumb)

	isa := newIsa(arch, thumb)
	ref := runObjdump(t, data, arch, thumb)
	got := decode(t, isa, data)

	if len(got) != len(ops) {
		t.Fatalf("decoder returned %d instructions, expected %d (wide thumb handled as one?)",
			len(got), len(ops))
	}

	buckets := map[string][]mismatch{}
	total := 0

	for i, op := range ops {
		want, ok := ref[offs[i]]
		if !ok {
			continue // objdump desynced or skipped this address
		}

		wantN, gotN := normRef(want), norm(got[i])

		//fmt.Printf("OP %08X, got %s, want %s\n", op, gotN, wantN)

		if isUndefinedText(wantN) {
			continue
		}

		if gotN != wantN {
			total++
			key := strings.Fields(wantN + " ?")[0]
			buckets[key] = append(buckets[key], mismatch{op, gotN, wantN})
		}
	}
	if total == 0 {
		return
	}
	keys := make([]string, 0, len(buckets))
	for k := range buckets {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(a, b int) bool { return len(buckets[keys[a]]) > len(buckets[keys[b]]) })

	instCnt := 1000
	mismatchesPerInst := 3

	var sb strings.Builder
	fmt.Fprintf(&sb, "%d / %d mismatches (%.3f%%)\n", total, len(ops), 100*float64(total)/float64(len(ops)))
	for _, k := range keys[:min(len(keys), instCnt)] {
		ms := buckets[k]
		fmt.Fprintf(&sb, "  [%s] %d mismatches:\n", k, len(ms))
		for _, m := range ms[:min(mismatchesPerInst, len(ms))] {
			if thumb && m.op>>16 == 0 {
				fmt.Fprintf(&sb, "    %04X\n", m.op)
			} else if thumb {
				fmt.Fprintf(&sb, "    %04X %04X\n", m.op>>16, m.op&0xFFFF)
			} else {
				fmt.Fprintf(&sb, "    %08X\n", m.op)
			}
			fmt.Fprintf(&sb, "      got : %s\n      want: %s\n", m.got, m.want)
		}
	}
	t.Error(sb.String())
}

func testArm(t *testing.T, arch Architecture) {
	r := rand.New(rand.NewSource(1))
	ops := make([]uint32, refCount)
	for i := range ops {
		op := r.Uint32()
		// bias 3/4 of samples toward cond=AL so the unconditional space isn't wasted
		if r.Intn(4) != 0 {
			op = op&0x0FFFFFFF | 0xE0000000
		}
		ops[i] = op
	}
	compare(t, arch, false, ops)
}

func testThumb(t *testing.T, arch Architecture) {
	ops := make([]uint32, 0, 65536)
	for i := range 65536 {
		// 0b11101 / 0b11110 / 0b11111 prefixes start 32-bit Thumb-2 encodings; skip
		// them since this decoder is 16-bit only, and they'd desync objdump.
		if top5 := i >> 11; top5 == 0b11101 || top5 == 0b11110 || top5 == 0b11111 {
			continue
		}
		ops = append(ops, uint32(i))
	}
	compare(t, arch, true, ops)
}

func TestArmARMv2(t *testing.T)   { testArm(t, ARMv2) }
func TestArmARMv2a(t *testing.T)  { testArm(t, ARMv2a) }
func TestArmARMv3(t *testing.T)   { testArm(t, ARMv3) }
func TestArmARMv3M(t *testing.T)  { testArm(t, ARMv3M) }
func TestArmARMv4(t *testing.T)   { testArm(t, ARMv4) }
func TestArmARMv4T(t *testing.T)  { testArm(t, ARMv4T) }
func TestArmARMv5(t *testing.T)   { testArm(t, ARMv5) }
func TestArmARMv5T(t *testing.T)  { testArm(t, ARMv5T) }
func TestArmARMv5TE(t *testing.T) { testArm(t, ARMv5TE) }

func TestThumbARMv4T(t *testing.T) { testThumb(t, ARMv4T) }
func TestThumbARMv5T(t *testing.T) { testThumb(t, ARMv5T) }

func TestThumbBL(t *testing.T) {
	arch := ARMv5TE
	rng := rand.New(rand.NewSource(1))

	ops := make([]uint32, 0, 200_000)
	for range cap(ops) {

		op := uint32(0xF000|rng.Intn(0x800)) << 16

		for _, suffix := range []uint32{0xF800, 0xE800} {
			op2 := suffix | uint32(rng.Intn(0x800))
			if suffix == 0xE800 {
				op2 &^= 1 // if unaligned, objdump splits into 2 instructions
			}
			ops = append(ops, op|op2)
		}
	}
	compare(t, arch, true, ops)
}
