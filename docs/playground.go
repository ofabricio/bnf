//go:build js && wasm

package main

import (
	"bytes"
	"syscall/js"

	"github.com/ofabricio/bnf"
)

func main() {
	js.Global().Set("bnfGet", bnfGet())
	<-make(chan struct{})
}

func bnfGet() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		INP := args[0].String()
		BNF := args[1].String()
		c := bnf.Compile(BNF)
		v := bnf.Parse(c, INP)
		var b bytes.Buffer
		bnf.Write(&b, v)
		return b.String()
	})
}
