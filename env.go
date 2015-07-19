package main

import "fmt"

type Binding map[Symbol]Data
type Env struct {
	vars  Binding
	outer *Env
}

func NewEnv(outer *Env) *Env {
	return &Env{
		vars:  make(Binding),
		outer: outer,
	}

}

func (e *Env) BindName(name string, i Data) {
	sym := internSymbol(name)
	e.vars[sym] = i
}

func (e *Env) Bind(sym Symbol, i Data) {
	if !Symbolp(sym) {
		panic(fmt.Errorf("sym is not a Symbol: %v", sym))
	}
	e.vars[sym] = i
}

func (e *Env) Var(sym Symbol) (Data, error) {
	if !Symbolp(sym) {
		panic(fmt.Errorf("sym is not a Symbol: %v", sym))
	}
	v, ok := e.vars[sym]
	if !ok {
		return nil, fmt.Errorf("Undefined symbol: %v", sym)
	}
	return v, nil
}

func (env *Env) Find(sym Symbol) *Env {
	if !Symbolp(sym) {
		panic(fmt.Errorf("sym is not a Symbol: %v", sym))
	}
	if _, ok := env.vars[sym]; ok {
		return env
	} else if env.outer != nil {
		return env.outer.Find(sym)
	} else {
		return env
	}
}

func (env *Env) FindVar(sym Symbol) (Data, error) {
	return env.Find(sym).Var(sym)
}
