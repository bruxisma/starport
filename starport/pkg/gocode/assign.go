package gocode

import (
	"fmt"
	"go/token"

	"github.com/dave/dst"
)

type Assignment struct {
	inner *dst.AssignStmt
}

func Assign(expr dst.Expr, exprs ...dst.Expr) *Assignment {
	lhs := []dst.Expr{}
	lhs = append(append(lhs, expr), exprs...)
	return &Assignment{
		inner: &dst.AssignStmt{Lhs: lhs, Tok: token.ASSIGN},
	}
}

func Assignf(format string, args ...interface{}) *Assignment {
	return AssignVariable(fmt.Sprintf(format, args...))
}

func AssignVariable(name string, fields ...string) *Assignment {
	return Assign(Identifier(name, fields...))
}

func AssignCheck(name string) *Assignment {
	return Assign(Identifier(name), Identifier("err"))
}

func AssignIndex(collection dst.Expr, index dst.Expr) *Assignment {
	return Assign(&dst.IndexExpr{X: collection, Index: index})
}

func Define(expr dst.Expr, exprs ...dst.Expr) *Assignment {
	node := Assign(expr, exprs...)
	node.inner.Tok = token.DEFINE
	return node
}

func Definef(format string, args ...interface{}) *Assignment {
	return DefineVariable(fmt.Sprintf(format, args...))
}

func DefineVariable(name string) *Assignment {
	return Define(Identifier(name))
}

func DefineCheck(name string) *Assignment {
	return Define(Identifier(name), Identifier("err"))
}

func (assignment *Assignment) To(expr dst.Expr, exprs ...dst.Expr) *dst.AssignStmt {
	rhs := []dst.Expr{}
	rhs = append(append(rhs, expr), exprs...)
	assignment.inner.Rhs = rhs
	return assignment.inner
}

func (assignment *Assignment) addTarget(name string) {
	assignment.inner.Lhs = append(assignment.inner.Lhs, Identifier(name))
}
