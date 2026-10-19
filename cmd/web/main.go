package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"regexp"
	"strconv"
	"strings"
	"syscall/js"

	"queso"
)

var (
	isa      *queso.Isa
	fileData []byte
	hexRegex = regexp.MustCompile(`[^0-9a-fA-F]+`)
	doc      = js.Global().Get("document")
)

func main() {
	println("entering wasm")

	api := js.Global().Get("Object").New()
	api.Set("Decode", js.FuncOf(Decode))
	api.Set("ProcessFileBytes", js.FuncOf(ProcessFileBytes))
	js.Global().Set("goApp", api)

	js.Global().Get("document").Call("dispatchEvent",
		js.Global().Get("CustomEvent").New("goready"))

	isa = &queso.Isa{
		Order:           binary.LittleEndian,
		RegisterNaming:  queso.ConventionNaming,
		Arch:            queso.ARMv5TE,
		UndefinedOutput: "unknown instruction",
	}

	initHtml()

	select {}
}

func initHtml() {
	doc.Call("getElementById", "thumb").Set("checked", isa.Thumb)
	doc.Call("getElementById", "output_address").Set("checked", isa.OutputAddress)
	doc.Call("getElementById", "output_opcode").Set("checked", isa.OutputOpcode)
	doc.Call("getElementById", "little").Set("checked", isa.Order == binary.LittleEndian)
	doc.Call("getElementById", "conventional").Set("checked", isa.RegisterNaming == queso.ConventionNaming)
	doc.Call("getElementById", "general").Set("checked", isa.RegisterNaming == queso.GeneralNaming)
	doc.Call("getElementById", "hexinput").Set("checked", true)
	doc.Call("getElementById", "offset_input").Set("value", strconv.FormatInt(int64(isa.Offset), 16))

	custom := isa.RegisterNaming != queso.ConventionNaming && isa.RegisterNaming != queso.GeneralNaming

	doc.Call("getElementById", "custom").Set("checked", custom)

	for _, reg := range queso.GeneralNaming {
		doc.Call("getElementById", reg).Set("disabled", !custom)
	}

	if custom {
		for i, reg := range queso.GeneralNaming {
			doc.Call("getElementById", reg).Set("value", isa.RegisterNaming[i])
		}
	}

	archs := doc.Call("getElementsByClassName", "arch_version")
	for i := range archs.Length() {
		arch := archs.Index(i)
		if arch.Get("value").String() == strings.ToLower(isa.Arch.Name) {
			arch.Set("selected", true)
			break
		}
	}
}

func UpdateIsa() {
	isa.Thumb = doc.Call("getElementById", "thumb").Get("checked").Bool()
	isa.OutputAddress = doc.Call("getElementById", "output_address").Get("checked").Bool()
	isa.OutputOpcode = doc.Call("getElementById", "output_opcode").Get("checked").Bool()

	offsetValue := doc.Call("getElementById", "offset_input").Get("value").String()
	offsetValue = strings.ToLower(offsetValue)
	offsetValue = strings.TrimPrefix(offsetValue, "0x")
	offsetValue = hexRegex.ReplaceAllString(offsetValue, "")
	offset, _ := strconv.ParseUint(offsetValue, 16, 32)
	isa.Offset = int(offset)

	if little := doc.Call("getElementById", "little").Get("checked").Bool(); little {
		isa.Order = binary.LittleEndian
	} else {
		isa.Order = binary.BigEndian
	}

	general := doc.Call("getElementById", "general").Get("checked").Bool()
	custom := doc.Call("getElementById", "custom").Get("checked").Bool()

	for _, reg := range queso.GeneralNaming {
		doc.Call("getElementById", reg).Set("disabled", !custom)
	}

	switch {
	case general:
		isa.RegisterNaming = queso.GeneralNaming
	case custom:

		isa.RegisterNaming = queso.RegisterNaming{}

		for i, reg := range queso.GeneralNaming {
			isa.RegisterNaming[i] = doc.Call("getElementById", reg).Get("value").String()
		}

	default:
		isa.RegisterNaming = queso.ConventionNaming
	}

	archs := doc.Call("getElementsByClassName", "arch_version")
	for i := range archs.Length() {
		if arch := archs.Index(i); arch.Get("selected").Bool() {
			isa.Arch = queso.ArchLookup[arch.Get("value").String()]
			break
		}
	}
}

func Decode(this js.Value, args []js.Value) any {
	disassembly := doc.Call("getElementById", "disassembly")

	UpdateIsa()

	var r io.Reader
	if isFileInput := js.Global().Get("isFileInput").Bool(); isFileInput {
		r = bytes.NewReader(fileData)
	} else {
		var err error
		hex := []string{doc.Call("getElementById", "hex_input").Get("value").String()}
		r, err = queso.NewInstReaderFromStrings(hex)
		if err != nil {
			disassembly.Set("textContent", err.Error())
			return nil
		}

		doc.Call("getElementById", "file_input_src").Set("value", "")
		fileData = []byte{}
	}

	if insts, err := isa.Decode(r); err != nil {
		disassembly.Set("textContent", err.Error())
	} else {
		disassembly.Set("textContent", strings.Join(insts, "\n"))
	}

	return nil
}

func ProcessFileBytes(this js.Value, args []js.Value) any {
	arr := args[0]

	fileData = make([]byte, arr.Get("length").Int())
	js.CopyBytesToGo(fileData, arr)
	return nil
}
