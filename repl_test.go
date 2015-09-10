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
		val, err := repl("", env)
		So(err, ShouldBeNil)
		So(val, ShouldBeNil)

		val, err = repl(" \t\r\n", env)
		So(err, ShouldBeNil)
		So(val, ShouldBeNil)

		val, err = repl("  ()   ", env)
		So(err, ShouldBeNil)
		So(val, ShouldEqual, Empty)

		val, err = repl("123", env)
		So(err, ShouldBeNil)
		So(val, ShouldEqual, 123)

		val, err = repl("\"123\"", env)
		So(err, ShouldBeNil)
		So(val, ShouldEqual, "123")

		val, err = repl("'-", env)
		So(err, ShouldBeNil)
		So(S(val), ShouldEqual, "-")

		val, err = repl("  456  ", env)
		So(err, ShouldBeNil)
		So(val, ShouldEqual, 456)

		val, err = repl(".123 ", env)
		So(err, ShouldBeNil)
		So(val, ShouldEqual, 0.123)

		val, err = repl("'(a . b)", env)
		So(err, ShouldBeNil)
		So(S(val), ShouldEqual, "(A . B)")

		val, err = repl(`"bad string`, env)
		So(err.Error(), ShouldContainSubstring, "unterminated string")

		val, err = repl(`"string with \" quote`, env)
		So(err.Error(), ShouldContainSubstring, "unterminated string")

		val, err = repl(`"string with \"`, env)
		So(err.Error(), ShouldContainSubstring, "unterminated string")

		val, err = repl(`"string with \" quote"`, env)
		So(err, ShouldBeNil)
		So(val, ShouldEqual, `string with " quote`)

	})

	Convey("testing", t, func() {
		env := DefaultEnv()
		Convey("Use unset variable n", func() {
			val, err := repl("(+ n n)", env)
			So(err.Error(), ShouldContainSubstring, "Undefined symbol: N")
			So(val, ShouldBeNil)
		})
		Convey("Calculate radius from undefined variable", func() {
			_, err := repl("(define radius (* pi (* r r)))", env)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Undefined symbol: R")
			_, err = repl("radius", env)
			So(err.Error(), ShouldContainSubstring, "Undefined symbol: RADIUS")
		})
		Convey("Test Integer literals", func() {
			val, err := repl("123", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 123)

			val, err = repl("-123", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, -123)
		})

		Convey("Test predicates", func() {
			Convey("null?", func() {
				val, err := repl("(null? '())", env)
				So(err, ShouldBeNil)
				So(val, ShouldEqual, False)

				val, err = repl("(null? 123)", env)
				So(err, ShouldBeNil)
				So(val, ShouldEqual, False)
			})
			Convey("equal?", func() {
				val, err := repl("(equal? 1 1)", env)
				So(err, ShouldBeNil)
				So(val, ShouldEqual, T)

				val, err = repl("(equal? 1 30)", env)
				So(err, ShouldBeNil)
				So(val, ShouldEqual, False)

			})
			Convey("pair?", func() {
				val, err := repl("(pair? '(a b))", env)
				So(err, ShouldBeNil)
				So(val, ShouldEqual, T)

				val, err = repl("(pair? 30)", env)
				So(err, ShouldBeNil)
				So(val, ShouldEqual, False)

			})
		})

		Convey("Test boolean literals", func() {
			val, err := repl("'#t", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, T)
			So(S(val), ShouldEqual, "#t")

			val, err = repl("'#f", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, False)
			So(S(val), ShouldEqual, "#f")
		})
		Convey("If statements", func() {
			val, err := repl("(if (<= 4 2) (* 10 2))", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, Empty)

			val, err = repl("(if (< 4 2) (* 10 2) (+ 1 2))", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 3)

			val, err = repl("(if (> 4 2) (* 10 2) (+ 1 2))", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 20)

			val, err = repl("(if '(1 2) (* 10 2) (+ 1 2))", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 20)

			val, err = repl("(if 'atom 'a 'b)", env)
			So(err, ShouldBeNil)
			So(S(val), ShouldResemble, "A")

			val, err = repl("(if '() 'a 'b)", env)
			So(err, ShouldBeNil)
			So(S(val), ShouldEqual, ("A"))

			val, err = repl("(if 0 'a 'b)", env)
			So(err, ShouldBeNil)
			So(S(val), ShouldEqual, "B")

			val, err = repl("(if \"\" 'a 'b)", env)
			So(err, ShouldBeNil)
			So(S(val), ShouldEqual, "A")

		})

		Convey("Test statements", func() {
			val, err := repl("(quote (1  1))", env)
			So(err, ShouldBeNil)
			So(S(val), ShouldEqual, "(1 1)")

			val, err = repl("'(1  1)", env)
			So(err, ShouldBeNil)
			So(S(val), ShouldEqual, "(1 1)")

			_, err = repl("(lambda () (+ 2 3) (+ 1 1))", env)
			So(err, ShouldBeNil)

			val, err = repl("((lambda () (+ 1 1)))", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 2)

			_, err = repl("(define foo ((lambda (x) (lambda () (set! x (+ x 1)) x)) 0))", env)
			So(err, ShouldBeNil)

			val, err = repl("(foo) (foo)", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 2)

			val, err = repl("(foo) (foo)", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 4)

			_, err = repl("(define plus (lambda (a b) (+ a b)))", env)
			So(err, ShouldBeNil)

			val, err = repl("(plus 10 -2)", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 8)

			val, err = repl("(plus 10)", env)
			So(err.Error(), ShouldContainSubstring, "parameter mismatch")

			val, err = repl("(not-func 10)", env)
			So(err.Error(), ShouldContainSubstring, "Undefined symbol: NOT-FUNC")

			Convey("begin", func() {
				val, err = repl("(begin (+ 2 3) (* 5 8))", env)
				So(err, ShouldBeNil)
				So(val, ShouldEqual, 40)

			})

		})
		Convey("Test arithmetic", func() {
			var a float64 = 30
			var b float64 = 40
			val, err := repl(fmt.Sprintf("(* %v %v)", a, b), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, a*b)
			val, err = repl(fmt.Sprintf("(/ %v %v)", a, b), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, a/b)
			val, err = repl(fmt.Sprintf("(+ %v %v)", a, b), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, a+b)
			val, err = repl(fmt.Sprintf("(- %v %v)", a, b), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, a-b)

			val, err = repl(fmt.Sprintf("(- 'atom %v)", a), env)
			So(err.Error(), ShouldContainSubstring, "Not a number: ATOM")
			val, err = repl(fmt.Sprintf("(- %v 'atom)", a), env)
			So(err.Error(), ShouldContainSubstring, "Not a number: ATOM")
		})
		Convey("Test numeric compare", func() {
			var a float64 = -10.1
			var b float64 = 40.123
			var c float64 = 70
			val, err := repl(fmt.Sprintf("(< %v 'atom)", a), env)
			So(err.Error(), ShouldContainSubstring, "Not a number: ATOM")
			val, err = repl(fmt.Sprintf("(< 'atom %v)", a), env)
			So(err.Error(), ShouldContainSubstring, "Not a number: ATOM")
			val, err = repl(fmt.Sprintf("(< %v %v)", a, b), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, a < b)
			val, err = repl(fmt.Sprintf("(> %v %v)", a, b), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, a > b)
			val, err = repl(fmt.Sprintf("(< %v %v %v)", a, b, c), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, a < b && b < c)
			val, err = repl(fmt.Sprintf("(> %v %v)", b, a), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, b > a)
			val, err = repl(fmt.Sprintf("(< %v %v)", b, a), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, b < a)
			val, err = repl(fmt.Sprintf("(<= %v %v)", a, a), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, a <= a)
			val, err = repl(fmt.Sprintf("(<= %v %v %v)", a, b, b), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, a <= b)
			val, err = repl(fmt.Sprintf("(<= %v %v %v)", b, b, a), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, b <= a)
			val, err = repl(fmt.Sprintf("(>= %v %v %v)", b, b, a), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, b >= a)
			val, err = repl(fmt.Sprintf("(>= %v %v %v)", a, b, b), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, a >= b)
			val, err = repl(fmt.Sprintf("(= %v %v %v)", a, a, a), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, a == a)
			val, err = repl(fmt.Sprintf("(= %v %v %v)", a, b, b), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, a == b)
			val, err = repl(fmt.Sprintf("(number? %v)", a), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, T)
			val, err = repl(fmt.Sprintf("(number? '(%v))", a), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, False)
			val, err = repl(fmt.Sprintf("(number? \"%v\")", a), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, False)
			val, err = repl(fmt.Sprintf("(number? 'atom)"), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, False)
		})
		Convey("Define and use variables", func() {
			Convey("Define r and n", func() {
				val, err := repl("(define r 10)\n(define n 12)", env)
				So(err, ShouldBeNil)
				So(val, ShouldEqual, "OK")
				Convey("Testr r and n", func() {
					val, err := repl("(+ r n)", env)
					So(err, ShouldBeNil)
					So(val, ShouldEqual, 22)
				})
				Convey("Calculate radius", func() {
					val, err := repl("(define radius (* pi (* r r)))", env)
					So(err, ShouldBeNil)
					So(val, ShouldEqual, "OK")
					val, err = repl("radius", env)
					So(err, ShouldBeNil)
					So(val, ShouldEqual, math.Pi*10*10)
					val, err = repl("(/ radius 10)", env)
					So(err, ShouldBeNil)
					So(val, ShouldEqual, math.Pi*10)
				})
			})
			Convey("Procedure definition", func() {
				val, err := repl("(define (double a)  (+ a a))", env)
				So(err, ShouldBeNil)
				So(val, ShouldEqual, "OK")
				val, err = repl("(double 22)", env)
				So(err, ShouldBeNil)
				So(val, ShouldEqual, 44)
			})

		})
		Convey("Test strings", func() {
			val, err := repl("\"asdf\"", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, "asdf")

		})
		Convey("Test cons", func() {
			val, err := repl("(cons 1 2)", env)
			So(err, ShouldBeNil)
			So(S(val), ShouldEqual, "(1 . 2)")

			val, err = repl("(cdr (cons 1 2))", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 2)

			val, err = repl("(cons 1 (cons 2 3))", env)
			So(err, ShouldBeNil)
			So(S(val), ShouldEqual, "(1 2 . 3)")

			val, err = repl("(cons 1 '(2 3))", env)
			So(err, ShouldBeNil)
			So(S(val), ShouldEqual, "(1 2 3)")

			val, err = repl("(car (cons 1 '(2 3)))", env)
			So(err, ShouldBeNil)
			So(S(val), ShouldEqual, "1")

			val, err = repl("(cdr (cons 1 '(2 3)))", env)
			So(err, ShouldBeNil)
			So(S(val), ShouldEqual, "(2 3)")

			val, err = repl("(cdr (cdr (cons 1 '(2 3))))", env)
			So(err, ShouldBeNil)
			So(S(val), ShouldEqual, "(3)")

			val, err = repl("(cdr (cdr (cdr (cons 1 '(2 3)))))", env)
			So(err, ShouldBeNil)
			So(val, ShouldResemble, Empty)
		})

		factorial := []TestCase{
			{"(define fact (lambda (n) (if (<= n 1) 1 (* n (fact (- n 1))))))", "OK", ""},
			{"(fact 10)", 3.6288e+06, ""},
		}
		doCases("factorial", factorial)

		letCases := []TestCase{
			{"(let ((a 1) (b 2)) (+ a b))", Number(3), ""},
			{"(let ((a 3) (b 2)) (* a b))", Number(6), ""},
			{"(let () (* a b))", nil, "Undefined symbol: A"},
			{"(let (()) (* a b))", nil, "(): value is not a pair"},
			{"(let ((a 1) ()) (* a b))", nil, "(): value is not a pair"},
			{"(let ((a 1)) )", nil, "(): value is not a pair"},
		}
		doCases("Test let statements", letCases)

	})
}

func TestEndofFile(t *testing.T) {
	Convey("fail when given imbalanced parens", t, func() {
		imbalanced := []TestCase{
			{"(", nil, "End of File"},
			{"(define counter (lambda (n) (lambda () (set! n (+ n 1))))", nil, "End of File"},
			{"(let ((a 1 ) (+ a a)))", nil, "Undefined symbol: A"},
		}
		doCases("fail when given imbalanced parens", imbalanced)
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
		So(err.Error(), ShouldContainSubstring, expectedErr)
	}
	So(val, ShouldEqual, value)
}

func doCases(name string, cases []TestCase) {
	Convey(name, func() {
		env := DefaultEnv()
		for _, c := range cases {
			doRepl(env, c.expr, c.value, c.expectedErr)
		}
	})
}
