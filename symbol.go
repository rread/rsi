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

type Data struct {
	Type  DataType
	Value interface{}
}

func (d *Data) String() string {
	switch d.Type {
	case SymbolType:
		return d.StringValue()
	}
	return ""
}

func (d *Data) StringValue() string {
	if d == nil {
		return ""
	}
	return d.Value.(string)
}

func SymbolWithName(n string) *Data {
	return &Data{
		Type:  SymbolType,
		Value: n,
	}
}

var internedSymbols = make(map[string]*Data, 1024)

func internSymbol(name string) *Data {
	// Force uppercase symbol names
	name = strings.ToUpper(name)
	sym, ok := internedSymbols[name]
	if !ok || sym == nil {
		sym = SymbolWithName(name)
		internedSymbols[name] = sym
	}
	return sym
}
