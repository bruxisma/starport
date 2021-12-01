package gocode

import (
	"fmt"
	"go/token"

	"github.com/dave/dst"
)

// Assignment is used as an AssignStmt node builder, and can be used to assign
// or define variables with some right hand expression.
type Assignment struct {
	inner *dst.AssignStmt
}

// Assign returns an Assignment where the LHS has been initialized with the
// provided expressions.
func Assign(expr dst.Expr, exprs ...dst.Expr) *Assignment {
	lhs := []dst.Expr{}
	lhs = append(append(lhs, expr), exprs...)
	return &Assignment{
		inner: &dst.AssignStmt{Lhs: lhs, Tok: token.ASSIGN},
	}
}

// Assignf returns an Assignment where the LHS has been initialized with an
// Identifier created from the given format string and values.
func Assignf(format string, args ...interface{}) *Assignment {
	return AssignVariable(fmt.Sprintf(format, args...))
}

// AssignVariable returns an Assignment where the LHS has been initialized with
// either an identifier or a "selector" expression (e.g., x.y)
func AssignVariable(name string, fields ...string) *Assignment {
	return Assign(Identifier(name, fields...))
}

// AssignCheck returns an Assignment where the LHS has been initialized with
// the provided identifier, and an identifier named 'err'.
//
// This is most useful when constructing an assignment equal to
//
//    variable, err = <expression>
//
func AssignCheck(name string) *Assignment {
	return Assign(Identifier(name), Identifier("err"))
}

// AssignIndex returns an Assignment where the provided collection and index
// expressions are used to initialize the LHS with an IndexExpr
func AssignIndex(collection dst.Expr, index dst.Expr) *Assignment {
	return Assign(&dst.IndexExpr{X: collection, Index: index})
}

// Define returns an Assignment where the LHS has been initialized with the
// provided expressions, and the := operator is used.
func Define(expr dst.Expr, exprs ...dst.Expr) *Assignment {
	node := Assign(expr, exprs...)
	node.inner.Tok = token.DEFINE
	return node
}

// Definef returns an Assignment where the LHS has been initialized with an
// Identifier created from the given format string and arguments, and the :=
// operator is used.
func Definef(format string, args ...interface{}) *Assignment {
	return DefineVariable(fmt.Sprintf(format, args...))
}

// DefineVariable returns an Assignment where the LHS has been initialized with
// the provided name as an Identifier.
func DefineVariable(name string) *Assignment {
	return Define(Identifier(name))
}

// DefineCheck returns an Assignment where the LHS has been initialized with
// the provided identifier, and an identifier named 'err'.
//
// This is most useful when constructing an assignment equal to
//
//    variable, err := <expression>
//
func DefineCheck(name string) *Assignment {
	return Define(Identifier(name), Identifier("err"))
}

// To returns an AssignStmt AST node after setting the RHS of an assignment to
// the provided expressions.
func (assignment *Assignment) To(expr dst.Expr, exprs ...dst.Expr) *dst.AssignStmt {
	rhs := []dst.Expr{}
	rhs = append(append(rhs, expr), exprs...)
	assignment.inner.Rhs = rhs
	return assignment.inner
}

func (assignment *Assignment) addTarget(name string) {
	assignment.inner.Lhs = append(assignment.inner.Lhs, Identifier(name))
}
