package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//ToUnexported changes Exported function name to unexported
func ToUnexported(name string) string {
	return strings.ToLower(name[0:1]) + name[1:]
}

//ModuleWriter writes the tmp module wrapper that will be build to shared object
type ModuleWriter struct {
	fset       *token.FileSet
	packageAst *ast.Package
	functions  []CodeWriter
}

//NewModuleWriter parses the go package and returns the FileSet and AST
func NewModuleWriter(packagePath string) (*ModuleWriter, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseDir(fset, packagePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse package: %s", err)
	}
	if len(f) > 1 {
		return nil, fmt.Errorf("More than one package in %s", packagePath)
	}
	packageAst, ok := f["main"]
	if !ok {
		return nil, fmt.Errorf("No package main in %s", packagePath)
	}
	//collect functions from the package
	funcVisitor := new(FuncVisitor)
	ast.Walk(funcVisitor, packageAst)
	if funcVisitor.err != nil {
		return nil, err
	}
	return &ModuleWriter{fset: fset, packageAst: packageAst, functions: funcVisitor.functions}, nil
}

//WriteModule writes the tmp module wrapper
func (mw *ModuleWriter) WriteModule() (string, error) {
	tempPackagePath, err := ioutil.TempDir("", plgo)
	if err != nil {
		return "", fmt.Errorf("Cannot get tempdir: %s", err)
	}
	err = mw.writeUserPackage(tempPackagePath)
	if err != nil {
		return "", err
	}
	err = mw.writeplgo(tempPackagePath)
	if err != nil {
		return "", err
	}
	err = mw.writeExportedMethods(tempPackagePath)
	if err != nil {
		return "", err
	}
	return tempPackagePath, nil
}

func (mw *ModuleWriter) writeUserPackage(tempPackagePath string) error {
	ast.Walk(new(Remover), mw.packageAst)
	packageFile, err := os.Create(filepath.Join(tempPackagePath, "package.go"))
	if err != nil {
		return fmt.Errorf("Cannot write file tempdir: %s", err)
	}
	if err = format.Node(packageFile, mw.fset, ast.MergePackageFiles(mw.packageAst, ast.FilterFuncDuplicates)); err != nil {
		return fmt.Errorf("Cannot format package %s", err)
	}
	err = packageFile.Close()
	if err != nil {
		return fmt.Errorf("Cannot write file tempdir: %s", err)
	}
	return nil
}

func (mw *ModuleWriter) writeplgo(tempPackagePath string) error {
	plgoPath := filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "microo8", "plgo", "pl.go")
	if _, err := os.Stat(plgoPath); os.IsNotExist(err) {
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
	for _, f := range mw.functions {
		funcdec += f.FuncDec()
	}
	plgoSource = strings.Replace(plgoSource, "//{funcdec}", funcdec, 1)
	err = ioutil.WriteFile(filepath.Join(tempPackagePath, "pl.go"), []byte(plgoSource), 0644)
	if err != nil {
		return fmt.Errorf("Cannot write file tempdir: %s", err)
	}
	return nil
}

func (mw *ModuleWriter) writeExportedMethods(tempPackagePath string) error {
	buf := bytes.NewBuffer(nil)
	_, err := buf.WriteString(`package main

/*
#include "postgres.h"
#include "fmgr.h"
*/
import "C"
`)
	if err != nil {
		return fmt.Errorf("Cannot write file tempdir: %s", err)
	}
	for _, f := range mw.functions {
		f.Code(buf)
	}
	//fmt.Println(buf.String())
	code, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(tempPackagePath, "methods.go"), code, 0644)
	if err != nil {
		return fmt.Errorf("Cannot write file tempdir: %s", err)
	}
	return nil
}
