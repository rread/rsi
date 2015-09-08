package main

import (
	"fmt"

	"github.com/rread/rsi/log"
)

type Pair struct {
	car Data
	cdr Data
}

func (p *Pair) String() string {
	ret := "("
	for {
		ret += fmt.Sprintf("%v", p.car)
		if nullp(p.cdr) {
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

func getList(d Data) (Data, error) {
	if nullp(d) {
		return Empty, nil
	}
	switch v := d.(type) {
	case *Pair:
		return v, nil
	case error:
		return nil, v
	default:
		return nil, fmt.Errorf("%v: value is not a pair", d)

	}
}

func getPair(d Data) (*Pair, error) {
	if nullp(d) {
		return nil, fmt.Errorf("%v: value is not a pair", d)
	}
	switch v := d.(type) {
	case *Pair:
		return v, nil
	case error:
		return nil, v
	default:
		return nil, fmt.Errorf("%v: value is not a pair", d)

	}
}

func cons(car, cdr Data) Data {
	if getError(car) != nil {
		return car
	}
	if getError(cdr) != nil {
		return cdr
	}
	return &Pair{car, cdr}
}

func cdr(d Data) Data {
	l, err := getPair(d)
	if err != nil {
		return err
	}

	return l.cdr
}

func car(d Data) Data {
	l, err := getPair(d)
	if err != nil {
		return err
	}
	return l.car
}

func cadr(d Data) Data {
	return car(cdr(d))
}

func cddr(d Data) Data {
	return cdr(cdr(d))
}

func caddr(l Data) Data {
	return car(cdr(cdr(l)))
}

func cadddr(l Data) Data {
	return car(cdr(cdr(cdr(l))))
}

func pairp(d Data) Boolean {
	_, ok := d.(*Pair)
	if !ok {
		log.Printf("pairp %T %v %v", d, d, ok)
	}
	return Boolean(ok)
}

func listNext(d Data) (*Pair, error) {
	l, err := getPair(d)
	if err != nil {
		return nil, err
	}
	d = cdr(l)
	return getPair(d)
}

func listLen(d Data) int {
	if nullp(d) {
		return 0
	}
	var i int
	for {
		i++
		d = cdr(d)
		if nullp(d) {
			break
		}
		if !pairp(d) {
			log.Fatal("expecting a list")
		}
	}
	return i
}

func reverse(d Data) Data {
	var li Data = Empty
	for {
		if nullp(d) {
			break
		}
		li = cons(car(d), li)
		d = cdr(d)
	}
	return li
}
