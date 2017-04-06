package main

import (
	"fmt"
	"go/ast"
	"strings"
)

var functionNames = make(map[string]*Function)

//Param the parameters of the functions
type Param struct {
	Name, Type string
}

//Function is a list of parameters and the return type
type Function struct {
	Params     []Param
	ReturnType string
}

func getParamList(paramList []*ast.Field) (Params []Param, err error) {
	for i, param := range paramList {
		//TODO array types
		switch p := param.Type.(type) {
		case *ast.Ident:
			for _, name := range param.Names {
				if _, ok := datumTypes[p.Name]; !ok {
					return nil, fmt.Errorf("not ok param type, note datumable") //TODO
				}
				Params = append(Params, Param{Name: name.Name, Type: p.Name})
			}
		case *ast.StarExpr:
			selector, ok := p.X.(*ast.SelectorExpr)
			if !ok {
				return nil, fmt.Errorf("not ok param type, star") //TODO
			}
			var pkg *ast.Ident
			pkg, ok = selector.X.(*ast.Ident)
			if !ok {
				return nil, fmt.Errorf("not ok param type, star X") //TODO
			}
			if pkg.Name != plgo || selector.Sel.Name != triggerData {
				return nil, fmt.Errorf("not ok param type, not *plgo.TriggerData") //TODO
			}
			if i != 0 {
				return nil, fmt.Errorf("*plgo.TriggerData type must be the first parameter") //TODO
			}
			if len(param.Names) > 1 {
				return nil, fmt.Errorf("*plgo.TriggerData must be just one parameter") //TODO
			}
			Params = append(Params, Param{Name: param.Names[0].Name, Type: "TriggerData"})
		default:
			return nil, fmt.Errorf("not ok param type") //TODO
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

//NewFunction returns a new Function, inicialized from ast.FuncDecl
func NewFunction(function *ast.FuncDecl) (*Function, error) {
	f := functionNames[function.Name.Name]
	if f == nil {
		f = new(Function)
	}
	var err error
	f.Params, err = getParamList(function.Type.Params.List)
	if err != nil {
		return nil, err
	}
	f.ReturnType, err = getReturnType(function.Name.Name, function.Type.Results)
	if err != nil {
		return nil, err
	}
	if f.Params[0].Type != "TriggerData" {
		return nil, fmt.Errorf("Function %s can return *plgo.TriggerRow when the first parameter will be *plgo.TriggerData", function.Name.Name)
	}
	return f, nil
}

//IsTrigger is true if the function has TriggerData as first argument and *plgo.TriggerRow return type
func (f *Function) IsTrigger() bool {
	return len(f.Params) > 0 && f.Params[0].Type == "TriggerData" && f.ReturnType == "TriggerRow"
}

//FuncVisitor is an function that can be used like Visitor interface for ast.Walk
type FuncVisitor struct{}

//Visit just calls itself
func (v *FuncVisitor) Visit(node ast.Node) ast.Visitor {
	function, ok := node.(*ast.FuncDecl)
	if !ok || !ast.IsExported(function.Name.Name) {
		return v
	}
	var err error
	if functionNames[function.Name.Name], err = NewFunction(function); err != nil {
		panic(err) //TODO
	}
	function.Name.Name = strings.ToLower(function.Name.Name[0:1]) + function.Name.Name[1:]
	return v
}
