package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type Token int
type Number float64
type Atom string
type Item interface{}
type ItemList []Item

var ErrorEOF = errors.New("End of File")

func (items ItemList) String() string {
	s := "("
	for n, i := range items {
		if n > 0 {
			s += " "
		}
		s += fmt.Sprintf("%v", i)
	}
	s += ")"
	return s
}

func read(s *Scanner) (Item, error) {
	tok, lit := s.Scan()
	if tok == WS {
		tok, lit = s.Scan()
	}
	// fmt.Println("scan:", tok, lit)
	switch tok {
	case LEFT_PAREN:
		return readList(s)
	case ATOM:
		return Atom(lit), nil
	case NUMBER:
		v, err := strconv.ParseFloat(lit, 64)
		if err != nil {
			log.Println("Number fail:", err)
		}
		return Number(v), nil
	case RIGHT_PAREN:
		return nil, nil
	case EOF:
		return nil, ErrorEOF
	}
	return nil, errors.New("Malformed input")
}

func readList(s *Scanner) (Item, error) {
	var l ItemList
	for {
		c, err := read(s)
		if err != nil {
			return nil, fmt.Errorf("Failed to complete list: %v\n", err)
		}
		if c == nil {
			break
		}
		l = append(l, c)

	}
	return l, nil
}

type Binding map[Atom]Item
type Env struct {
	vars  Binding
	outer *Env
}

func NewEnv(outer *Env) *Env {
	return &Env{
		vars:  make(Binding),
		outer: outer,
	}

}

func (env *Env) Find(a Atom) *Env {
	//	fmt.Println(a)
	if _, ok := env.vars[a]; ok {
		return env
	} else if env.outer != nil {
		return env.outer.Find(a)
	} else {
		return env
	}
}

func eval(expr Item, env *Env) (interface{}, error) {
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
	case ItemList:
		switch car, _ := e[0].(Atom); car {
		case "quote":
			return e[1], nil
		case "define":
			v, err := eval(e[2], env)
			if err != nil {
				return nil, err
			}
			env.vars[e[1].(Atom)] = v
			return env.vars[e[1].(Atom)], nil
		case "set!":
			a := e[1].(Atom)
			env.Find(a).vars[a], _ = eval(e[2], env)
			return env.Find(a).vars[a], nil
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

		case "quit":
			os.Exit(0)
		case "lambda":
			return evalLambda(e, env)
		default:
			proc, err := eval(e[0], env)
			if err != nil {
				log.Println("Error1", err)
				return nil, err
			}
			args := make(ItemList, len(e)-1)
			for i, a := range e[1:] {
				args[i], err = eval(a, env)
				if err != nil {
					log.Println("Error2", err)
					return nil, err
				}
			}
			return apply(proc, args, env)
		}
	}
	return nil, fmt.Errorf("Unparsable expression: %v", expr)
}

type Lambda struct {
	params []Atom
	body   Item
	envt   *Env
}

func (l *Lambda) String() string {
	return fmt.Sprintf("<function %p: %v>", l, l.body)
}

func evalLambda(expr ItemList, env *Env) (interface{}, error) {
	l := Lambda{}
	params, ok := expr[1].(ItemList)
	if !ok {
		return nil, fmt.Errorf("bad params:", expr[1])
	}
	for _, x := range params {
		switch p := x.(type) {
		case Atom:
			l.params = append(l.params, p)
		case []interface{}:
			log.Fatal("combo param not supported:", x)
		}

	}
	//	log.Printf("lambda %#v\n", expr[2])
	l.body = expr[2]
	l.envt = NewEnv(env)
	return &l, nil

}
func apply(proc Item, args ItemList, env *Env) (Item, error) {
	//	log.Printf("apply: %v args: %v\n", proc, args)
	switch f := proc.(type) {
	case func(...Item) Item:
		return f(args...), nil
	case *Lambda:
		if len(f.params) != len(args) {
			log.Fatalf("parameter mismatch %v != %v", f.params, args)
		}
		for i, p := range f.params {
			f.envt.vars[p] = args[i]
		}
		return eval(f.body, f.envt)
	default:
		log.Fatalf("apply to a non function: %#v", proc)
	}
	return nil, nil
}

func repl(in string, env *Env) (interface{}, error) {
	s := NewScanner(strings.NewReader(in))
	var result interface{}
	for {
		var err error
		expr, err := read(s)
		if err != nil {
			return result, err
		}
		//fmt.Println(expr)

		result, err = eval(expr, env)
		if err != nil {
			return result, err
		}
	}
}

func replCLI(env *Env) {
	reader := (bufio.NewReader(os.Stdin))
	for {
		fmt.Print(">>> ")
		text, _ := reader.ReadString('\n')
		result, err := repl(text, env)
		if err != nil && err != ErrorEOF {
			fmt.Println("New Error:", err)
		}
		fmt.Println("===>", result)
	}
}

var defaultEnv *Env

func ApplyNumeric(f func(Number, Number) Number) func(a ...Item) Item {
	return func(a ...Item) Item {
		v, ok := a[0].(Number)
		if !ok {
			log.Fatalf("Not a number: %v", a[0])
		}
		for _, n := range a[1:] {
			i, ok := n.(Number)
			if !ok {
				log.Fatalf("Not a number: %v", a[0])
			}
			v = f(v, i)
		}
		return v
	}
}

func DefaultEnv() *Env {
	return &Env{
		Binding{
			"*": ApplyNumeric(func(x, y Number) Number {
				return x * y
			}),
			"/": ApplyNumeric(func(x, y Number) Number {
				return x / y
			}),
			"+": ApplyNumeric(func(x, y Number) Number {
				return x + y
			}),
			"-": ApplyNumeric(func(x, y Number) Number {
				return x - y
			}),
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
func init() {
	defaultEnv = DefaultEnv()
}

func main() {
	env := defaultEnv
	/*
		repl("(define r 10)\n(define n 12)", env)
		repl("(begin (define r 10)\n(define n 12))", env)
		repl("(define radius (* pi (* r r)))", env)
		repl("(if (<= 4 2) (* 10 2))", env)
		repl("(if (<= 4 2) (* 10 2) (+ 1 2))", env)
		repl("(/ radius 10)", env)
		repl("(quote (1  1))", env)
		repl("123", env)
		repl("(lambda () (+ 1 1))", env)
		repl("((lambda () (+ 1 1)))", env)
		repl("((lambda () (+ 1 1)))", env)
		repl("(define foo (begin (define count 0) (lambda () (set! count (+ count 1)))))", env)
		repl("(foo) (foo)", env)
		repl("(foo) (foo)", env)
		repl("(define plus (lambda (a b) (+ a b)))", env)
		repl("(define counter (lambda (n) (lambda () (set! n (+ n 1)))))", env)
	*/
	replCLI(env)
}
