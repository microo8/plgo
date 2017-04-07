package main

import (
	"fmt"
	"go/ast"
	"strings"
)

//ToUnexported changes Exported function name to unexported
func ToUnexported(name string) string {
	return strings.ToLower(name[0:1]) + name[1:]
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

//NewCode parses the ast.FuncDecl and returns a new Function or An TriggerFunction
func NewCode(function *ast.FuncDecl) (CodeWriter, error) {
	params, err := getParamList(function.Type.Params.List)
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
		return &TriggerFunction{VoidFunction: VoidFunction{Name: function.Name.Name, Params: params[1:]}}, nil
	}
	if returnType == "" {
		return &VoidFunction{Name: function.Name.Name, Params: params}, nil
	}
	return &Function{VoidFunction: VoidFunction{Name: function.Name.Name, Params: params}, ReturnType: returnType}, nil
}

//FuncVisitor is an function that can be used like Visitor interface for ast.Walk
type FuncVisitor struct {
	err       error
	functions []CodeWriter
}

//Visit just calls itself
func (v *FuncVisitor) Visit(node ast.Node) ast.Visitor {
	function, ok := node.(*ast.FuncDecl)
	if !ok || !ast.IsExported(function.Name.Name) {
		return v
	}
	var code CodeWriter
	code, v.err = NewCode(function)
	if v.err != nil {
		return nil
	}
	v.functions = append(v.functions, code)
	function.Name.Name = ToUnexported(function.Name.Name)
	return v
}
