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

func PrintBNF(n BNF) {
	WriteBNF(os.Stdout, n)
}

func WriteBNF(w io.Writer, n BNF) {
	writeBNF(w, n, 0)
}

func writeBNF(w io.Writer, n BNF, depth int) {
	if n == nil {
		return
	}
	for range depth {
		fmt.Fprint(w, "    ")
	}
	switch n := n.(type) {
	case Root:
		fmt.Fprintf(w, "[Root]\n")
		for _, v := range n.Stmts {
			writeBNF(w, v, depth+1)
		}
	case Stmt:
		fmt.Fprintf(w, "[Stmt] %s\n", n.Ident.Text)
		writeBNF(w, n.Expr, depth+1)
	case ExprAnd:
		fmt.Fprintf(w, "[And]\n")
		for _, v := range n.And {
			writeBNF(w, v, depth+1)
		}
	case ExprOr:
		fmt.Fprintf(w, "[Or]\n")
		for _, v := range n.Or {
			writeBNF(w, v, depth+1)
		}
	case Quantifier:
		fmt.Fprintf(w, "[Quantifier] %s\n", n.Name)
		writeBNF(w, n.Expr, depth+1)
	case Function:
		fmt.Fprintf(w, "[Function] %s\n", n.Name)
		writeBNF(w, n.Expr, depth+1)
	case Type:
		fmt.Fprintf(w, "[Type] %s\n", n.Ident.Text)
		writeBNF(w, n.Expr, depth+1)
	case Ignore:
		fmt.Fprintf(w, "[Ignore]\n")
		writeBNF(w, n.Expr, depth+1)
	case Plain:
		fmt.Fprintf(w, "[Plain] %s\n", n.Text)
	case Regex:
		fmt.Fprintf(w, "[Regex] %s\n", n.Text)
	case Ident:
		fmt.Fprintf(w, "[Ident] %s\n", n.Text)
	}
}
