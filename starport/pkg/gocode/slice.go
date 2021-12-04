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

// Append returns the same Slice that receives this method.
//
// This function takes one of an integer, string, bool, dst.Expr, or
// gocode.Builder. Any other types will cause a panic
func (slice *Slice) Append(item interface{}) *Slice {
	return slice.AppendExpr(Item(item))
}

// Extend returns the same Slice that it receives after extending the Slice by
// appending each argument to it.
func (slice *Slice) Extend(args ...interface{}) *Slice {
	for _, item := range args {
		slice.Append(item)
	}
	return slice
}

// AppendExpr returns the received Slice after appending raw dst.Expr to the
// inner composite literal
func (slice *Slice) AppendExpr(expr dst.Expr, exprs ...dst.Expr) *Slice {
	slice.inner.Elts = append(append(slice.inner.Elts, expr), exprs...)
	return slice
}

func (slice *Slice) Node() *dst.CompositeLit {
	return slice.inner
}

func (slice *Slice) Build() dst.Expr {
	return slice.Node()
}
