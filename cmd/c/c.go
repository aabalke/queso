package main

/*
#include <stdlib.h>
#include "libqueso.h"
*/
import "C"

import (
	"bytes"
	"encoding/binary"
	"unsafe"

	"queso"
)

var archs = [...]queso.Architecture{
	C.QUESO_ARMv1:     queso.ARMv1,
	C.QUESO_ARMv2:     queso.ARMv2,
	C.QUESO_ARMv2a:    queso.ARMv2a,
	C.QUESO_ARMv3:     queso.ARMv3,
	C.QUESO_ARMv3M:    queso.ARMv3M,
	C.QUESO_ARMv4xM:   queso.ARMv4xM,
	C.QUESO_ARMv4:     queso.ARMv4,
	C.QUESO_ARMv4TxM:  queso.ARMv4TxM,
	C.QUESO_ARMv4T:    queso.ARMv4T,
	C.QUESO_ARMv5xM:   queso.ARMv5xM,
	C.QUESO_ARMv5:     queso.ARMv5,
	C.QUESO_ARMv5TxM:  queso.ARMv5TxM,
	C.QUESO_ARMv5T:    queso.ARMv5T,
	C.QUESO_ARMv5TExP: queso.ARMv5TExP,
	C.QUESO_ARMv5TE:   queso.ARMv5TE,
}

var orders = [...]binary.ByteOrder{
	C.QUESO_BIG_ENDIAN:    binary.BigEndian,
	C.QUESO_LITTLE_ENDIAN: binary.LittleEndian,
	C.QUESO_NATIVE_ENDIAN: binary.NativeEndian,
}

//export queso_decode
func queso_decode(opts *C.queso_isa_opts, data_bytes *C.uint8_t, n C.int, res *C.queso_result) C.int {
	var naming queso.RegisterNaming
	switch opts.naming {
	case C.QUESO_CONVENTIONAL:
		naming = queso.ConventionNaming
	case C.QUESO_GENERAL:
		naming = queso.GeneralNaming
	case C.QUESO_CUSTOM:
		naming = getCustomNames(opts)
	}

	isa := queso.Isa{
		Arch:            archs[opts.arch],
		Order:           orders[opts.order],
		RegisterNaming:  naming,
		Thumb:           bool(opts.thumb),
		OutputAddress:   bool(opts.output_address),
		OutputOpcode:    bool(opts.output_opcode),
		Offset:          uint32(opts.offset),
		UndefinedOutput: C.GoString(opts.undefined_output),
	}

	b := unsafe.Slice((*byte)(unsafe.Pointer(data_bytes)), int(n))

	vs, err := isa.Decode(bytes.NewReader(b))

	if len(vs) != 0 {
		arr := (**C.char)(C.calloc(C.size_t(len(vs)), C.size_t(unsafe.Sizeof(uintptr(0)))))
		if arr == nil {
			res.err = C.CString("error allocating string")
			return -1
		}
		slots := unsafe.Slice(arr, len(vs))
		for i, s := range vs {
			slots[i] = C.CString(s)
		}
		res.lines = arr
		res.len = C.size_t(len(vs))
	}

	if err != nil {
		res.err = C.CString(err.Error())
		return -1
	}

	return 0
}

//export queso_free_result
func queso_free_result(r *C.queso_result) {
	if r == nil {
		return
	}
	if r.lines != nil {
		for _, p := range unsafe.Slice(r.lines, int(r.len)) {
			C.free(unsafe.Pointer(p))
		}
		C.free(unsafe.Pointer(r.lines))
	}
	C.free(unsafe.Pointer(r.err))
	*r = C.queso_result{}
}

func getCustomNames(opts *C.queso_isa_opts) queso.RegisterNaming {
	var out queso.RegisterNaming
	for i, p := range opts.custom_names {
		if p == nil {
			out[i] = queso.ConventionNaming[i]
			continue
		}

		out[i] = C.GoString(p)
	}
	return queso.RegisterNaming(out)
}

func main() {}
