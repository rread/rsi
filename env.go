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

func ExtendEnv(names []Symbol, values Data, outer *Env) (*Env, error) {
	env := NewEnv(outer)
	if len(names) != listLen(values) {
		return nil, fmt.Errorf("parameter mismatch %v != %v", names, values)
	}
	for _, name := range names {

		val := car(values)

		if err := getError(val); err != nil {
			return nil, err
		}
		env.Bind(name, val)

		values = cdr(values)
		if err := getError(val); err != nil {
			return nil, err
		}

	}
	return env, nil
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
