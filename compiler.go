package bnf

import (
	"regexp"
	"strings"

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
	if c.ws(); c.ident(&i) && c.stmtMarker() && c.expr(out) {
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
		compact(out)
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
		compact(out)
		return true
	}
	return false
}

func (c *Compiler) factor(out *AST) bool {
	c.ws()
	if c.s.MatchChar("(") && c.expr(out) && c.s.MatchChar(")") || c.function(out) {
		c.typ(out)
		c.quantifier(out)
		c.ignoreFlag(out)
		return true
	}
	if c.value(out) {
		c.typ(out)
		_ = c.quantifier(out) && c.ignoreFlag(out)
		return true
	}
	return false
}

func (c *Compiler) function(out *AST) bool {
	var arg AST
	if c.s.Match("ROOT") && c.s.MatchChar("(") && c.expr(&arg) && c.s.MatchChar(")") {
		*out = AST{Type: "ROOT", Next: []AST{arg}}
		return true
	}
	if c.s.Match("GROUP") && c.s.MatchChar("(") && c.expr(&arg) && c.s.MatchChar(")") {
		*out = AST{Type: "GROUP", Next: []AST{arg}}
		return true
	}
	if c.s.Match("ANYNOT") && c.s.MatchChar("(") && c.expr(&arg) && c.s.MatchChar(")") {
		*out = AST{Type: "ANYNOT", Next: []AST{arg}}
		return true
	}
	if c.s.Match("JOIN") && c.s.MatchChar("(") && c.expr(&arg) && c.s.MatchChar(")") {
		*out = AST{Type: "JOIN", Next: []AST{arg}}
		return true
	}
	if c.s.Match("MATCH") && c.s.MatchChar("(") && c.expr(&arg) && c.s.MatchChar(")") {
		*out = AST{Type: "MATCH", Next: []AST{arg}}
		return true
	}
	if c.s.Match("TEXT") && c.s.MatchChar("(") && (c.expr(&arg) || true) && c.s.MatchChar(")") {
		*out = AST{Type: "TEXT", Next: []AST{arg}}
		return true
	}
	if c.s.Match("SAVE") && c.s.MatchChar("(") && c.expr(&arg) && c.s.MatchChar(")") {
		*out = AST{Type: "SAVE", Next: []AST{arg}}
		return true
	}
	if c.s.Match("LOAD") && c.s.MatchChar("(") && c.s.MatchChar(")") {
		*out = AST{Type: "LOAD"}
		return true
	}
	if c.s.Match("SCAN") && c.s.MatchChar("(") && c.expr(&arg) && c.s.MatchChar(")") {
		*out = AST{Type: "SCAN", Next: []AST{arg}}
		return true
	}
	return false
}

func (c *Compiler) quantifier(out *AST) bool {
	c.ws()
	if m := c.s.Mark(); c.s.MatchChar("*+?") {
		*out = AST{Type: c.s.Text(m), Next: []AST{*out}}
		return true
	}
	return false
}

func (c *Compiler) typ(out *AST) bool {
	if c.s.MatchChar(":") {
		var id AST
		if c.ident(&id) {
			*out = AST{Type: "Type", Text: id.Text, Next: []AST{*out}}
			swapRoot(out)
			return true
		}
	}
	return false
}

func (c *Compiler) value(out *AST) bool {
	if m := c.s.Mark(); c.ident(out) {
		if c.stmtMarker() {
			c.s.Move(m)
			return false
		}
		return true
	}
	return c.text(out)
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
		t = strings.NewReplacer(`\'`, `'`, `\\`, `\`).Replace(t)
		if t == "" {
			return false
		}
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

func (c *Compiler) stmtMarker() bool {
	return c.ws() && c.s.MatchChar("=")
}

func (c *Compiler) ws() bool {
	return c.s.WhileChar(" \t\n\r") || true
}

var reText = regexp.MustCompile(`^'([^'\\]|\\.)*'`)
var reIden = regexp.MustCompile(`^\w+`)
