package main

import (
	"fmt"
	"go/ast"
	"io"
	"reflect"
	"regexp"
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
	"time.Time":   "timestamptz",
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
	"[]time.Time": "timestamptz[]",
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
	returnType, isStar, err := getReturnType(function.Name.Name, function.Type.Results)
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
	return &Function{VoidFunction: VoidFunction{Name: function.Name.Name, Params: params, Doc: function.Doc.Text()}, ReturnType: returnType, IsStar: isStar}, nil
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
			case *ast.SelectorExpr:
				//other selector
				var pkg *ast.Ident
				pkg, ok := paramType.X.(*ast.Ident)
				if !ok {
					return nil, fmt.Errorf("Function %s, parameter %s: type not supported, could not get identifier", function.Name.Name, paramName.Name)
				}
				typeName := pkg.Name + "." + paramType.Sel.Name
				if _, ok := datumTypes[typeName]; !ok {
					return nil, fmt.Errorf("Function %s, parameter %s: type %s not supported", function.Name.Name, paramName.Name, typeName)
				}
				Params = append(Params, Param{Name: paramName.Name, Type: typeName})
			case *ast.ArrayType:
				//built in array type
				arrayType, ok := paramType.Elt.(*ast.Ident)
				if !ok {
					return nil, fmt.Errorf("Function %s, parameter %s: array type not supported", function.Name.Name, paramName.Name)
				}
				if _, ok := datumTypes["[]"+arrayType.Name]; !ok {
					return nil, fmt.Errorf("Function %s, parameter %s: array type not supported", function.Name.Name, paramName.Name)
				}
				Params = append(Params, Param{Name: paramName.Name, Type: "[]" + arrayType.Name})
			case *ast.StarExpr:
				//*plgo.TriggerData
				selector, ok := paramType.X.(*ast.SelectorExpr)
				if !ok {
					return nil, fmt.Errorf("Function %s, parameter %s: type not supported, could not get selector", function.Name.Name, paramName.Name)
				}
				var pkg *ast.Ident
				pkg, ok = selector.X.(*ast.Ident)
				if !ok {
					return nil, fmt.Errorf("Function %s, parameter %s: type not supported, could not get identifier", function.Name.Name, paramName.Name)
				}
				if pkg.Name != plgo || selector.Sel.Name != triggerData {
					// type is not plgo.TriggerData, check for other supported types
					typeName := pkg.Name + "." + selector.Sel.Name
					if _, ok := datumTypes[typeName]; ok {
						Params = append(Params, Param{Name: paramName.Name, Type: typeName, Null: true})
						continue
					}
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
				return nil, fmt.Errorf("Function %s, parameter %s: type not supported, no match for %s", function.Name.Name, paramName.Name, reflect.TypeOf(param.Type))
			}
		}
	}
	return
}

func getReturnType(functionName string, results *ast.FieldList) (string, bool, error) {
	//Result is void
	if results == nil {
		return "", false, nil
	}
	if len(results.List) > 1 {
		return "", false, fmt.Errorf("Function %s has multiple return types", functionName)
	}
	switch res := results.List[0].Type.(type) {
	case *ast.StarExpr:
		var selector *ast.SelectorExpr
		ident, ok := res.X.(*ast.Ident)
		if ok {
			if _, ok := datumTypes[ident.Name]; !ok {
				return "", false, fmt.Errorf("Function %s has not suported return type", functionName)
			}
			return ident.Name, true, nil
		}
		selector, ok = res.X.(*ast.SelectorExpr)
		if !ok {
			return "", false, fmt.Errorf("Function %s has not supported return type", functionName)
		}
		var pkg *ast.Ident
		pkg, ok = selector.X.(*ast.Ident)
		if !ok {
			return "", false, fmt.Errorf("Function %s has not supported return type", functionName)
		}
		if pkg.Name != plgo || selector.Sel.Name != triggerRow {
			return "", false, fmt.Errorf("Function %s has not supported return type", functionName)
		}
		return "TriggerRow", false, nil
	case *ast.Ident:
		if _, ok := datumTypes[res.Name]; !ok {
			return "", false, fmt.Errorf("Function %s has not suported return type", functionName)
		}
		return res.Name, false, nil
	case *ast.ArrayType:
		ident, ok := res.Elt.(*ast.Ident)
		if !ok {
			return "", false, fmt.Errorf("Function %s has not supported return type", functionName)
		}
		return "[]" + ident.Name, false, nil
	default:
		return "", false, fmt.Errorf("Function %s has not suported return type", functionName)
	}
}

//Param the parameters of the functions
type Param struct {
	Name string
	Type string
	Null bool
}

func (p *Param) sql() string {
	return p.Name + " " + datumTypes[p.Type]
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
	w.Write([]byte("__" + f.Name + "(\n"))
	for _, p := range f.Params {
		if p.Null {
			w.Write([]byte("&" + p.Name + ",\n"))
		} else {
			w.Write([]byte(p.Name + ",\n"))
		}
	}
	w.Write([]byte(")\n"))
	w.Write([]byte("return toDatum(nil)\n"))
	w.Write([]byte("}\n"))
}

//SQL writes the SQL command that creates the function in DB
func (f *VoidFunction) SQL(packageName string, w io.Writer) {
	w.Write([]byte("CREATE OR REPLACE FUNCTION " + toPgName(f.Name) + "("))
	var paramStrings []string
	isNull := false
	for _, p := range f.Params {
		paramStrings = append(paramStrings, p.sql())
		isNull = isNull || p.Null
	}
	w.Write([]byte(strings.Join(paramStrings, ",")))
	w.Write([]byte(")\n"))
	w.Write([]byte("RETURNS VOID AS\n"))
	w.Write([]byte("'$libdir/" + packageName + "', '" + f.Name + "'\n"))
	if isNull {
		w.Write([]byte("LANGUAGE c VOLATILE;\n"))
	} else {
		w.Write([]byte("LANGUAGE c VOLATILE STRICT;\n"))
	}
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
	w.Write([]byte("COMMENT ON FUNCTION " + toPgName(f.Name) + "(" + strings.Join(paramTypes, ",") + ") IS '" + f.Doc + "';\n\n"))
}

//Function is a list of parameters and the return type
type Function struct {
	VoidFunction
	ReturnType string
	IsStar     bool
}

//Code writes the wrapper function
func (f *Function) Code(w io.Writer) {
	w.Write([]byte("//export " + f.Name + "\nfunc " + f.Name + "(fcinfo *funcInfo) Datum {\n"))
	if len(f.Params) > 0 {
		for _, p := range f.Params {
			w.Write([]byte("var " + p.Name + " " + p.Type + "\n"))
		}
		w.Write([]byte("err:=fcinfo.Scan(\n"))
		for _, p := range f.Params {
			w.Write([]byte("&" + p.Name + ",\n"))
		}
		w.Write([]byte(")\n"))
		w.Write([]byte(`
		if(err!=nil){
			C.elog_error(C.CString(
				err.Error(),
			))
		}
		`))
	}
	w.Write([]byte("ret := "))
	w.Write([]byte("__" + f.Name + "(\n"))
	for _, p := range f.Params {
		if p.Null {
			w.Write([]byte("&" + p.Name + ",\n"))
		} else {
			w.Write([]byte(p.Name + ",\n"))
		}
	}
	w.Write([]byte(")\n"))
	if f.IsStar {
		w.Write([]byte(`
		if(ret==nil){
			fcinfo.isnull=C.char(1);
			return toDatum(nil)
		}
		return toDatum(*ret)
		`))
	} else {
		w.Write([]byte("return toDatum(ret)\n"))
	}
	w.Write([]byte("}\n"))

}

//SQL writes the SQL command that creates the function in DB
func (f *Function) SQL(packageName string, w io.Writer) {
	w.Write([]byte("CREATE OR REPLACE FUNCTION " + toPgName(f.Name) + "("))
	var paramsString []string
	isNull := false
	for _, p := range f.Params {
		paramsString = append(paramsString, p.sql())
		isNull = isNull || p.Null
	}
	w.Write([]byte(strings.Join(paramsString, ",")))
	w.Write([]byte(")\n"))
	switch {
	case f.ReturnType == "[]byte":
		w.Write([]byte("RETURNS bytea AS\n"))
	case strings.HasPrefix(f.ReturnType[:2], "[]"):
		w.Write([]byte("RETURNS " + datumTypes[f.ReturnType[2:len(f.ReturnType)]] + "[] AS\n"))
	default:
		w.Write([]byte("RETURNS " + datumTypes[f.ReturnType] + " AS\n"))
	}
	w.Write([]byte("'$libdir/" + packageName + "', '" + f.Name + "'\n"))
	if isNull {
		w.Write([]byte("LANGUAGE c VOLATILE;\n"))
	} else {
		w.Write([]byte("LANGUAGE c VOLATILE STRICT;\n"))
	}
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
	w.Write([]byte("__" + f.Name + "(\nfcinfo.TriggerData(),\n"))
	for _, p := range f.Params {
		w.Write([]byte(p.Name + ",\n"))
	}
	w.Write([]byte(")\n"))
	w.Write([]byte("return toDatum(ret)\n"))
	w.Write([]byte("}\n"))
}

//SQL writes the SQL command that creates the function in DB
func (f *TriggerFunction) SQL(packageName string, w io.Writer) {
	w.Write([]byte("CREATE OR REPLACE FUNCTION " + toPgName(f.Name) + "("))
	var paramsString []string
	for _, p := range f.Params {
		paramsString = append(paramsString, p.sql())
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

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toPgName(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
