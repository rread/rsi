package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Token int

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
	STRING
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
	case STRING:
		return "STRING"
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
	return unicode.IsNumber(ch) || ch == '.'
}

func isAtom(ch rune) bool {
	return isLetter(ch) || isNumber(ch) || ch == '_'
}

type TokenItem struct {
	token Token
	lit   string
}

func (tok *TokenItem) String() string {
	return fmt.Sprintf("%s '%s'", tok.token, tok.lit)
}

type Lexer struct {
	name  string
	input string
	start int
	pos   int
	width int
	items chan *TokenItem
}

func NewLexer(name, input string) *Lexer {
	l := &Lexer{
		name:  name,
		input: input,
		items: make(chan *TokenItem)}
	go l.run()
	return l
}

func (l *Lexer) NextItem() *TokenItem {
	return <-l.items
}

func (l *Lexer) peek() rune {
	ch := l.next()
	l.rewind()
	return ch
}

func (l *Lexer) next() (ch rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	ch, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return ch
}

func (l *Lexer) ignore() {
	l.start = l.pos
}

func (l *Lexer) emit(t Token) {
	l.items <- &TokenItem{token: t,
		lit: l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *Lexer) rewind() {
	l.pos -= l.width
	l.width = 0
}

func (l *Lexer) run() {
	for state := lexBase; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *Lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.rewind()
	return false
}

func (l *Lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.rewind()
}

func (l *Lexer) acceptRunFn(test func(rune) bool) {
	for test(l.next()) {
	}
	l.rewind()
}

type stateFn func(l *Lexer) stateFn

func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- &TokenItem{
		ILLEGAL,
		fmt.Sprintf(format, args...),
	}
	return nil
}

func lexBase(l *Lexer) stateFn {
	for {
		switch ch := l.next(); {
		case ch == eof:
			l.emit(EOF)
			return nil
		case isWhitespace(ch):
			l.ignore()
		case ch == '(':
			l.emit(LEFT_PAREN)
			return lexBase
		case ch == ')':
			l.emit(RIGHT_PAREN)
			return lexBase
		case ch == '\'':
			l.emit(QUOTE)
			return lexBase
		case ch == '-' || ch == '+':
			return atomOrNumber
		case ch == '"':
			l.ignore()
			return lexString
		case isNumber(ch):
			return lexNumber
		case isAtom(ch):
			return lexAtom
		}
	}
	return lexBase
}

func atomOrNumber(l *Lexer) stateFn {
	ch := l.peek()
	if ch == eof {
		l.emit(EOF)
		return nil
	}
	if isNumber(ch) {
		return lexNumber
	}
	return lexAtom
}

func lexNumber(l *Lexer) stateFn {
	l.acceptRun("0123456789.")
	l.emit(NUMBER)
	return lexBase
}

func lexAtom(l *Lexer) stateFn {
	l.acceptRunFn(isAtom)
	l.emit(ATOM)
	return lexBase
}

func lexString(l *Lexer) stateFn {
	for {
		switch ch := l.next(); {
		case ch == eof:
			return l.errorf("unterminated string")
		case ch == '\\':
			ch := l.next()
			if ch == eof {
				return l.errorf("unterminated string")

			}
		case ch == '"':
			l.rewind()
			l.emit(STRING)
			l.next()
			l.ignore()
			return lexBase
		}
	}
}
