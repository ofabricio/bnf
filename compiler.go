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

func (c *Compiler) stmt(n *AST) bool {
	var i AST
	if c.s.Spaces(); c.ident(&i) && c.ost() && c.s.Match("=") && c.ost() {
		var expr AST
		if c.expr(&expr) {
			*n = AST{Type: "Stmt", Next: []AST{i, expr}}
			return true
		}
	}
	return false
}

func (c *Compiler) expr(expr *AST) bool {
	var l, r AST
	if c.term(&l) {
		if c.ost() && c.s.Match("|") && c.ost() && c.expr(&r) {
			*expr = AST{Type: "Or", Next: []AST{l, r}}
			return true
		}
		*expr = l
		return true
	}
	return false
}

func (c *Compiler) term(expr *AST) bool {
	var l, r AST
	if c.qualifiedFactor(&l) {
		if c.st() && c.term(&r) {
			*expr = AST{Type: "And", Next: []AST{l, r}}
			return true
		}
		*expr = l
		return true
	}
	return false
}

func (c *Compiler) qualifiedFactor(expr *AST) bool {
	if c.factor(expr) {
		var q AST
		if c.quantifier(&q) {
			*expr = AST{Type: "Quant", Text: q.Text, Next: []AST{*expr}}
		}
		return true
	}
	return false
}

func (c *Compiler) factor(expr *AST) bool {
	if c.function(expr) {
		return true
	}
	if c.s.Match("(") && c.ost() && c.expr(expr) && c.ost() && c.s.Match(")") {
		return true
	}
	var i AST
	if c.value(&i) {
		*expr = i
		return true
	}
	return false
}

func (c *Compiler) function(expr *AST) bool {
	if c.s.Match("EXPR1") {
		var l, root, r AST
		if c.s.Match("(") && c.ost() && c.factor(&l) && c.ost() && c.factor(&root) && c.ost() && c.factor(&r) && c.ost() && c.s.Match(")") {
			*expr = AST{Type: "Func", Text: "EXPR1", Next: []AST{l, root, r}}
			return true
		}
	}
	if c.s.Match("GROUP") {
		if c.s.Match("(") {
			var vs []AST
			var v AST
			for c.ost(); c.factor(&v); c.ost() {
				vs = append(vs, v)
			}
			if c.s.Match(")") {
				*expr = AST{Type: "Func", Text: "GROUP", Next: vs}
				return true
			}
		}
	}
	return false
}

func (c *Compiler) quantifier(q *AST) bool {
	m := c.s.Mark()
	c.ost()
	if mm := c.s.Mark(); c.s.MatchChar("*+?") {
		q.Text = c.s.Text(mm)
		return true
	}
	c.s.Move(m)
	return false
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
