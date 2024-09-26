# bnf

Parse text using a BNF-like notation.

**Note:** this is still in a very early, unusable state.

## Example 1

Parsing a simple expression. [Go Playground](https://go.dev/play/p/nxZDwYEy42G)

```go
import "github.com/ofabricio/bnf"

func main() {

	theINP := `6+5*(4+3)*2`

	theBNF := `
	    expr   = EXPR1(term   '+' expr) | term
	    term   = EXPR1(factor '*' term) | factor
	    factor = I'(' expr I')' | value
	    value  = R'\d+'
	`

	b := bnf.Compile(theBNF)
	v := bnf.Parse(b, theINP)

	bnf.Print(v)

	// Output:
	// [Expr] +
	//     [Iden] 6
	//     [Expr] *
	//         [Iden] 5
	//         [Expr] *
	//             [Expr] +
	//                 [Iden] 4
	//                 [Iden] 3
	//             [Iden] 2
}
```

## Example 2

Parsing a simple Go function. [Go Playground](https://go.dev/play/p/v2VwFfrwPhl)

```go
import "github.com/ofabricio/bnf"

func main() {

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
| `R'...'` | The string is a Regular Expression. |
| `I'...'` | Ignore the token (do not emit it). |
| `EXPR1(a b c)` | Make the tokens `a b c` an expression where `b` will be the root token. |
| `GROUP(a b c)` | Group the tokens `a b c`. |
