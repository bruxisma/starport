package gocode

import (
	"fmt"
	"go/token"

	"github.com/dave/dst"
)

type RangeStatement struct {
	inner *dst.RangeStmt
}

// ForEachItem is used to construct a range statement where the key, value is
// _, <value>.
//
// This is mostly used for generating statements that iterate over slices
// directly, or maps where the key is ignored.
func ForEachItem(value string) *RangeStatement {
	rs := &RangeStatement{
		inner: &dst.RangeStmt{
			Tok: token.DEFINE,
			Body: &dst.BlockStmt{
				List: []dst.Stmt{},
			},
		},
	}
	return rs.Value(value).Key("_")
}

// ForEach is used to construct a range statement where the key and value are
// the ones provided in the function call.
//
// This is mostly used when generating range loops over maps or where knowing
// the index of a slice is necessary.
func ForEach(key, value string) *RangeStatement {
	return ForEachItem(value).Key(key)
}

// In returns RangeStatement as part of the builder pattern and is used to
// construct a range over some identifier or selector.
//
//This should be used when the name is some identifier or selector expression.
//More complex range loops should use RangeStatement.Of to pass a manually
//constructed expression.
func (rs *RangeStatement) In(name string, fields ...string) *RangeStatement {
	return rs.Of(Identifier(name, fields...))
}

// Of returns RangeStatement as part of the builder pattern and is used to
// construct a range over some arbitrary expression.
//
// NOTE: Javascript uses `for ... of` for ranges of data, and thus we chose the
// same name here for familiarity for the more common set of programmers out
// there. If golang were to allow basic function overloading, this wouldn't be
// an issue in the first place.
func (rs *RangeStatement) Of(expr dst.Expr) *RangeStatement {
	rs.inner.X = expr
	return rs
}

// Do returns RangeStatement as part of the builder pattern and is used to set the body of the range statement directly.
//
// This function does not take a Block directly, but instead takes a callback
// to more closely mimic the blocks a user would write with braces when writing
// a for range loop in golang. To pass a Block directly, call
// RangeStatement.With
func (rs *RangeStatement) Do(body func(*Block)) *RangeStatement {
	block := &Block{
		inner: &dst.BlockStmt{
			List: []dst.Stmt{},
		},
	}
	body(block)
	return rs.With(block)
}

// With returns RangeStatement as part of the builder pattern and is used to
// set the body of the range statement directly.
//
// This function is used to set pre-existing Block statements as the body of a
// range loop.
func (rs *RangeStatement) With(block *Block) *RangeStatement {
	rs.inner.Body = block.inner
	return rs
}

// Value sets the Value member of the inner dst.RangeStmt
func (rs *RangeStatement) Value(name string, fields ...string) *RangeStatement {
	rs.inner.Value = Identifier(name, fields...)
	return rs
}

func (rs *RangeStatement) Key(name string, fields ...string) *RangeStatement {
	rs.inner.Key = Identifier(name, fields...)
	return rs
}

// Assignment changes the form of assignment for the range statement to
// assignment instead of definition.
//
// By default, all RangeStatements will use definition assignment, over
// variable assignment
func (rs *RangeStatement) Assignment() *RangeStatement {
	rs.inner.Tok = token.ASSIGN
	return rs
}

func (rs *RangeStatement) PrependComment(format string, args ...interface{}) *RangeStatement {
	rs.inner.Decorations().Start.Append("//%s\n", fmt.Sprintf(format, args))
	return rs
}

// Node returns the inner dst.RangeStmt that has been constructed, and marks
// that the range statement is no longer being modified.
//
// NOTE: The RangeStmt is currently returned by reference, and is not cloned in
// a deep manner. This might change in the future to reduce users accidentally
// mutating a RangeStmt after calling Done()
func (rs *RangeStatement) Node() *dst.RangeStmt {
	return rs.inner
}

func (rs *RangeStatement) Done() *dst.RangeStmt {
	return rs.Node()
}
