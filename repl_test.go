package main

import (
	"fmt"
	"math"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func S(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

var result Data

func BenchmarkFactorial(b *testing.B) {
	b.StopTimer()
	env := DefaultEnv()
	value, err := repl("(define fact (lambda (n) (if (<= n 1) 1 (* n (fact (- n 1))))))", env)
	if err != nil {
		panic(err)
	}
	result = value
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		val, err := repl("(fact 100)", env)
		if err != nil {
			panic(err)
		}
		result = val
	}
}

func TestRepl(t *testing.T) {
	Convey("basic lexer testing", t, func() {
		env := EmptyEnv()
		lexer := []TestCase{
			{"", nil, ""},
			{" \t\r\n", nil, ""},
			{"  ()   ", Empty, ""},
			{"123", 123, ""},
			{`"123"`, `"123"`, ""},
			{"'-", "-", ""},
			{"  456  ", 456, ""},
			{".123 ", 0.123, ""},
			{"'(a . b)", "(A . B)", ""},
			{`"bad string`, nil, "unterminated string"},
			{`"string with \" quote`, nil, "unterminated string"},
			{`"string with \"`, nil, "unterminated string"},
			{`"string with \" quote"`, `"string with " quote"`, ""},
			{"123 ; comment", 123, ""},
			{"123 ; comment\n", 123, ""},
			{"#n", nil, "unsupported hash code #n"},
			{`"\`, nil, "unterminated string"},
		}

		doCases("Lexer Tests", lexer, env)

	})

	Convey("testing", t, func() {
		env := DefaultEnv()

		literals := []TestCase{
			{"#t", T, ""},
			{"#f", False, ""},
			{"123", 123, ""},
			{"-123", -123, ""},
			{"-123.123", -123.123, ""},
		}
		doCases("Testing Literals", literals, env)

		Convey("Test predicates", func() {
			predicates := []TestCase{
				{"(null? ())", T, ""},
				{"(null? 123)", False, ""},
				{"(equal? 1 1)", T, ""},
				{"(equal? 1 30)", False, ""},
				{"(pair? '(a b))", T, ""},
				{"(pair? 30)", False, ""},
			}
			doCases("Predicates", predicates, env)

		})

		Convey("If statements", func() {
			conditionals := []TestCase{
				{"(if (< 4 2) (* 10 2) (+ 1 2))", 3, ""},
				{"(if (> 4 2) (* 10 2) (+ 1 2))", 20, ""},
				{"(if '(1 2) (* 10 2) (+ 1 2))", 20, ""},
				{"(if 'atom 'a 'b)", "A", ""},
				{"(if '() 'a 'b)", "A", ""},
				{"(if 0 'a 'b)", "B", ""},
				{"(if \"\" 'a 'b)", "A", ""},
			}
			doCases("Conditionals", conditionals, env)

		})

		Convey("Test statements", func() {
			statements := []TestCase{
				{"(quote (1  1))", "(1 1)", ""},
				{"'(1  1)", "(1 1)", ""},
				{"(lambda () (+ 2 3) (+ 1 1))", "_", ""},
				{"((lambda () (+ 1 1)))", 2, ""},
				{"(define foo ((lambda (x) (lambda () (set! x (+ x 1)) x)) 0))", "OK", ""},
				{"(foo) (foo)", 2, ""},
				{"(foo) (foo)", 4, ""},
				{"(define plus (lambda (a b) (+ a b)))", "OK", ""},
				{"(plus 10 -2)", 8, ""},
				{"(plus 10)", nil, "parameter mismatch"},
				{"(not-func 10)", nil, "Undefined symbol: NOT-FUNC"},
				{"(begin (+ 2 3) (* 5 8))", 40, ""},
			}
			doCases("Statements", statements, env)
		})

		Convey("Test arithmetic", func() {
			var a float64 = 30
			var b float64 = 40
			arithmetic := []TestCase{
				{fmt.Sprintf("(* %v %v)", a, b), a * b, ""},
				{fmt.Sprintf("(/ %v %v)", a, b), a / b, ""},
				{fmt.Sprintf("(+ %v %v)", a, b), a + b, ""},
				{fmt.Sprintf("(- %v %v)", a, b), a - b, ""},
				{fmt.Sprintf("(- 'atom %v)", a), nil, "Not a number: ATOM"},
				{fmt.Sprintf("(- %v 'atom)", a), nil, "Not a number: ATOM"},
			}
			doCases("Arithmetic", arithmetic, env)
		})
		Convey("Test numeric compare", func() {
			var a float64 = -10.1
			var b float64 = 40.123
			var c float64 = 70

			numerics := []TestCase{
				{fmt.Sprintf("(< %v 'atom)", a), nil, "Not a number: ATOM"},
				{fmt.Sprintf("(< 'atom %v)", a), nil, "Not a number: ATOM"},
				{fmt.Sprintf("(< %v %v)", a, b), a < b, ""},
				{fmt.Sprintf("(> %v %v)", a, b), a > b, ""},
				{fmt.Sprintf("(< %v %v %v)", a, b, c), a < b && b < c, ""},
				{fmt.Sprintf("(> %v %v)", b, a), b > a, ""},
				{fmt.Sprintf("(< %v %v)", b, a), b < a, ""},
				{fmt.Sprintf("(<= %v %v)", a, a), a <= a, ""},
				{fmt.Sprintf("(<= %v %v %v)", a, b, b), a <= b && b <= b, ""},
				{fmt.Sprintf("(<= %v %v %v)", b, b, a), b <= a && b <= a, ""},
				{fmt.Sprintf("(>= %v %v %v)", b, b, a), b >= a && b >= a, ""},
				{fmt.Sprintf("(>= %v %v %v)", a, b, b), a >= b && b >= b, ""},
				{fmt.Sprintf("(= %v %v %v)", a, a, a), a == a, ""},
				{fmt.Sprintf("(= %v %v %v)", a, b, b), a == b, ""},
				{fmt.Sprintf("(number? %v)", a), T, ""},
				{fmt.Sprintf("(number? '(%v))", a), False, ""},
				{fmt.Sprintf("(number? \"%v\")", a), False, ""},
				{fmt.Sprintf("(number? 'atom)"), False, ""},
			}
			doCases("Numeric Comparisons", numerics, env)
		})

		Convey("Define tests", func() {
			defines := []TestCase{
				{"(+ n n)", nil, "Undefined symbol: N"},
				{"(define radius (* pi (* r r)))", nil, "Undefined symbol: R"},
				{"radius", nil, "Undefined symbol: RADIUS"},
				{"(define r 10)\n(define n 12)", "OK", ""},
				{"(+ r n)", 22, ""},
				{"(define radius (* pi (* r r)))", "OK", ""},
				{"radius", math.Pi * 10 * 10, ""},
				{"(/ radius 10)", math.Pi * 10, ""},
			}
			doCases("Define variables", defines, env)

			procs := []TestCase{
				{"(define (double a)  (+ a a))", "OK", ""},
				{"(double 22)", 44, ""},
			}
			doCases("Define procedures", procs, env)
		})

		strings := []TestCase{
			{`"asdf"`, `"asdf"`, ""},
			// {`"asdf\""`, `"asdf\""`, ""},
		}
		doCases("Test Strings", strings, env)

		pairs := []TestCase{
			{"(cons 1 2)", "(1 . 2)", ""},
			{"(cdr (cons 1 2))", 2, ""},
			{"(cons 1 (cons 2 3))", "(1 2 . 3)", ""},
			{"(cons 1 '(2 3))", "(1 2 3)", ""},
			{"(car (cons 1 '(2 3)))", "1", ""},
			{"(cdr (cons 1 '(2 3)))", "(2 3)", ""},
			{"(cdr (cdr (cons 1 '(2 3))))", "(3)", ""},
			{"(cdr (cdr (cdr (cons 1 '(2 3)))))", Empty, ""},
		}
		doCases("Test Pairs", pairs, env)

		factorial := []TestCase{
			{"(define fact (lambda (n) (if (<= n 1) 1 (* n (fact (- n 1))))))", "OK", ""},
			{"(fact 10)", 3.6288e+06, ""},
		}
		doCases("factorial", factorial, env)

		letCases := []TestCase{
			{"(let ((a 1) (b 2)) (+ a b))", Number(3), ""},
			{"(let ((a 3) (b 2)) (* a b))", Number(6), ""},
			{"(let () (* a b))", nil, "Undefined symbol: A"},
			{"(let (()) (* a b))", nil, "(): value is not a pair"},
			{"(let ((a 1) ()) (* a b))", nil, "(): value is not a pair"},
			{"(let ((a 1)) )", nil, "(): value is not a pair"},
		}
		doCases("Test let statements", letCases, env)

	})
}

func TestEndofFile(t *testing.T) {
	Convey("fail when given imbalanced parens", t, func() {
		env := DefaultEnv()
		imbalanced := []TestCase{
			{"(", nil, "End of File"},
			{`(define counter ` +
				`(lambda (n)` +
				` (lambda () (set! n (+ n 1))))`,
				nil, "End of File"},
			{"(let ((a 1 ) (+ a a)))", nil, "Undefined symbol: A"},
		}
		doCases("fail when given imbalanced parens", imbalanced, env)
	})
}

type TestCase struct {
	expr        string
	value       Data
	expectedErr string
}

func doRepl(env *Env, exp string, value Data, expectedErr string) {
	if expectedErr == "" {
		Printf("%v => %v\n", exp, value)
	} else {
		Printf("%v !FAIL> \"%v\"\n", exp, expectedErr)
	}
	val, err := repl(exp, env)
	if err != nil {
		if expectedErr == "" {
			expectedErr = "UNEXPECTED ERROR!"
		}
		So(err.Error(), ShouldContainSubstring, expectedErr)
	} else if err == nil && expectedErr != "" {
		So(fmt.Sprintf("%v: No Error Returned", exp), ShouldContainSubstring, expectedErr)
	}
	if s, ok := value.(string); ok {
		// Expecting "_" means ignore
		if s != "_" {
			So(S(val), ShouldEqual, s)
		}
	} else {
		So(val, ShouldEqual, value)
	}
}

func doCases(name string, cases []TestCase, env *Env) {
	Convey(name, func() {
		for _, c := range cases {
			doRepl(env, c.expr, c.value, c.expectedErr)
		}
	})
}
