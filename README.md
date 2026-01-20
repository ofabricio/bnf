# BNF

Parse text using a [BNF](https://en.wikipedia.org/wiki/Backus%E2%80%93Naur_form)-like notation.

[![playground](https://img.shields.io/badge/play-ground-blueviolet.svg?logo=github&labelColor=2b3137)](https://ofabricio.github.io/bnf/)

## Note

This is still in a very early, unusable state.

## Install

```
go get github.com/ofabricio/bnf
```

## Example 1

Parsing a simple expression. [Go Playground](https://go.dev/play/p/HFlZ7h_OlGl)

```go
import "github.com/ofabricio/bnf"

func main() {

    INP := `6+5*(4+3)*2`

    BNF := `
        expr = ROOT(term '+' expr) | term
        term = ROOT(fact '*' term) | fact
        fact = '('i expr ')'i | numb
        numb = '\d+'r
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
[Go Playground](https://go.dev/play/p/HAr1ULSu8q3)

```go
import "github.com/ofabricio/bnf"

func main() {

    INP := `{"name":"John","addresses":[{"zip":"111"},{"zip":"222"}]}`

    BNF := `
        val = obj | arr | str
        obj = GROUP( '{'i okv (','i okv)* '}'i ):Object
        arr = GROUP( '['i val (','i val)* ']'i ):Array
        okv = str ':'i val
        str = MATCH( '"' ( ANYNOT('"' | '\\') | '\\' any )* '"' )
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

# Quick Reference

| Operator | Description |
| --- | --- |
| `'...'` | Match a plain text string. This emits a token. |
| `'...'r` | The string is a Regular Expression. |
| `'...'i` | Ignore the token (do not emit it). |
| `ROOT(a b c)` | Make `b` a root token. |
| `GROUP(a)` | Group the tokens. |
| `ANYNOT(a)` | Match any character that is not `a`. |
| `JOIN(a)` | Join nodes into one. |
| `MATCH(a)` | Match nodes into one. |
| `TEXT(a)` | Add a text node in the tree. |
| `SAVE(a)` | Save a token to be loaded with `LOAD()`. |
| `LOAD()` | Load a token saved with `SAVE()`. |
| `FIND(a)` | Scan through the input text until a match is found and emit it. |
| `REVERSE(a b)` | Reverse the token positions. |

**Note:** there is support for custom functions by adding them to `bnf.DefaultFuncs` map.

### Default identifiers

These identifiers don't need to be defined in the BNF.
If defined they will be overridden.

| Identifier | Description |
| --- | --- |
| `WS` | Match a whitespace character `'\s'ri` |
| `SP` | Match a space character `' 'i` |
| `ST` | Match a space or tab character `'[ \t]'ri` |
| `NL` | Match a newline character `'\n'ri` |
| `TB` | Match a tab character `'\t'ri` |
| `ANY` | Match any character `'.'ri` |
| `any` | Match any character `'.'r` |
| `EOF` | Match if the scanner is at the end. |
| `MORE` | Match if the scanner has more to scan. |
| `SKIPLINE` | Skip a line. |

# Introduction

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
- Every terminal symbol emits a token (a node in the tree).
- When the parsing fails, a node of type `Error` is returned; it has the position where the parser failed.
- There can be recursive calls, when a statement calls itself (see [Example 1](#example-1)).

This is the struct returned by the `bnf.Parse(b, src)` function:

```go
type AST struct {
    Type string // The type to identify a group of nodes.
    Text string // The token.
    Pos  int    // Position of the token.
    Next []AST  // Children nodes.
}
```

We can then parse that resulting tree to create something nicer, like a programming language;
or we can simply use it to collect data from text like a (maybe more readable) regular expression.

# Documentation

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
All of them need to succeed for the matching to continue.
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

By default all matching tokens are emitted separatelly, one node for each matching token.
This function groups these nodes together.

Try removing the `GROUP` functions in the example below and see how the result changes.

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

### JOIN

By default all matching tokens are emitted separatelly, one node for each matching token.
This function joins these nodes into one. The order of the nodes matter.

In the example below, note:
- How `1+1` are emitted separatelly by default.
- How `2+2` are joined into one node.
- How `3+3` are joined in a strange order due to the [ROOT](#root) function,
  that changes the order of the nodes.
- How the `+` gets removed from `4+4` due to the [Ignore](#ignore) flag `i`.

```go
INP := `1+1 2+2 3+3 4+4`

BNF := `
    root = (num '+' num) SP JOIN(num '+' num) SP JOIN(ROOT(num '+' num)) SP JOIN(num '+'i num)
     num = '\d+'r
`

b := bnf.Compile(BNF)
v := bnf.Parse(b, INP)

bnf.Print(v)

// Output:
// [Group]
//     [Ident] 1
//     [Ident] +
//     [Ident] 1
//     [Ident] 2+2
//     [Ident] +33
//     [Ident] 44
```

### MATCH

By default all matching tokens are emitted separatelly, one node for each matching token.
This function matches these nodes into one.
It takes the text between the first and the last matching token,
returning a node with it.

It is more efficient than [JOIN](#join), but some operators
don't work with it, for example the [Ignore](#ignore) flag.

In the example below, note:

- How `1+1` are emitted separatelly by default.
- How `2+2` are joined into one node.
- How `3+3` are joined into one node even though [ROOT](#root) changes the order of the nodes.
- How the `+` is not removed from `4+4` even though [Ignore](#ignore) flag `i` is present.

```go
INP := `1+1 2+2 3+3 4+4`

BNF := `
 root = (num '+' num) SP MATCH(num '+' num) SP MATCH(ROOT(num '+' num)) SP MATCH(num '+'i num)
  num = '\d+'r
`

b := bnf.Compile(BNF)
v := bnf.Parse(b, INP)

bnf.Print(v)

// Output:
// [Group]
//     [Ident] 1
//     [Ident] +
//     [Ident] 1
//     [Ident] 2+2
//     [Ident] 3+3
//     [Ident] 4+4
```

### ANYNOT

This function matches any character that is not the argument.

```go
INP := `A`

BNF := `
    root = ANYNOT('B')
`

b := bnf.Compile(BNF)
v := bnf.Parse(b, INP)

bnf.Print(v)

// Output:
// [Ident] A
```

### ROOT

This function takes 3 arguments and when all of them match it makes the second argument a root node.

```go
INP := `2+3`

BNF := `
    root = ROOT('2' '+' '3')
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

Use `:type` suffix in any node to rename its type.

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

### TEXT

This function adds a text node in the tree; this is useful to add a default value.

It accepts either a plain string argument `TEXT('example')` or no argument `TEXT()` to add an empty node.

```go
INP := `
    Key1=Value1
    Key2=
    Key3=Value3
`

BNF := `
    root = FIND( pair )+
    pair = val '='i ( val | TEXT('Default Value') )
     val = '\w+'r
`

b := bnf.Compile(BNF)
v := bnf.Parse(b, INP)

bnf.Print(v)

// Output:
// [Group]
//     [Ident] Key1
//     [Ident] Value1
//     [Ident] Key2
//     [Ident] Default Value
//     [Ident] Key3
//     [Ident] Value3
```

### SAVE and LOAD (Backreference)

These functions work together to allow for Backreference.

In this example we guarantee that the closing tags have the same name of the opening tags.

```go
INP := `<a>hello<b>world</b></a>`

BNF := `
    tag = GROUP( MATCH('<' SAVE(w) '>') ( w | tag )* MATCH('</' LOAD() '>') )
      w = '\w+'r
`

b := bnf.Compile(BNF)
v := bnf.Parse(b, INP)

bnf.Print(v)

// Output:
// [Group]
//     [Ident] <a>
//     [Ident] hello
//     [Group]
//         [Ident] <b>
//         [Ident] world
//         [Ident] </b>
//     [Ident] </a>
```

### FIND

This function scans through the input text until a match is found and emit it.

```go
INP := `The standard chunk of Lorem Ipsum used since the 1500s
is reproduced below for those interested. Sections 1.10.32 and
1.10.33 from "de Finibus Bonorum et Malorum" by Cicero are also
reproduced in their exact original form, accompanied by English
versions from the 1914 translation by H. Rackham.`

BNF := `
    root = FIND(ver | num)+
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

### REVERSE

This function reverses the token positions.

```go
INP := `OneTwoThree`

BNF := `
    root = REVERSE('One' 'Two' 'Three')
`

b := bnf.Compile(BNF)
v := bnf.Parse(b, INP)

bnf.Print(v)

// Output:
// [Group]
//     [Ident] Three
//     [Ident] Two
//     [Ident] One
```
