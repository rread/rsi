package main

import (
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
	// temporarily removed this as error handling is integrated into Data
	// String() string
}

type Symbol string
type Boolean bool
type Number float64
type String string
type InternalFunc func(Data) (Data, error)

func (sym Symbol) String() string {
	return string(sym)
}

func (num Number) String() string {
	return fmt.Sprintf("%v", float64(num))
}

func (fun InternalFunc) String() string {
	return fmt.Sprintf("native-func '%v'", ((func(Data) (Data, error))(fun)))
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

type Null byte

var Empty Null = 0xfe // Null is nothing so it can be anything

func (n Null) String() string {
	return "()"
}

func nullp(d Data) Boolean {
	if v, ok := d.(Null); ok {
		return v == Empty
	}
	return false
}

type Tokenizer interface {
	NextItem() *TokenItem
}

var (
	T     = Boolean(true)
	False = Boolean(false)
)

var ErrorEOF = errors.New("End of File")

func read(l Tokenizer) (Data, error) {
	t := l.NextItem()
	if t == nil {
		return nil, ErrorEOF
	}
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
	case DOT:
		return _dot, nil
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
	return cons(internSymbol("quote"), cons(c, Empty)), nil
}

func readList2(lex Tokenizer) (Data, error) {
	c, err := read(lex)
	if err != nil {
		return nil, fmt.Errorf("Failed to complete list: %v\n", err)
	}
	if c == nil {
		return Empty, nil
	}

	// handle (a b . c) but (a b . c d) is an error.
	if c == _dot {
		last, err := read(lex)
		if err != nil {
			return nil, err
		}
		end, _ := read(lex)
		if end != nil {
			return nil, fmt.Errorf("More than one object follows .")
		}

		log.Printf("last %v", last)
		return last, nil
	}

	rest, err := readList2(lex)
	if err != nil {
		return nil, err
	}
	return cons(c, rest), nil
}

func readList(lex Tokenizer) (Data, error) {
	li, err := readList2(lex)
	if err != nil {
		return nil, err
	}
	return li, nil
}

var (
	_quote  = internSymbol("quote")
	_define = internSymbol("define")
	_dot    = internSymbol("::dot::")
	_set    = internSymbol("set!")
	_if     = internSymbol("if")
	_begin  = internSymbol("begin")
	_quit   = internSymbol("quit")
	_lambda = internSymbol("lambda")
	_let    = internSymbol("let")
	_vars   = internSymbol(":vars")
	_ok     = internSymbol("ok")
)

func eval(expr Data, env *Env) (Data, error) {
	log.Printf("eval: %T: %v\n", expr, expr)
	for {
		switch e := expr.(type) {
		case Boolean:
			return e, nil
		case Symbol:
			return env.FindVar(e)
		case Number:
			return e, nil
		case String:
			return e, nil
		case Null:
			return e, nil
		case *Pair:
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
				d, err := getSymbol(cadr(e))
				if err != nil {
					return nil, err
				}
				val, err := eval(caddr(e), env)
				if err != nil {
					return nil, err
				}
				env.Find(d).Bind(d, val)
				return nil, nil
			case _if:
				test, _ := eval(cadr(e), env)
				if isTrue(test) {
					expr = caddr(e)
				} else if listLen(e) > 3 {
					expr = cadddr(e)
				} else {
					return Empty, nil
				}
			case _let:
				letExpr, err := let(cdr(e), env)
				if err != nil {
					return nil, err
				}
				return eval(letExpr, env)
			case _begin:
				e, err := getPair(cdr(e))
				if err != nil {
					return nil, err
				}
				for !nullp(e) {
					if nullp(cdr(e)) {
						expr = car(e)
						break
					}
					_, err = eval(car(e), env)
					if err != nil {
						return nil, err
					}
					e, err = listNext(e)
					if err != nil {
						return nil, err
					}

				}

			case _quit:
				os.Exit(0)
			case _lambda:
				params, err := getList(cadr(e))
				if err != nil {
					return nil, fmt.Errorf("bad params: %v", err)
				}
				body, err := getList(cddr(e))
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
				switch f := proc.(type) {
				case InternalFunc:
					return f(args)
				case *Lambda:
					var err error
					env, err = ExtendEnv(f.params, args, f.envt)
					if err != nil {
						return nil, err
					}
					expr = cons(_begin, f.body)
				default:
					return nil, fmt.Errorf("apply to a non function: %#v %v", proc, args)
				}

			}
		case nil:
			log.Fatal("parsed a nil?")
			return nil, nil
		}
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
	body   Data
	envt   *Env
}

func (l *Lambda) String() string {
	return fmt.Sprintf("<function %p: %v>", l, l.body)
}

func evalLambda(params Data, body Data, env *Env) (Data, error) {
	l := Lambda{}
	for params != Empty {
		var err error
		p, err := getPair(params)
		if err != nil {
			return nil, fmt.Errorf("not a pair here %v", params)
		}
		switch arg := car(p).(type) {
		case Symbol:
			l.params = append(l.params, arg)
		case *Pair:
			return nil, fmt.Errorf("combo param not supported: %v", arg)
		}
		params = cdr(p)
	}
	l.body = body
	l.envt = NewEnv(env)
	return &l, nil

}

func evalArgs(next Data, env *Env) (Data, error) {
	if nullp(next) {
		return Empty, nil
	}
	e, err := getPair(next)
	if err != nil {
		return nil, err
	}

	val, err := eval(car(e), env)
	if err != nil {
		return nil, err
	}

	if nullp(cdr(e)) {
		return cons(val, Empty), nil
	} else {
		rest, err := evalArgs(cdr(e), env)
		if err != nil {
			return nil, err
		}
		return cons(val, rest), nil
	}
}

func getError(d Data) error {
	v, ok := d.(error)
	if ok {
		return v
	}
	return nil
}

func _map(f func(Data) Data, d Data) Data {
	if nullp(d) {
		return Empty
	}
	lst, err := getPair(d)
	if err != nil {
		return err
	}
	if nullp(lst) {
		return Empty
	}
	e := f(car(lst))
	if err := getError(e); err != nil {
		return err
	}
	rest := cdr(lst)
	if err := getError(rest); err != nil {
		return err
	}
	return cons(e, _map(f, rest))
}

func let(expr Data, env *Env) (Data, error) {
	arguments := _map(car, car(expr))
	if err := getError(arguments); err != nil {
		return nil, err
	}
	values := _map(cadr, car(expr))
	if err := getError(values); err != nil {
		return nil, err
	}

	body := cdr(expr)
	if err := getError(body); err != nil {
		return nil, err
	}

	result := cons(cons(_lambda, cons(arguments, body)), values)
	log.Printf("lambda %v", result)

	return eval(result, env)
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

func ApplyNumeric(f func(Number, Number) Number) InternalFunc {
	return func(d Data) (Data, error) {
		if nullp(d) {
			return 0, nil
		}
		a, err := getPair(d)
		if err != nil {
			return nil, err
		}
		v, ok := car(a).(Number)
		if !ok {
			return nil, fmt.Errorf("Not a number: %v", car(a))
		}
		for {
			var err error
			if nullp(cdr(a)) {
				break
			}
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
	return func(d Data) (Data, error) {
		var ret Boolean
		if nullp(d) {
			return 0, nil
		}
		a, err := getPair(d)
		if err != nil {
			return nil, err
		}
		v, ok := car(a).(Number)
		if !ok {
			return nil, fmt.Errorf("Not a number: %v", car(a))
		}
		for {
			var err error
			if nullp(cdr(a)) {
				break
			}
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
	return func(args Data) (Data, error) {
		if listLen(args) != 1 {
			return nil, fmt.Errorf("Expected 1 arguments, received %d", listLen(args))
		}
		return f(car(args))
	}
}

func Apply2(f func(Data, Data) (Data, error)) InternalFunc {
	return func(args Data) (Data, error) {
		if listLen(args) != 2 {
			return nil, fmt.Errorf("Expected 2 arguments, received %d", listLen(args))
		}

		a, err := getPair(args)
		if err != nil {
			return nil, err
		}
		return f(car(a), cadr(a))
	}
}

func isTrue(i Data) Boolean {
	log.Printf("isTrue %T %v", i, i)
	if b, ok := i.(Boolean); ok {
		return b
	}
	if _, ok := i.(Symbol); ok {
		return true
	}
	if n, ok := i.(Number); ok {
		return !(n == 0)
	}
	if _, ok := i.(*Pair); ok {
		return true
	}
	if _, ok := i.(Null); ok {
		return true
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
	env.BindName("number?", InternalFunc(func(args Data) (Data, error) {
		if listLen(args) > 1 {
			return nil, fmt.Errorf("Too many arguments for number?: %v", args)
		}
		a, err := getPair(args)
		if err != nil {
			return nil, err
		}
		if _, ok := car(a).(Number); ok {
			return T, nil
		}
		return False, nil
	}))
	env.BindName("equal?", InternalFunc(func(args Data) (Data, error) {
		if listLen(args) != 2 {
			return nil, fmt.Errorf("wrong number of arguments given to equal?")
		}
		a, err := getPair(args)
		if err != nil {
			return nil, err
		}
		log.Printf("%v", a)
		return Boolean(reflect.DeepEqual(car(a), cadr(a))), nil
	}))
	env.BindName("pi", Number(math.Pi))
	env.BindName("cons", Apply2(_cons))
	env.BindName("car", Apply1(_car))
	env.BindName("cdr", Apply1(_cdr))
	env.BindName("null?", Apply1(_nullp))
	env.BindName("pair?", Apply1(_pairp))
	return env
}

func _cons(a Data, b Data) (Data, error) {
	return cons(a, b), nil
}

func _car(a Data) (Data, error) {
	p, err := getPair(a)
	if err != nil {
		return nil, fmt.Errorf("car received: %v", err)
	}
	return car(p), nil
}

func _cdr(a Data) (Data, error) {
	p, err := getPair(a)
	if err != nil {
		return nil, fmt.Errorf("cdr received: %v", err)
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
