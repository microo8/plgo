package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

const methods = `package main

/*
#include "postgres.h"
#include "fmgr.h"
*/
import "C"

{{range $funcName, $funcParams := .}}
//export {{$funcName}}
func {{$funcName}}(fcinfo *FuncInfo) Datum {
	{{range $funcParams}}var {{.Name}} {{.Type}}
	{{end}}
	fcinfo.Scan(
		{{range $funcParams}}&{{.Name}},
		{{end}})
	ret := {{$funcName | ToLower }}(
		{{range $funcParams}}{{.Name}},
		{{end}})
	return ToDatum(ret)
}
{{end}}
`

const sql = `{{range . }}
CREATE OR REPLACE FUNCTION {{.Schema}}.{{.Name}}({{range $funcParams}}{{.Name}} {{.Type}}, {{end}})
RETURNS {{.ReturnType}} AS
'$libdir/{{..Package}}', '{{.Name}}'
LANGUAGE c IMMUTABLE STRICT;
{{end}}`

//Param represents the parameters of the functions
type Param struct {
	Name, Type string
}

var functionNames = make(map[string][]Param)

//FuncVisitor is an function that can be used like Visitor interface for ast.Walk
type FuncVisitor struct{}

//Visit just calls itself
func (v *FuncVisitor) Visit(node ast.Node) ast.Visitor {
	//is exported function
	function, ok := node.(*ast.FuncDecl)
	if !ok || !ast.IsExported(function.Name.Name) {
		return v
	}

	for _, param := range function.Type.Params.List {
		paramType, ok := param.Type.(*ast.Ident)
		if !ok {
			panic("not ok param type") //TODO
		}
		for _, name := range param.Names {
			functionNames[function.Name.Name] = append(functionNames[function.Name.Name], Param{Name: name.Name, Type: paramType.Name})
		}
	}
	function.Name.Name = strings.ToLower(function.Name.Name[0:1]) + function.Name.Name[1:]

	//TODO test param type, so that the type is supported string, int, float64 etc.
	//params := function.Type.Params
	//for i, param := range params.List {
	//TODO test
	//}

	return v
}

//ImportVisitor is an function that can be used like Visitor interface for ast.Walk
type ImportVisitor struct{}

//Visit removes plgo imports
func (v *ImportVisitor) Visit(node ast.Node) ast.Visitor {
	imp, ok := node.(*ast.ImportSpec)
	if !ok || imp.Path.Value != "\"github.com/microo8/plgo\"" {
		return v
	}
	imp.Path.Value = ""
	return v
}

//PLGOVisitor is an function that can be used like Visitor interface for ast.Walk
type PLGOVisitor struct{}

//Visit removes plgo imports
func (v *PLGOVisitor) Visit(node ast.Node) ast.Visitor {
	call, ok := node.(*ast.CallExpr)
	if !ok {
		return v
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return v
	}
	expr, ok := selector.X.(*ast.Ident)
	if !ok || expr.Name != "plgo" {
		return v
	}
	call.Fun = selector.Sel
	return v
}

//parsePackage parses the go package and returns the FileSet and AST
func parsePackage(packagePath string) (*token.FileSet, *ast.Package, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseDir(fset, packagePath, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("Cannot parse package: %s", err)
	}
	if len(f) > 1 {
		return nil, nil, fmt.Errorf("More than one package in %s", packagePath)
	}
	packageAst, ok := f["main"]
	if !ok {
		return nil, nil, fmt.Errorf("No package main in %s", packagePath)
	}
	return fset, packageAst, nil
}

//writeFiles creates users module, the plgo library and the methods wrapper in the temporary directory
func writeFiles(packagePath string, fset *token.FileSet, packageAst *ast.Package) error {
	//write users package
	packageFile, err := os.Create(path.Join(packagePath, "package.go"))
	if err != nil {
		return fmt.Errorf("Cannot write file tempdir: %s", err)
	}
	if err = format.Node(packageFile, fset, ast.MergePackageFiles(packageAst, ast.FilterFuncDuplicates)); err != nil {
		return fmt.Errorf("Cannot format package %s", err)
	}
	err = packageFile.Close()
	if err != nil {
		return fmt.Errorf("Cannot write file tempdir: %s", err)
	}
	//write and modify plgo package
	plgoPath := path.Join(os.Getenv("GOPATH"), "src", "github.com", "microo8", "plgo", "pl.go")
	fmt.Println("plgoPath:", plgoPath)
	if _, err = os.Stat(plgoPath); os.IsNotExist(err) {
		return fmt.Errorf("Package github.com/microo8/plgo not installed\nplease install it with: go get -u github.com/microo8/plgo/... ")
	}
	plgoSourceBin, err := ioutil.ReadFile(plgoPath)
	if err != nil {
		return fmt.Errorf("Cannot read plgo package: %s", err)
	}
	plgoSource := string(plgoSourceBin)
	plgoSource = "package main\n\n" + plgoSource[12:]
	postgresIncludeDir, err := exec.Command("pg_config", "--includedir-server").CombinedOutput()
	if err != nil {
		return fmt.Errorf("Cannot run pg_config: %s", err)
	}
	plgoSource = strings.Replace(plgoSource, "/usr/include/postgresql/server", string(postgresIncludeDir), 1)
	var funcdec string
	for funcName := range functionNames {
		funcdec += "PG_FUNCTION_INFO_V1(" + funcName + ");"
	}
	plgoSource = strings.Replace(plgoSource, "//{funcdec}", funcdec, 1)
	err = ioutil.WriteFile(path.Join(packagePath, "pl.go"), []byte(plgoSource), 0644)
	if err != nil {
		return fmt.Errorf("Cannot write file tempdir: %s", err)
	}
	//create the exported methods file
	methodsFile, err := os.Create(path.Join(packagePath, "methods.go"))
	if err != nil {
		return fmt.Errorf("Cannot write file tempdir: %s", err)
	}
	funcMap := template.FuncMap{
		"ToLower": func(str string) string { return strings.ToLower(str[0:1]) + str[1:] },
	}
	t := template.Must(template.New("").Funcs(funcMap).Parse(methods))
	err = t.Execute(methodsFile, functionNames)
	if err != nil {
		return fmt.Errorf("Cannot write file tempdir: %s", err)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func buildPackage(buildPath, packagePath string) error {
	goBuild := exec.Command("go", "build", "-buildmode=c-shared",
		"-o", path.Join(buildPath, path.Base(packagePath)+".so"),
		path.Join(buildPath, "package.go"),
		path.Join(buildPath, "methods.go"),
		path.Join(buildPath, "pl.go"),
	)
	//TODO test the compile errors and print it out
	goBuild.Stdout = os.Stdout
	goBuild.Stderr = os.Stderr
	err := goBuild.Run()
	if err != nil {
		return fmt.Errorf("Cannot build package: %s", err)
	}
	return nil
}

func printUsage() {
	fmt.Println(`Usage:
plgo build [path/to/package]
or
plgo install [path/to/package]`)
}

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 || len(flag.Args()) > 2 || (flag.Arg(0) != "build" && flag.Arg(0) != "install") {
		printUsage()
		return
	}
	//set package path
	packagePath := "."
	if len(flag.Args()) == 2 {
		packagePath = flag.Arg(1)
	}
	fset, packageAst, err := parsePackage(packagePath)
	if err != nil {
		fmt.Println(err)
		printUsage()
		return
	}
	ast.Walk(new(FuncVisitor), packageAst)
	ast.Walk(new(ImportVisitor), packageAst)
	ast.Walk(new(PLGOVisitor), packageAst)
	tempPackagePath, err := ioutil.TempDir("", "plgo")
	if err != nil {
		fmt.Println("Cannot get tempdir:", err)
		return
	}
	err = writeFiles(tempPackagePath, fset, packageAst)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = buildPackage(tempPackagePath, packagePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch flag.Arg(0) {
	case "build":
		err = copyFile(
			path.Join(tempPackagePath, path.Base(packagePath)+".so"),
			path.Join(".", path.Base(packagePath)+".so"),
		)
		if err != nil {
			fmt.Println(err)
			return
		}
	case "install":
		pglibdirBin, err := exec.Command("pg_config", "--pkglibdir").CombinedOutput()
		if err != nil {
			fmt.Println("Cannot get postgresql libdir:", err)
			return
		}
		pglibdir := strings.TrimSpace(string(pglibdirBin))
		err = copyFile(
			path.Join(tempPackagePath, path.Base(packagePath)+".so"),
			path.Join(pglibdir, path.Base(packagePath)+".so"),
		)
		if err != nil && os.IsPermission(err) {
			cp := exec.Command("sudo", "cp",
				path.Join(tempPackagePath, path.Base(packagePath)+".so"),
				path.Join(pglibdir, path.Base(packagePath)+".so"),
			)
			cp.Stdin = os.Stdin
			cp.Stderr = os.Stderr
			cp.Stdout = os.Stdout
			fmt.Println("Copying the shared objects file to postgres libdir, must get permissions")
			err = cp.Run()
		}
		if err != nil {
			fmt.Println("Cannot copy the shared objects file to postgres libdir:", err)
			return
		}
		//TODO create sql and run it in db
	}
}
