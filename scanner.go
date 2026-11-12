package bnf

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

func (s *Scanner) Regex(re *regexp.Regexp) bool {
	return s.Advance(re.FindString(s.Tail()))
}

func (s *Scanner) MatchFunc(f func(rune) bool) bool {
	r, sz := s.Rune()
	return f(r) && s.AdvanceRune(r, sz)
}

func (s *Scanner) Match(v string) bool {
	return s.Equal(v) && s.Advance(v)
}

func (s *Scanner) Equal(v string) bool {
	return strings.HasPrefix(s.Tail(), v)
}

func (s *Scanner) Tail() string {
	return s.src[s.pos:]
}

func (s *Scanner) Mark() int {
	return s.pos
}

func (s *Scanner) Text(m int) string {
	return s.src[m:s.pos]
}

func (s *Scanner) Move(m int) {
	s.pos = m
}

func (s *Scanner) Rune() (rune, int) {
	return utf8.DecodeRuneInString(s.Tail())
}

func (s *Scanner) Next() bool {
	return s.AdvanceRune(s.Rune())
}

func (s *Scanner) Advance(v string) bool {
	s.pos += len(v)
	return len(v) > 0
}

func (s *Scanner) AdvanceRune(r rune, size int) bool {
	s.pos += size
	return size > 0
}

func (s *Scanner) More() bool {
	return s.pos < len(s.src)
}

func (s *Scanner) Pos() int {
	return s.pos
}

type Scanner struct {
	src string
	pos int
}
