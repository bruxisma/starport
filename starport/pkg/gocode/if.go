package gocode

import (
	"go/token"

	"github.com/dave/dst"
)

// IfStatement is used to represent the beginning of an if-else chain.
type IfStatement struct {
	inner *dst.IfStmt
}

// IfVar returns an IfStatement used for checking a single identifier or
// selector as the start of the condition of the IfStatement.
func IfVar(name string) *IfStatement {
	return If(Identifier(name))
}

// If returns an IfStatement for checking the provided dst.Expr as the
// condition
func If(expr dst.Expr) *IfStatement {
	return &IfStatement{
		inner: &dst.IfStmt{Cond: expr},
	}
}

// Then is used to provide a body for the IfStatement
func (is *IfStatement) Then(body func(*Block)) *IfStatement {
	block := &Block{
		inner: &dst.BlockStmt{
			List: []dst.Stmt{},
		},
	}
	body(block)
	is.inner.Body = block.inner
	return is
}

// IsTrue checks the current condition of the IfStatement and if possible,
// will change it to be <current-condition> == true
func (is *IfStatement) IsTrue() *IfStatement {
	switch is.condition().(type) {
	case *dst.Ident, *dst.ParenExpr:
		return is
	default:
		return is.IsEqualTo(True())
	}
}

// IsEqualTo will take the provided dst.Expr and turns the condition into
// <current> == <expr>
func (is *IfStatement) IsEqualTo(expr dst.Expr) *IfStatement {
	switch value := is.condition().(type) {
	case *dst.Ident, *dst.ParenExpr:
		is.inner.Cond = &dst.BinaryExpr{X: value, Op: token.EQL, Y: expr}
	default:
		is.inner.Cond = &dst.ParenExpr{X: value}
		return is.IsEqualTo(expr)
	}
	return is
}

// IsGreaterOrEqualTo works like IsEqualTo but turns the condition into
// <current> >= <expr>
func (is *IfStatement) IsGreaterOrEqualTo(expr dst.Expr) *IfStatement {
	switch value := is.condition().(type) {
	case *dst.Ident, *dst.SelectorExpr, *dst.ParenExpr:
		is.inner.Cond = &dst.BinaryExpr{X: value, Op: token.GEQ, Y: expr}
	default:
		is.inner.Cond = &dst.ParenExpr{X: value}
		return is.IsGreaterOrEqualTo(expr)
	}
	return is
}

// IsGreaterOrEqualToVar is like IsGreaterOrEqualTo, but the right hand side is a
// selector or identifier
func (is *IfStatement) IsGreaterOrEqualToVar(name string) *IfStatement {
	return is.IsGreaterOrEqualTo(Identifier(name))
}

// IsGreaterOrEqualToVarf is like IsGreaterOrEqualToVar, but uses the provided
// format specificer to create an identifier
func (is *IfStatement) IsGreaterOrEqualToVarf(format string, args ...interface{}) *IfStatement {
	return is.IsGreaterOrEqualTo(Name(format, args...))
}

func (is *IfStatement) condition() dst.Expr {
	return is.inner.Cond
}
