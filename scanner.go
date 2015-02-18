package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"unicode"
)

const (
	ILLEGAL Token = iota
	EOF
	WS
	ATOM
	NUMBER
	LEFT_PAREN
	RIGHT_PAREN
	QUOTE
	DQUOTE
)

const eof = rune(0)

func (t Token) String() string {
	switch t {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case WS:
		return "WS"
	case ATOM:
		return "ATOM"
	case NUMBER:
		return "NUMBER"
	case LEFT_PAREN:
		return "LEFT_PAREN"
	case RIGHT_PAREN:
		return "RIGHT_PAREN"
	case QUOTE:
		return "QUOTE"
	case DQUOTE:
		return "DQUOTE"
	}
	return "Unknown token: " + fmt.Sprintf("%d", t)
}

func isWhitespace(ch rune) bool {
	return unicode.IsSpace(ch)
}

func isLetter(ch rune) bool {
	if ch == '(' || ch == ')' || ch == '\'' || ch == '"' {
		return false
	}

	return unicode.IsLetter(ch) || unicode.IsPunct(ch) || unicode.IsSymbol(ch)
}

func isNumber(ch rune) bool {
	return unicode.IsNumber(ch)
}

type Scanner struct {
	r *bufio.Reader
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (s *Scanner) unread() {
	_ = s.r.UnreadRune()
}

func (s *Scanner) Scan() (tok Token, lit string) {
	ch := s.read()

	if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if isLetter(ch) {
		s.unread()
		return s.scanAtom()
	} else if isNumber(ch) {
		s.unread()
		return s.scanNumber()
	}

	switch ch {
	case eof:
		return EOF, ""
	case '(':
		return LEFT_PAREN, string(ch)
	case ')':
		return RIGHT_PAREN, string(ch)
	case '\'':
		return QUOTE, string(ch)
	case '"':
		return DQUOTE, string(ch)
	}

	fmt.Println("illegal: ", string(ch))
	return ILLEGAL, string(ch)
}

func (s *Scanner) scanWhitespace() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}
	return WS, buf.String()
}

func (s *Scanner) scanAtom() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isNumber(ch) && ch != '_' {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}
	return ATOM, buf.String()
}

func (s *Scanner) scanNumber() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isNumber(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}
	return NUMBER, buf.String()
}
