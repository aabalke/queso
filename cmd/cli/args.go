package main

import (
	"encoding/binary"
	"io"
	"strconv"
)

type ArgsUint32 struct {
	thumb bool
	args  []uint32
	idx   int
	buf   [4]byte
	cur   []byte
}

func NewArgsUint32(ops []string, thumb bool) *ArgsUint32 {
	a := &ArgsUint32{thumb: thumb}

	for _, s := range ops {
		v, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			panic(err)
		}
		a.args = append(a.args, uint32(v))
	}

	return a
}

func (a *ArgsUint32) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	for n < len(p) {
		if len(a.cur) == 0 {
			if a.idx >= len(a.args) {
				break
			}
			if a.thumb {
				binary.BigEndian.PutUint16(a.buf[:2], uint16(a.args[a.idx]))
				a.cur = a.buf[:2]
			} else {
				binary.BigEndian.PutUint32(a.buf[:4], a.args[a.idx])
				a.cur = a.buf[:4]
			}
			a.idx++
		}

		c := copy(p[n:], a.cur)
		n += c
		a.cur = a.cur[c:]
	}

	if n == 0 {
		return 0, io.EOF
	}
	return n, nil
}
