package bnf

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type AST struct {
	Type string
	Text string
	Next []AST
}

func (a AST) NextOrRoot() []AST {
	if len(a.Next) == 0 {
		return []AST{a}
	}
	return a.Next
}

// Compact replaces the root node with the child
// node when there is only one child node.
func (a *AST) Compact() {
	if len(a.Next) == 1 {
		*a = a.Next[0]
	}
}

func (a AST) Empty() bool {
	return len(a.Next)+len(a.Type) == 0
}

func String(a AST) string {
	var b strings.Builder
	print(&b, a, "")
	return b.String()
}

func Print(a AST) {
	print(os.Stdout, a, "")
}

func print(out io.Writer, a AST, pad string) {
	if len(a.Text) == 0 {
		fmt.Fprintf(out, "%s[%s]\n", pad, a.Type)
	} else {
		fmt.Fprintf(out, "%s[%s] %s\n", pad, a.Type, a.Text)
	}
	for _, n := range a.Next {
		print(out, n, pad+"    ")
	}
}
