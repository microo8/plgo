package main

import (
	"fmt"
	"go/ast"
	"io"
	"strings"
)

const (
	triggerData = "TriggerData"
	triggerRow  = "TriggerRow"
)

var datumTypes = map[string]string{
	"error":       "text",
	"string":      "text",
	"[]byte":      "bytea",
	"int16":       "smallint",
	"uint16":      "smallint",
	"int32":       "integer",
	"uint32":      "integer",
	"int64":       "bigint",
	"int":         "bigint",
	"uint":        "bigint",
	"float32":     "real",
	"float64":     "double precision",
	"time.Time":   "timestamp with timezone",
	"bool":        "boolean",
	"[]string":    "text[]",
	"[]int16":     "smallint[]",
	"[]uint16":    "smallint[]",
	"[]int32":     "integer[]",
	"[]uint32":    "integer[]",
	"[]int64":     "bigint[]",
	"[]int":       "bigint[]",
	"[]uint":      "bigint[]",
	"[]float32":   "real[]",
	"[]float64":   "double precision[]",
	"[]bool":      "boolean[]",
	"[]time.Time": "timestamp with timezone[]",
	"TriggerRow":  "trigger",
}

//CodeWriter is an interface of an object that can print its code
type CodeWriter interface {
	FuncDec() string
	Code(w io.Writer)
	SQL(packageName string, w io.Writer)
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
		for _, paramName := range param.Names {
			switch paramType := param.Type.(type) {
			case *ast.Ident:
				//built in type
				if _, ok := datumTypes[paramType.Name]; !ok {
					return nil, fmt.Errorf("Function %s, parameter %s: type %s not supported", function.Name.Name, paramName.Name, paramType.Name)
				}
				Params = append(Params, Param{Name: paramName.Name, Type: paramType.Name})
			case *ast.ArrayType:
				//built in array type
				arrayType, ok := paramType.Elt.(*ast.Ident)
				if !ok {
					return nil, fmt.Errorf("Function %s, parameter %s: array type not supported", function.Name.Name, paramName.Name)
				}
				if _, ok := datumTypes[arrayType.Name]; !ok {
					return nil, fmt.Errorf("Function %s, parameter %s: array type not supported", function.Name.Name, paramName.Name)
				}
				Params = append(Params, Param{Name: paramName.Name, Type: "[]" + arrayType.Name})
			case *ast.StarExpr:
				//*plgo.TriggerData
				selector, ok := paramType.X.(*ast.SelectorExpr)
				if !ok {
					return nil, fmt.Errorf("Function %s, parameter %s: type not supported", function.Name.Name, paramName.Name)
				}
				var pkg *ast.Ident
				pkg, ok = selector.X.(*ast.Ident)
				if !ok {
					return nil, fmt.Errorf("Function %s, parameter %s: type not supported", function.Name.Name, paramName.Name)
				}
				if pkg.Name != plgo || selector.Sel.Name != triggerData {
					return nil, fmt.Errorf("Function %s, parameter %s: type not supported", function.Name.Name, paramName.Name)
				}
				if i != 0 {
					return nil, fmt.Errorf("Function %s, parameter %s: *plgo.TriggerData type must be the first parameter", function.Name.Name, paramName.Name)
				}
				if len(param.Names) > 1 {
					return nil, fmt.Errorf("Function %s, parameter %s: *plgo.TriggerData must be just one parameter", function.Name.Name, paramName.Name)
				}
				Params = append(Params, Param{Name: param.Names[0].Name, Type: "TriggerData"})
			default:
				return nil, fmt.Errorf("Function %s, parameter %s: type not supported", function.Name.Name, paramName.Name)
			}
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

//SQL writes the SQL command that creates the function in DB
func (f *VoidFunction) SQL(packageName string, w io.Writer) {
	w.Write([]byte("CREATE OR REPLACE FUNCTION " + f.Name + "("))
	var paramStrings []string
	for _, p := range f.Params {
		paramStrings = append(paramStrings, p.Name+" "+datumTypes[p.Type])
	}
	w.Write([]byte(strings.Join(paramStrings, ",")))
	w.Write([]byte(")\n"))
	w.Write([]byte("RETURNS VOID AS\n"))
	w.Write([]byte("'$libdir/" + packageName + "', '" + f.Name + "'\n"))
	w.Write([]byte("LANGUAGE c VOLATILE STRICT;\n"))
	if f.Doc == "" {
		w.Write([]byte("\n"))
		return
	}
	f.Comment(w)
}

//Comment writes the Doc comment of the golang function as an DB comment for that function
func (f *VoidFunction) Comment(w io.Writer) {
	var paramTypes []string
	for _, p := range f.Params {
		paramTypes = append(paramTypes, datumTypes[p.Type])
	}
	w.Write([]byte("COMMENT ON FUNCTION " + f.Name + "(" + strings.Join(paramTypes, ",") + ") IS '" + f.Doc + "';\n\n"))
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

//SQL writes the SQL command that creates the function in DB
func (f *Function) SQL(packageName string, w io.Writer) {
	w.Write([]byte("CREATE OR REPLACE FUNCTION " + f.Name + "("))
	var paramsString []string
	for _, p := range f.Params {
		paramsString = append(paramsString, p.Name+" "+datumTypes[p.Type])
	}
	w.Write([]byte(strings.Join(paramsString, ",")))
	w.Write([]byte(")\n"))
	w.Write([]byte("RETURNS " + datumTypes[f.ReturnType] + " AS\n"))
	w.Write([]byte("'$libdir/" + packageName + "', '" + f.Name + "'\n"))
	w.Write([]byte("LANGUAGE c VOLATILE STRICT;\n"))
	if f.Doc == "" {
		w.Write([]byte("\n"))
		return
	}
	f.Comment(w)
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

//SQL writes the SQL command that creates the function in DB
func (f *TriggerFunction) SQL(packageName string, w io.Writer) {
	w.Write([]byte("CREATE OR REPLACE FUNCTION " + f.Name + "("))
	var paramsString []string
	for _, p := range f.Params {
		paramsString = append(paramsString, p.Name+" "+datumTypes[p.Type])
	}
	w.Write([]byte(strings.Join(paramsString, ",")))
	w.Write([]byte(")\n"))
	w.Write([]byte("RETURNS TRIGGER AS\n"))
	w.Write([]byte("'$libdir/" + packageName + "', '" + f.Name + "'\n"))
	w.Write([]byte("LANGUAGE c VOLATILE STRICT;\n"))
	if f.Doc == "" {
		w.Write([]byte("\n"))
		return
	}
	f.Comment(w)
}
