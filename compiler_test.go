package bnf_test

import (
	"github.com/ofabricio/bnf"
)

func ExampleCompile() {

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
		m = FUNCA() FUNCB(a) FUNCC(a b)
		n = 'a'i 'b'r 'c'ri
		o = FUNC(a)i (b)i c+i
		p = a:id
	`

	v := bnf.Compile(s)

	bnf.PrintBNF(v)

	// Output:
	// [Root]
	//     [Stmt] a
	//         [Ident] a
	//     [Stmt] b
	//         [And]
	//             [Ident] a
	//             [Ident] b
	//     [Stmt] c
	//         [Or]
	//             [Ident] a
	//             [Ident] b
	//     [Stmt] d
	//         [And]
	//             [Ident] a
	//             [Ident] b
	//             [Ident] c
	//     [Stmt] e
	//         [Or]
	//             [And]
	//                 [Ident] a
	//                 [Ident] b
	//             [Ident] c
	//     [Stmt] f
	//         [Or]
	//             [Ident] a
	//             [And]
	//                 [Ident] b
	//                 [Ident] c
	//     [Stmt] g
	//         [Or]
	//             [Ident] a
	//             [Ident] b
	//             [Ident] c
	//     [Stmt] h
	//         [Or]
	//             [And]
	//                 [Ident] a
	//                 [Ident] b
	//             [And]
	//                 [Ident] c
	//                 [Ident] d
	//     [Stmt] i
	//         [Or]
	//             [Ident] a
	//             [And]
	//                 [Ident] b
	//                 [Ident] c
	//             [Ident] d
	//     [Stmt] j
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
	//     [Stmt] k
	//         [And]
	//             [Or]
	//                 [Ident] a
	//                 [Ident] b
	//             [Or]
	//                 [Ident] c
	//                 [Ident] d
	//     [Stmt] l
	//         [And]
	//             [Quantifier] *
	//                 [Ident] a
	//             [Quantifier] +
	//                 [Ident] b
	//             [Quantifier] *
	//                 [Or]
	//                     [Quantifier] ?
	//                         [And]
	//                             [Ident] c
	//                             [Ident] d
	//                     [Ident] e
	//     [Stmt] m
	//         [And]
	//             [Function] FUNCA
	//             [Function] FUNCB
	//                 [Ident] a
	//             [Function] FUNCC
	//                 [And]
	//                     [Ident] a
	//                     [Ident] b
	//     [Stmt] n
	//         [And]
	//             [Ignore]
	//                 [Plain] a
	//             [Regex] ^b
	//             [Ignore]
	//                 [Regex] ^c
	//     [Stmt] o
	//         [And]
	//             [Ignore]
	//                 [Function] FUNC
	//                     [Ident] a
	//             [Ignore]
	//                 [Ident] b
	//             [Ignore]
	//                 [Quantifier] +
	//                     [Ident] c
	//     [Stmt] p
	//         [Type] id
	//             [Ident] a
}
