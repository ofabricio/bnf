package bnf

import (
	"regexp"

	"github.com/ofabricio/scan"
)

func Compile(bnf string) AST {
	return NewCompiler(bnf).Compile()
}

func NewCompiler(bnf string) *Compiler {
	return &Compiler{s: scan.Bytes(bnf)}
}

type Compiler struct {
	s scan.Bytes
}

func (c *Compiler) Compile() AST {
	var stmts []AST
	c.stmts(&stmts)
	return AST{Type: "Root", Next: stmts}
}

func (c *Compiler) stmts(out *[]AST) bool {
	var stmt AST
	for c.stmt(&stmt) {
		*out = append(*out, stmt)
	}
	return len(*out) > 0
}

func (c *Compiler) stmt(out *AST) bool {
	var i AST
	if c.s.Spaces(); c.ident(&i) && c.ost() && c.s.MatchChar("=") && c.expr(out) {
		*out = AST{Type: "Stmt", Next: []AST{i, *out}}
		return true
	}
	return false
}

func (c *Compiler) expr(out *AST) bool {
	var or []AST
	var v AST
	for c.term(&v) {
		or = append(or, v)
		if !c.s.MatchChar("|") {
			break
		}
	}
	if len(or) > 0 {
		*out = AST{Type: "Or", Next: or}
		out.Compact()
		return true
	}
	return false
}

func (c *Compiler) term(out *AST) bool {
	var and []AST
	var v AST
	for c.factor(&v) {
		and = append(and, v)
	}
	if len(and) > 0 {
		*out = AST{Type: "And", Next: and}
		out.Compact()
		return true
	}
	return false
}

func (c *Compiler) factor(out *AST) bool {
	c.st()
	return (c.s.MatchChar("(") && c.expr(out) && c.s.MatchChar(")") ||
		c.function(out) ||
		c.value(out)) &&
		(c.quantifier(out) || true)
}

func (c *Compiler) function(out *AST) bool {
	if c.s.Match("ROOT") && c.s.MatchChar("(") && c.expr(out) && c.s.MatchChar(")") {
		*out = AST{Type: "ROOT", Next: []AST{*out}}
		return true
	}
	if c.s.Match("GROUP") && c.s.MatchChar("(") && c.expr(out) && c.s.MatchChar(")") {
		*out = AST{Type: "GROUP", Next: []AST{*out}}
		return true
	}
	if c.s.Match("NOT") && c.s.MatchChar("(") && c.expr(out) && c.s.MatchChar(")") {
		*out = AST{Type: "NOT", Next: []AST{*out}}
		return true
	}
	if c.s.Match("MATCH") && c.s.MatchChar("(") && c.expr(out) && c.s.MatchChar(")") {
		*out = AST{Type: "MATCH", Next: []AST{*out}}
		return true
	}
	if c.s.Match("SCAN") && c.s.MatchChar("(") && c.expr(out) && c.s.MatchChar(")") {
		n := AST{Type: "Ident", Text: "any"}
		o := AST{Type: "Or", Next: []AST{*out, n}}
		q := AST{Type: "*", Next: []AST{o}}
		*out = q
		return true
	}
	return false
}

func (c *Compiler) quantifier(out *AST) bool {
	c.st()
	if m := c.s.Mark(); c.s.MatchChar("*+?") {
		*out = AST{Type: c.s.Text(m), Next: []AST{*out}}
		return true
	}
	return false
}

func (c *Compiler) value(out *AST) bool {
	return c.ident(out) || c.text(out)
}

func (c *Compiler) ident(out *AST) bool {
	if m := c.s.Mark(); c.matchRegex(reIden) {
		*out = AST{Type: "Ident", Text: c.s.Text(m)}
		return true
	}
	return false
}

func (c *Compiler) text(out *AST) bool {
	if c.plain(out) {
		c.regexFlag(out)
		c.ignoreFlag(out)
		return true
	}
	return false
}

func (c *Compiler) plain(out *AST) bool {
	if m := c.s.Mark(); c.matchRegex(reText) {
		t := c.s.Text(m)
		t = t[1 : len(t)-1]
		*out = AST{Type: "Plain", Text: t}
		return true
	}
	return false
}

func (c *Compiler) regexFlag(out *AST) bool {
	if c.s.MatchChar("r") {
		*out = AST{Type: "Regex", Text: "^" + out.Text}
		return true
	}
	return false
}

func (c *Compiler) ignoreFlag(out *AST) bool {
	if c.s.MatchChar("i") {
		*out = AST{Type: "Ignore", Next: []AST{*out}}
		return true
	}
	return false
}

func (c *Compiler) matchRegex(re *regexp.Regexp) bool {
	if v := re.FindIndex(c.s); v != nil {
		return c.s.Advance(v[1])
	}
	return false
}

// Optional Space or Tab.
func (c *Compiler) ost() bool {
	return c.st() || true
}

// Space or Tab.
func (c *Compiler) st() bool {
	return c.s.WhileChar(" \t")
}

var reText = regexp.MustCompile(`^'([^'\\]|\\.)*'`)
var reIden = regexp.MustCompile(`^\w+`)
