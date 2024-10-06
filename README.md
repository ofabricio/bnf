# bnf

Parse text using a [BNF](https://en.wikipedia.org/wiki/Backus%E2%80%93Naur_form)-like notation.

[![playground](https://img.shields.io/badge/play-ground-blueviolet.svg?logo=github&labelColor=2b3137)](https://ofabricio.github.io/bnf/)

## Note

This is still in a very early, unusable state.

## Example 1

Parsing a simple expression. [Go Playground](https://go.dev/play/p/HFlZ7h_OlGl)

```go
import "github.com/ofabricio/bnf"

func main() {

	INP := `6+5*(4+3)*2`

	BNF := `
	    expr   = term   ROOT('+') expr | term
	    term   = factor ROOT('*') term | factor
	    factor = '('i expr ')'i | value
	    value  = '\d+'r
	`

	b := bnf.Compile(BNF)
	v := bnf.Parse(b, INP)

	bnf.Print(v)

	// Output:
	// [Ident] +
	//     [Ident] 6
	//     [Ident] *
	//         [Ident] 5
	//         [Ident] *
	//             [Ident] +
	//                 [Ident] 4
	//                 [Ident] 3
	//             [Ident] 2
}
```

Homework: try adding support for whitespaces, minus, division and negative numbers.

## Example 2

Parsing a simple JSON that supports only strings and no whitespaces.
[Go Playground](https://go.dev/play/p/BBIwl0pENfW)

```go
import "github.com/ofabricio/bnf"

func main() {

	INP := `{"name":"John","addresses":[{"zip":"111"},{"zip":"222"}]}`

	BNF := `
 	    val = obj | arr | str
	    obj = GROUP( '{'i okv (','i okv)* '}'i ):Object
	    arr = GROUP( '['i val (','i val)* ']'i ):Array
	    okv = str ':'i val
	    str = MATCH( '"' ( NOT('"' | '\\') | '\\' any )* '"' )
	`

	b := bnf.Compile(BNF)
	v := bnf.Parse(b, INP)

	bnf.Print(v)

	// Output:
	// [Object]
	//     [Ident] "name"
	//     [Ident] "John"
	//     [Ident] "addresses"
	//     [Array]
	//         [Object]
	//             [Ident] "zip"
	//             [Ident] "111"
	//         [Object]
	//             [Ident] "zip"
	//             [Ident] "222"
}
```

Homework: try adding support for whitespaces, numbers, booleans and null.

## Quick Reference

| Operator | Description |
| --- | --- |
| `'...'` | Match a plain text string. This emits a token. |
| `'...'r` | The string is a Regular Expression. |
| `'...'i` | Ignore the token (do not emit it). |
| `ROOT(a)` | Make the token a root token. Works only in a logical AND operation. |
| `GROUP(a)` | Group the tokens. |
| `NOT(a)` | Match any character that is not `a`. |
| `MATCH(a)` | Return the matched text instead of a tree of tokens. |
| `SCAN(a)` | Scan through the entire input text. Useful to collect data. |

### Default identifiers

These identifiers don't need to be defined in the BNF.
If defined they will be overridden.

| Identifier | Description |
| --- | --- |
| `ws` | Match a whitespace character `'\s'ri` |
| `sp` | Match a space character `' 'i` |
| `st` | Match a space or tab character `'[ \t]'ri` |
| `nl` | Match a newline character `'\n'ri` |
| `tb` | Match a tab character `'\t'ri` |
| `any` | Match any character `'.'ri` |
| `EOF` | Match if the scanner is at the end. |

## Introduction

[BNF](https://en.wikipedia.org/wiki/Backus%E2%80%93Naur_form) is a notation used to describe
the syntax of programming languages or other formal languages. It describes how to combine
different symbols to produce a syntactically correct sequence of tokens.

It consists of three components: (1) a set of *nonterminal* symbols; (2) a set of *terminal* symbols;
and (3) rules for replacing nonterminal symbols with a sequence of symbols.

These so-called "derivation rules" are written as `identifier = expression`, where `identifier`
is a nonterminal symbol and `expression` consists of one or more sequences of either terminal
or nonterminal symbols.

Consider this possible BNF for a math expression:

```go
expr = num '+' num
num  = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9'
```

The strings `'...'` are terminal symbols and everything else is nonterminal symbol (or identifier).
These identifiers will be replaced by the terminal symbols as the parser advances.

Now consider this expression as input:

```
6+5*(4+3)
```

By applying the previous BNF to the input above we get as result:

```
[Group]
    [Ident] 6
    [Ident] +
    [Ident] 5
```

Not very exciting at first, but if we get a result it means the parsing succeeded
and we have a syntactically correct sequence of tokens. Though we haven't exhausted the whole input.

Noteworthy points:

- We always get a flat tree like the above as result.
  Only by using functions we can get a real tree out of it; they are described in the next section.
- The parser starts from upper left (`expr = num ...`) and goes rightward.
- Every terminal symbol emits a token.
- There can be recursive calls, when a statement calls itself (see [Example 1](#example-1)).

This is the struct returned by the `bnf.Parse(b, src)` function:

```go
type AST struct {
    Type string
    Text string
    Next []AST
}
```

`Type` is the node type, that can be used to identify a group of nodes.
`Text` is the token emited for a type.
`Next` has the children nodes.

We can then parse that resulting tree to create something nicer, like a programming language;
or we can simply use it to collect data from text like a (maybe more readable) regular expression.

## Documentation

### String

Strings `'...'` emit nodes whenever they match.

```go
INP := `Hello World`

BNF := `
    root = 'Hello' ' ' 'World'
`

b := bnf.Compile(BNF)
v := bnf.Parse(b, INP)

bnf.Print(v)

// Output:
// [Group]
//     [Ident] Hello
//     [Ident]  
//     [Ident] World
```

### Regex

The `r` flag makes a [String](#string) a regular expression instead.

```go
INP := `Hello World`

BNF := `
    root = '\w+'r '\s'r '\w+'r
`

b := bnf.Compile(BNF)
v := bnf.Parse(b, INP)

bnf.Print(v)

// Output:
// [Group]
//     [Ident] Hello
//     [Ident]  
//     [Ident] World
```

### Ignore

The `i` flag prevents a node from being emitted.

```go
INP := `Hello World`

BNF := `
    root = 'Hello' ' 'i 'World'
`

b := bnf.Compile(BNF)
v := bnf.Parse(b, INP)

bnf.Print(v)

// Output:
// [Group]
//     [Ident] Hello
//     [Ident] World
```

### And

Any sequence of symbols is considered an And expression.
For example `a b c` reads as match `a` AND `b` AND `c`.
All of them needs to succeed for the matching to continue.
`And` has precedence over [Or](#or).

```go
INP := `abc`

BNF := `
    root = 'a' 'b' 'c'
`

b := bnf.Compile(BNF)
v := bnf.Parse(b, INP)

bnf.Print(v)

// Output:
// [Group]
//     [Ident] a
//     [Ident] b
//     [Ident] c
```

### Or

Any sequence of symbols separeted by a `|` is considered an Or expression.
For example `a | b | c` reads as match `a` OR `b` OR `c`.
If one of them succeeds the matching continues.
[And](#and) has precedence over `Or`.

```go
INP := `cba`

BNF := `
    root = 'a' | 'b' | 'c'
`

b := bnf.Compile(BNF)
v := bnf.Parse(b, INP)

bnf.Print(v)

// Output:
// [Ident] c
```

### GROUP

This function groups nodes together.

```go
INP := `OneTwoThreeFourFiveSixSevenEight`

BNF := `
    root = 'One' GROUP('Two' 'Three') GROUP('Four' GROUP('Five' 'Six') 'Seven') 'Eight'
`

b := bnf.Compile(BNF)
v := bnf.Parse(b, INP)

bnf.Print(v)

// Output:
// [Group]
//     [Ident] One
//     [Group]
//         [Ident] Two
//         [Ident] Three
//     [Group]
//         [Ident] Four
//         [Group]
//             [Ident] Five
//             [Ident] Six
//         [Ident] Seven
//     [Ident] Eight
```

### MATCH

This function emits a node with the matched text instead.

```go
INP := `OneTwoThreeFour`

BNF := `
    root = 'One' 'Two' MATCH('Three' 'Four')
`

b := bnf.Compile(BNF)
v := bnf.Parse(b, INP)

bnf.Print(v)

// Output:
// [Group]
//     [Ident] One
//     [Ident] Two
//     [Ident] ThreeFour
```

### NOT

This function matches any character that is not the argument.

```go
INP := `A`

BNF := `
    root = NOT('B')
`

b := bnf.Compile(BNF)
v := bnf.Parse(b, INP)

bnf.Print(v)

// Output:
// [Ident] A
```

### ROOT

This function makes the node a root node of a branch.
It works only in a logical [And](#and).
It is not possible to have two `ROOT` functions in the same `And` sequence; use parentheses to start
a new sequence.

```go
INP := `2+3`

BNF := `
    root = '2' ROOT('+') '3'
`

b := bnf.Compile(BNF)
v := bnf.Parse(b, INP)

bnf.Print(v)

// Output:
// [Ident] +
//     [Ident] 2
//     [Ident] 3
```

### Quantifiers

Quantifiers have the same behavior as in a regular expression.

- `?` matches zero or one token.
- `*` matches zero or many tokens.
- `+` matches one or many tokens.

```go
INP := `OneOneTwo`

BNF := `
    root = 'One'+
`

b := bnf.Compile(BNF)
v := bnf.Parse(b, INP)

bnf.Print(v)

// Output:
// [Group]
//     [Ident] One
//     [Ident] One
```

### Type

Use `:type` suffix in any node to rename the type.

```go
INP := `2+3`

BNF := `
    root = num '+':Opr num
     num = '\w+'r:Val
`

b := bnf.Compile(BNF)
v := bnf.Parse(b, INP)

bnf.Print(v)

// Output:
// [Group]
//     [Val] 2
//     [Opr] +
//     [Val] 3
```

### SCAN

This function scans through the entire text input; this is useful to collect data.

```go
INP := `The standard chunk of Lorem Ipsum used since the 1500s
is reproduced below for those interested. Sections 1.10.32 and
1.10.33 from "de Finibus Bonorum et Malorum" by Cicero are also
reproduced in their exact original form, accompanied by English
versions from the 1914 translation by H. Rackham.`

BNF := `
    root = SCAN(ver | num)
     num = '\d+'r
     ver = MATCH(num '.' num '.' num)
`

b := bnf.Compile(BNF)
v := bnf.Parse(b, INP)

bnf.Print(v)

// Output:
// [Group]
//     [Ident] 1500
//     [Ident] 1.10.32
//     [Ident] 1.10.33
//     [Ident] 1914
```
