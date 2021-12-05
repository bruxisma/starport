package gocode

import (
	"fmt"
	"go/token"
	"strconv"
	"strings"

	"github.com/dave/dst"
)

var (
	Newlines      = dst.NodeDecs{Before: dst.NewLine, After: dst.NewLine}
	EmptyLines    = dst.NodeDecs{Before: dst.EmptyLine, After: dst.EmptyLine}
	CompositeDecs = dst.CompositeLitDecorations{NodeDecs: Newlines}
	KVDecs        = dst.KeyValueExprDecorations{NodeDecs: Newlines}
)

// Builder is defined as any type in the gocode package that is used with the
// so-called (and aptly named) Builder pattern. This is primarily used
// internally by gocode, but is still available to users in cases where it
// might be necessary
type Builder interface {
	Build() dst.Expr
}

type Expression interface {
	AsExpression() dst.Expr
}

type Statement interface {
	AsStatement() dst.Stmt
}

type MakeMap struct {
	inner *dst.MapType
}

// BasicStringf returns a basic string literal from the format specifiers
func BasicStringf(format string, args ...interface{}) *dst.BasicLit {
	return BasicString(fmt.Sprintf(format, args...))
}

// BasicString returns a basic string literal
func BasicString(value string) *dst.BasicLit {
	return &dst.BasicLit{
		Kind:  token.STRING,
		Value: strconv.Quote(value),
	}
}

// BasicInt returns a basic integer literal
func BasicInt(value int64) *dst.BasicLit {
	return &dst.BasicLit{
		Kind:  token.INT,
		Value: fmt.Sprintf("%d", value),
	}
}

// MakeSliceOf returns a CallExpr equivalent to `make([]name.fields)`
func MakeSliceOf(name string) *dst.CallExpr {
	arrayType := &dst.ArrayType{Elt: Identifier(name)}
	return Call("make").WithParameters(arrayType).Node()
}

// MakeMapOf returns a node builder that is used to construct a very basic call
// to `make(map[T]U)`
func MakeMapOf(name string) *MakeMap {
	return &MakeMap{
		inner: &dst.MapType{
			Key: Identifier(name),
		},
	}
}

// WithIndexOf returns a CallExpr where the CallExpr contains the provided
// values as an identifier or selector expression such that the CallExpr
// provided is equivalent to `make(map[T]U)`
func (mm *MakeMap) WithIndexOf(name string) *dst.CallExpr {
	mm.inner.Value = Identifier(name)
	return Call("make").WithParameters(mm.inner).Node()
}

// Identifier returns either a dst.Ident or a dst.SelectorExpr
//
// This is how non-package qualified names are generated, and how access to
// struct fields are created.
func Identifier(name string) dst.Expr {
	components := strings.Split(name, ".")
	if len(components) == 1 {
		return dst.NewIdent(name)
	}
	return selector(components[0], components[1], components[2:]...)
}

// Name returns a dst.Ident constructed from the formatted string.
//
// This function is mostly provided as a convenience.
func Name(format string, args ...interface{}) *dst.Ident {
	return dst.NewIdent(fmt.Sprintf(format, args...))
}

// False returns the `false` identifier
func False() *dst.Ident {
	return dst.NewIdent("false")
}

// True returns the `true` identifier
func True() *dst.Ident {
	return dst.NewIdent("true")
}

// Nil returns the `nil` keyword
func Nil() *dst.Ident {
	return dst.NewIdent("nil")
}

// Item returns a dst.Expr constructed from either an integer, string, boolean,
// dst.Expr, or Builder.
//
// This same function is used by Slice, and Structure, as well as KeyValue to
// convert values to valid Exprs. If an invalid parameter is passed in, Item
// will panic
func Item(item interface{}) dst.Expr {
	switch value := item.(type) {
	case int:
		return BasicInt(int64(value))
	case int8:
		return BasicInt(int64(value))
	case int16:
		return BasicInt(int64(value))
	case int32:
		return BasicInt(int64(value))
	case int64:
		return BasicInt(value)
	case string:
		return BasicString(value)
	case bool:
		return Name("%t", value)
	case Builder:
		return value.Build()
	case dst.Expr:
		return value
	default:
		panic(fmt.Sprintf("Expression conversion for '%[1]T' is not yet implemented: %#[1]v\n", item))
	}
}

// Uninitialized returns a *single* uninitialized variable declartion
func UninitializedVar(name, typename string) *dst.GenDecl {
	return &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names: []*dst.Ident{Name(name)},
				Type:  Identifier(typename),
			},
		},
	}
}

// TypedGlobal is used for creating ValueSpecs with an explicit type
func TypedGlobal(name, typename string, item interface{}) *dst.ValueSpec {
	spec := Global(name, item)
	spec.Type = Identifier(typename)
	return spec
}

// Global is used for construct ValueSpecs in a simpler way
func Global(name string, item interface{}) *dst.ValueSpec {
	return &dst.ValueSpec{
		Names:  []*dst.Ident{Name(name)},
		Values: []dst.Expr{Item(item)},
	}
}

// AddressOf takes a string, fmt.Stringer, or a dst.Expr and returns a UnaryExpr with
// the address of operator prepended. If a string is passed, it will converted
// to a  dst.Ident first
func AddressOf(item interface{}) *dst.UnaryExpr {
	switch value := item.(type) {
	case string:
		return AddressOf(Identifier(value))
	case dst.Expr:
		return &dst.UnaryExpr{Op: token.AND, X: value}
	case fmt.Stringer:
		return AddressOf(value.String())
	default:
		panic(fmt.Sprintf("gocode.AddressOf only takes a string or dst.Expr. Received %[1]T: %[1]v", item))
	}
}

// Errorf acts much like fmt.Errorf, except that it returns a *FunctionCall
// that turns *into* the provided string.
//
// NOTE: This function does not delay the format specifiers.
func Errorf(format string, args ...interface{}) *FunctionCall {
	return Call("fmt.Errorf").WithStringf(format, args...)
}

// selectors returns a SelectorExpr such that the parameters passed will create
// a selector of the form <name>.<field>[.<fields>]+
//
// This function can be a bit confusing as we have to construct a selector as
// each "parent" (the field is named 'X') along the way until the fields
// parameter is exhausted.
func selector(name, field string, fields ...string) *dst.SelectorExpr {
	selector := &dst.SelectorExpr{X: dst.NewIdent(name), Sel: dst.NewIdent(field)}
	for _, field := range fields {
		selector = &dst.SelectorExpr{X: selector, Sel: dst.NewIdent(field)}
	}
	return selector
}
