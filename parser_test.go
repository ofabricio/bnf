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
	    factor = '('i expr ')'i | value
	    value  = '\d+'r
	`

	b := bnf.Compile(theBNF)
	v := bnf.Parse(b, theINP)

	bnf.Print(v)

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
	    func = 'func'i ws name '('i args ')'i ws '{'i ws body ws '}'i
	    args = name ws name
	    body = name '.' name '('i name ')'i
	    name = '\w+'r
	      ws = '\s*'ri
	`

	b := bnf.Compile(theBNF)
	v := bnf.Parse(b, theINP)

	bnf.Print(v)

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
	bnf.Print(v)
	fmt.Println("FLATTEN 0:")
	bnf.Print(bnf.AST{Type: "Flatten", Next: bnf.Flatten(v, 0)})
	fmt.Println("FLATTEN 1:")
	bnf.Print(bnf.AST{Type: "Flatten", Next: bnf.Flatten(v, 1)})
	fmt.Println("FLATTEN 2:")
	bnf.Print(bnf.AST{Type: "Flatten", Next: bnf.Flatten(v, 2)})
	fmt.Println("FLATTEN 3:")
	bnf.Print(bnf.AST{Type: "Flatten", Next: bnf.Flatten(v, 3)})

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
           s = ' 'i
	`

	b := bnf.Compile(theBNF)
	v := bnf.Parse(b, theINP)

	bnf.Print(v)

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
