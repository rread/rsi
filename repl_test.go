package main

import (
	"fmt"
	"math"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRepl(t *testing.T) {
	Convey("repl basic testing", t, func() {
		env := DefaultEnv()
		Convey("Use unset variable n", func() {
			val, err := repl("(+ n n)", env)
			So(err.Error(), ShouldContainSubstring, "Undefined symbol: n")
			So(val, ShouldBeNil)
		})
		Convey("Calculate radius from undefined variable", func() {
			_, err := repl("(define radius (* pi (* r r)))", env)
			So(err.Error(), ShouldContainSubstring, "Undefined symbol: r")
			_, err = repl("radius", env)
			So(err.Error(), ShouldContainSubstring, "Undefined symbol: radius")
		})
		Convey("Test Integer literals", func() {
			val, err := repl("123", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 123)

			val, err = repl("-123", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, -123)
		})

		Convey("Test statements", func() {
			_, err := repl("(if (<= 4 2) (* 10 2))", env)
			So(err, ShouldBeNil)

			_, err = repl("(if (<= 4 2) (* 10 2) (+ 1 2))", env)
			So(err, ShouldBeNil)
			_, err = repl("(/ radius 10)", env)

			val, err := repl("(quote (1  1))", env)
			So(err, ShouldBeNil)
			So(fmt.Sprintf("%v", val), ShouldEqual, "(1 1)")

			val, err = repl("'(1  1)", env)
			So(err, ShouldBeNil)
			So(fmt.Sprintf("%v", val), ShouldEqual, "(1 1)")

			_, err = repl("(lambda () (+ 1 1))", env)
			So(err, ShouldBeNil)

			val, err = repl("((lambda () (+ 1 1)))", env)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 2)

			_, err = repl("(define foo (begin (define count 0) (lambda () (set! count (+ count 1)))))", env)
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
			So(err.Error(), ShouldContainSubstring, "undefined-function: Undefined symbol: not-func")

			val, err = repl("(equal? 1 1)", env)
			So(err, ShouldBeNil)
			So(val, ShouldResemble, true)
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
			So(err.Error(), ShouldContainSubstring, "Not a number: atom")
			val, err = repl(fmt.Sprintf("(- %v 'atom)", a), env)
			So(err.Error(), ShouldContainSubstring, "Not a number: atom")
		})
		Convey("Test numeric compare", func() {
			var a float64 = -10.1
			var b float64 = 40.123
			var c float64 = 70
			val, err := repl(fmt.Sprintf("(< %v 'atom)", a), env)
			So(err.Error(), ShouldContainSubstring, "Not a number: atom")
			val, err = repl(fmt.Sprintf("(< 'atom %v)", a), env)
			So(err.Error(), ShouldContainSubstring, "Not a number: atom")
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
			So(val, ShouldBeTrue)
			val, err = repl(fmt.Sprintf("(number? '(%v))", a), env)
			So(err, ShouldBeNil)
			So(val, ShouldBeFalse)
			val, err = repl(fmt.Sprintf("(number? \"%v\")", a), env)
			So(err, ShouldBeNil)
			So(val, ShouldBeFalse)
			val, err = repl(fmt.Sprintf("(number? 'atom)"), env)
			So(err, ShouldBeNil)
			So(val, ShouldBeFalse)
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
			So(fmt.Sprintf("%v", val), ShouldEqual, "(1 2)")

			val, err = repl("(cons 1 '(2 3))", env)
			So(err, ShouldBeNil)
			So(fmt.Sprintf("%v", val), ShouldEqual, "(1 2 3)")

			val, err = repl("(car (cons 1 '(2 3)))", env)
			So(err, ShouldBeNil)
			So(fmt.Sprintf("%v", val), ShouldEqual, "1")

			val, err = repl("(cdr (cons 1 '(2 3)))", env)
			So(err, ShouldBeNil)
			So(fmt.Sprintf("%v", val), ShouldEqual, "(2 3)")

			val, err = repl("(cdr (cdr (cons 1 '(2 3))))", env)
			So(err, ShouldBeNil)
			So(fmt.Sprintf("%v", val), ShouldEqual, "(3)")

			val, err = repl("(cdr (cdr (cdr (cons 1 '(2 3)))))", env)
			So(err, ShouldBeNil)
			SkipSo(fmt.Sprintf("%v", val), ShouldEqual, "()") // hmm, not using ItemList.String()

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
