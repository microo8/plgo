package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"os"
)

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
	//import "C"
	for _, file := range packageAst.Files {
		file.Decls = append([]ast.Decl{
			&ast.GenDecl{
				Doc: &ast.CommentGroup{
					List: []*ast.Comment{
						&ast.Comment{
							Text: "/*\n#include \"postgres.h\"\n#include \"fmgr.h\"\n*/",
						},
					},
				},
				Tok:   token.IMPORT,
				Specs: []ast.Spec{&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: "\"C\""}}},
			},
		}, file.Decls...)
		break
	}

	//TODO add this after all imports
	/*
		&ast.FuncDecl{
			Name: ast.NewIdent("main"),
			Type: &ast.FuncType{Params: &ast.FieldList{}},
			Body: &ast.BlockStmt{},
		},
	*/

	ast.Inspect(packageAst, func(n ast.Node) bool {
		//is exported function
		function, ok := n.(*ast.FuncDecl)
		if !ok || !ast.IsExported(function.Name.Name) {
			return true
		}

		params := function.Type.Params
		//TODO test param types, so that the type is supported string, int, float64 etc.

		//function params are just fcinfo *FuncInfo and return value is Datum
		function.Type.Params = &ast.FieldList{
			List: []*ast.Field{
				&ast.Field{
					Names: []*ast.Ident{ast.NewIdent("fcinfo")},
					Type:  ast.NewIdent("*FuncInfo"),
				},
			},
		}

		function.Type.Results = &ast.FieldList{
			List: []*ast.Field{
				&ast.Field{
					Type: ast.NewIdent("Datum"),
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

		return true
	})

	//all return statements wrapped by ToDatum()
	ast.Inspect(packageAst, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		ret.Results = []ast.Expr{
			&ast.CallExpr{
				Fun:  ast.NewIdent("ToDatum"),
				Args: ret.Results,
			},
		}
		return true
	})
	ast.Print(fset, packageAst)
	if err := printer.Fprint(os.Stdout, fset, ast.MergePackageFiles(packageAst, ast.FilterFuncDuplicates)); err != nil {
		panic(err)
	}
}
