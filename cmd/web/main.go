package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"
	"syscall/js"

	"queso"
)

var isa *queso.Isa

func main() {
	println("entering wasm")

	api := js.Global().Get("Object").New()
	api.Set("Decode", js.FuncOf(Decode))
	api.Set("ProcessFileUpload", js.FuncOf(ProcessFileUpload))
	js.Global().Set("goApp", api)

	isa = &queso.Isa{
		Order:           binary.LittleEndian,
		RegisterNaming:  queso.ConventionNaming,
		Arch:            queso.ARMv5TE,
		UndefinedOutput: "unknown instruction",
		Offset:          0x0,
	}

	js.Global().Get("document").Call("dispatchEvent",
		js.Global().Get("CustomEvent").New("goready"))

	initHtml()

	select {}
}

var regs = []string{
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

func initHtml() {
	doc := js.Global().Get("document")
	doc.Call("getElementById", "thumb").Set("checked", isa.Thumb)
	doc.Call("getElementById", "output address").Set("checked", isa.OutputAddress)
	doc.Call("getElementById", "output opcode").Set("checked", isa.OutputOpcode)
	doc.Call("getElementById", "little").Set("checked", isa.Order == binary.LittleEndian)
	doc.Call("getElementById", "conventional").Set("checked", isa.RegisterNaming == queso.ConventionNaming)
	doc.Call("getElementById", "general").Set("checked", isa.RegisterNaming == queso.GeneralNaming)
	doc.Call("getElementById", "hexinput").Set("checked", true)
	doc.Call("getElementById", "offset_input").Set("value", strconv.FormatInt(int64(isa.Offset), 16))

	custom := isa.RegisterNaming != queso.ConventionNaming && isa.RegisterNaming != queso.GeneralNaming

	doc.Call("getElementById", "custom").Set("checked", custom)

	for _, reg := range regs {
		doc.Call("getElementById", reg).Set("disabled", !custom)
	}

	if custom {
		for i, reg := range regs {
			doc.Call("getElementById", reg).Set("value", isa.RegisterNaming[i])
		}
	}

	archs := doc.Call("getElementsByName", "architecture version")
	for i := range archs.Length() {
		arch := archs.Index(i)
		if arch.Get("value").String() == isa.Arch.Name {
			arch.Set("checked", true)
			break
		}
	}

	// TODO: init values of additional options, offset etc etc
}

func UpdateIsa() {
	doc := js.Global().Get("document")

	isa.Thumb = doc.Call("getElementById", "thumb").Get("checked").Bool()
	isa.OutputAddress = doc.Call("getElementById", "output address").Get("checked").Bool()
	isa.OutputOpcode = doc.Call("getElementById", "output opcode").Get("checked").Bool()

	offsetValue := doc.Call("getElementById", "offset_input").Get("value").String()
	offset, _ := strconv.ParseUint(offsetValue, 16, 32)
	isa.Offset = int(offset)

	little := doc.Call("getElementById", "little").Get("checked").Bool()
	if little {
		isa.Order = binary.LittleEndian
	} else {
		isa.Order = binary.BigEndian
	}

	conventional := doc.Call("getElementById", "conventional").Get("checked").Bool()
	general := doc.Call("getElementById", "general").Get("checked").Bool()
	custom := doc.Call("getElementById", "custom").Get("checked").Bool()

	for _, reg := range regs {
		doc.Call("getElementById", reg).Set("disabled", !custom)
	}

	switch {
	case conventional:
		isa.RegisterNaming = queso.ConventionNaming
	case general:
		isa.RegisterNaming = queso.GeneralNaming
	case custom:

		isa.RegisterNaming = queso.RegisterNaming{}

		for i, reg := range regs {
			isa.RegisterNaming[i] = doc.Call("getElementById", reg).Get("value").String()
		}

	default:
		isa.RegisterNaming = queso.ConventionNaming
	}

	archs := doc.Call("getElementsByName", "architecture version")
	for i := range archs.Length() {
		arch := archs.Index(i)
		if arch.Get("checked").Bool() {
			// NOTE: better way of doing this? in perfect world shouldn't use
			// "value" but everything is client side so maybe dont care
			isa.Arch = queso.ArchLookup[arch.Get("value").String()]
			break
		}
	}
}

func Decode(this js.Value, args []js.Value) any {
	doc := js.Global().Get("document")

	UpdateIsa()

	var r io.Reader
	if isFileInput := js.Global().Get("isFileInput").Bool(); isFileInput {
		r = bytes.NewReader(fileData)
	} else {
		hex := doc.Call("getElementById", "hex input").Get("value").String()
		r = NewInstReaderFromString(hex, isa.Thumb)

		doc.Call("getElementById", "file_input_src").Set("value", "")
		fileData = []byte{}
	}

	insts, err := isa.Decode(r)
	if err != nil {
		doc.Call("getElementById", "disassembly").Set("textContent", err.Error())
		return nil
	}

	s := strings.Join(insts, "\n")
	doc.Call("getElementById", "disassembly").Set("innerHTML", s)

	return nil
}

var fileData []byte

func ProcessFileUpload(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return "missing args"
	}
	name := args[0].String()
	arr := args[1]

	fileData = make([]byte, arr.Get("length").Int())
	js.CopyBytesToGo(fileData, arr)
	return fmt.Sprintf("%s: %d bytes", name, len(fileData))
}
