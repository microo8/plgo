package main

import (
	"fmt"
	"go/ast"
	"io"
)

const (
	triggerData = "TriggerData"
	triggerRow  = "TriggerRow"
)

//CodeWriter is an interface of an object that can print its code
type CodeWriter interface {
	FuncDec() string
	Code(w io.Writer)
}

//NewCode parses the ast.FuncDecl and returns a new Function or An TriggerFunction
func NewCode(function *ast.FuncDecl) (CodeWriter, error) {
	params, err := getParamList(function)
	if err != nil {
		return nil, err
	}
	returnType, err := getReturnType(function.Name.Name, function.Type.Results)
	if err != nil {
		return nil, err
	}

	if returnType == triggerRow {
		if len(params) == 0 || params[0].Type != triggerData {
			return nil, fmt.Errorf("Function %s can return *plgo.TriggerRow when the first parameter will be *plgo.TriggerData", function.Name.Name)
		}
		return &TriggerFunction{VoidFunction: VoidFunction{Name: function.Name.Name, Params: params[1:], Doc: function.Doc.Text()}}, nil
	}
	if returnType == "" {
		return &VoidFunction{Name: function.Name.Name, Params: params, Doc: function.Doc.Text()}, nil
	}
	return &Function{VoidFunction: VoidFunction{Name: function.Name.Name, Params: params, Doc: function.Doc.Text()}, ReturnType: returnType}, nil
}

func getParamList(function *ast.FuncDecl) (Params []Param, err error) {
	for i, param := range function.Type.Params.List {
		//TODO array types
		switch p := param.Type.(type) {
		case *ast.Ident:
			for _, name := range param.Names {
				if _, ok := datumTypes[p.Name]; !ok {
					return nil, fmt.Errorf("Function %s, parameter position %d: type %s not supported", function.Name.Name, i, p.Name)
				}
				Params = append(Params, Param{Name: name.Name, Type: p.Name})
			}
		case *ast.StarExpr:
			selector, ok := p.X.(*ast.SelectorExpr)
			if !ok {
				return nil, fmt.Errorf("Function %s, parameter position %d: type not supported", function.Name.Name, i)
			}
			var pkg *ast.Ident
			pkg, ok = selector.X.(*ast.Ident)
			if !ok {
				return nil, fmt.Errorf("Function %s, parameter position %d: type not supported", function.Name.Name, i)
			}
			if pkg.Name != plgo || selector.Sel.Name != triggerData {
				return nil, fmt.Errorf("Function %s, parameter position %d: type not supported", function.Name.Name, i)
			}
			if i != 0 {
				return nil, fmt.Errorf("Function %s, parameter position %d: *plgo.TriggerData type must be the first parameter", function.Name.Name, i)
			}
			if len(param.Names) > 1 {
				return nil, fmt.Errorf("Function %s, parameter position %d: *plgo.TriggerData must be just one parameter", function.Name.Name, i)
			}
			Params = append(Params, Param{Name: param.Names[0].Name, Type: "TriggerData"})
		default:
			return nil, fmt.Errorf("Function %s, parameter position %d: type not supported", function.Name.Name, i)
		}
	}
	return
}

func getReturnType(functionName string, results *ast.FieldList) (string, error) {
	//Result is void
	if results == nil {
		return "", nil
	}
	if len(results.List) > 1 {
		return "", fmt.Errorf("Function %s has multiple return types", functionName)
	}
	switch res := results.List[0].Type.(type) {
	case *ast.StarExpr:
		var selector *ast.SelectorExpr
		selector, ok := res.X.(*ast.SelectorExpr)
		if !ok {
			return "", fmt.Errorf("Function %s has not supported return type", functionName)
		}
		var pkg *ast.Ident
		pkg, ok = selector.X.(*ast.Ident)
		if !ok {
			return "", fmt.Errorf("Function %s has not supported return type", functionName)
		}
		if pkg.Name != plgo || selector.Sel.Name != triggerRow {
			return "", fmt.Errorf("Function %s has not supported return type", functionName)
		}
		return "TriggerRow", nil
	case *ast.Ident:
		if _, ok := datumTypes[res.Name]; !ok {
			return "", fmt.Errorf("Function %s has not suported return type", functionName)
		}
		return res.Name, nil
	default:
		return "", fmt.Errorf("Function %s has not suported return type", functionName)
	}
}

//Param the parameters of the functions
type Param struct {
	Name, Type string
}

//VoidFunction is an function with no return type
type VoidFunction struct {
	Name   string
	Params []Param
	Doc    string
}

//FuncDec returns the PG INFO_V1 macro
func (f *VoidFunction) FuncDec() string {
	return "PG_FUNCTION_INFO_V1(" + f.Name + ");"
}

//Code writes the wrapper function
func (f *VoidFunction) Code(w io.Writer) {
	w.Write([]byte("//export " + f.Name + "\nfunc " + f.Name + "(fcinfo *funcInfo) Datum {\n"))
	if len(f.Params) > 0 {
		for _, p := range f.Params {
			w.Write([]byte("var " + p.Name + " " + p.Type + "\n"))
		}
		w.Write([]byte("fcinfo.Scan(\n"))
		for _, p := range f.Params {
			w.Write([]byte("&" + p.Name + ",\n"))
		}
		w.Write([]byte(")\n"))
	}
	w.Write([]byte(ToUnexported(f.Name) + "(\n"))
	for _, p := range f.Params {
		w.Write([]byte(p.Name + ",\n"))
	}
	w.Write([]byte(")\n"))
	w.Write([]byte("return toDatum(nil)\n"))
	w.Write([]byte("}\n"))
}

//Function is a list of parameters and the return type
type Function struct {
	VoidFunction
	ReturnType string
}

//Code writes the wrapper function
func (f *Function) Code(w io.Writer) {
	w.Write([]byte("//export " + f.Name + "\nfunc " + f.Name + "(fcinfo *funcInfo) Datum {\n"))
	if len(f.Params) > 0 {
		for _, p := range f.Params {
			w.Write([]byte("var " + p.Name + " " + p.Type + "\n"))
		}
		w.Write([]byte("fcinfo.Scan(\n"))
		for _, p := range f.Params {
			w.Write([]byte("&" + p.Name + ",\n"))
		}
		w.Write([]byte(")\n"))
	}
	w.Write([]byte("ret := "))
	w.Write([]byte(ToUnexported(f.Name) + "(\n"))
	for _, p := range f.Params {
		w.Write([]byte(p.Name + ",\n"))
	}
	w.Write([]byte(")\n"))
	w.Write([]byte("return toDatum(ret)\n"))
	w.Write([]byte("}\n"))
}

//TriggerFunction a special type of function, it takes TriggerData as the first argument and TriggerRow as return type
type TriggerFunction struct {
	VoidFunction
}

//Code writes the wrapper function
func (f *TriggerFunction) Code(w io.Writer) {
	w.Write([]byte("//export " + f.Name + "\nfunc " + f.Name + "(fcinfo *funcInfo) Datum {\n"))
	if len(f.Params) > 0 {
		//TODO scan from fcinfo may not work, TEST IT!
		for _, p := range f.Params {
			w.Write([]byte("var " + p.Name + " " + p.Type + "\n"))
		}
		w.Write([]byte("fcinfo.Scan(\n"))
		for _, p := range f.Params {
			w.Write([]byte("&" + p.Name + ",\n"))
		}
		w.Write([]byte(")\n"))
	}
	w.Write([]byte("ret := "))
	w.Write([]byte(ToUnexported(f.Name) + "(\nfcinfo.TriggerData(),\n"))
	for _, p := range f.Params {
		w.Write([]byte(p.Name + ",\n"))
	}
	w.Write([]byte(")\n"))
	w.Write([]byte("return toDatum(ret)\n"))
	w.Write([]byte("}\n"))
}
