# bnf

Parse text using a BNF-like notation.

**Note:** this is still in a very early, unusable state.

## Example 1

Parsing a simple expression. [Go Playground](https://go.dev/play/p/w7o5Dtxw34I)

```go
import "github.com/ofabricio/bnf"

func main() {

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

Parsing a simple Go function. [Go Playground](https://go.dev/play/p/eEtoeR1eCFy)

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
```

## Operators

| Operator | Description |
| --- | --- |
| `'...'` | Match a plain text string. This emits a token. |
| `'...'r` | The string is a Regular Expression. |
| `'...'i` | Ignore the token (do not emit it). |
| `EXPR1(a b c)` | Make the tokens `a b c` an expression where `b` will be the root token. |
| `GROUP(a b c)` | Group the tokens `a b c`. |
| `UNTIL(a)` | Match any character until `a` and return the text until it. Does not include `a`. |
| `MATCH(a)` | Match `a` and return the matched text. |


## Default identifiers

These are identifiers that don't need to be defined in the BNF.
If defined they will be overridden:

```
ws = ' \t\r\n'i
sp = ' 'i
st = ' \t'i
nl = 'nl'i
tb = '\t'i
EOF
```
