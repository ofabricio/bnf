package bnf

import (
	"regexp"
	"strings"

	"github.com/ofabricio/scan"
)

func Compile(bnf string) Root {
	return NewCompiler(bnf).Compile()
}

func NewCompiler(bnf string) *Compiler {
	return &Compiler{s: scan.Bytes(bnf)}
}

type Compiler struct {
	s scan.Bytes
}

func (c *Compiler) Compile() Root {
	var stmts []Stmt
	if c.stmts(&stmts) {
		return Root{Stmts: stmts, Stmtm: c.buildIdentMap(stmts)}
	}
	return Root{}
}

func (c *Compiler) buildIdentMap(stmts []Stmt) map[string]Stmt {
	m := make(map[string]Stmt, len(stmts))
	for _, stmt := range stmts {
		m[stmt.Ident.Text] = stmt
	}
	return m
}

func (c *Compiler) stmts(out *[]Stmt) bool {
	var stmt Stmt
	for c.stmt(&stmt) {
		*out = append(*out, stmt)
	}
	return len(*out) > 0
}

func (c *Compiler) stmt(out *Stmt) bool {
	c.ws()
	return c.ident(&out.Ident) && c.stmtMarker() && c.expr(&out.Expr)
}

func (c *Compiler) expr(out *BNF) bool {
	var or []BNF
	var v BNF
	for c.term(&v) {
		or = append(or, v)
		if !c.s.MatchChar("|") {
			break
		}
	}
	if len(or) > 0 {
		if len(or) == 1 {
			*out = or[0]
		} else {
			*out = ExprOr{Or: or}
		}
	}
	return len(or) > 0
}

func (c *Compiler) term(out *BNF) bool {
	var and []BNF
	var v BNF
	for c.factor(&v) {
		and = append(and, v)
	}
	if len(and) > 0 {
		if len(and) == 1 {
			*out = and[0]
		} else {
			*out = ExprAnd{And: and}
		}
	}
	return len(and) > 0
}

func (c *Compiler) factor(out *BNF) bool {
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

func (c *Compiler) function(out *BNF) bool {
	var arg BNF
	if c.s.Match("ROOT") && c.s.MatchChar("(") && c.expr(&arg) && c.s.MatchChar(")") {
		*out = Function{Name: "ROOT", Expr: arg}
		return true
	}
	if c.s.Match("GROUP") && c.s.MatchChar("(") && c.expr(&arg) && c.s.MatchChar(")") {
		*out = Function{Name: "GROUP", Expr: arg}
		return true
	}
	if c.s.Match("ANYNOT") && c.s.MatchChar("(") && c.expr(&arg) && c.s.MatchChar(")") {
		*out = Function{Name: "ANYNOT", Expr: arg}
		return true
	}
	if c.s.Match("JOIN") && c.s.MatchChar("(") && c.expr(&arg) && c.s.MatchChar(")") {
		*out = Function{Name: "JOIN", Expr: arg}
		return true
	}
	if c.s.Match("MATCH") && c.s.MatchChar("(") && c.expr(&arg) && c.s.MatchChar(")") {
		*out = Function{Name: "MATCH", Expr: arg}
		return true
	}
	if c.s.Match("TEXT") && c.s.MatchChar("(") && (c.expr(&arg) || true) && c.s.MatchChar(")") {
		*out = Function{Name: "TEXT", Expr: arg}
		return true
	}
	if c.s.Match("SAVE") && c.s.MatchChar("(") && c.expr(&arg) && c.s.MatchChar(")") {
		*out = Function{Name: "SAVE", Expr: arg}
		return true
	}
	if c.s.Match("LOAD") && c.s.MatchChar("(") && c.s.MatchChar(")") {
		*out = Function{Name: "LOAD"}
		return true
	}
	if c.s.Match("FIND") && c.s.MatchChar("(") && c.expr(&arg) && c.s.MatchChar(")") {
		*out = Function{Name: "FIND", Expr: arg}
		return true
	}
	if c.s.Match("REVERSE") && c.s.MatchChar("(") && c.expr(&arg) && c.s.MatchChar(")") {
		*out = Function{Name: "REVERSE", Expr: arg}
		return true
	}
	return false
}

func (c *Compiler) quantifier(out *BNF) bool {
	c.ws()
	if m := c.s.Mark(); c.s.MatchChar("*+?") {
		*out = Quantifier{Name: c.s.Text(m), Expr: *out}
		return true
	}
	return false
}

func (c *Compiler) typ(out *BNF) bool {
	if c.s.MatchChar(":") {
		var id Ident
		if c.ident(&id) {
			*out = Type{Ident: id, Expr: *out}
			return true
		}
	}
	return false
}

func (c *Compiler) value(out *BNF) bool {
	var i Ident
	if m := c.s.Mark(); c.ident(&i) {
		if c.stmtMarker() {
			c.s.Move(m)
			return false
		}
		*out = i
		return true
	}
	return c.text(out)
}

func (c *Compiler) ident(out *Ident) bool {
	if m := c.s.Mark(); c.matchRegex(reIden) {
		*out = Ident{Text: c.s.Text(m)}
		return true
	}
	return false
}

func (c *Compiler) text(out *BNF) bool {
	if c.plain(out) {
		c.regexFlag(out)
		c.ignoreFlag(out)
		return true
	}
	return false
}

func (c *Compiler) plain(out *BNF) bool {
	if m := c.s.Mark(); c.matchRegex(reText) {
		t := c.s.Text(m)
		t = t[1 : len(t)-1]
		t = strings.NewReplacer(`\'`, `'`, `\\`, `\`).Replace(t)
		if t == "" {
			return false
		}
		*out = Plain{Text: t}
		return true
	}
	return false
}

func (c *Compiler) regexFlag(out *BNF) bool {
	if c.s.MatchChar("r") {
		txt := "^" + (*out).(Plain).Text
		*out = Regex{Text: txt, Regx: regexp.MustCompile(txt)}
		return true
	}
	return false
}

func (c *Compiler) ignoreFlag(out *BNF) bool {
	if c.s.MatchChar("i") {
		*out = Ignore{Expr: *out}
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
