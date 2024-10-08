package bnf

import (
	"regexp"
	"unicode"

	"github.com/ofabricio/scan"
)

func Parse(bnf AST, src string) AST {
	var out []AST
	if !NewParser(bnf, src).Parse(&out) {
		return AST{Type: "Group", Next: nil}
	}
	if len(out) == 1 {
		return out[0]
	}
	return AST{Type: "Group", Next: out}
}

func NewParser(bnf AST, src string) *Parser {
	return &Parser{s: scan.Bytes(src), bnf: bnf}
}

type Parser struct {
	s   scan.Bytes
	bnf AST // Root BNF.
	sav string
}

func (p *Parser) Parse(out *[]AST) bool {
	return p.parse(p.bnf, out)
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
	case "ROOT":
		return p.parse(bnf.Next[0], out)
	case "SAVE":
		var v []AST
		if p.parse(bnf.Next[0], &v) {
			p.sav = v[0].Text
			*out = append(*out, v...)
			return true
		}
	case "LOAD":
		return p.parse(AST{Type: "Plain", Text: p.sav}, out)
	case "TEXT":
		*out = append(*out, AST{Type: "Ident", Text: bnf.Next[0].Text})
		return true
	case "JOIN":
		var v []AST
		if p.parse(bnf.Next[0], &v) {
			*out = append(*out, AST{Type: "Ident", Text: join(v)})
			return true
		}
	case "GROUP":
		m := p.s.Mark()
		var v []AST
		for _, n := range bnf.Next {
			if !p.parse(n, &v) {
				p.s.Move(m)
				return false
			}
		}
		*out = append(*out, AST{Type: "Group", Next: v})
		return true
	case "And":
		m := p.s.Mark()
		var v, r []AST
		for _, n := range bnf.Next {
			vr := &v
			if n.Type == "ROOT" {
				vr = &r
			}
			if !p.parse(n, vr) {
				p.s.Move(m)
				return false
			}
		}
		if len(r) > 0 {
			*out = append(*out, AST{Type: r[0].Type, Text: r[0].Text, Next: v})
		} else {
			*out = append(*out, v...)
		}
		return true
	case "Or":
		for _, n := range bnf.Next {
			var v []AST
			if p.parse(n, &v) {
				*out = append(*out, v...)
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
		if p.parse(bnf.Next[0], &v) {
			v[0].Type = bnf.Text
			*out = append(*out, v[0])
			return true
		}
	case "Ident":
		return p.parseIdent(bnf, out) || p.matchDefaultIdent(bnf) || p.matchDefaultIdentThatEmit(bnf, out)
	case "Ignore":
		var v []AST
		if p.parse(bnf.Next[0], &v) {
			return true
		}
	case "Regex":
		if v := regexp.MustCompile(bnf.Text).FindIndex(p.s); v != nil {
			m := p.s.Mark()
			p.s.Advance(v[1])
			*out = append(*out, AST{Type: "Ident", Text: p.s.Text(m)})
			return true
		}
	case "NOT":
		var v []AST
		if m := p.s.Mark(); p.parse(bnf.Next[0], &v) {
			p.s.Move(m)
			return false
		} else if p.s.Next() {
			*out = append(*out, AST{Type: "Ident", Text: p.s.Text(m)})
			return true
		}
	case "Plain":
		if m := p.s.Mark(); p.s.Match(bnf.Text) {
			*out = append(*out, AST{Type: "Ident", Text: p.s.Text(m)})
			return true
		}
	}
	return false
}

func (p *Parser) parseIdent(bnf AST, out *[]AST) bool {
	for _, stmt := range p.bnf.Next {
		if stmt.Next[0].Text == bnf.Text {
			return p.parse(stmt, out)
		}
	}
	return false
}

// matchDefaultIdent match identifiers that do not emit a token.
func (p *Parser) matchDefaultIdent(bnf AST) bool {
	switch bnf.Text {
	case "EOF":
		return p.s.Empty()
	case "MORE":
		return p.s.More()
	case "any":
		return p.s.Next()
	case "ws":
		return p.s.MatchFunc(unicode.IsSpace) // ' \t\r\n'
	case "nl":
		return p.s.MatchChar("\n")
	case "sp":
		return p.s.MatchChar(" ")
	case "st":
		return p.s.MatchChar(" \t")
	case "tb":
		return p.s.MatchChar("\t")
	}
	return false
}

// matchDefaultIdentThatEmit match identifiers that emit a token.
func (p *Parser) matchDefaultIdentThatEmit(bnf AST, out *[]AST) bool {
	switch bnf.Text {
	case "ANY":
		if m := p.s.Mark(); p.s.Next() {
			*out = append(*out, AST{Type: "Ident", Text: p.s.Text(m)})
			return true
		}
	}
	return false
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

func join(tree []AST) string {
	var out string
	for _, n := range tree {
		out += n.Text + join(n.Next)
	}
	return out
}
