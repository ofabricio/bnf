# bnf

Parse text using a [BNF](https://en.wikipedia.org/wiki/Backus%E2%80%93Naur_form)-like notation.

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

## Example 2

Parsing a simple Go function. [Go Playground](https://go.dev/play/p/s3hGGgIT8fh)

```go
import "github.com/ofabricio/bnf"

func main() {

	INP := `
	    func Say(msg string) {
	        fmt.Println(msg)
	    }
	`

	BNF := `
	    root = ws func
	    func = 'func'i ws name '('i GROUP(args) ')'i ws '{'i ws GROUP(body) ws '}'i
	    args = name ws name
	    body = name '.' name '('i name ')'i
	    name = '\w+'r
	      ws = '\s*'ri
	`

	b := bnf.Compile(BNF)
	v := bnf.Parse(b, INP)

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
```

## Operators

| Operator | Description |
| --- | --- |
| `'...'` | Match a plain text string. This emits a token. |
| `'...'r` | The string is a Regular Expression. |
| `'...'i` | Ignore the token (do not emit it). |
| `ROOT(a)` | Make the token `a` the root token. Works only in a logical AND operation. |
| `GROUP(a b c)` | Group the tokens `a b c`. |
| `NOT(a)` | Match any character that is not `a`. |
| `MATCH(a)` | Match `a` and return the matched text. |

## Default identifiers

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
