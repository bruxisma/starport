package ast

import (
	"bytes"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/jennifer/jen"
)

// IntoStmt works like IntoStmts, but only returns the first dst.Stmt
func IntoStmt(stmt *jen.Statement) (dst.Stmt, error) {
	stmts, err := IntoStmts(stmt)
	return stmts[0], err
}

// IntoStmts converts *most* jen.Statement trees into a []dst.Stmt slice
//
// It does this by creating a fake package, and inserting the statements inside
// of a dummy 'main' function, converting all of this into a string, parsing it
// with decorator.Parse, and then extracting the nodes directly.
//
// This function is necessary as the manual creation of AST nodes is
// *currently* very painful. However as time goes on, more constructs will most
// likely be used to reduce the amount of work needed to construct AST nodes
// directly.
func IntoStmts(stmt *jen.Statement) ([]dst.Stmt, error) {
	buffer := &bytes.Buffer{}
	file := jen.NewFile("main")
	file.Func().Id("main").Params().Block(stmt)
	if err := file.Render(buffer); err != nil {
		return nil, err
	}
	tree, err := decorator.Parse(buffer.String())
	if err != nil {
		return nil, err
	}
	fn, err := FindFunction(tree, "main")
	if err != nil {
		return nil, err
	}
	return fn.Body.List, nil
}
