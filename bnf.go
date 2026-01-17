package bnf

import "regexp"

type AST struct {
	Type string
	Text string
	Pos  int
	Next []AST
}

type BNF any

type Root struct {
	Stmts []Stmt
	Stmtm map[string]Stmt
}

type Stmt struct {
	Ident Ident
	Expr  BNF
}

type Ident struct {
	Text string
}

type ExprAnd struct {
	And []BNF
}

type ExprOr struct {
	Or []BNF
}

type Quantifier struct {
	Name string
	Expr BNF
}

type Function struct {
	Name string
	Expr BNF
}

type Ignore struct {
	Expr BNF
}

type Plain struct {
	Text string
}

type Regex struct {
	Text string
	Regx *regexp.Regexp
}

type Type struct {
	Ident Ident
	Expr  BNF
}
