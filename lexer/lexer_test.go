package lexer

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type Case struct {
	value    string
	expected string
	err      string
}

func TestLexer(t *testing.T) {
	Convey("basic lexer testing", t, func() {
		for _, c := range []Case{
			{`"`, `ILLEGAL "unterminated string: ''"`, ""},
			{"", `EOF ""`, ""},
			{"\n  ", `EOF ""`, ""},
			{"; asdf", `COMMENT "; asdf"`, ""},
			{"; asdf\n", `COMMENT "; asdf\n"`, ""},
			{"(", `LEFT_PAREN "("`, ""},
			{")", `RIGHT_PAREN ")"`, ""},
			{" . ", `DOT "."`, ""},
			{"'", `QUOTE "'"`, ""},
			{"-", `SYMBOL "-"`, ""},
			{"-1", `NUMBER "-1"`, ""},
			{".", `NUMBER "."`, ""},
			{"123", `NUMBER "123"`, ""},
			{"-abc", `SYMBOL "-abc"`, ""},
			{"abc", `SYMBOL "abc"`, ""},
			{`"abc"`, `STRING "abc"`, ""},
			{`"abc\`, `ILLEGAL "unterminated string: \"abc\""`, ""},
			{"#t", `TRUE "t"`, ""},
			{"#f", `FALSE "f"`, ""},
			{"#n", `ILLEGAL "unsupported hash code #n"`, ""},
			{"a(", `SYMBOL "a"`, ""},
			{"12(", `NUMBER "12"`, ""},
		} {
			doLex(c.value, c.expected, c.err)
		}

		Convey("Unknown tokeen", func() {
			t := Token(500)
			So(t.String(), ShouldEqual, "Unknown token: 500")
		})

	})
}

func doLex(expr, expectedValue, expectedError string) {
	l := New("test", expr)
	ti := l.NextItem()
	So(ti.String(), ShouldEqual, expectedValue)

}
