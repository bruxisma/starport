package gocode

import (
	"go/token"

	"github.com/dave/dst"
)

type IfStatement struct {
	inner *dst.IfStmt
}

func IfVar(name string, fields ...string) *IfStatement {
	return If(Identifier(name, fields...))
}

func If(expr dst.Expr) *IfStatement {
	return &IfStatement{
		inner: &dst.IfStmt{Cond: expr},
	}
}

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

func (is *IfStatement) IsTrue() *IfStatement {
	switch is.condition().(type) {
	case *dst.Ident, *dst.ParenExpr:
		return is
	default:
		return is.IsEqualTo(Identifier("true"))
	}
}

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

func (is *IfStatement) IsGreaterOrEqualToVar(name string, fields ...string) *IfStatement {
	return is.IsGreaterOrEqualTo(Identifier(name, fields...))
}

func (is *IfStatement) IsGreaterOrEqualToVarf(format string, args ...interface{}) *IfStatement {
	return is.IsGreaterOrEqualTo(Name(format, args...))
}

func (is *IfStatement) condition() dst.Expr {
	return is.inner.Cond
}
