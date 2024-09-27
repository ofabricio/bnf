package bnf

import (
	"regexp"

	"github.com/ofabricio/scan"
)

func Parse(bnf AST, src string) AST {
	p := &Parser{s: scan.Bytes(src), bnf: bnf}
	var a AST
	p.Parse(bnf, &a)
	return a
}

type Parser struct {
	s   scan.Bytes
	bnf AST
}

func (p *Parser) Parse(bnf AST, out *AST) bool {
	m := p.s.Mark()
	switch bnf.Type {
	case "Root":
		for _, stmt := range bnf.Next {
			return p.Parse(stmt, out)
		}
	case "Stmt":
		return p.Parse(bnf.Next[1], out)
	case "Func":
		switch bnf.Text {
		case "EXPR1":
			var l, root, r AST
			if p.Parse(bnf.Next[0], &l) && p.Parse(bnf.Next[1], &root) && p.Parse(bnf.Next[2], &r) {
				*out = AST{Type: "Expr", Text: root.Text, Next: []AST{l, r}}
				return true
			}
		case "GROUP":
			var g []AST
			for _, n := range bnf.Next {
				var v AST
				if !p.Parse(n, &v) {
					return false
				}
				if !v.Empty() {
					g = append(g, v)
				}
			}
			*out = AST{Type: "Group", Next: g}
			return true
		}
	case "And":
		for _, n := range bnf.Next {
			var v AST
			if !p.Parse(n, &v) {
				return false
			}
			if !v.Empty() {
				out.Next = append(out.Next, v)
			}
		}
		if len(out.Next) == 1 {
			*out = out.Next[0]
		} else {
			out.Type = "Group"
		}
		return true
	case "Or":
		for _, n := range bnf.Next {
			if p.Parse(n, out) {
				return true
			}
		}
	case "Quant":
		c := 0
		for {
			var v AST
			if !p.Parse(bnf.Next[0], &v) {
				break
			}
			if !v.Empty() {
				out.Next = append(out.Next, v)
			}
			c++
		}
		out.Compact()
		if len(out.Next) > 0 {
			out.Type = "Group"
		}
		switch bnf.Text {
		case "?":
			return c <= 1
		case "*":
			return c >= 0
		case "+":
			return c > 0
		}
	case "Plain":
		if p.s.Match(bnf.Text) {
			out.Text = p.s.Text(m)
			out.Type = "Ident"
			return true
		}
	case "Regex":
		re := regexp.MustCompile("^" + bnf.Text)
		if v := re.FindIndex(p.s); v != nil {
			p.s.Advance(v[1])
			out.Text = p.s.Text(m)
			out.Type = "Ident"
			return true
		}
	case "Ident":
		if stmt, ok := p.findStmt(bnf.Text); ok {
			return p.Parse(stmt, out)
		}
	case "Ignore":
		var ignored AST
		return p.Parse(bnf.Next[0], &ignored)
	}
	p.s.Move(m)
	return false
}

func (p *Parser) findStmt(id string) (AST, bool) {
	for _, stmt := range p.bnf.Next {
		if stmt.Next[0].Text == id {
			return stmt, true
		}
	}
	return AST{}, false
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
