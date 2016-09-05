package main

import (
	"flag"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
)

var (
	filename = flag.String("f", "procedures.go", "go source files with stored procedures")
)

func main() {
	flag.Parse()

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, *filename, nil, 0)
	if err != nil {
		panic(err)
	}
	ast.Inspect(f, func(n ast.Node) bool {
		//is exported function
		if function, ok := n.(*ast.FuncDecl); ok && ast.IsExported(function.Name.Name) {
			params := function.Type.Params

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
			for i, param := range params.List {
				for j, name := range param.Names {
					log.Println(i, j, function.Name.Name, name.Name)
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

			//err := fcinfo.Scan
			function.Body.List = append([]ast.Stmt{
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
				}}, function.Body.List...)

			//if err panic TODO

			//prepend all parameter variables
			function.Body.List = append(paramDecs, function.Body.List...)

		}
		return true
	})
	ast.Print(fset, f)
	if err := format.Node(os.Stdout, fset, f); err != nil {
		panic(err)
	}
}
