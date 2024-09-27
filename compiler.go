package bnf

import (
	"fmt"
	"regexp"

	"github.com/ofabricio/scan"
)

func Compile(bnf string) AST {
	c := &Compiler{
		s: scan.Bytes(bnf),
	}
	return c.Compile()
}

type Compiler struct {
	s scan.Bytes
}

func (c *Compiler) Compile() AST {
	var n []AST
	c.stmts(&n)
	return AST{Type: "Root", Next: n}
}

func (c *Compiler) stmts(stmts *[]AST) bool {
	for {
		var stmt AST
		if !c.stmt(&stmt) {
			break
		}
		*stmts = append(*stmts, stmt)
	}
	return len(*stmts) > 0
}

func (c *Compiler) stmt(stmt *AST) bool {
	var i AST
	if c.s.Spaces(); c.ident(&i) && c.ost() && c.s.Match("=") && c.ost() {
		var expr AST
		if c.expr(&expr) {
			*stmt = AST{Type: "Stmt", Next: []AST{i, expr}}
			return true
		}
	}
	return false
}

func (c *Compiler) expr(expr *AST) bool {
	var or []AST
	for {
		var v AST
		if !c.term(&v) {
			break
		}
		or = append(or, v)
		if !(c.ost() && c.s.Match("|") && c.ost()) {
			break
		}
	}
	if len(or) == 1 {
		*expr = or[0]
		return true
	}
	if len(or) > 0 {
		*expr = AST{Type: "Or", Next: or}
		return true
	}
	return false
}

func (c *Compiler) term(expr *AST) bool {
	var and []AST
	for {
		var v AST
		if !c.factor(&v) {
			break
		}
		and = append(and, v)
	}
	if len(and) == 1 {
		*expr = and[0]
		return true
	}
	if len(and) > 0 {
		*expr = AST{Type: "And", Next: and}
		return true
	}
	return false
}

func (c *Compiler) factor(expr *AST) bool {
	c.st()
	defer c.quantifier(expr)
	if c.function(expr) {
		return true
	}
	if c.s.Match("(") && c.ost() && c.expr(expr) && c.ost() && c.s.Match(")") {
		return true
	}
	if c.value(expr) {
		return true
	}
	return false
}

func (c *Compiler) function(expr *AST) bool {
	if c.s.Match("EXPR1") {
		var v AST
		if c.s.Match("(") && c.term(&v) && c.s.Match(")") {
			if len(v.Next) != 3 {
				return false
			}
			*expr = AST{Type: "Func", Text: "EXPR1", Next: v.Next}
			return true
		}
	}
	if c.s.Match("GROUP") {
		var v AST
		if c.s.Match("(") && c.term(&v) && c.s.Match(")") {
			*expr = AST{Type: "Func", Text: "GROUP", Next: v.NextOrRoot()}
			return true
		}
	}
	return false
}

func (c *Compiler) quantifier(q *AST) bool {
	m := c.s.Mark()
	c.ost()
	if mm := c.s.Mark(); c.s.MatchChar("*+?") {
		*q = AST{Type: "Quant", Text: c.s.Text(mm), Next: []AST{*q}}
		return true
	}
	c.s.Move(m)
	return true
}

func (c *Compiler) value(v *AST) bool {
	if c.text(v) {
		return true
	}
	if c.ident(v) {
		return true
	}
	return false
}

func (c *Compiler) ident(v *AST) bool {
	var s string
	if c.s.TextWith(&s, func(*scan.Bytes) bool {
		return c.regex(`^\w+`)
	}) {
		*v = AST{Type: "Ident", Text: s}
		return true
	}
	return false
}

func (c *Compiler) text(v *AST) bool {
	m := c.s.Mark()
	ignore, regexp := c.s.Match("I"), c.s.Match("R")
	var s string
	if c.s.TextWith(&s, func(*scan.Bytes) bool {
		return c.regex(`^'([^'\\]|\\.)*'`)
	}) {
		s = s[1 : len(s)-1]
		a := AST{Type: "Plain", Text: s}
		if regexp {
			a.Type = "Regex"
		}
		if ignore {
			a = AST{Type: "Ignore", Next: []AST{a}}
		}
		*v = a
		return true
	}
	c.s.Move(m)
	return false
}

func (c *Compiler) regex(r string) bool {
	re := regexp.MustCompile(r)
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

func Print(a AST) {
	print(a, "")
}

func print(a AST, pad string) {
	if a.Text == "" {
		fmt.Printf("%s[%s]\n", pad, a.Type)
	} else {
		fmt.Printf("%s[%s] %s\n", pad, a.Type, a.Text)
	}
	for _, n := range a.Next {
		print(n, pad+"    ")
	}
}

type AST struct {
	Type string
	Text string
	Next []AST
}

func (a AST) NextOrRoot() []AST {
	if len(a.Next) == 0 {
		return []AST{a}
	}
	return a.Next
}

// Compact replaces the root node with the
// child node when there is only one child
// node.
func (a *AST) Compact() {
	if len(a.Next) == 1 {
		*a = a.Next[0]
	}
}

func (a AST) Empty() bool {
	return len(a.Next)+len(a.Type) == 0
}
