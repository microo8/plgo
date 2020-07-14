package main

import "go/ast"

const plgo = "plgo"

//FuncVisitor collects all definitions of exported functions in an packate
type FuncVisitor struct {
	err       error
	functions []CodeWriter
}

//Visit checks if the functions is exported and creates and Code object from it
func (v *FuncVisitor) Visit(node ast.Node) ast.Visitor {
	function, ok := node.(*ast.FuncDecl)
	if !ok || !ast.IsExported(function.Name.Name) {
		return v
	}
	var code CodeWriter
	code, v.err = NewCode(function)
	if v.err != nil {
		return nil
	}
	v.functions = append(v.functions, code)
	function.Name.Name = "__" + function.Name.Name
	return v
}

//Remover is an visitor that removes all plgo usages
type Remover struct{}

//Visit removes plgo selectors and plgo import
func (v *Remover) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.ImportSpec:
		if n.Path.Value == "\"github.com/paulhatch/plgo\"" {
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
