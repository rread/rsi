package main

import (
	"flag"
	"io/ioutil"

	"github.com/rread/unlisp/log"
)

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
