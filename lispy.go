package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Item is an Atom, Number, String, or Function
type Item interface{}

// ItemList is a fundamental data type.
type ItemList []Item

type Number float64
type Atom string
type String string
type NativeFunc func(...Item) (Item, error)

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

func (s String) String() string {
	return fmt.Sprintf("\"%s\"", string(s))
}

func read(s *Scanner) (Item, error) {
	tok, lit := s.Scan()
	if tok == WS {
		tok, lit = s.Scan()
	}
	//log.Println("scan:", tok, lit)
	switch tok {
	case LEFT_PAREN:
		return readList(s)
	case ATOM:
		return Atom(lit), nil
	case QUOTE:
		return readQuote(s)
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
	case STRING:
		return String(lit), nil
	}
	return nil, errors.New("Malformed input")
}

func readQuote(s *Scanner) (Item, error) {
	l := ItemList{Atom("quote")}
	c, err := read(s)
	if err != nil {
		return nil, fmt.Errorf("Failed to complete list: %v\n", err)
	}
	return append(l, c), nil
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
	case String:
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
				return nil, fmt.Errorf("undefined-function: %v", err)
			}
			args := make(ItemList, len(e)-1)
			for i, a := range e[1:] {
				args[i], err = eval(a, env)
				if err != nil {
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
			return nil, fmt.Errorf("combo param not supported: %v", x)
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
	case NativeFunc:
		return f(args...)
	case *Lambda:
		if len(f.params) != len(args) {
			return nil, fmt.Errorf("parameter mismatch %v != %v", f.params, args)
		}
		for i, p := range f.params {
			f.envt.vars[p] = args[i]
		}
		return eval(f.body, f.envt)
	default:
		return nil, fmt.Errorf("apply to a non function: %#v", proc)
	}
	return nil, nil
}

func replReader(in io.Reader, env *Env) (interface{}, error) {
	s := NewScanner(in)
	var result interface{}
	for {
		var err error
		expr, err := read(s)
		if err != nil {
			if err == ErrorEOF {
				return result, nil
			}
			return result, err
		}
		// fmt.Println(expr)

		result, err = eval(expr, env)
		if err != nil {
			return result, err
		}
	}
}

func repl(in string, env *Env) (interface{}, error) {
	return replReader(strings.NewReader(in), env)
}

func replCLI(env *Env) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("* ")
		text, _ := reader.ReadString('\n')
		result, err := repl(text, env)
		if err != nil && err != ErrorEOF {
			fmt.Println("Error:", err)
		} else {
			fmt.Println(result)
		}
	}
}

func ApplyNumeric(f func(Number, Number) Number) NativeFunc {
	return func(a ...Item) (Item, error) {
		v, ok := a[0].(Number)
		if !ok {
			return nil, fmt.Errorf("Not a number: %v", a[0])
		}
		for _, n := range a[1:] {
			i, ok := n.(Number)
			if !ok {
				return nil, fmt.Errorf("Not a number: %v", n)
			}
			v = f(v, i)
		}
		return v, nil
	}
}

func ApplyNumericBool(f func(Number, Number) bool) NativeFunc {
	return func(a ...Item) (Item, error) {
		var ret bool
		v, ok := a[0].(Number)
		if !ok {
			return nil, fmt.Errorf("Not a number: %v", a[0])
		}
		for _, n := range a[1:] {
			i, ok := n.(Number)
			if !ok {
				return nil, fmt.Errorf("Not a number: %v", n)
			}
			ret = f(v, i)
			v = i
			if !ret {
				return ret, nil
			}
		}
		return Item(ret), nil
	}
}

func isItemList(i Item) bool {
	if _, ok := i.(ItemList); ok {
		return true
	}
	return false
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
			"<": ApplyNumericBool(func(x, y Number) bool {
				return x < y
			}),
			"<=": ApplyNumericBool(func(x, y Number) bool {
				return x <= y
			}),
			">": ApplyNumericBool(func(x, y Number) bool {
				return x > y
			}),
			">=": ApplyNumericBool(func(x, y Number) bool {
				return x >= y
			}),
			"=": ApplyNumericBool(func(x, y Number) bool {
				return x == y
			}),
			"number?": NativeFunc(func(a ...Item) (Item, error) {
				if len(a) > 1 {
					return nil, fmt.Errorf("Too many arguments for number?: %v", a)
				}
				if _, ok := a[0].(Number); ok {
					return Item(true), nil
				}
				return Item(false), nil
			}),
			"car": NativeFunc(func(a ...Item) (Item, error) {
				if !isItemList(a[0]) {
					return nil, fmt.Errorf("Not a list: %#v", a[0])
				}
				il := a[0].(ItemList)
				if len(il) == 0 {
					return []ItemList{}, nil
				}
				return a[0].(ItemList)[0], nil
			}),
			"cdr": NativeFunc(func(a ...Item) (Item, error) {
				if !isItemList(a[0]) {
					return nil, fmt.Errorf("Not a list: %#v", a[0])
				}
				il := a[0].(ItemList)
				if len(il) < 2 {
					return []ItemList{}, nil
				}
				return il[1:], nil
			}),
			"cons": NativeFunc(func(a ...Item) (Item, error) {
				if len(a) != 2 {
					return nil, fmt.Errorf("wrong number of arguments given to cons")
				}
				l := ItemList{a[0]}
				if il, ok := a[1].(ItemList); ok {
					l = append(l, il...)
				} else {
					l = append(l, a[1])
				}
				return l, nil
			}),
			"equal?": NativeFunc(func(a ...Item) (Item, error) {
				return reflect.DeepEqual(a[0], a[1]), nil
			}),
			"pi": Number(math.Pi),
		},
		nil,
	}
}

func main() {
	replCLI(DefaultEnv())
}
