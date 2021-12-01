package gocode

import "github.com/dave/dst"

type Slice struct {
	inner *dst.CompositeLit
}

func SliceOf(name string, fields ...string) *Slice {
	return &Slice{
		inner: &dst.CompositeLit{
			Type: &dst.ArrayType{
				Elt: Identifier(name, fields...),
			},
			Elts: []dst.Expr{},
		},
	}
}

func (slice *Slice) AppendExpr(expr dst.Expr, exprs ...dst.Expr) *Slice {
	slice.inner.Elts = append(append(slice.inner.Elts, expr), exprs...)
	return slice
}

func (slice *Slice) Node() *dst.CompositeLit {
	return slice.inner
}
