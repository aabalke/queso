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

func newIsa(thumb bool) *Isa {
	return &Isa{
		Order:              binary.LittleEndian,
		Thumb:              thumb,
		RegisterNaming:     ConventionNaming,
		BlockSuffixes:      StandardBlockSuffixes,
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

func runObjdump(t *testing.T, data []byte, thumb bool) map[int]string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "in.bin")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	args := []string{"-D", "-z", "-b", "binary", "-marm", "--endian=little"}
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

	return res
}

type mismatch struct {
	op        uint32
	got, want string
}

func compare(t *testing.T, thumb bool, ops []uint32) {
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
	ref := runObjdump(t, data, thumb)
	got := decode(t, newIsa(thumb), data)

	buckets := map[string][]mismatch{}
	total := 0

	for i, op := range ops {
		want, ok := ref[i*w]
		if !ok {
			continue
		}

		wantN, gotN := normRef(want), norm(got[i])
		// skip things objdump itself calls undefined/unpredictable
		if strings.HasPrefix(wantN, "udf") || strings.Contains(wantN, "undefined") ||
			strings.Contains(wantN, "unpredictable") || strings.HasPrefix(wantN, ".word") ||
			strings.HasPrefix(wantN, ".short") {
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

func TestArm(t *testing.T) {
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
	compare(t, false, ops)
}

func TestThumb(t *testing.T) {
	ops := make([]uint32, 0, 65536)
	for i := range 65536 {
		// 0b11101 / 0b11110 / 0b11111 prefixes start 32-bit Thumb-2 encodings; skip
		// them since this decoder is 16-bit only, and they'd desync objdump.
		if top5 := i >> 11; top5 == 0b11101 || top5 == 0b11110 || top5 == 0b11111 {
			continue
		}
		ops = append(ops, uint32(i))
	}
	compare(t, true, ops)
}

func TestThumb2(t *testing.T) {
	ops := make([]uint32, 0, (1<<11)*3)

	tops := []uint32{0b11101 << 11, 0b11110 << 11, 0b11111 << 11}

	for _, top := range tops {
		for i := range uint32(1 << 11) {
			ops = append(ops, uint32(top|i))
		}
	}

	compare(t, true, ops)
}
