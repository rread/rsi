package main

import (
	"flag"
	"io/ioutil"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()
	flag.Args()
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
