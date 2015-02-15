package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

type Token int
type Number float64
type Atom string
type Item interface{}

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

func scan(s *Scanner) (Item, error) {
	tok, lit := s.Scan()
	if tok == WS {
		tok, lit = s.Scan()
	}
	//	fmt.Println("scan:", tok, lit)
	switch tok {
	case LEFT_PAREN:
		return scanList(s), nil
	case ATOM:
		return Atom(lit), nil
	case NUMBER:
		v, err := strconv.ParseFloat(lit, 64)
		if err != nil {
			log.Println("Number fail:", err)
		}
		return Number(v), nil
	case EOF:
		return nil, errors.New("End of File")
	}
	return nil, nil
}

func scanList(s *Scanner) Item {
	var l []Item
	for {
		c, _ := scan(s)
		if c == nil {
			break
		}
		l = append(l, c)

	}
	return l
}

type foo map[Atom]Item
type Env struct {
	vars  foo
	outer *Env
}

func (env *Env) Find(a Atom) *Env {
	if _, ok := env.vars[a]; ok {
		return env
	} else {
		return env.outer.Find(a)
	}
}

func eval(expr interface{}, env Env) (interface{}, error) {
	//	fmt.Println(expr)
	switch e := expr.(type) {
	case Atom:
		v, ok := env.Find(e).vars[e]
		if ok {
			return v, nil
		} else {
			return nil, fmt.Errorf("Undefined symbol: %v", e)
		}
	case Number:
		return e, nil
	case []Item:
		switch car, _ := e[0].(Atom); car {
		case "quote":
			return e[1], nil
		case "define":
			env.vars[e[1].(Atom)], _ = eval(e[2], env)
			return env.vars[e[1].(Atom)], nil
		case "set!":
			a := e[1].(Atom)
			env.Find(a).vars[a], _ = eval(e[2], env)
			return env.vars[e[1].(Atom)], nil
		case "if":
			test, _ := eval(e[1], env)
			if test.(bool) {
				return eval(e[2], env)
			} else if len(e) > 3 {
				return eval(e[3], env)
			}
			return Atom("nil"), nil
		case "begin":
			var v Item
			for _, e := range e[1:] {
				v, _ = eval(e, env)
			}
			return v, nil

		default:
			proc, err := eval(e[0], env)
			if err != nil {
				log.Println("Error1", err)
			}
			args := make([]Item, len(e)-1)
			for i, a := range e[1:] {
				args[i], err = eval(a, env)
				if err != nil {
					log.Println("Error2", err)
				}
			}
			return apply(proc, args, env), nil
		}
	}
	return nil, fmt.Errorf("Unparsable expression: %v", expr)
}

func apply(proc Item, args []Item, env Env) Item {
	fn := proc.(func(...Item) Item)
	return fn(args...)
}

func repl(in string, env Env) {
	s := NewScanner(strings.NewReader(in))
	for {
		expr, err := scan(s)
		if err != nil {
			break
		}
		fmt.Println(expr)
		result, err := eval(expr, env)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println(">>>", result)
		}
	}
}

var defaultEnv Env

func init() {
	defaultEnv = Env{
		foo{"*": func(a ...Item) Item {
			v := a[0].(Number)
			for _, n := range a[1:] {
				v *= n.(Number)
			}
			return v
		},
			"/": func(a ...Item) Item {
				v := a[0].(Number)
				for _, n := range a[1:] {
					v /= n.(Number)
				}
				return v
			},
			"+": func(a ...Item) Item {
				v := a[0].(Number)
				for _, n := range a[1:] {
					v += n.(Number)
				}
				return v
			},
			"-": func(a ...Item) Item {
				v := a[0].(Number)
				for _, n := range a[1:] {
					v -= n.(Number)
				}
				return v
			},
			"<=": func(a ...Item) Item {
				return a[0].(Number) <= a[1].(Number)
			},
			"equal?": func(a ...Item) Item {
				return reflect.DeepEqual(a[0], a[1])
			},
			"pi": Number(math.Pi),
		},
		nil,
	}
}

func main() {
	env := defaultEnv
	repl("(define r 10)\n(define n 12)", env)
	repl("(begin (define r 10)\n(define n 12))", env)
	repl("(define radius (* pi (* r r)))", env)
	repl("(if (<= 4 2) (* 10 2))", env)
	repl("(if (<= 4 2) (* 10 2) (+ 1 2))", env)
	repl("(/ radius 10)", env)
	repl("(quote (1  1))", env)
	repl("123", env)
}
