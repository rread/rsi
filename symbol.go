package main

import "strings"

type DataType int

const (
	ConsType = DataType(iota)
	SymbolType
	NumberType
	StringType
	FuncType
	NativeType
)

func StringWithValue(v string) String {
	return String(v)
}

func SymbolWithName(n string) Symbol {
	return Symbol(strings.ToUpper(n))
}

func Symbolp(i Data) bool {
	_, ok := i.(Symbol)
	return ok
}

func Stringp(i Data) bool {
	_, ok := i.(String)
	return ok
}

var internedSymbols = make(map[string]Symbol, 1024)

func internSymbol(name string) Symbol {
	// Force uppercase symbol names
	name = strings.ToUpper(name)
	sym, ok := internedSymbols[name]
	if !ok || sym == "" {
		sym = SymbolWithName(name)
		internedSymbols[name] = sym
	}
	return sym
}
