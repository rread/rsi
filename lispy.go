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

// Data is a fundamental type (symbol, string, number, list, etc)
type Data interface {
	String() string
}

// DataList is a list of Data items.
type DataList []Data

type Symbol string
type Boolean bool
type Number float64
type String string
type InternalFunc func(...Data) (Data, error)

func (sym Symbol) String() string {
	return string(sym)
}

func (num Number) String() string {
	return fmt.Sprintf("%v", float64(num))
}

func (fun InternalFunc) String() string {
	return fmt.Sprintf("native-func '%v'", ((func(...Data) (Data, error))(fun)))
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
	Nil   = DataList(nil)
)

var ErrorEOF = errors.New("End of File")

func (items DataList) String() string {
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

func Car(items DataList) Data {
	return items[0]
}

func Cdr(items DataList) DataList {
	return items[1:]
}

func Cadr(items DataList) Data {
	return items[1]
}

func Caddr(items DataList) Data {
	return items[2]
}

func Cadddr(items DataList) Data {
	return items[3]
}

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
	l := DataList{internSymbol("quote")}
	c, err := read(lex)
	if err != nil {
		return nil, fmt.Errorf("Failed to complete list: %v\n", err)
	}
	return append(l, c), nil
}

func readList(lex Tokenizer) (Data, error) {
	var l DataList
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

var (
	_quote  = internSymbol("quote")
	_define = internSymbol("define")
	_set    = internSymbol("set!")
	_if     = internSymbol("if")
	_begin  = internSymbol("begin")
	_quit   = internSymbol("quit")
	_lambda = internSymbol("lambda")
	_vars   = internSymbol(":vars")
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
	case DataList:
		if len(e) == 0 {
			return Nil, nil
		}
		switch car, _ := Car(e).(Symbol); car {
		case _quote:
			return Cadr(e), nil
		case _define:
			v, err := eval(Caddr(e), env)
			if err != nil {
				return nil, err
			}
			env.Bind(e[1].(Symbol), v)
			return env.Var(e[1].(Symbol))
		case _set:
			d := Cadr(e).(Symbol)
			var err error
			val, err := eval(Caddr(e), env)
			if err != nil {
				return nil, err
			}
			env.Find(d).Bind(d, val)
			return val, nil
		case _if:
			test, _ := eval(Cadr(e), env)
			if isTrue(test) {
				return eval(Caddr(e), env)
			} else if len(e) > 3 {
				return eval(Cadddr(e), env)
			}
			return Nil, nil
		case _begin:
			var v Data
			for _, e := range Cdr(e) {
				v, _ = eval(e, env)
			}
			return v, nil

		case _quit:
			os.Exit(0)
		case _lambda:
			return evalLambda(e, env)
		case _vars:
			for k, v := range env.vars {
				log.Printf("%v: %v\n", k, v)
			}
			return nil, nil

		default:
			proc, err := eval(Car(e), env)
			if err != nil {
				return nil, fmt.Errorf("undefined-function: %v", err)
			}
			args := make(DataList, len(e)-1)
			for i, a := range Cdr(e) {
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
	params []Symbol
	body   Data
	envt   *Env
}

func (l *Lambda) String() string {
	return fmt.Sprintf("<function %p: %v>", l, l.body)
}

func evalLambda(expr DataList, env *Env) (Data, error) {
	l := Lambda{}
	params, ok := Cadr(expr).(DataList)
	if !ok {
		return nil, fmt.Errorf("bad params: %v", expr[1])
	}
	for _, x := range params {
		switch p := x.(type) {
		case Symbol:
			l.params = append(l.params, p)
		case DataList:
			return nil, fmt.Errorf("combo param not supported: %v", x)
		}

	}
	//	log.Printf("lambda %#v\n", expr[2])
	l.body = expr[2]
	l.envt = NewEnv(env)
	return &l, nil

}

func apply(proc Data, args DataList, env *Env) (Data, error) {
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
		//	case *Data:
		//		return eval(proc, env)
	default:
		return nil, fmt.Errorf("apply to a non function: %#v", proc)
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
	return func(a ...Data) (Data, error) {
		v, ok := Car(a).(Number)
		if !ok {
			return nil, fmt.Errorf("Not a number: %v", Car(a))
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

func ApplyNumericBool(f func(Number, Number) Boolean) InternalFunc {
	return func(a ...Data) (Data, error) {
		var ret Boolean
		v, ok := Car(a).(Number)
		if !ok {
			return nil, fmt.Errorf("Not a number: %v", Car(a))
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
		return Boolean(ret), nil
	}
}

func Apply2(f func(Data, Data) (Data, error)) InternalFunc {
	return func(a ...Data) (Data, error) {
		if len(a) != 2 {
			return nil, fmt.Errorf("Expected 2 arguments, received %d", len(a))
		}
		return f(a[0], a[1])
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
	if il, ok := i.(DataList); ok {
		return len(il) > 0
	}
	if _, ok := i.(String); ok {
		return true
	}
	return false
}

func isDataList(i Data) Boolean {
	if _, ok := i.(DataList); ok {
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
	env.BindName("number?", InternalFunc(func(a ...Data) (Data, error) {
		if len(a) > 1 {
			return nil, fmt.Errorf("Too many arguments for number?: %v", a)
		}
		if _, ok := a[0].(Number); ok {
			return Boolean(true), nil
		}
		return Boolean(false), nil
	}))
	env.BindName("car", InternalFunc(func(a ...Data) (Data, error) {
		if il, ok := a[0].(DataList); ok {
			if len(il) < 1 {
				return Nil, nil
			}
			return Car(il), nil
		}
		return nil, fmt.Errorf("Not a list: %#v", a[0])
	}))
	env.BindName("cdr", InternalFunc(func(a ...Data) (Data, error) {
		if il, ok := a[0].(DataList); ok {
			if len(il) < 2 {
				return Nil, nil
			}
			return Cdr(il), nil
		}
		return nil, fmt.Errorf("Not a list: %#v", a[0])
	}))
	env.BindName("cons", InternalFunc(func(a ...Data) (Data, error) {
		if len(a) != 2 {
			return nil, fmt.Errorf("wrong number of arguments given to cons")
		}
		l := DataList{a[0]}
		if a[1] == nil {
			return l, nil
		}
		if il, ok := a[1].(DataList); ok {
			l = append(l, il...)
		} else {
			l = append(l, a[1])
		}
		return l, nil
	}))
	env.BindName("equal?", InternalFunc(func(a ...Data) (Data, error) {
		return Boolean(reflect.DeepEqual(a[0], a[1])), nil
	}))
	env.BindName("null?", InternalFunc(func(a ...Data) (Data, error) {
		if len(a) != 1 {
			return nil, fmt.Errorf("wrong number of arguments passed")
		}
		if a[0] == nil {
			return T, nil
		}

		if l, ok := a[0].(DataList); ok {
			if len(l) == 0 {
				return T, nil
			}
		}
		return False, nil
	}))
	env.BindName("pi", Number(math.Pi))
	env.BindName("mcons", Apply2(cons))
	env.BindName("mcar", Primitive(car))
	env.BindName("mcdr", Primitive(cdr))
	env.Bind(internSymbol("nil"), nil)
	return env
}

func Primitive(f func(Data) Data) InternalFunc {
	return func(a ...Data) (Data, error) {
		return f(a[0]), nil
	}
}

type Pair struct {
	car Data
	cdr Data
}

func cons(car, cdr Data) (Data, error) {
	return &Pair{car, cdr}, nil
}

func PairP(d Data) bool {
	_, ok := d.(*Pair)
	if !ok {
		log.Printf("consp %T %v %v", d, d, ok)
	}
	return ok
}

func car(l Data) Data {
	if isNil(l) {
		return Nil
	}
	if PairP(l) {
		return l.(*Pair).car
	}
	log.Fatalf("Not a list: %T", l)
	return Nil
}

func isNil(l Data) bool {
	//	log.Printf("isNil %T %v", l, l == nil)
	return l == nil
}

func cdr(l Data) Data {
	if isNil(l) {
		return Nil
	}
	if PairP(l) {
		return l.(*Pair).cdr
	}
	log.Fatalf("Not a list: %T", l)
	return Nil
}

func (p *Pair) String() string {
	ret := "("
	//	var d Data = p
	//	log.Printf("Pair.String:  %T %v %v\n", d, car(d), cdr(d))
	for {
		ret += fmt.Sprintf("%v", p.car)
		if isNil(p.cdr) {
			break
		}
		if PairP(p.cdr) {
			ret += " "
			p = p.cdr.(*Pair)
			continue
		} else {
			ret += " . " + fmt.Sprintf("%v", p.cdr)
			break
		}
		break
	}
	ret = ret + ")"
	return ret
}
