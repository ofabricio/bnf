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
		return p.parse(bnf.Next[1], out)
	case "ROOT":
		return p.parse(bnf.Next[0], out)
	case "TEXT":
		*out = append(*out, AST{Type: "Ident", Text: bnf.Next[0].Text})
		return true
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
		return p.parseIdent(bnf, out) || p.matchDefaultIdent(bnf)
	case "NOT", "MATCH", "Plain", "Regex":
		if m := p.s.Mark(); p.match(bnf) {
			*out = append(*out, AST{Type: "Ident", Text: p.s.Text(m)})
			return true
		}
	case "Ignore":
		return p.match(bnf.Next[0])
	}
	return false
}

func (p *Parser) match(bnf AST) bool {
	switch bnf.Type {
	case "Root":
		for _, stmt := range bnf.Next {
			return p.match(stmt)
		}
	case "Stmt":
		return p.match(bnf.Next[1])
	case "ROOT":
		return p.match(bnf.Next[0])
	case "NOT":
		if m := p.s.Mark(); p.match(bnf.Next[0]) {
			p.s.Move(m)
			return false
		}
		return p.s.Next()
	case "TEXT":
		return true
	case "MATCH":
		return p.match(bnf.Next[0])
	case "And", "GROUP":
		m := p.s.Mark()
		for _, n := range bnf.Next {
			if !p.match(n) {
				p.s.Move(m)
				return false
			}
		}
		return true
	case "Or":
		for _, n := range bnf.Next {
			if p.match(n) {
				return true
			}
		}
	case "?":
		return p.match(bnf.Next[0]) || true
	case "*":
		c := 0
		for p.match(bnf.Next[0]) {
			c++
		}
		return c >= 0
	case "+":
		c := 0
		for p.match(bnf.Next[0]) {
			c++
		}
		return c > 0
	case "Plain":
		return p.s.Match(bnf.Text)
	case "Regex":
		if v := regexp.MustCompile(bnf.Text).FindIndex(p.s); v != nil {
			return p.s.Advance(v[1])
		}
	case "Type":
		return p.match(bnf.Next[0])
	case "Ident":
		return p.matchIdent(bnf) || p.matchDefaultIdent(bnf)
	case "Ignore":
		return p.match(bnf.Next[0])
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

func (p *Parser) matchIdent(bnf AST) bool {
	for _, stmt := range p.bnf.Next {
		if stmt.Next[0].Text == bnf.Text {
			return p.match(stmt)
		}
	}
	return false
}

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
