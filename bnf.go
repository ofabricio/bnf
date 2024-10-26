package bnf

type AST struct {
	Type string
	Text string
	Pos  int
	Next []AST
}

func compact(a *AST) {
	if len(a.Next) == 1 {
		*a = a.Next[0]
	}
}

func swapRoot(a *AST) {
	if len(a.Next) == 1 && a.Next[0].Type == "ROOT" {
		// inp: Type[Root[Ident]]
		// out: Root[Type[Ident]]
		a.Next[0], a.Next[0].Next[0], *a = a.Next[0].Next[0], *a, a.Next[0]
	}
}
