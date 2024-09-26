package bnf_test

import (
	"fmt"

	"github.com/ofabricio/bnf"
)

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

func ExampleFlatten() {

	theINP := `abcd`

	theBNF := `
	    root = GROUP( 'a' GROUP( 'b' GROUP( 'c' GROUP( 'd' ) ) ) )
	`

	b := bnf.Compile(theBNF)
	v := bnf.Parse(b, theINP)

	fmt.Println("AST:")
	bnf.Print(v, 0)
	fmt.Println("FLATTEN 0:")
	bnf.Print(bnf.AST{Type: "Flatten", Next: bnf.Flatten(v, 0)}, 0)
	fmt.Println("FLATTEN 1:")
	bnf.Print(bnf.AST{Type: "Flatten", Next: bnf.Flatten(v, 1)}, 0)
	fmt.Println("FLATTEN 2:")
	bnf.Print(bnf.AST{Type: "Flatten", Next: bnf.Flatten(v, 2)}, 0)
	fmt.Println("FLATTEN 3:")
	bnf.Print(bnf.AST{Type: "Flatten", Next: bnf.Flatten(v, 3)}, 0)

	// Output:
	// AST:
	// [Group]
	//     [Ident] a
	//     [Group]
	//         [Ident] b
	//         [Group]
	//             [Ident] c
	//             [Group]
	//                 [Ident] d
	// FLATTEN 0:
	// [Flatten]
	//     [Ident] a
	//     [Ident] b
	//     [Ident] c
	//     [Ident] d
	// FLATTEN 1:
	// [Flatten]
	//     [Ident] a
	//     [Ident] b
	//     [Group]
	//         [Ident] c
	//         [Group]
	//             [Ident] d
	// FLATTEN 2:
	// [Flatten]
	//     [Ident] a
	//     [Ident] b
	//     [Ident] c
	//     [Group]
	//         [Ident] d
	// FLATTEN 3:
	// [Flatten]
	//     [Ident] a
	//     [Ident] b
	//     [Ident] c
	//     [Ident] d
}

func Example_quantifier() {

	theINP := `a aa b bb c d`

	theBNF := `
	    root = 'a'+ s 'a'+ s 'a'* 'b'* s 'b'* s 'c'? s 'c'? 'd'
           s = I' '
	`

	b := bnf.Compile(theBNF)
	v := bnf.Parse(b, theINP)

	bnf.Print(v, 0)

	// Output:
	// [Group]
	//     [Ident] a
	//     [Group]
	//         [Ident] a
	//         [Ident] a
	//     [Ident] b
	//     [Group]
	//         [Ident] b
	//         [Ident] b
	//     [Ident] c
	//     [Ident] d
}
