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

// Item is an *Data, Number, String, or Function
type Item interface{}

// ItemList is a fundamental data type.
type ItemList []Item

type Number float64
type String string
type InternalFunc func(...Item) (Item, error)

type Tokenizer interface {
	NextItem() *TokenItem
}

var (
	T   *Data
	Nil *Data
)

func init() {
	Nil = internSymbol("NIL")
	T = internSymbol("t")
}

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
	return fmt.Sprintf(`"%s"`, string(s))
}

func read(l Tokenizer) (Item, error) {
	t := l.NextItem()
	if t.token == WS {
		t = l.NextItem()
	}
	log.Printf("scan: %v\n", t)
	switch t.token {
	case LEFT_PAREN:
		return readList(l)
	case RIGHT_PAREN:
		return nil, nil
	case SYMBOL:
		return internSymbol(t.lit), nil
	case QUOTE:
		return readQuote(l)
	case NUMBER:
		v, err := strconv.ParseFloat(t.lit, 64)
		if err != nil {
			log.Fatal("Number fail:", err)
		}
		return Number(v), nil
	case EOF:
		return nil, ErrorEOF
	case STRING:
		return String(t.lit), nil
	case ILLEGAL:
		return nil, errors.New(t.lit)
	}
	return nil, errors.New("Malformed input")
}

func readQuote(lex Tokenizer) (Item, error) {
	l := ItemList{internSymbol("quote")}
	c, err := read(lex)
	if err != nil {
		return nil, fmt.Errorf("Failed to complete list: %v\n", err)
	}
	return append(l, c), nil
}

func readList(lex Tokenizer) (Item, error) {
	var l ItemList
	for {
		c, err := read(lex)
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

func eval(expr Item, env *Env) (interface{}, error) {
	log.Printf("%v\n", expr)
	switch e := expr.(type) {
	case *Data:
		switch e.Type {
		case SymbolType:
			return env.FindVar(e)
		}
	case Number, String:
		return e, nil
	case ItemList:
		if len(e) == 0 {
			return Nil, nil
		}
		switch car, _ := e[0].(*Data); car.StringValue() {
		case "QUOTE":
			return e[1], nil
		case "DEFINE":
			v, err := eval(e[2], env)
			if err != nil {
				return nil, err
			}
			env.Bind(e[1].(*Data), v)
			return env.Var(e[1].(*Data))
		case "SET!":
			d := e[1].(*Data)
			var err error
			val, err := eval(e[2], env)
			if err != nil {
				return nil, err
			}
			env.Find(d).Bind(d, val)
			return val, nil
		case "IF":
			test, _ := eval(e[1], env)
			if isTrue(test) {
				return eval(e[2], env)
			} else if len(e) > 3 {
				return eval(e[3], env)
			}
			return Nil, nil
		case "BEGIN":
			var v Item
			for _, e := range e[1:] {
				v, _ = eval(e, env)
			}
			return v, nil

		case "QUIT":
			os.Exit(0)
		case "LAMBDA":
			return evalLambda(e, env)
		case ":VARS":
			for k, v := range env.vars {
				log.Printf("%v: %v\n", k, v)
			}
			return nil, nil

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
	params []*Data
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
		return nil, fmt.Errorf("bad params: %v", expr[1])
	}
	for _, x := range params {
		switch p := x.(type) {
		case *Data:
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
	case InternalFunc:
		return f(args...)
	case *Lambda:
		if len(f.params) != len(args) {
			return nil, fmt.Errorf("parameter mismatch %v != %v", f.params, args)
		}
		for i, p := range f.params {
			f.envt.Bind(p, args[i])
		}
		return eval(f.body, f.envt)
	case *Data:
		return eval(proc, env)
	default:
		return nil, fmt.Errorf("apply to a non function: %#v", proc)
	}
}

func replReader(in io.Reader, env *Env) (interface{}, error) {
	//	l := NewScanner(in)
	buf := make([]byte, 1024)
	n, _ := in.Read(buf)
	l := NewLexer("lispy", string(buf[:n]))
	var result interface{}
	for {
		var err error
		expr, err := read(l)
		//log.Println(expr, err)
		if err != nil {
			if err == ErrorEOF {
				return result, nil
			}
			return result, err
		}

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
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		result, err := repl(text, env)
		if err != nil && err != ErrorEOF {
			fmt.Println("Error:", err)
		} else {
			fmt.Println(result)
		}
	}
}

func ApplyNumeric(f func(Number, Number) Number) InternalFunc {
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

func ApplyNumericBool(f func(Number, Number) bool) InternalFunc {
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

func isTrue(i Item) bool {
	if b, ok := i.(bool); ok {
		return b
	}
	if a, ok := i.(*Data); ok {
		if a == Nil {
			return false
		}
		return true
	}
	if n, ok := i.(Number); ok {
		return !(n == 0)
	}

	if il, ok := i.(ItemList); ok {
		return len(il) > 0
	}
	if _, ok := i.(String); ok {
		return true
	}
	return false
}

func isItemList(i Item) bool {
	if _, ok := i.(ItemList); ok {
		return true
	}
	return false
}
func EmptyEnv() *Env {
	return NewEnv(nil)
}

func DefaultEnv() *Env {
	env := EmptyEnv()
	env.BindName("*", ApplyNumeric(func(x, y Number) Number {
		return x * y
	}))

	env.BindName("-", ApplyNumeric(func(x, y Number) Number {
		return x - y
	}))

	env.BindName("/", ApplyNumeric(func(x, y Number) Number {
		return x / y
	}))
	env.BindName("+", ApplyNumeric(func(x, y Number) Number {
		return x + y
	}))
	env.BindName("<", ApplyNumericBool(func(x, y Number) bool {
		return x < y
	}))
	env.BindName("<=", ApplyNumericBool(func(x, y Number) bool {
		return x <= y
	}))
	env.BindName(">", ApplyNumericBool(func(x, y Number) bool {
		return x > y
	}))
	env.BindName(">=", ApplyNumericBool(func(x, y Number) bool {
		return x >= y
	}))
	env.BindName("=", ApplyNumericBool(func(x, y Number) bool {
		return x == y
	}))
	env.BindName("number?", InternalFunc(func(a ...Item) (Item, error) {
		if len(a) > 1 {
			return nil, fmt.Errorf("Too many arguments for number?: %v", a)
		}
		if _, ok := a[0].(Number); ok {
			return Item(true), nil
		}
		return Item(false), nil
	}))
	env.BindName("car", InternalFunc(func(a ...Item) (Item, error) {
		if !isItemList(a[0]) {
			return nil, fmt.Errorf("Not a list: %#v", a[0])
		}
		il := a[0].(ItemList)
		if len(il) == 0 {
			return Nil, nil
		}
		return a[0].(ItemList)[0], nil
	}))
	env.BindName("cdr", InternalFunc(func(a ...Item) (Item, error) {
		if !isItemList(a[0]) {
			return nil, fmt.Errorf("Not a list: %#v", a[0])
		}
		il := a[0].(ItemList)
		if len(il) < 2 {
			return Nil, nil
		}
		return il[1:], nil
	}))
	env.BindName("cons", InternalFunc(func(a ...Item) (Item, error) {
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
	}))
	env.BindName("equal?", InternalFunc(func(a ...Item) (Item, error) {
		return reflect.DeepEqual(a[0], a[1]), nil
	}))
	env.BindName("pi", Number(math.Pi))
	return env
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	replCLI(DefaultEnv())
}
