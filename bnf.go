package bnf

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
