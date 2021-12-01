package gocode

import (
	"fmt"

	"github.com/dave/dst"
)

type FunctionCall struct {
	inner *dst.CallExpr
}

// Call is used to construct a CallStatement with the provided identifier.
//
// Additional arguments can be added with the FunctionCall.WithParameters or
// FunctionCall.WithParameter function.
func Call(name string, fields ...string) *FunctionCall {
	return &FunctionCall{
		inner: &dst.CallExpr{
			Fun:  Identifier(name, fields...),
			Args: []dst.Expr{},
		},
	}
}

func (fc *FunctionCall) WithParameters(exprs ...dst.Expr) *FunctionCall {
	fc.inner.Args = append(fc.inner.Args, exprs...)
	return fc
}

func (fc *FunctionCall) WithArgument(name string, fields ...string) *FunctionCall {
	return fc.WithParameters(Identifier(name, fields...))
}

func (fc *FunctionCall) WithString(text string) *FunctionCall {
	return fc.WithParameters(BasicString(text))
}

func (fc *FunctionCall) WithVariadicArgument(name string, fields ...string) *dst.CallExpr {
	return fc.WithVariadicExpression(Identifier(name, fields...))
}

func (fc *FunctionCall) WithVariadicExpression(expr dst.Expr) *dst.CallExpr {
	fc.WithParameters(expr)
	fc.inner.Ellipsis = true
	return fc.Node()
}

func (fc *FunctionCall) AsStatement() *dst.ExprStmt {
	return &dst.ExprStmt{X: fc.inner}
}

func (fc *FunctionCall) PrependComment(format string, args ...interface{}) *FunctionCall {
	fc.inner.Decorations().Start.Append("//%s\n", fmt.Sprintf(format, args))
	return fc
}

func (fc *FunctionCall) Node() *dst.CallExpr {
	return fc.inner
}
