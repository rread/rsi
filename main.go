package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"unicode"

	"github.com/bobappleyard/readline"
	"github.com/rread/rsi/log"
)

func validSexp(s string) bool {
	var parens int
	var notEmpty bool
	for _, c := range s {
		switch c {
		case '(':
			parens++
		case ')':
			parens--
		}
		if notEmpty || !unicode.IsSpace(c) {
			notEmpty = true
		}
	}
	return notEmpty && parens == 0
}

func replCLI(env *Env) {
	defer fmt.Println("\nbye!")
	counter := readline.HistorySize()
	for {
		buf := bytes.Buffer{}
		prompt := fmt.Sprintf("[%d]-> ", counter)
		for {
			l, err := readline.String(prompt)
			if err == io.EOF {
				return
			}
			buf.WriteString(l)
			if validSexp(buf.String()) {
				break
			}
			buf.WriteString("\n")
			prompt = ": "
		}
		result, err := repl(buf.String(), env)
		if err != nil && err != ErrorEOF {
			fmt.Println("Error:", err)
		} else {
			fmt.Println(result)
		}
		readline.AddHistory(buf.String())
		counter++

	}
}

func main() {
	debug := flag.Bool("debug", false, "Enable debugging")
	flag.Parse()
	flag.Args()
	if *debug {
		log.SetLevel(log.Debug)
	}
	env := DefaultEnv()

	if flag.NArg() > 0 {
		for _, f := range flag.Args() {
			buf, err := ioutil.ReadFile(f)
			if err != nil {
				log.Fatal(err)
			}
			repl(string(buf), env)
		}
	}
	replCLI(env)
}
