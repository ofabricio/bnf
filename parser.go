package bnf

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

func Parse(bnf Root, src string) AST {
	return NewParser(bnf, src).Parse()
}

func NewParser(bnf Root, src string) *Parser {
	return &Parser{scn: Scanner{src: src}, bnf: bnf, fun: DefaultFuncs}
}

type Parser struct {
	scn Scanner
	bnf Root
	sav string
	fun map[string]Func
}

func (p *Parser) Parse() AST {
	var v []AST
	if !p.ParseNode(p.bnf, &v) {
		return AST{Type: "Error", Pos: p.scn.Pos()}
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
		m := p.scn.Mark()
		var v []AST
		for _, n := range bnf.And {
			if !p.ParseNode(n, &v) {
				p.scn.Move(m)
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
			p.Emit(out, v...)
		}
		return ok
	case Ignore:
		return p.ParseNode(bnf.Expr, nil)
	case Plain:
		if m := p.scn.Mark(); p.scn.Match(bnf.Text) {
			p.EmitIdent(out, p.scn.Text(m))
			return true
		}
	case Regex:
		if m := p.scn.Mark(); p.scn.Regex(bnf.Regx) {
			p.EmitIdent(out, p.scn.Text(m))
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
	if out != nil {
		*out = append(*out, AST{Type: "Ident", Text: text, Pos: p.scn.Pos() - len(text)})
	}
}

func (p *Parser) Emit(out *[]AST, n ...AST) {
	if out != nil {
		*out = append(*out, n...)
	}
}

// matchDefaultIdent match identifiers that do not emit a token.
func (p *Parser) matchDefaultIdent(ident string) bool {
	switch ident {
	case "EOF":
		return !p.scn.More()
	case "MORE":
		return p.scn.More()
	case "SKIPLINE":
		return p.scn.Regex(reLINE)
	case "ANY":
		return p.scn.Next()
	case "WS":
		return p.scn.MatchFunc(unicode.IsSpace)
	case "NL":
		return p.scn.Match("\n")
	case "SP":
		return p.scn.Match(" ")
	case "ST":
		return p.scn.Match(" ") || p.scn.Match("\t")
	case "TB":
		return p.scn.Match("\t")
	}
	return false
}

// matchDefaultIdentThatEmit match identifiers that emit a token.
func (p *Parser) matchDefaultIdentThatEmit(ident string, out *[]AST) bool {
	switch ident {
	case "any":
		if m := p.scn.Mark(); p.scn.Next() {
			p.EmitIdent(out, p.scn.Text(m))
			return true
		}
	}
	return false
}

var DefaultFuncs = map[string]Func{
	"ANYNOT":  funcANYNOT,
	"EXPR":    funcEXPR,
	"FIND":    funcFIND,
	"GROUP":   funcGROUP,
	"JOIN":    funcJOIN,
	"LOAD":    funcLOAD,
	"MATCH":   funcMATCH,
	"REVERSE": funcREVERSE,
	"ROOT":    funcROOT,
	"SAVE":    funcSAVE,
	"SUM":     funcSUM,
	"TEXT":    funcTEXT,
}

type Func func(p *Parser, arg BNF, out *[]AST) bool

func funcANYNOT(p *Parser, arg BNF, out *[]AST) bool {
	if m := p.scn.Mark(); p.ParseNode(arg, nil) {
		p.scn.Move(m)
		return false
	} else if p.scn.Next() {
		p.EmitIdent(out, p.scn.Text(m))
		return true
	}
	return false
}

func funcEXPR(p *Parser, arg BNF, out *[]AST) bool {
	var v []AST
	if p.ParseNode(arg, &v) {
		expr := funcExprEvaluator(v[0])
		p.EmitIdent(out, strconv.FormatFloat(expr, 'f', -1, 64))
		return true
	}
	return false
}

func funcExprEvaluator(n AST) float64 {
	switch n.Text {
	case "+":
		return funcExprEvaluator(n.Next[0]) + funcExprEvaluator(n.Next[1])
	case "-":
		return funcExprEvaluator(n.Next[0]) + funcExprEvaluator(n.Next[1])
	case "*":
		return funcExprEvaluator(n.Next[0]) * funcExprEvaluator(n.Next[1])
	case "/":
		return funcExprEvaluator(n.Next[0]) / funcExprEvaluator(n.Next[1])
	default:
		v, _ := strconv.ParseFloat(n.Text, 64)
		return v
	}
}

func funcFIND(p *Parser, arg BNF, out *[]AST) bool {
	for ; p.scn.More(); p.scn.Next() {
		if p.ParseNode(arg, out) {
			return true
		}
	}
	return false
}

func funcGROUP(p *Parser, arg BNF, out *[]AST) bool {
	var v []AST
	if !p.ParseNode(arg, &v) {
		return false
	}
	p.Emit(out, AST{Type: "Group", Next: v})
	return true
}

func funcJOIN(p *Parser, arg BNF, out *[]AST) bool {
	var v []AST
	if p.ParseNode(arg, &v) {
		var sb strings.Builder
		sb.Grow(64)
		Walks(v, func(n AST) {
			sb.WriteString(n.Text)
		})
		p.EmitIdent(out, sb.String())
		return true
	}
	return false
}

func funcLOAD(p *Parser, arg BNF, out *[]AST) bool {
	return p.ParseNode(Plain{Text: p.sav}, out)
}

func funcMATCH(p *Parser, arg BNF, out *[]AST) bool {
	if m := p.scn.Mark(); p.ParseNode(arg, nil) {
		p.EmitIdent(out, p.scn.Text(m))
		return true
	}
	return false
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

func funcSAVE(p *Parser, arg BNF, out *[]AST) bool {
	var v []AST
	if p.ParseNode(arg, &v) {
		p.sav = v[0].Text
		p.Emit(out, v...)
		return true
	}
	return false
}

func funcSUM(p *Parser, arg BNF, out *[]AST) bool {
	var v []AST
	if p.ParseNode(arg, &v) {
		sum := 0.0
		Walks(v, func(n AST) {
			v, _ := strconv.ParseFloat(n.Text, 64)
			sum += v
		})
		p.EmitIdent(out, strconv.FormatFloat(sum, 'f', -1, 64))
		return true
	}
	return false
}

func funcTEXT(p *Parser, arg BNF, out *[]AST) bool {
	if v, ok := arg.(Plain); ok {
		p.EmitIdent(out, v.Text)
	} else {
		p.EmitIdent(out, "")
	}
	return true
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

func Walk(tree AST, fn func(AST)) {
	fn(tree)
	Walks(tree.Next, fn)
}

func Walks(ns []AST, fn func(AST)) {
	for _, n := range ns {
		Walk(n, fn)
	}
}

var reLINE = regexp.MustCompile(".*\n?")
