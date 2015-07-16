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
	SYMBOL
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
	case SYMBOL:
		return "SYMBOL"
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
	return fmt.Sprintf("%s %#v", tok.token, tok.lit)
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

// peek returns the next rune but leaves the range unchanged.
func (l *Lexer) peek() rune {
	ch := l.next()
	l.rewind()
	return ch
}

// next returns the next rune and extends current input range.
func (l *Lexer) next() (ch rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	ch, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return ch
}

// skip removes the current token from the input in a rather ineffecient way.
func (l *Lexer) skip() {
	w := l.width
	l.rewind()
	l.input = l.input[:l.pos] + l.input[l.pos+w:]
}

// ignore skips current input range up to current rune.
func (l *Lexer) ignore() {
	l.start = l.pos
}

// emit sends the current range as a t token and resets
// the range.
func (l *Lexer) emit(t Token) {
	l.items <- &TokenItem{token: t,
		lit: l.input[l.start:l.pos]}
	l.start = l.pos
}

// rewind moves end of range to previous rune.
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
	l.acceptRunFn(func(r rune) bool {
		return strings.IndexRune(valid, r) >= 0
	})
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
}

func atomOrNumber(l *Lexer) stateFn {
	ch := l.peek()
	if ch == eof {
		l.emit(SYMBOL)
		return lexBase
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
	l.emit(SYMBOL)
	return lexBase
}

func lexString(l *Lexer) stateFn {
	for {
		switch ch := l.next(); {
		case ch == eof:
			return l.errorf("unterminated string: '%v'", l.input[l.start:l.pos])
		case ch == '\\':
			l.skip()
			ch := l.next()
			if ch == eof {
				return l.errorf("unterminated string: %#v", l.input[l.start:l.pos])
			}
		case ch == '"':
			l.skip()
			l.emit(STRING)
			return lexBase
		}
	}
}
