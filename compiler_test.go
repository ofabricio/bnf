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
		m = ROOT( a b c )
		n = GROUP( a b c )
		o = JOIN( a b c )
		p = 'a'i 'b'r 'c'ri
		q = MATCH( a )i ( b )i c+i
		r = a:id
		s = TEXT() TEXT(a)
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
	//         [Function] ROOT
	//             [And]
	//                 [Ident] a
	//                 [Ident] b
	//                 [Ident] c
	//     [Stmt] n
	//         [Function] GROUP
	//             [And]
	//                 [Ident] a
	//                 [Ident] b
	//                 [Ident] c
	//     [Stmt] o
	//         [Function] JOIN
	//             [And]
	//                 [Ident] a
	//                 [Ident] b
	//                 [Ident] c
	//     [Stmt] p
	//         [And]
	//             [Ignore]
	//                 [Plain] a
	//             [Regex] ^b
	//             [Ignore]
	//                 [Regex] ^c
	//     [Stmt] q
	//         [And]
	//             [Ignore]
	//                 [Function] MATCH
	//                     [Ident] a
	//             [Ignore]
	//                 [Ident] b
	//             [Ignore]
	//                 [Quantifier] +
	//                     [Ident] c
	//     [Stmt] r
	//         [Type] id
	//             [Ident] a
	//     [Stmt] s
	//         [And]
	//             [Function] TEXT
	//             [Function] TEXT
	//                 [Ident] a
}
