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
	if c.s.Spaces(); c.ident(&i) && c.ost() && c.s.Match("=") && c.expr(out) {
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
		if !c.s.Match("|") {
			break
		}
	}
	if len(or) > 0 {
		v = AST{Type: "Or", Next: or}
		v.Compact()
		*out = v
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
		v = AST{Type: "And", Next: and}
		v.Compact()
		*out = v
		return true
	}
	return false
}

func (c *Compiler) factor(out *AST) bool {
	c.st()
	defer c.quantifier(out)
	return c.function(out) ||
		c.s.Match("(") && c.expr(out) && c.s.Match(")") ||
		c.value(out)
}

func (c *Compiler) function(out *AST) bool {
	if c.s.Match("EXPR1") {
		var v AST
		if c.factor(&v) && len(v.Next) == 3 {
			*out = AST{Type: "Func", Text: "EXPR1", Next: v.Next}
			return true
		}
	}
	if c.s.Match("GROUP") {
		var v AST
		if c.factor(&v) {
			*out = AST{Type: "Func", Text: "GROUP", Next: v.NextOrRoot()}
			return true
		}
	}
	return false
}

func (c *Compiler) quantifier(out *AST) bool {
	m := c.s.Mark()
	c.st()
	if m := c.s.Mark(); c.s.MatchChar("*+?") {
		*out = AST{Type: "Quant", Text: c.s.Text(m), Next: []AST{*out}}
		return true
	}
	c.s.Move(m)
	return true
}

func (c *Compiler) value(out *AST) bool {
	return c.ident(out) || c.text(out)
}

func (c *Compiler) ident(out *AST) bool {
	if m := c.s.Mark(); c.regex(reIden) {
		*out = AST{Type: "Ident", Text: c.s.Text(m)}
		return true
	}
	return false
}

func (c *Compiler) text(out *AST) bool {
	if m := c.s.Mark(); c.regex(reText) {
		t := c.s.Text(m)
		t = t[1 : len(t)-1]
		v := AST{Type: "Plain", Text: t}
		if c.s.MatchChar("r") {
			v = AST{Type: "Regex", Text: "^" + t}
		}
		if c.s.MatchChar("i") {
			v = AST{Type: "Ignore", Next: []AST{v}}
		}
		*out = v
		return true
	}
	return false
}

func (c *Compiler) regex(re *regexp.Regexp) bool {
	if v := re.FindIndex(c.s); v != nil {
		return c.s.Advance(v[1])
	}
	return false
}

// Space or Tab.
func (c *Compiler) st() bool {
	return c.s.WhileChar(" \t")
}

// Optional Space or Tab.
func (c *Compiler) ost() bool {
	return c.st() || true
}

var reText = regexp.MustCompile(`^'([^'\\]|\\.)*'`)
var reIden = regexp.MustCompile(`^\w+`)
