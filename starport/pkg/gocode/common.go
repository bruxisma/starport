package gocode

import (
	"fmt"
	"go/token"
	"strconv"

	"github.com/dave/dst"
)

var (
	Newlines      = dst.NodeDecs{Before: dst.NewLine, After: dst.NewLine}
	EmptyLines    = dst.NodeDecs{Before: dst.EmptyLine, After: dst.EmptyLine}
	CompositeDecs = dst.CompositeLitDecorations{NodeDecs: Newlines}
	KVDecs        = dst.KeyValueExprDecorations{NodeDecs: Newlines}
)

type Expression interface {
	AsExpression() dst.Expr
}

type Statement interface {
	AsStatement() dst.Stmt
}

type MakeMap struct {
	inner *dst.MapType
}

func BasicString(value string) *dst.BasicLit {
	return &dst.BasicLit{
		Kind:  token.STRING,
		Value: strconv.Quote(value),
	}
}

func BasicInt(value int64) *dst.BasicLit {
	return &dst.BasicLit{
		Kind:  token.INT,
		Value: fmt.Sprintf("%d", value),
	}
}

func MakeSliceOf(name string, fields ...string) *dst.CallExpr {
	arrayType := &dst.ArrayType{Elt: Identifier(name, fields...)}
	return Call("make").WithParameters(arrayType).Node()
}

func MakeMapOf(name string, fields ...string) *MakeMap {
	return &MakeMap{
		inner: &dst.MapType{
			Key: Identifier(name, fields...),
		},
	}
}

func (mm *MakeMap) WithIndexOf(name string, fields ...string) *dst.CallExpr {
	mm.inner.Value = Identifier(name, fields...)
	return Call("make").WithParameters(mm.inner).Node()
}

// Identifier returns either a dst.Ident or a dst.SelectorExpr
//
// This is how non-package qualified names are generated, and how access to
// struct fields are created.
func Identifier(name string, fields ...string) dst.Expr {
	if len(fields) == 0 {
		return dst.NewIdent(name)
	}
	return selector(name, fields[0], fields[1:]...)
}

func Name(format string, args ...interface{}) *dst.Ident {
	return dst.NewIdent(fmt.Sprintf(format, args...))
}

func False() *dst.Ident {
	return dst.NewIdent("false")
}

func True() *dst.Ident {
	return dst.NewIdent("true")
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
