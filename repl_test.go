package main

import (
	"fmt"
	"math"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRepl(t *testing.T) {
	Convey("repl basic testing", t, func() {
		Convey("Bunch of repl tests", func() {
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
			Convey("rest of stuff", func() {
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
			})
		})
		Convey("Test arithmetic", func() {
			env := DefaultEnv()
			var a float64 = 30
			var b float64 = 40
			val, err := repl(fmt.Sprintf("(* %v %v)", a, b), env)
			So(err, ShouldEqual, ErrorEOF)
			So(val, ShouldEqual, a*b)
			val, err = repl(fmt.Sprintf("(/ %v %v)", a, b), env)
			So(err, ShouldEqual, ErrorEOF)
			So(val, ShouldEqual, a/b)
			val, err = repl(fmt.Sprintf("(+ %v %v)", a, b), env)
			So(err, ShouldEqual, ErrorEOF)
			So(val, ShouldEqual, a+b)
			val, err = repl(fmt.Sprintf("(- %v %v)", a, b), env)
			So(err, ShouldEqual, ErrorEOF)
			So(val, ShouldEqual, a-b)
		})
		Convey("Define and use variables", func() {
			env := DefaultEnv()
			Convey("Define r and n", func() {
				val, err := repl("(define r 10)\n(define n 12)", env)
				So(err, ShouldEqual, ErrorEOF)
				So(val, ShouldEqual, 12)
				Convey("Testr r and n", func() {
					val, err := repl("(+ r n)", env)
					So(err, ShouldEqual, ErrorEOF)
					So(val, ShouldEqual, 22)
				})
				Convey("Calculate radius", func() {
					val, err := repl("(define radius (* pi (* r r)))", env)
					So(err, ShouldEqual, ErrorEOF)
					So(val, ShouldEqual, math.Pi*10*10)
					val, err = repl("(/ radius 10)", env)
					So(err, ShouldEqual, ErrorEOF)
					So(val, ShouldEqual, math.Pi*10)
				})
			})

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
