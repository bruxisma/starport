package gocode

import (
	"fmt"

	"github.com/dave/dst"
)

// FunctionCall is used as a CallExpr node builder, and can be used to call a
// function as either an expression OR as a statement by using the
// FunctionCall.AsStatement function.
type FunctionCall struct {
	inner *dst.CallExpr
}

func Callf(format string, args ...interface{}) *FunctionCall {
	return Call(fmt.Sprintf(format, args...))
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

// WithParameters returns the received FunctionCall after appending the given
// expressions as arguments to the function.
func (fc *FunctionCall) WithParameters(exprs ...dst.Expr) *FunctionCall {
	fc.inner.Args = append(fc.inner.Args, exprs...)
	return fc
}

func (fc *FunctionCall) WithArgumentf(format string, args ...interface{}) *FunctionCall {
	return fc.WithArgument(fmt.Sprintf(format, args...))
}

// WithArgument returns the received FunctionCall after appending the given
// values as either an identifier or selector expression.
func (fc *FunctionCall) WithArgument(name string, fields ...string) *FunctionCall {
	return fc.WithParameters(Identifier(name, fields...))
}

// WithArgument returns the received FunctionCall after appending the given
// argument as a string literal.
func (fc *FunctionCall) WithString(text string) *FunctionCall {
	return fc.WithParameters(BasicString(text))
}

// WithVariadicArgument returns the constructed CallExpr after appending the
// given argument and adding an ellipsis to it.
//
// NOTE: This function returns a CallExpr because variadic arguments are the
// last parameter in a function.
func (fc *FunctionCall) WithVariadicArgument(name string, fields ...string) *dst.CallExpr {
	return fc.WithVariadicExpression(Identifier(name, fields...))
}

// WithVariadicExpression returns the constructed CallExpr after appending the
// given argument and adding an ellipsis to it.
//
// NOTE: This function returns a CallExpr because variadic arguments are the
// last parameter permitted in a function.
func (fc *FunctionCall) WithVariadicExpression(expr dst.Expr) *dst.CallExpr {
	fc.WithParameters(expr)
	fc.inner.Ellipsis = true
	return fc.Node()
}

// AsStatement returns an ExprStmt containing the constructed function call for
// use in places where statements, and not expressions, are expected.
func (fc *FunctionCall) AsStatement() *dst.ExprStmt {
	return &dst.ExprStmt{X: fc.inner}
}

// PrependCOmment returns the received FunctionCall after prepending the given
// format string as a single line comment.
func (fc *FunctionCall) PrependComment(format string, args ...interface{}) *FunctionCall {
	fc.inner.Decorations().Start.Append("//%s\n", fmt.Sprintf(format, args...))
	return fc
}

// Node returns the CallExpr, signalling that the FunctionCall builder is
// complete.
func (fc *FunctionCall) Node() *dst.CallExpr {
	return fc.inner
}

// Build retrns a dst.Expr by calling Node().
func (fc *FunctionCall) Build() dst.Expr {
	return fc.Node()
}
