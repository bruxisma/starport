package gocode

import (
	"fmt"

	"github.com/dave/dst"
)

// SliceBuilder is used to construct a slice literal
type SliceBuilder struct {
	inner *dst.CompositeLit
}

func Slicef(format string, args ...interface{}) *SliceBuilder {
	return SliceOf(fmt.Sprintf(format, args...))
}

func SliceOf(name string) *SliceBuilder {
	return &SliceBuilder{
		inner: &dst.CompositeLit{
			Type: &dst.ArrayType{
				Elt: Identifier(name),
			},
			Elts: []dst.Expr{},
		},
	}
}

// Append returns the same Slice that receives this method.
//
// This function takes one of an integer, string, bool, dst.Expr, or
// gocode.Builder. Any other types will cause a panic
func (slice *SliceBuilder) Append(item interface{}) *SliceBuilder {
	return slice.AppendExpr(Item(item))
}

// Extend returns the same Slice that it receives after extending the Slice by
// appending each argument to it.
func (slice *SliceBuilder) Extend(args ...interface{}) *SliceBuilder {
	for _, item := range args {
		slice.Append(item)
	}
	return slice
}

// AppendExpr returns the received Slice after appending raw dst.Expr to the
// inner composite literal
func (slice *SliceBuilder) AppendExpr(expr dst.Expr, exprs ...dst.Expr) *SliceBuilder {
	slice.inner.Elts = append(append(slice.inner.Elts, expr), exprs...)
	return slice
}

func (slice *SliceBuilder) Node() *dst.CompositeLit {
	return slice.inner
}

func (slice *SliceBuilder) Build() dst.Expr {
	return slice.Node()
}
