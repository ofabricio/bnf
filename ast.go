package bnf

import (
	"fmt"
	"io"
	"os"
)

type AST struct {
	Type string
	Text string
	Next []AST
}

// Compact replaces the root node with the child
// node when there is only one child node.
func (a *AST) Compact() {
	if len(a.Next) == 1 {
		*a = a.Next[0]
	}
}

func Print(a AST) {
	Write(os.Stdout, a)
}

func Write(w io.Writer, a AST) {
	write(w, a, 0)
}

func write(w io.Writer, a AST, depth int) {
	for i := 0; i < depth; i++ {
		fmt.Fprint(w, "    ")
	}
	if len(a.Text) == 0 {
		fmt.Fprintf(w, "[%s]\n", a.Type)
	} else {
		fmt.Fprintf(w, "[%s] %s\n", a.Type, a.Text)
	}
	for _, n := range a.Next {
		write(w, n, depth+1)
	}
}
