package bnf

import "fmt"

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

func Print(a AST) {
	print(a, "")
}

func print(a AST, pad string) {
	if a.Text == "" {
		fmt.Printf("%s[%s]\n", pad, a.Type)
	} else {
		fmt.Printf("%s[%s] %s\n", pad, a.Type, a.Text)
	}
	for _, n := range a.Next {
		print(n, pad+"    ")
	}
}
