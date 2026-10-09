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

func newIsa(arch Arch, thumb bool) *Isa {
	return &Isa{
		Arch:               arch,
		Order:              binary.LittleEndian,
		Thumb:              thumb,
		RegisterNaming:     ConventionNaming,
		ImmediateValueBase: DEC_BASE,
	}
}

// norm makes comparisons tolerant of whitespace/case differences.
func norm(s string) string {
	s = strings.ToLower(s)
	return strings.Join(strings.Fields(strings.ReplaceAll(s, "\t", " ")), " ")
}

func decode(t testing.TB, isa *Isa, b []byte) []string {
	t.Helper()
	return isa.Decode(strings.NewReader(string(b)))
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

func runObjdump(t *testing.T, data []byte, arch Arch, thumb bool) map[int]string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "in.bin")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	args := []string{"-D", "-z", "-b", "binary", "-marm", "--endian=little", "-m", string(arch)}
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
		//fmt.Printf("%q\n", sc.Text())
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
		strings.HasPrefix(s, ".short")
}

// markUnassemblable runs every disassembled line through arm-none-eabi-as
// for the same arch and replaces the text of lines gas rejects.
func markUnassemblable(t *testing.T, res map[int]string, arch Arch, thumb bool) {
	t.Helper()

	addrs := make([]int, 0, len(res))
	for a := range res {
		addrs = append(addrs, a)
	}
	sort.Ints(addrs)

	var sb strings.Builder
	sb.WriteString(".syntax unified\n") // line 1
	if thumb {
		sb.WriteString(".thumb\n") // line 2
	} else {
		sb.WriteString(".arm\n") // line 2
	}
	const header = 2 // instruction i lives on line header+1+i

	for _, a := range addrs {
		txt := res[a]
		if i := strings.IndexAny(txt, ";@"); i >= 0 { // strip objdump comments
			txt = txt[:i]
		}
		txt = strings.TrimSpace(txt)
		if isUndefinedText(txt) {
			txt = "" // keep line numbering aligned
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
	out, _ := exec.Command(asBin, "-march="+string(arch), src,
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
	t.Logf("%s: %d/%d lines not assemblable", arch, bad, len(addrs))
}

type mismatch struct {
	op        uint32
	got, want string
}

func compare(t *testing.T, arch Arch, thumb bool, ops []uint32) {
	w := 4
	if thumb {
		w = 2
	}
	data := make([]byte, 0, len(ops)*w)
	for _, op := range ops {
		if thumb {
			data = binary.LittleEndian.AppendUint16(data, uint16(op))
		} else {
			data = binary.LittleEndian.AppendUint32(data, op)
		}
	}

	isa := newIsa(arch, thumb)
	ref := runObjdump(t, data, arch, thumb)
	got := decode(t, isa, data)

	buckets := map[string][]mismatch{}
	total := 0

	for i, op := range ops {
		want, ok := ref[i*w]
		if !ok {
			continue
		}

		wantN, gotN := normRef(want), norm(got[i])
		// skip things objdump itself calls undefined/unpredictable
		// hlt is v8 but for some reason objdump includes in all
		if strings.HasPrefix(wantN, "udf") || strings.Contains(wantN, "undefined") ||
			strings.Contains(wantN, "unpredictable") || strings.HasPrefix(wantN, ".word") ||
			strings.HasPrefix(wantN, ".short") || strings.Contains(wantN, notAssemblable) || strings.HasPrefix(wantN, "hlt") {
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

	var sb strings.Builder
	fmt.Fprintf(&sb, "%d / %d mismatches (%.3f%%)\n", total, len(ops), 100*float64(total)/float64(len(ops)))
	for _, k := range keys[:min(len(keys), 1000)] {
		ms := buckets[k]
		fmt.Fprintf(&sb, "  [%s] %d mismatches:\n", k, len(ms))
		for _, m := range ms[:min(3, len(ms))] {
			fmt.Fprintf(&sb, "    %08X\n      got : %s\n      want: %s\n", m.op, m.got, m.want)
		}
	}
	t.Error(sb.String())
}

func testArm(t *testing.T, arch Arch) {
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

func testThumb(t *testing.T, arch Arch) {
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

func TestArmV2(t *testing.T)   { testArm(t, ARCH_ARMv2) }
func TestArmV4T(t *testing.T)  { testArm(t, ARCH_ARMv4t) }
func TestArmV5TE(t *testing.T) { testArm(t, ARCH_ARMv5te) }

func TestThumbV4T(t *testing.T)  { testThumb(t, ARCH_ARMv4t) }
func TestThumbV5TE(t *testing.T) { testThumb(t, ARCH_ARMv5te) }

// TODO: thumb-2 (32bit instructions)

//func TestThumb2(t *testing.T) {
//	arch := ARCH_ARM_ANY
//	ops := make([]uint32, 0, (1<<11)*3)
//
//	tops := []uint32{0b11101 << 11, 0b11110 << 11, 0b11111 << 11}
//
//	for _, top := range tops {
//		for i := range uint32(1 << 11) {
//			ops = append(ops, uint32(top|i))
//		}
//	}
//
//	compare(t, arch, true, ops)
//}
