package main

import (
	"io"
)

//CodeWriter is an interface of an object that can print its code
type CodeWriter interface {
	FuncDec() string
	Header(w io.Writer)
	ParamScan(w io.Writer)
	FunctionCall(w io.Writer)
	Return(w io.Writer)
	Footer(w io.Writer)
}

//Param the parameters of the functions
type Param struct {
	Name, Type string
}

//VoidFunction is an function with no return type
type VoidFunction struct {
	Name   string
	Params []Param
}

//FuncDec returns the PG INFO_V1 macro
func (f *VoidFunction) FuncDec() string {
	return "PG_FUNCTION_INFO_V1(" + f.Name + ");"
}

//Header writes the header ofthe wrapper function
func (f *VoidFunction) Header(w io.Writer) {
	w.Write([]byte("//export " + f.Name + "\nfunc " + f.Name + "(fcinfo *funcInfo) Datum {\n"))
}

//ParamScan writes the scan parameters part of the wrapper function
func (f *VoidFunction) ParamScan(w io.Writer) {
	if len(f.Params) == 0 {
		return
	}
	for _, p := range f.Params {
		w.Write([]byte("var " + p.Name + " " + p.Type + "\n"))
	}
	w.Write([]byte("fcinfo.Scan(\n"))
	for _, p := range f.Params {
		w.Write([]byte("&" + p.Name + ",\n"))
	}
	w.Write([]byte(")\n"))
}

//FunctionCall writes the call of the wrapped function
func (f *VoidFunction) FunctionCall(w io.Writer) {
	w.Write([]byte(ToUnexported(f.Name) + "(\n"))
	for _, p := range f.Params {
		w.Write([]byte(p.Name + ",\n"))
	}
	w.Write([]byte(")\n"))
}

//Return writes the return statement
func (f *VoidFunction) Return(w io.Writer) {
	w.Write([]byte("return toDatum(nil)\n"))
}

//Footer writes the return statement
func (f *VoidFunction) Footer(w io.Writer) {
	w.Write([]byte("}\n"))
}

//Function is a list of parameters and the return type
type Function struct {
	VoidFunction
	ReturnType string
}

//FunctionCall writes the call of the wrapped function
func (f *Function) FunctionCall(w io.Writer) {
	w.Write([]byte("ret := "))
	f.VoidFunction.FunctionCall(w)
}

//Return writes the return statement
func (f *Function) Return(w io.Writer) {
	w.Write([]byte("return toDatum(ret)\n"))
}

//TriggerFunction a special type of function, it takes TriggerData as the first argument and TriggerRow as return type
type TriggerFunction struct {
	VoidFunction
}

//FunctionCall writes the call of the wrapped function
func (f *TriggerFunction) FunctionCall(w io.Writer) {
	w.Write([]byte("ret := "))
	w.Write([]byte(ToUnexported(f.Name) + "(\nfcinfo.TriggerData(),\n"))
	for _, p := range f.Params {
		w.Write([]byte(p.Name + ",\n"))
	}
	w.Write([]byte(")\n"))
}

//Return writes the return statement
func (f *TriggerFunction) Return(w io.Writer) {
	w.Write([]byte("return toDatum(ret)\n"))
}
