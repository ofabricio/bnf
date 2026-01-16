package bnf

import (
	"fmt"
	"io"
	"os"
)

func Print(a AST) {
	Write(os.Stdout, a)
}

func Write(w io.Writer, a AST) {
	write(w, a, 0)
}

func write(w io.Writer, a AST, depth int) {
	for range depth {
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
