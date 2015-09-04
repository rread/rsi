package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/rread/rsi/log"
)

// Data is a fundamental type (symbol, boolean, string, number, pair, func)
type Data interface {
	String() string
}

type Symbol string
type Boolean bool
type Number float64
type String string
type InternalFunc func(*Pair) (Data, error)

func (sym Symbol) String() string {
	return string(sym)
}

func (num Number) String() string {
	return fmt.Sprintf("%v", float64(num))
}

func (fun InternalFunc) String() string {
	return fmt.Sprintf("native-func '%v'", ((func(*Pair) (Data, error))(fun)))
}

func (b Boolean) String() string {
	if b {
		return "#t"
	} else {
		return "#f"
	}
}

func (s String) String() string {
	return fmt.Sprintf(`"%s"`, string(s))
}

type Tokenizer interface {
	NextItem() *TokenItem
}

var (
	T     = Boolean(true)
	False = Boolean(false)
	Nil   = (*Pair)(nil)
)

var ErrorEOF = errors.New("End of File")

func read(l Tokenizer) (Data, error) {
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
		return StringWithValue(t.lit), nil
	case TRUE:
		return T, nil
	case FALSE:
		return False, nil
	case ILLEGAL:
		return nil, errors.New(t.lit)
	}
	return nil, errors.New("Malformed input")
}

func readQuote(lex Tokenizer) (Data, error) {
	c, err := read(lex)
	if err != nil {
		return nil, fmt.Errorf("Failed to complete list: %v\n", err)
	}
	return cons(internSymbol("quote"), cons(c, Nil)), nil
}

func readList(lex Tokenizer) (Data, error) {
	var l []Data
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
	return list2cons(l), nil
}

func list2cons(l []Data) *Pair {
	p := Nil
	for i := len(l) - 1; i >= 0; i-- {
		p = cons(l[i], p)
	}
	return p
}

var (
	_quote  = internSymbol("quote")
	_define = internSymbol("define")
	_set    = internSymbol("set!")
	_if     = internSymbol("if")
	_begin  = internSymbol("begin")
	_quit   = internSymbol("quit")
	_lambda = internSymbol("lambda")
	_vars   = internSymbol(":vars")
	_ok     = internSymbol("ok")
)

func eval(expr Data, env *Env) (Data, error) {
	log.Printf("eval: %T: %v\n", expr, expr)
	switch e := expr.(type) {
	case Symbol:
		return env.FindVar(e)
	case Number:
		return e, nil
	case String:
		return e, nil
	case *Pair:
		if e == nil {
			return Nil, nil
		}
		c, _ := getSymbol(car(e))
		/* non-Symbols fall through to default */
		switch c {
		case _quote:
			return cadr(e), nil
		case _define:
			expr, err := getPair(cdr(e))
			if err != nil {
				return nil, err
			}
			err = definition(expr, env)
			if err != nil {
				return nil, err
			}
			// Return value of define is undefined
			return _ok, nil
		case _set:
			d := cadr(e).(Symbol)
			val, err := eval(caddr(e), env)
			if err != nil {
				return nil, err
			}
			env.Find(d).Bind(d, val)
			return nil, nil
		case _if:
			test, _ := eval(cadr(e), env)
			if isTrue(test) {
				return eval(caddr(e), env)
			} else if listLen(e) > 3 {
				return eval(cadddr(e), env)
			}
			return Nil, nil
		case _begin:
			var v Data
			e, err := getPair(cdr(e))
			if err != nil {
				return nil, err
			}
			for e != Nil {
				log.Printf("begin: %v", e)
				v, err = eval(car(e), env)
				if err != nil {
					return nil, err
				}
				e, err = listNext(e)
				if err != nil {
					return nil, err
				}

			}
			return v, nil

		case _quit:
			os.Exit(0)
		case _lambda:
			params, err := getPair(cadr(e))
			if err != nil {
				return nil, fmt.Errorf("bad params: %v", err)
			}
			body, err := getPair(cddr(e))
			if err != nil {
				return nil, fmt.Errorf("bad body: %v", err)
			}
			return evalLambda(params, body, env)
		case _vars:
			for k, v := range env.vars {
				log.Printf("%v: %v\n", k, v)
			}
			return nil, nil
		default:
			log.Printf("procedure call %v", e)
			proc, err := eval(car(e), env)
			if err != nil {
				return nil, err
			}
			args, err := evalArgs(cdr(e), env)
			if err != nil {
				return nil, err
			}
			return apply(proc, args, env)
		}
	case nil:
		log.Errorln("parsed a nil?")
		return Nil, nil
	}
	return nil, fmt.Errorf("Unparsable expression: %v", expr)
}

func definition(defn *Pair, env *Env) error {
	var value Data
	var name Symbol
	var err error
	switch e := car(defn).(type) {
	// (define var value)
	case Symbol:
		name = e
		value, err = eval(cadr(defn), env)
		if err != nil {
			return err
		}
	// (define (proc a b) (body))
	case *Pair:
		name, err = getSymbol(car(e))
		if err != nil {
			return err
		}
		params, err := getPair(cdr(e))
		if err != nil {
			return err
		}
		body, err := getPair(cdr(defn))
		if err != nil {
			return err
		}
		value, err = evalLambda(params, body, env)
		if err != nil {
			return err
		}
	}

	env.Bind(name, value)
	return nil
}

type Lambda struct {
	params []Symbol
	body   []Data
	envt   *Env
}

func (l *Lambda) String() string {
	return fmt.Sprintf("<function %p: %v>", l, l.body)
}

func evalLambda(params *Pair, body *Pair, env *Env) (Data, error) {
	l := Lambda{}
	for params != nil {
		var err error
		switch p := car(params).(type) {
		case Symbol:
			l.params = append(l.params, p)
		case *Pair:
			return nil, fmt.Errorf("combo param not supported: %v", p)
		}
		d := cdr(params)
		if d == nil {
			break
		}
		if params, err = getPair(d); err != nil {
			return nil, fmt.Errorf("not a pair here %v", d)
		}

	}
	for {
		if body == nil {
			break
		}
		l.body = append(l.body, car(body))
		body, _ = listNext(body)
	}
	l.envt = NewEnv(env)
	return &l, nil

}

func evalArgs(next Data, env *Env) (*Pair, error) {
	e, err := getPair(next)
	if err != nil {
		return nil, err
	}

	if e == nil {
		return Nil, nil
	}

	val, err := eval(car(e), env)
	if err != nil {
		return nil, err
	}

	rest, err := evalArgs(cdr(e), env)
	if err != nil {
		return nil, err
	}
	return cons(val, rest), nil
}

func apply(proc Data, args *Pair, env *Env) (Data, error) {
	//	log.Printf("apply: %v args: %v\n", proc, args)
	switch f := proc.(type) {
	case InternalFunc:
		return f(args)
	case *Lambda:
		if len(f.params) != listLen(args) {
			return nil, fmt.Errorf("parameter mismatch %v != %v", f.params, args)
		}
		for _, p := range f.params {
			var err error
			f.envt.Bind(p, car(args))
			args, err = listNext(args)
			if err != nil {
				return nil, err
			}
		}
		var result Data
		var err error
		for _, expr := range f.body {
			result, err = eval(expr, f.envt)
			if err != nil {
				return nil, err
			}
		}
		return result, nil
	default:
		return nil, fmt.Errorf("apply to a non function: %#v %v", proc, args)
	}
}

func replReader(in io.Reader, env *Env) (Data, error) {
	//	l := NewScanner(in)
	buf := make([]byte, 1024)
	n, _ := in.Read(buf)
	l := NewLexer("lispy", string(buf[:n]))
	var result Data
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

func repl(in string, env *Env) (Data, error) {
	return replReader(strings.NewReader(in), env)
}

func replCLI(env *Env) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("-> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("\nbye!")
			os.Exit(0)
		}
		result, err := repl(text, env)
		if err != nil && err != ErrorEOF {
			fmt.Println("Error:", err)
		} else {
			fmt.Println(result)
		}
	}
}

func ApplyNumeric(f func(Number, Number) Number) InternalFunc {
	return func(a *Pair) (Data, error) {
		v, ok := car(a).(Number)
		if !ok {
			return nil, fmt.Errorf("Not a number: %v", car(a))
		}
		for {
			var err error
			a, err = listNext(a)
			if err != nil {
				return nil, err
			}
			if a == nil {
				break
			}
			i, ok := car(a).(Number)
			if !ok {
				return nil, fmt.Errorf("Not a number: %v", car(a))
			}
			v = f(v, i)
		}
		return v, nil
	}
}

func ApplyNumericBool(f func(Number, Number) Boolean) InternalFunc {
	return func(a *Pair) (Data, error) {
		var ret Boolean
		v, ok := car(a).(Number)
		if !ok {
			return nil, fmt.Errorf("Not a number: %v", car(a))
		}
		for {
			var err error
			a, err = listNext(a)
			if err != nil {
				return nil, err
			}
			if a == nil {
				break
			}
			i, ok := car(a).(Number)
			if !ok {
				return nil, fmt.Errorf("Not a number: %v", car(a))
			}
			ret = f(v, i)
			if !ret {
				return Boolean(ret), nil
			}
			v = i
		}
		return Boolean(ret), nil
	}
}

func Apply1(f func(Data) (Data, error)) InternalFunc {
	return func(args *Pair) (Data, error) {
		if listLen(args) != 1 {
			return nil, fmt.Errorf("Expected 1 arguments, received %d", listLen(args))
		}
		return f(car(args))
	}
}

func Apply2(f func(Data, Data) (Data, error)) InternalFunc {
	return func(args *Pair) (Data, error) {
		if listLen(args) != 2 {
			return nil, fmt.Errorf("Expected 2 arguments, received %d", listLen(args))
		}
		return f(car(args), cadr(args))
	}
}

func isTrue(i Data) Boolean {
	if b, ok := i.(Boolean); ok {
		return b
	}
	if a, ok := i.(Symbol); ok {
		if a == internSymbol("nil") {
			return false
		}
		return true
	}
	if n, ok := i.(Number); ok {
		return !(n == 0)
	}
	if il, ok := i.(*Pair); ok {
		/* FIXME: empty list if false, like Lisp */
		return listLen(il) > 0
	}
	if _, ok := i.(String); ok {
		return true
	}
	return false
}

func EmptyEnv() *Env {
	return NewEnv(nil)
}

func DefaultEnv() *Env {
	env := EmptyEnv()
	//	env.Bind(internSymbol("#t"), T)
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
	env.BindName("<", ApplyNumericBool(func(x, y Number) Boolean {
		return x < y
	}))
	env.BindName("<=", ApplyNumericBool(func(x, y Number) Boolean {
		return x <= y
	}))
	env.BindName(">", ApplyNumericBool(func(x, y Number) Boolean {
		return x > y
	}))
	env.BindName(">=", ApplyNumericBool(func(x, y Number) Boolean {
		return x >= y
	}))
	env.BindName("=", ApplyNumericBool(func(x, y Number) Boolean {
		return x == y
	}))
	env.BindName("number?", InternalFunc(func(a *Pair) (Data, error) {
		if listLen(a) > 1 {
			return nil, fmt.Errorf("Too many arguments for number?: %v", a)
		}
		if _, ok := car(a).(Number); ok {
			return T, nil
		}
		return False, nil
	}))
	env.BindName("equal?", InternalFunc(func(args *Pair) (Data, error) {
		if listLen(args) != 2 {
			return nil, fmt.Errorf("wrong number of arguments given to equal?")
		}
		return Boolean(reflect.DeepEqual(car(args), cadr(args))), nil
	}))
	env.BindName("pi", Number(math.Pi))
	env.BindName("cons", Apply2(_cons))
	env.BindName("car", Apply1(_car))
	env.BindName("cdr", Apply1(_cdr))
	env.BindName("null?", Apply1(_nullp))
	env.BindName("pair?", Apply1(_pairp))
	env.Bind(internSymbol("nil"), nil)
	return env
}

func _cons(a Data, b Data) (Data, error) {
	return cons(a, b), nil
}

func _car(a Data) (Data, error) {
	p, err := getPair(a)
	if err != nil {
		return Nil, fmt.Errorf("car received: %v", err)
	}
	return car(p), nil
}

func _cdr(a Data) (Data, error) {
	p, err := getPair(a)
	if err != nil {
		return Nil, fmt.Errorf("cdr received: %v", err)
	}
	return cdr(p), nil
}

func _nullp(a Data) (Data, error) {
	if p, ok := a.(*Pair); ok {
		return nullp(p), nil
	}
	return False, nil
}

func _pairp(a Data) (Data, error) {
	return pairp(a), nil
}
