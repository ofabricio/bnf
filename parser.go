package bnf

import (
	"slices"
	"strings"
	"unicode"

	"github.com/ofabricio/scan"
)

func Parse(bnf Root, src string) AST {
	return NewParser(bnf, src).Parse()
}

func NewParser(bnf Root, src string) *Parser {
	return &Parser{src: src, cur: scan.Bytes(src), bnf: bnf, fun: DefaultFuncs}
}

type Parser struct {
	src string
	cur scan.Bytes
	bnf Root
	sav string
	fun map[string]Func
}

func (p *Parser) Parse() AST {
	var v []AST
	if !p.ParseNode(p.bnf, &v) {
		return AST{Type: "Error", Pos: p.pos()}
	}
	if len(v) == 1 {
		return v[0]
	}
	return AST{Type: "Group", Next: v}
}

func (p *Parser) ParseNode(bnf BNF, out *[]AST) bool {
	switch bnf := bnf.(type) {
	case Root:
		for _, stmt := range bnf.Stmts {
			return p.ParseNode(stmt, out)
		}
	case Stmt:
		s := p.sav
		v := p.ParseNode(bnf.Expr, out)
		p.sav = s
		return v
	case ExprAnd:
		m := p.cur.Mark()
		var v []AST
		for _, n := range bnf.And {
			if !p.ParseNode(n, &v) {
				p.cur.Move(m)
				return false
			}
		}
		p.Emit(out, v...)
		return true
	case ExprOr:
		for _, n := range bnf.Or {
			var v []AST
			if p.ParseNode(n, &v) {
				p.Emit(out, v...)
				return true
			}
		}
	case Quantifier:
		switch bnf.Name {
		case "?":
			return p.ParseNode(bnf.Expr, out) || true
		case "*":
			c := 0
			for p.ParseNode(bnf.Expr, out) {
				c++
			}
			return c >= 0
		case "+":
			c := 0
			for p.ParseNode(bnf.Expr, out) {
				c++
			}
			return c > 0
		}
	case Function:
		if fn, ok := p.fun[bnf.Name]; ok {
			return fn(p, bnf.Expr, out)
		}
	case Type:
		var v []AST
		ok := p.ParseNode(bnf.Expr, &v)
		if ok && len(v) > 0 {
			v[0].Type = bnf.Ident.Text
			p.Emit(out, v[0])
		}
		return ok
	case Ignore:
		var v []AST
		return p.ParseNode(bnf.Expr, &v)
	case Plain:
		if m := p.cur.Mark(); p.cur.Match(bnf.Text) {
			p.EmitIdent(out, p.cur.Text(m))
			return true
		}
	case Regex:
		if v := bnf.Regx.FindIndex(p.cur); v != nil {
			m := p.cur.Mark()
			p.cur.Advance(v[1])
			p.EmitIdent(out, p.cur.Text(m))
			return true
		}
	case Ident:
		return p.parseIdent(bnf.Text, out) ||
			p.matchDefaultIdent(bnf.Text) ||
			p.matchDefaultIdentThatEmit(bnf.Text, out)
	}
	return false
}

func (p *Parser) parseIdent(ident string, out *[]AST) bool {
	if stmt, ok := p.bnf.Stmtm[ident]; ok {
		return p.ParseNode(stmt, out)
	}
	return false
}

func (p *Parser) EmitIdent(out *[]AST, text string) {
	*out = append(*out, AST{Type: "Ident", Text: text, Pos: p.pos() - len(text)})
}

func (p *Parser) Emit(out *[]AST, n ...AST) {
	*out = append(*out, n...)
}

// matchDefaultIdent match identifiers that do not emit a token.
func (p *Parser) matchDefaultIdent(ident string) bool {
	switch ident {
	case "EOF":
		return p.cur.Empty()
	case "MORE":
		return p.cur.More()
	case "SKIPLINE":
		p.cur.UntilChar("\n")
		return p.cur.MatchChar("\n") || p.cur.Empty()
	case "ANY":
		return p.cur.Next()
	case "WS":
		return p.cur.MatchFunc(unicode.IsSpace) // ' \t\r\n'
	case "NL":
		return p.cur.MatchChar("\n")
	case "SP":
		return p.cur.MatchChar(" ")
	case "ST":
		return p.cur.MatchChar(" \t")
	case "TB":
		return p.cur.MatchChar("\t")
	}
	return false
}

// matchDefaultIdentThatEmit match identifiers that emit a token.
func (p *Parser) matchDefaultIdentThatEmit(ident string, out *[]AST) bool {
	switch ident {
	case "any":
		if m := p.cur.Mark(); p.cur.Next() {
			p.EmitIdent(out, p.cur.Text(m))
			return true
		}
	}
	return false
}

func (p *Parser) pos() int {
	return len(p.src) - len(p.cur)
}

// Flatten returns a flat list of nodes. The
// depth parameter specifies the maximum depth
// to flatten; zero flattens the entire tree.
func Flatten(tree AST, depth int) []AST {
	if depth <= 0 {
		depth = -2
	}
	return flatten(tree, depth+1)
}

func flatten(tree AST, depth int) []AST {
	if len(tree.Next) == 0 || depth == 0 {
		return []AST{tree}
	}
	var f []AST
	for _, n := range tree.Next {
		f = append(f, flatten(n, depth-1)...)
	}
	return f
}

var DefaultFuncs = map[string]Func{
	"FIND":    funcFIND,
	"SAVE":    funcSAVE,
	"LOAD":    funcLOAD,
	"TEXT":    funcTEXT,
	"REVERSE": funcREVERSE,
	"JOIN":    funcJOIN,
	"MATCH":   funcMATCH,
	"ROOT":    funcROOT,
	"GROUP":   funcGROUP,
	"ANYNOT":  funcANYNOT,
}

type Func func(p *Parser, arg BNF, out *[]AST) bool

func funcFIND(p *Parser, arg BNF, out *[]AST) bool {
	for ; p.cur.More(); p.cur.Next() {
		if p.ParseNode(arg, out) {
			return true
		}
	}
	return false
}

func funcSAVE(p *Parser, arg BNF, out *[]AST) bool {
	var v []AST
	if p.ParseNode(arg, &v) {
		p.sav = v[0].Text
		p.Emit(out, v...)
		return true
	}
	return false
}

func funcLOAD(p *Parser, arg BNF, out *[]AST) bool {
	return p.ParseNode(Plain{Text: p.sav}, out)
}

func funcTEXT(p *Parser, arg BNF, out *[]AST) bool {
	if v, ok := arg.(Plain); ok {
		p.EmitIdent(out, v.Text)
	} else {
		p.EmitIdent(out, "")
	}
	return true
}

func funcREVERSE(p *Parser, arg BNF, out *[]AST) bool {
	var v []AST
	if !p.ParseNode(arg, &v) {
		return false
	}
	slices.Reverse(v)
	p.Emit(out, v...)
	return true
}

func funcJOIN(p *Parser, arg BNF, out *[]AST) bool {
	var v []AST
	if p.ParseNode(arg, &v) {
		p.EmitIdent(out, Join(v))
		return true
	}
	return false
}

func funcMATCH(p *Parser, arg BNF, out *[]AST) bool {
	var v []AST
	if m := p.cur.Mark(); p.ParseNode(arg, &v) {
		p.EmitIdent(out, p.cur.Text(m))
		return true
	}
	return false
}

func funcROOT(p *Parser, arg BNF, out *[]AST) bool {
	var v []AST
	if !p.ParseNode(arg, &v) {
		return false
	}
	if len(v) >= 2 {
		r := v[1]
		r.Next = append(r.Next, v[0])
		r.Next = append(r.Next, v[2:]...)
		p.Emit(out, r)
		return true
	}
	p.Emit(out, v...)
	return true
}

func funcGROUP(p *Parser, arg BNF, out *[]AST) bool {
	var v []AST
	if !p.ParseNode(arg, &v) {
		return false
	}
	p.Emit(out, AST{Type: "Group", Next: v})
	return true
}

func funcANYNOT(p *Parser, arg BNF, out *[]AST) bool {
	var v []AST
	if m := p.cur.Mark(); p.ParseNode(arg, &v) {
		p.cur.Move(m)
		return false
	} else if p.cur.Next() {
		p.EmitIdent(out, p.cur.Text(m))
		return true
	}
	return false
}

func Join(tree []AST) string {
	var out strings.Builder
	for _, n := range tree {
		out.WriteString(n.Text)
		out.WriteString(Join(n.Next))
	}
	return out.String()
}
