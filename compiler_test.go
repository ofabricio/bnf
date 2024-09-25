package bnf_test

import (
	"github.com/ofabricio/bnf"
)

func Example_all() {

	s := `
		a = a
		b = a b
		c = a | b
		d = a b c
		e = a b | c
		f = a | b c
		g = a | b | c
		h = a b | c d
		i = a | b c | d
		j = a ( b ( c | d ) | ( e | f ) g )
		k = ( a | b ) ( c | d )
		l = a * b+ ( ( c d ) ? | e )*
		m = EXPR1( a b c )
		n = GROUP( a b c )
	`

	v := bnf.Compile(s)

	bnf.Print(v, 0)

	// Output:
	// [Root]
	//     [Stmt]
	//         [Ident] a
	//         [Ident] a
	//     [Stmt]
	//         [Ident] b
	//         [And]
	//             [Ident] a
	//             [Ident] b
	//     [Stmt]
	//         [Ident] c
	//         [Or]
	//             [Ident] a
	//             [Ident] b
	//     [Stmt]
	//         [Ident] d
	//         [And]
	//             [Ident] a
	//             [And]
	//                 [Ident] b
	//                 [Ident] c
	//     [Stmt]
	//         [Ident] e
	//         [Or]
	//             [And]
	//                 [Ident] a
	//                 [Ident] b
	//             [Ident] c
	//     [Stmt]
	//         [Ident] f
	//         [Or]
	//             [Ident] a
	//             [And]
	//                 [Ident] b
	//                 [Ident] c
	//     [Stmt]
	//         [Ident] g
	//         [Or]
	//             [Ident] a
	//             [Or]
	//                 [Ident] b
	//                 [Ident] c
	//     [Stmt]
	//         [Ident] h
	//         [Or]
	//             [And]
	//                 [Ident] a
	//                 [Ident] b
	//             [And]
	//                 [Ident] c
	//                 [Ident] d
	//     [Stmt]
	//         [Ident] i
	//         [Or]
	//             [Ident] a
	//             [Or]
	//                 [And]
	//                     [Ident] b
	//                     [Ident] c
	//                 [Ident] d
	//     [Stmt]
	//         [Ident] j
	//         [And]
	//             [Ident] a
	//             [Or]
	//                 [And]
	//                     [Ident] b
	//                     [Or]
	//                         [Ident] c
	//                         [Ident] d
	//                 [And]
	//                     [Or]
	//                         [Ident] e
	//                         [Ident] f
	//                     [Ident] g
	//     [Stmt]
	//         [Ident] k
	//         [And]
	//             [Or]
	//                 [Ident] a
	//                 [Ident] b
	//             [Or]
	//                 [Ident] c
	//                 [Ident] d
	//     [Stmt]
	//         [Ident] l
	//         [And]
	//             [Quant] *
	//                 [Ident] a
	//             [And]
	//                 [Quant] +
	//                     [Ident] b
	//                 [Quant] *
	//                     [Or]
	//                         [Quant] ?
	//                             [And]
	//                                 [Ident] c
	//                                 [Ident] d
	//                         [Ident] e
	//     [Stmt]
	//         [Ident] m
	//         [Func] EXPR1
	//             [Ident] a
	//             [Ident] b
	//             [Ident] c
	//     [Stmt]
	//         [Ident] n
	//         [Func] GROUP
	//             [Ident] a
	//             [Ident] b
	//             [Ident] c
}
