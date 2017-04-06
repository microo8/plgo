package main

import (
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
	"text/template"
)

//Remover is an function that can be used like Visitor interface for ast.Walk
type Remover struct{}

//Visit removes plgo selectors and plgo import
func (v *Remover) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.ImportSpec:
		if n.Path.Value == "\"github.com/microo8/plgo\"" {
			n.Path.Value = ""
		}
	case *ast.CallExpr:
		selector, ok := n.Fun.(*ast.SelectorExpr)
		if !ok {
			break
		}
		expr, ok := selector.X.(*ast.Ident)
		if !ok || expr.Name != plgo {
			break
		}
		n.Fun = selector.Sel
	case *ast.StarExpr:
		sel, ok := n.X.(*ast.SelectorExpr)
		if !ok {
			break
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok || ident.Name != plgo {
			break
		}
		n.X = sel.Sel
	}
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

func writeUserPackage(packagePath string, fset *token.FileSet, packageAst *ast.Package) error {
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
	return nil
}

func writeplgo(packagePath string) error {
	plgoPath := path.Join(os.Getenv("GOPATH"), "src", "github.com", "microo8", "plgo", "pl.go")
	fmt.Println("plgoPath:", plgoPath)
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
	for funcName := range functionNames {
		funcdec += "PG_FUNCTION_INFO_V1(" + funcName + ");"
	}
	plgoSource = strings.Replace(plgoSource, "//{funcdec}", funcdec, 1)
	err = ioutil.WriteFile(path.Join(packagePath, "pl.go"), []byte(plgoSource), 0644)
	if err != nil {
		return fmt.Errorf("Cannot write file tempdir: %s", err)
	}
	return nil
}

func writeExportedMethods(packagePath string) error {
	methodsFile, err := os.Create(path.Join(packagePath, "methods.go"))
	if err != nil {
		return fmt.Errorf("Cannot write file tempdir: %s", err)
	}
	funcMap := template.FuncMap{
		"ToLower": func(str string) string { return strings.ToLower(str[0:1]) + str[1:] },
	}
	t := template.Must(template.New("").Funcs(funcMap).Parse(methods))
	err = t.Execute(methodsFile, functionNames) //TODO format the output
	if err != nil {
		return fmt.Errorf("Cannot write file tempdir: %s", err)
	}
	return nil
}

//writeFiles creates users module, the plgo library and the methods wrapper in the temporary directory
func writeFiles(packagePath string, fset *token.FileSet, packageAst *ast.Package) error {
	err := writeUserPackage(packagePath, fset, packageAst)
	if err != nil {
		return err
	}
	err = writeplgo(packagePath)
	if err != nil {
		return err
	}
	return writeExportedMethods(packagePath)
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
	ast.Walk(new(FuncVisitor), packageAst)
	ast.Walk(new(Remover), packageAst)
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
