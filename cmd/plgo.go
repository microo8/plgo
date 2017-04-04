package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

var functionNames = make(map[string]*ast.FieldList)

//FuncVisitor is an function that can be used like Visitor interface for ast.Walk
type FuncVisitor struct{}

//Visit just calls itself
func (v *FuncVisitor) Visit(node ast.Node) ast.Visitor {
	//is exported function
	function, ok := node.(*ast.FuncDecl)
	if !ok || !ast.IsExported(function.Name.Name) {
		return v
	}

	functionNames[function.Name.Name] = function.Type.Params
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
	plgoPath := path.Join(os.Getenv("GOPATH"), "src", "github.com", "microo8", "plgo", "pl.go")
	//TODO check plgo exists + how to go get the plgo lib
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
	plgoSource = strings.Replace(plgoSource, "{funcdec}", funcdec, 1)
	err = ioutil.WriteFile(path.Join(packagePath, "pl.go"), []byte(plgoSource), 0644)
	if err != nil {
		return fmt.Errorf("Cannot write file tempdir: %s", err)
	}
	return nil
}

func buildPackage(buildPath, packagePath string) error {
	goBuild := exec.Command("go", "build", "-buildmode=c-shared",
		"-o", path.Join(buildPath, path.Base(packagePath)+".so"),
		path.Join(buildPath, "package.go"),
		path.Join(buildPath, "pl.go"),
	)
	err := goBuild.Start()
	if err != nil {
		return fmt.Errorf("Cannot build package: %s", err)
	}
	return nil
}

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 || len(flag.Args()) > 2 || flag.Arg(0) != "build" {
		fmt.Println(`Usage:\nplgo build [path/to/package]`)
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
}
