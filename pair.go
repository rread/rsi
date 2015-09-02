package main

import (
	"fmt"

	"github.com/rread/unlisp/log"
)

type Pair struct {
	car Data
	cdr Data
}

func (p *Pair) String() string {
	ret := "("
	for {
		ret += fmt.Sprintf("%v", p.car)
		if p.cdr == Nil {
			break
		}
		if pairp(p.cdr) {
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

func getPair(d Data) (*Pair, error) {
	if d == Nil {
		return Nil, nil
	}
	if p, ok := d.(*Pair); ok {
		return p, nil
	}
	return Nil, fmt.Errorf("%v: data is not a pair", d)
}

func cons(car, cdr Data) *Pair {
	return &Pair{car, cdr}
}

func car(l *Pair) Data {
	if l == nil {
		return Nil
	}
	return l.car
}

func cdr(l *Pair) Data {
	if l == nil {
		return Nil
	}
	return l.cdr
}

func cadr(l *Pair) Data {
	d := cdr(l)
	if p, ok := d.(*Pair); ok {
		return car(p)
	}
	log.Fatalf("pair expected: %v", l)
	return Nil
}

func caddr(l *Pair) Data {
	d := cdr(l)
	if p, ok := d.(*Pair); ok {
		d := cdr(p)
		if p, ok := d.(*Pair); ok {
			return car(p)
		}
	}

	log.Fatalf("pair expected: %v", l)
	return Nil
}

func cadddr(l *Pair) Data {
	d := cdr(l)
	if p, ok := d.(*Pair); ok {
		d := cdr(p)
		if p, ok := d.(*Pair); ok {
			d := cdr(p)
			if p, ok := d.(*Pair); ok {
				return car(p)
			}
		}
	}
	log.Fatal("pair expected")
	return Nil
}

func nullp(p *Pair) Boolean {
	return p == Nil
}

func pairp(d Data) Boolean {
	_, ok := d.(*Pair)
	if !ok {
		log.Printf("consp %T %v %v", d, d, ok)
	}
	return Boolean(ok)
}

func listNext(l *Pair) (*Pair, error) {
	d := cdr(l)
	return getPair(d)
}

func listLen(p *Pair) int {
	var i int
	if p == Nil {
		return i
	}
	for {
		i++
		d := cdr(p)
		if d == Nil {
			break
		}

		var ok bool
		if p, ok = d.(*Pair); !ok {
			log.Fatal("expecting a list")
		}
	}
	return i
}
