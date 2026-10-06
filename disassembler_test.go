package queso

import (
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
)

func newIsa(thumb bool) *Isa {
	return &Isa{
		Order:              binary.LittleEndian,
		Thumb:              thumb,
		RegisterNaming:     ConventionNaming,
		BlockSuffixes:      StandardBlockSuffixes,
		ImmediateValueBase: HEX_BASE,
	}
}

func words32LE(ws ...uint32) []byte {
	b := make([]byte, 0, len(ws)*4)
	for _, w := range ws {
		b = binary.LittleEndian.AppendUint32(b, w)
	}
	return b
}

func words16LE(ws ...uint16) []byte {
	b := make([]byte, 0, len(ws)*2)
	for _, w := range ws {
		b = binary.LittleEndian.AppendUint16(b, w)
	}
	return b
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

// ---- golden table (adjust expected strings to match YOUR formatting) ----

func TestGoldenARM(t *testing.T) {
	tests := []struct {
		op   uint32
		want string
	}{
		{0xE0810002, "add r0, r1, r2"},
		{0xE3A00001, "mov r0, #1"},
		{0xE12FFF1E, "bx lr"},
		{0xE5921000, "ldr r1, [r2]"},
		{0xE92D4010, "push {r4, lr}"},
		{0xE8BD8010, "pop {r4, pc}"},
		{0xEA000000, "b #8"}, // branch target formatting is yours to define
	}
	isa := newIsa(false)
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%08X", tc.op), func(t *testing.T) {
			got := decode(t, isa, words32LE(tc.op))
			if len(got) != 1 || norm(got[0]) != norm(tc.want) {
				t.Fatalf("op %08X: got %q want %q", tc.op, got, tc.want)
			}
		})
	}
}

func TestGoldenThumb(t *testing.T) {
	tests := []struct {
		op   uint16
		want string
	}{
		{0x1888, "adds r0, r1, r2"},
		{0x2001, "movs r0, #1"},
		{0x4770, "bx lr"},
		{0xB510, "push {r4, lr}"},
		{0xBD10, "pop {r4, pc}"},
		{0xBF00, "nop"},
	}
	isa := newIsa(true)
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%04X", tc.op), func(t *testing.T) {
			got := decode(t, isa, words16LE(tc.op))
			if len(got) != 1 || norm(got[0]) != norm(tc.want) {
				t.Fatalf("op %04X: got %q want %q", tc.op, got, tc.want)
			}
		})
	}
}

// ---- structural properties ----

func TestEmptyInput(t *testing.T) {
	for _, thumb := range []bool{false, true} {
		if got := decode(t, newIsa(thumb), nil); len(got) != 0 {
			t.Fatalf("thumb=%v: expected no output, got %v", thumb, got)
		}
	}
}

func TestMisalignedLengthPanics(t *testing.T) {
	cases := []struct {
		thumb bool
		n     int
	}{{false, 1}, {false, 2}, {false, 3}, {false, 5}, {true, 1}, {true, 3}}
	for _, c := range cases {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("thumb=%v n=%d: expected panic", c.thumb, c.n)
				}
			}()
			decode(t, newIsa(c.thumb), make([]byte, c.n))
		}()
	}
}

func TestOutputCountAndNonEmpty(t *testing.T) {
	for _, thumb := range []bool{false, true} {
		w := 4
		if thumb {
			w = 2
		}
		buf := make([]byte, w*1000)
		for i := range buf {
			buf[i] = byte(i*31 + 7)
		}
		out := decode(t, newIsa(thumb), buf)
		if len(out) != 1000 {
			t.Fatalf("thumb=%v: got %d lines", thumb, len(out))
		}
		for i, s := range out {
			if strings.TrimSpace(s) == "" {
				t.Fatalf("thumb=%v: empty output at %d", thumb, i)
			}
		}
	}
}

// Decoding a stream must equal decoding each instruction separately
// (catches state leaking between instructions via the shared buffer).
func TestStreamEqualsIndividual(t *testing.T) {
	isa := newIsa(false)
	ops := []uint32{0xE0810002, 0xE3A00001, 0xE12FFF1E, 0xE5921000, 0xE92D4010, 0xDEADBEEF, 0x00000000}
	stream := decode(t, isa, words32LE(ops...))
	for i, op := range ops {
		one := decode(t, isa, words32LE(op))
		if stream[i] != one[0] {
			t.Errorf("%08X: stream %q vs single %q", op, stream[i], one[0])
		}
	}
}

func TestBigEndianMatchesLittleEndian(t *testing.T) {
	op := uint32(0xE0810002)
	le := newIsa(false)
	be := newIsa(false)
	be.Order = binary.BigEndian
	leOut := decode(t, le, words32LE(op))
	beOut := decode(t, be, binary.BigEndian.AppendUint32(nil, op))
	if leOut[0] != beOut[0] {
		t.Fatalf("LE %q != BE %q", leOut[0], beOut[0])
	}
}

// ---- exhaustive: every 16-bit Thumb opcode must decode without panic ----

func TestThumbExhaustiveNoPanic(t *testing.T) {
	isa := newIsa(true)
	buf := make([]byte, 0, 65536*2)
	for i := 0; i < 65536; i++ {
		buf = binary.LittleEndian.AppendUint16(buf, uint16(i))
	}
	out := decode(t, isa, buf)
	if len(out) != 65536 {
		t.Fatalf("got %d", len(out))
	}
	for i, s := range out {
		if s == "" {
			t.Fatalf("empty output for %04X", i)
		}
	}
}

// ---- random ARM: no panic, deterministic ----

func TestARMRandomDeterministicNoPanic(t *testing.T) {
	isa := newIsa(false)
	buf := make([]byte, 0, 1<<22)
	x := uint32(0x12345678)
	for i := 0; i < 1<<20; i++ {
		x ^= x << 13 // xorshift32
		x ^= x >> 17
		x ^= x << 5
		buf = binary.LittleEndian.AppendUint32(buf, x)
	}
	a := decode(t, isa, buf)
	b := decode(t, isa, buf)
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("nondeterministic at %d: %q vs %q", i, a[i], b[i])
		}
	}
}

func FuzzDecode(f *testing.F) {
	f.Add(words32LE(0xE0810002, 0xE12FFF1E), false)
	f.Add(words16LE(0x1888, 0x4770), true)
	f.Fuzz(func(t *testing.T, data []byte, thumb bool) {
		w := 4
		if thumb {
			w = 2
		}
		data = data[:len(data)/w*w]
		out := decode(t, newIsa(thumb), data)
		if len(out) != len(data)/w {
			t.Fatalf("len mismatch: %d vs %d", len(out), len(data)/w)
		}
	})
}
