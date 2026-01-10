package bnf

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/ofabricio/scan"
)

func Parse(bnf AST, src string) AST {
	return NewParser(bnf, src).Parse()
}

func NewParser(bnf AST, src string) *Parser {
	return &Parser{src: src, cur: scan.Bytes(src), bnf: bnf}
}

type Parser struct {
	src string
	cur scan.Bytes
	bnf AST
	sav string
}

func (p *Parser) Parse() AST {
	var v []AST
	if !p.parse(p.bnf, &v) {
		return AST{Type: "Error", Pos: p.pos()}
	}
	if len(v) == 1 {
		return v[0]
	}
	return AST{Type: "Group", Next: v}
}

func (p *Parser) parse(bnf AST, out *[]AST) bool {
	switch bnf.Type {
	case "Root":
		for _, stmt := range bnf.Next {
			return p.parse(stmt, out)
		}
	case "Stmt":
		s := p.sav
		v := p.parse(bnf.Next[1], out)
		p.sav = s
		return v
	case "FIND":
		for ; p.cur.More(); p.cur.Next() {
			if p.parse(bnf.Next[0], out) {
				return true
			}
		}
	case "ROOT":
		return p.parse(bnf.Next[0], out)
	case "SAVE":
		var v []AST
		if p.parse(bnf.Next[0], &v) {
			p.sav = v[0].Text
			p.emit(out, v...)
			return true
		}
	case "LOAD":
		return p.parse(AST{Type: "Plain", Text: p.sav}, out)
	case "TEXT":
		p.emitIdent(out, bnf.Next[0].Text)
		return true
	case "JOIN":
		var v []AST
		if p.parse(bnf.Next[0], &v) {
			p.emitIdent(out, Join(v))
			return true
		}
	case "MATCH":
		var v []AST
		if m := p.cur.Mark(); p.parse(bnf.Next[0], &v) {
			p.emitIdent(out, p.cur.Text(m))
			return true
		}
	case "GROUP":
		var v []AST
		for _, n := range bnf.Next {
			if !p.parse(n, &v) {
				return false
			}
		}
		p.emit(out, AST{Type: "Group", Next: v})
		return true
	case "And":
		m := p.cur.Mark()
		var v, r []AST
		for _, n := range bnf.Next {
			vr := &v
			if n.Type == "ROOT" {
				vr = &r
			}
			if !p.parse(n, vr) {
				p.cur.Move(m)
				return false
			}
		}
		if len(r) > 0 {
			r[0].Next = v
			v = r[:1]
		}
		p.emit(out, v...)
		return true
	case "Or":
		for _, n := range bnf.Next {
			var v []AST
			if p.parse(n, &v) {
				p.emit(out, v...)
				return true
			}
		}
	case "?":
		return p.parse(bnf.Next[0], out) || true
	case "*":
		c := 0
		for p.parse(bnf.Next[0], out) {
			c++
		}
		return c >= 0
	case "+":
		c := 0
		for p.parse(bnf.Next[0], out) {
			c++
		}
		return c > 0
	case "Type":
		var v []AST
		ok := p.parse(bnf.Next[0], &v)
		if ok && len(v) > 0 {
			v[0].Type = bnf.Text
			p.emit(out, v[0])
		}
		return ok
	case "ANYNOT":
		var v []AST
		if m := p.cur.Mark(); p.parse(bnf.Next[0], &v) {
			p.cur.Move(m)
			return false
		} else if p.cur.Next() {
			p.emitIdent(out, p.cur.Text(m))
			return true
		}
	case "Ignore":
		var v []AST
		return p.parse(bnf.Next[0], &v)
	case "Ident":
		return p.parseIdent(bnf.Text, out) ||
			p.matchDefaultIdent(bnf.Text) ||
			p.matchDefaultIdentThatEmit(bnf.Text, out)
	case "Regex":
		if v := regexp.MustCompile(bnf.Text).FindIndex(p.cur); v != nil {
			m := p.cur.Mark()
			p.cur.Advance(v[1])
			p.emitIdent(out, p.cur.Text(m))
			return true
		}
	case "Plain":
		if m := p.cur.Mark(); p.cur.Match(bnf.Text) {
			p.emitIdent(out, p.cur.Text(m))
			return true
		}
	}
	return false
}

func (p *Parser) parseIdent(ident string, out *[]AST) bool {
	for _, stmt := range p.bnf.Next {
		if stmt.Next[0].Text == ident {
			return p.parse(stmt, out)
		}
	}
	return false
}

func (p *Parser) emitIdent(out *[]AST, text string) {
	*out = append(*out, AST{Type: "Ident", Text: text, Pos: p.pos() - len(text)})
}

func (p *Parser) emit(out *[]AST, n ...AST) {
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
			p.emitIdent(out, p.cur.Text(m))
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

func Join(tree []AST) string {
	var out strings.Builder
	for _, n := range tree {
		out.WriteString(n.Text)
		out.WriteString(Join(n.Next))
	}
	return out.String()
}
