package bnf_test

import "github.com/ofabricio/bnf"

func Example_expr() {

	theINP := `6+5*(4+3)*2`

	theBNF := `
	    expr   = EXPR1(term   '+' expr) | term
	    term   = EXPR1(factor '*' term) | factor
	    factor = I'(' expr I')' | value
	    value  = R'\d+'
	`

	b := bnf.Compile(theBNF)
	v := bnf.Parse(b, theINP)

	bnf.Print(v, 0)

	// Output:
	// [Expr] +
	//     [Ident] 6
	//     [Expr] *
	//         [Ident] 5
	//         [Expr] *
	//             [Expr] +
	//                 [Ident] 4
	//                 [Ident] 3
	//             [Ident] 2
}

func Example_group() {

	theINP := `
	    func Say(msg string) {
	        fmt.Println(msg)
	    }
	`

	theBNF := `
	    root = ws func
	    func = I'func' ws name I'(' args I')' ws I'{' ws body ws I'}'
	    args = name ws name
	    body = name '.' name I'(' name I')'
	    name = R'\w+'
	      ws = IR'\s+'
	`

	b := bnf.Compile(theBNF)
	v := bnf.Parse(b, theINP)

	bnf.Print(v, 0)

	// Output:
	// [Group]
	//     [Ident] Say
	//     [Group]
	//         [Ident] msg
	//         [Ident] string
	//     [Group]
	//         [Ident] fmt
	//         [Ident] .
	//         [Ident] Println
	//         [Ident] msg
}
