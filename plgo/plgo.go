package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

//ReturnVisitor is an function that can be used like Visitor interface for ast.Walk
type ReturnVisitor struct{}

//Visit just calls inself
func (v *ReturnVisitor) Visit(node ast.Node) ast.Visitor {
	ret, ok := node.(*ast.ReturnStmt)
	if !ok {
		return v
	}
	ret.Results = []ast.Expr{
		&ast.CallExpr{
			Fun:  ast.NewIdent("ToDatum"),
			Args: ret.Results,
		},
	}
	return v
}

//FuncVisitor is an function that can be used like Visitor interface for ast.Walk
type FuncVisitor struct{}

//Visit just calls inself
func (v *FuncVisitor) Visit(node ast.Node) ast.Visitor {
	//is exported function
	function, ok := node.(*ast.FuncDecl)
	if !ok || !ast.IsExported(function.Name.Name) {
		return v
	}

	params := function.Type.Params
	//TODO test param types, so that the type is supported string, int, float64 etc.

	//function params are just fcinfo *FuncInfo and return value is Datum
	function.Type.Params = &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Names: []*ast.Ident{ast.NewIdent("functionCallInfoData")},
				Type:  ast.NewIdent("*C.FunctionCallInfoData"),
			},
		},
	}

	function.Type.Results = &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Type: ast.NewIdent("C.Datum"),
			},
		},
	}

	cast := []ast.Stmt{
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("fcinfo")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.ParenExpr{
						X: &ast.SelectorExpr{
							X:   ast.NewIdent("plgo"),
							Sel: ast.NewIdent("FuncInfo"),
						},
					},
					Args: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("unsafe"),
								Sel: ast.NewIdent("Pointer"),
							},
							Args: []ast.Expr{ast.NewIdent("functionCallInfoData")},
						},
					},
				},
			},
		},
	}

	//declarations of parameter variables
	paramDecs := []ast.Stmt{}
	//fcinfo.Scan args
	scanArgs := []ast.Expr{}
	for _, param := range params.List {
		for _, name := range param.Names {
			paramDecs = append(paramDecs, &ast.DeclStmt{
				Decl: &ast.GenDecl{
					Tok: token.VAR,
					Specs: []ast.Spec{
						&ast.ValueSpec{
							Names: []*ast.Ident{ast.NewIdent(name.Name)},
							Type:  ast.NewIdent(types.ExprString(param.Type)),
						},
					},
				},
			})
			scanArgs = append(scanArgs, &ast.UnaryExpr{
				Op: token.AND,
				X:  ast.NewIdent(name.Name),
			})
		}
	}

	function.Body.List = append([]ast.Stmt{
		//err := fcinfo.Scan
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("err")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent("fcinfo"),
						Sel: ast.NewIdent("Scan"),
					},
					Args: scanArgs,
				},
			},
		},
		//if err != nil panic
		&ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  ast.NewIdent("err"),
				Op: token.NEQ,
				Y:  ast.NewIdent("nil"),
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ExprStmt{
						X: &ast.CallExpr{
							Fun:  ast.NewIdent("panic"),
							Args: []ast.Expr{ast.NewIdent("err")},
						},
					},
				},
			},
		},
	}, function.Body.List...)

	//prepend all parameter variables
	function.Body.List = append(paramDecs, function.Body.List...)
	function.Body.List = append(cast, function.Body.List...)

	return v
}

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 || len(flag.Args()) > 2 || flag.Arg(0) != "build" {
		fmt.Println(`Usage:\nplgo build [path/to/package]`)
		return
	}

	dir := "."
	if len(flag.Args()) == 2 {
		dir = flag.Arg(1)
	}

	fset := token.NewFileSet()
	f, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	if len(f) > 1 {
		fmt.Println("More than one package in", dir)
		return
	}

	packageAst, ok := f["main"]
	if !ok {
		fmt.Println("No package main in", dir)
		return
	}

	ast.Walk(new(FuncVisitor), packageAst)
	ast.Walk(new(ReturnVisitor), packageAst)
	//ast.Print(fset, packageAst)

	buf := bytes.NewBuffer(nil)
	if err = format.Node(buf, fset, ast.MergePackageFiles(packageAst, ast.FilterFuncDuplicates)); err != nil {
		panic(err)
	}
	src := buf.String()
	src = src[:14] + "/*\n#include \"postgres.h\"\n#include \"fmgr.h\"\n*/\nimport \"C\"\n" + src[14:]
	print(src)

	//save to files
	tmp, err := ioutil.TempDir("", "build")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmp)
	ioutil.WriteFile(path.Join(tmp, "tmp.go"), []byte(src), os.ModePerm)
	//TODO write pl.go as main package

	//go build -v -buildmode=c-shared -o my_procedures.so my_procedures.go pl.go
	goBuild := exec.Command("go", "build", "-v", "-buildmode=c-shared", "-o", path.Base(dir)+".so")
	stderr, err := goBuild.StderrPipe()
	if err != nil {
		panic(err)
	}
	stdout, err := goBuild.StdoutPipe()
	if err != nil {
		panic(err)
	}
	err = goBuild.Start()
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stderr, stderr)
	io.Copy(os.Stdout, stdout)
}
