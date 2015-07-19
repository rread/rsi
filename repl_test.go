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

func TestRepl(t *testing.T) {

	Convey("basic lexer testing", t, func() {
		env := EmptyEnv()
		val, err := repl("", env)
		So(err, ShouldBeNil)
		So(val, ShouldBeNil)

		val, err = repl(" \t\r\n", env)
		So(err, ShouldBeNil)
		So(val, ShouldBeNil)

		val, err = repl("()", env)
		So(err, ShouldBeNil)
		So(val, ShouldEqual, Nil)

		val, err = repl("123", env)
		So(err, ShouldBeNil)
		So(val, ShouldEqual, 123)

		val, err = repl("\"123\"", env)
		So(err, ShouldBeNil)
		So(val, ShouldEqual, "123")

		val, err = repl("'-", env)
		So(err, ShouldBeNil)
		So(S(val), ShouldResemble, "-")

		val, err = repl("  456  ", env)
		So(err, ShouldBeNil)
		So(val, ShouldEqual, 456)

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

		Convey("If statements", func() {
			val, err := repl("(if (<= 4 2) (* 10 2))", env)
			So(err, ShouldBeNil)
			So(S(val), ShouldEqual, "NIL")

			val, err = repl("(if (< 4 2) (* 10 2) (+ 1 2))", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 3)

			val, err = repl("(if (> 4 2) (* 10 2) (+ 1 2))", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 20)

			val, err = repl("(if '(1 2 ) (* 10 2) (+ 1 2))", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 20)

			val, err = repl("(if 'atom 'a 'b)", env)
			So(err, ShouldBeNil)
			So(S(val), ShouldResemble, "A")

			val, err = repl("(if '() 'a 'b)", env)
			So(err, ShouldBeNil)
			So(S(val), ShouldEqual, ("B"))

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

			_, err = repl("(lambda () (+ 1 1))", env)
			So(err, ShouldBeNil)

			val, err = repl("((lambda () (+ 1 1)))", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 2)

			_, err = repl("(define  foo (begin (define count 0) (lambda () (set! count (+ count 1)))))", env)
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
			So(err.Error(), ShouldContainSubstring, "undefined-function: Undefined symbol: NOT-FUNC")

			val, err = repl("(equal? 1 1)", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, true)
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
			So(val, ShouldEqual, true)
			val, err = repl(fmt.Sprintf("(number? '(%v))", a), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, false)
			val, err = repl(fmt.Sprintf("(number? \"%v\")", a), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, false)
			val, err = repl(fmt.Sprintf("(number? 'atom)"), env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, false)
		})
		Convey("Define and use variables", func() {
			Convey("Define r and n", func() {
				val, err := repl("(define r 10)\n(define n 12)", env)
				So(err, ShouldBeNil)
				So(val, ShouldEqual, 12)
				Convey("Testr r and n", func() {
					val, err := repl("(+ r n)", env)
					So(err, ShouldBeNil)
					So(val, ShouldEqual, 22)
				})
				Convey("Calculate radius", func() {
					val, err := repl("(define radius (* pi (* r r)))", env)
					So(err, ShouldBeNil)
					So(val, ShouldEqual, math.Pi*10*10)
					val, err = repl("(/ radius 10)", env)
					So(err, ShouldBeNil)
					So(val, ShouldEqual, math.Pi*10)
				})
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
			So(S(val), ShouldEqual, "(1 2)")

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
			So(S(val), ShouldEqual, "NIL")

		})
	})
}

func TestEndofFile(t *testing.T) {
	env := DefaultEnv()
	Convey("fail when given imbalanced parens", t, func() {
		_, err := repl("(define counter (lambda (n) (lambda () (set! n (+ n 1))))", env)
		if err == nil {
			t.Fail()
		}
		So(err.Error(), ShouldContainSubstring, "End of File")
	})
}
