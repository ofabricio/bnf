package bnf_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ofabricio/bnf"
)

func Example_customFunction() {

	INP := `One TWO`

	BNF := `root = LOWERCASE('One') SP LOWERCASE('TWO')`

	bnf.DefaultFuncs["LOWERCASE"] = func(p *bnf.Parser, arg bnf.BNF, out *[]bnf.AST) bool {
		var v []bnf.AST
		if p.ParseNode(arg, &v) {
			p.EmitIdent(out, strings.ToLower(v[0].Text))
			return true
		}
		return false
	}

	b := bnf.Compile(BNF)
	v := bnf.Parse(b, INP)

	bnf.Print(v)

	// Output:
	// [Group]
	//     [Ident] one
	//     [Ident] two
}

func Example_pos() {

	INP := `One Two Three`

	BNF := `
	    root = 'One' WS 'Two' WS 'Three'
	`

	b := bnf.Compile(BNF)
	v := bnf.Parse(b, INP)

	for _, v := range v.Next {
		fmt.Println(v.Text, v.Pos)
	}

	// Output:
	// One 0
	// Two 4
	// Three 8
}

func ExampleFlatten() {

	INP := `abcd`

	BNF := `
	    root = GROUP( 'a' GROUP( 'b' GROUP( 'c' GROUP( 'd' ) ) ) )
	`

	b := bnf.Compile(BNF)
	v := bnf.Parse(b, INP)

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

func ExampleParse_testDataBNF() {

	// This tests guarantees that the testdata parsing is correct.

	INP := `
[Test]
One

[Give]
Two

[When]
Three

[Then]
Four

[Test]
Five

[Give]
Six

[When]
Seven

[Then]
Eight
	`

	b := bnf.Compile(testDataBNF)
	v := bnf.Parse(b, INP)

	bnf.Print(v)

	// Output:
	// [Group]
	//     [Group]
	//         [Ident] One
	//         [Ident] Two
	//         [Ident] Three
	//         [Ident] Four
	//     [Group]
	//         [Ident] Five
	//         [Ident] Six
	//         [Ident] Seven
	//         [Ident] Eight
}

func TestParser(t *testing.T) {

	INP, err := os.ReadFile("testdata/parser.tests")
	if err != nil {
		t.Fatal(err)
	}

	b := bnf.Compile(testDataBNF)
	v := bnf.Parse(b, string(INP))

	if len(v.Next) == 0 {
		t.Fatal("No tests found")
	}

	var out strings.Builder
	out.Grow(512)
	for _, n := range v.Next {

		desc := n.Next[0].Text
		give := n.Next[1].Text
		when := n.Next[2].Text
		then := n.Next[3].Text

		b := bnf.Compile(when)
		v := bnf.Parse(b, give)

		bnf.Write(&out, v)
		got := strings.TrimSpace(out.String())
		exp := strings.TrimSpace(then)
		if got != exp {
			t.Errorf("\n[Test]\n%s\n[Give]\n%s\n[When]\n%s\n[Got]\n%s\n[Exp]\n%s\n", desc, give, when, got, exp)
		}
		out.Reset()
	}
}

var testDataBNF = `
    root = GROUP(section section section section)+
 section = ws tag '\n'ri MATCH( ANYNOT(ws (tag | EOF))+ )
     tag = '[Test]'i | '[Give]'i | '[When]'i | '[Then]'i
      ws = '\s*'ri
`
