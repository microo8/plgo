package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

//ToUnexported changes Exported function name to unexported
func ToUnexported(name string) string {
	return strings.ToLower(name[0:1]) + name[1:]
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

func writeUserPackage(tempPackagePath string, fset *token.FileSet, packageAst *ast.Package) error {
	packageFile, err := os.Create(path.Join(tempPackagePath, "package.go"))
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
	return nil
}

func writeplgo(tempPackagePath string, functions []CodeWriter) error {
	plgoPath := path.Join(os.Getenv("GOPATH"), "src", "github.com", "microo8", "plgo", "pl.go")
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
	for _, f := range functions {
		funcdec += f.FuncDec()
	}
	plgoSource = strings.Replace(plgoSource, "//{funcdec}", funcdec, 1)
	err = ioutil.WriteFile(path.Join(tempPackagePath, "pl.go"), []byte(plgoSource), 0644)
	if err != nil {
		return fmt.Errorf("Cannot write file tempdir: %s", err)
	}
	return nil
}

func writeExportedMethods(tempPackagePath string, functions []CodeWriter) error {
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
	for _, f := range functions {
		f.Code(buf)
	}
	//fmt.Println(buf.String())
	code, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(path.Join(tempPackagePath, "methods.go"), code, 0644)
	if err != nil {
		return fmt.Errorf("Cannot write file tempdir: %s", err)
	}
	return nil
}

//writeFiles creates users module, the plgo library and the methods wrapper in the temporary directory
func writeFiles(tempPackagePath string, fset *token.FileSet, packageAst *ast.Package, functions []CodeWriter) error {
	err := writeUserPackage(tempPackagePath, fset, packageAst)
	if err != nil {
		return err
	}
	err = writeplgo(tempPackagePath, functions)
	if err != nil {
		return err
	}
	return writeExportedMethods(tempPackagePath, functions)
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
	//collect functions from the package
	funcVisitor := new(FuncVisitor)
	ast.Walk(funcVisitor, packageAst)
	if funcVisitor.err != nil {
		fmt.Println(err)
		return
	}
	//remove plgo usages
	ast.Walk(new(Remover), packageAst)
	//write package and its wrappers in an temp dir
	tempPackagePath, err := ioutil.TempDir("", "plgo")
	if err != nil {
		fmt.Println("Cannot get tempdir:", err)
		return
	}
	err = writeFiles(tempPackagePath, fset, packageAst, funcVisitor.functions)
	if err != nil {
		fmt.Println(err)
		return
	}
	//build the package
	err = buildPackage(tempPackagePath, packagePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch flag.Arg(0) {
	case "build":
		//TODO create build dir
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
