package main

import "fmt"

type Binding map[*Data]Item
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

func (e *Env) BindName(name string, i Item) {
	sym := internSymbol(name)
	e.vars[sym] = i
}

func Symbolp(sym *Data) bool {
	return sym.Type == SymbolType
}

func (e *Env) Bind(sym *Data, i Item) {
	if !Symbolp(sym) {
		panic(fmt.Errorf("sym is not a Symbol: %v", sym))
	}
	e.vars[sym] = i
}

func (e *Env) Var(sym *Data) (Item, error) {
	if !Symbolp(sym) {
		panic(fmt.Errorf("sym is not a Symbol: %v", sym))
	}
	v, ok := e.vars[sym]
	if !ok {
		return nil, fmt.Errorf("Undefined symbol: %v", sym)
	}
	return v, nil
}

func (env *Env) Find(sym *Data) *Env {
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

func (env *Env) FindVar(sym *Data) (Item, error) {
	return env.Find(sym).Var(sym)
}
