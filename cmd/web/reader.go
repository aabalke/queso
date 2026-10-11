package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type InstReader struct {
	thumb        bool
	src          *bufio.Reader
	hexCharWidth int // hex chars per value: 8 normally, 4 in thumb mode
	chunk        []byte
	buf          [4]byte
	cur          []byte
	err          error
}

// NewInstReader reads a continuous hex string (e.g. "EA00FF00FF342534") from r.
func NewInstReader(r io.Reader, thumb bool) *InstReader {
	w := 8
	if thumb {
		w = 4
	}
	return &InstReader{
		thumb:        thumb,
		src:          bufio.NewReader(r),
		hexCharWidth: w,
		chunk:        make([]byte, 0, w),
	}
}

func NewInstReaderFromString(s string, thumb bool) *InstReader {
	return NewInstReader(strings.NewReader(s), thumb)
}

// fill reads the next width hex chars and loads them into a.cur.
func (a *InstReader) fill() bool {
	a.chunk = a.chunk[:0]
	for len(a.chunk) < a.hexCharWidth {
		c, err := a.src.ReadByte()
		if err != nil {
			if err == io.EOF {
				if len(a.chunk) == 0 {
					a.err = io.EOF // clean end on a value boundary
				} else {
					a.err = fmt.Errorf("truncated input: got %d hex chars, want %d: %w",
						len(a.chunk), a.hexCharWidth, io.ErrUnexpectedEOF)
				}
			} else {
				a.err = err
			}
			return false
		}
		switch c {
		case ' ', '\t', '\r', '\n':
			continue // tolerate stray whitespace/newlines
		}
		a.chunk = append(a.chunk, c)
	}

	v, err := strconv.ParseUint(string(a.chunk), 16, 32)
	if err != nil {
		a.err = fmt.Errorf("parsing %q: %w", a.chunk, err)
		return false
	}

	if a.thumb {
		binary.BigEndian.PutUint16(a.buf[:2], uint16(v))
		a.cur = a.buf[:2]
	} else {
		binary.BigEndian.PutUint32(a.buf[:4], uint32(v))
		a.cur = a.buf[:4]
	}
	return true
}

func (a *InstReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	for n < len(p) {
		if len(a.cur) == 0 {
			if a.err != nil || !a.fill() {
				break
			}
		}
		c := copy(p[n:], a.cur)
		n += c
		a.cur = a.cur[c:]
	}

	if n > 0 {
		return n, nil
	}
	return 0, a.err
}
