package gocode

import "github.com/dave/dst"

// Function represents a function literal (i.e., an anonymous function declared
// in place)
type Function struct {
	inner *dst.FuncLit
}

// Func returns a Function with an empty body, and no parameters
func Func() *Function {
	return &Function{
		inner: &dst.FuncLit{
			Type: &dst.FuncType{
				Params:  &dst.FieldList{Opening: true, Closing: true},
				Results: &dst.FieldList{},
				Func:    true,
			},
			Body: &dst.BlockStmt{},
		},
	}
}

// Parameter takes an identifier and typename to add a parameter to the
// function's argument list. It then returns the Function that received this
// method
func (function *Function) Parameter(name *dst.Ident, typename dst.Expr) *Function {
	params := function.inner.Type.Params
	params.List = append(params.List, &dst.Field{
		Names: []*dst.Ident{name},
		Type:  typename,
	})
	return function
}

// Do takes a function that operates on a Block to add statements to the
// function's body
func (function *Function) Do(body func(*Block)) *Function {
	body(&Block{function.inner.Body})
	return function
}

// Build returns a dst.Expr as part of implementing the Builder interface
func (function *Function) Build() dst.Expr {
	if len(function.inner.Type.Results.List) >= 2 {
		function.inner.Type.Results.Opening = true
		function.inner.Type.Results.Closing = true
	}
	return function.inner
}
