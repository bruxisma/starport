package gocode

import "github.com/dave/dst"

type Function struct {
	inner *dst.FuncLit
}

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

func (function *Function) Parameter(name *dst.Ident, typename dst.Expr) *Function {
	params := function.inner.Type.Params
	params.List = append(params.List, &dst.Field{
		Names: []*dst.Ident{name},
		Type:  typename,
	})
	return function
}

func (function *Function) Do(body func(*Block)) *Function {
	body(&Block{function.inner.Body})
	return function
}

func (function *Function) Build() dst.Expr {
	if len(function.inner.Type.Results.List) >= 2 {
		function.inner.Type.Results.Opening = true
		function.inner.Type.Results.Closing = true
	}
	return function.inner
}
