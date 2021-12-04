package gocode

import (
	"fmt"
	"go/token"
	"reflect"

	"github.com/dave/dst"
)

var functionCallType = reflect.TypeOf((*FunctionCall)(nil))

type Structure struct {
	inner *dst.CompositeLit
}

// Anonymous is an opaque type over a map[string]interface{} that implements
// Builder.
//
// This interface allows uses to write a map[string]interface{} literal that is
// then able to be passed into any API that takes a Builder (or with a call to
// .Build() a dst.Expr)
type Anonymous map[string]interface{}

func Structf(format string, args ...interface{}) *Structure {
	return Struct(fmt.Sprintf(format, args...))
}

func Struct(name string, fields ...string) *Structure {
	structure := AnonymousStruct()
	structure.inner.Type = Identifier(name, fields...)
	return structure
}

func AnonymousStruct() *Structure {
	return &Structure{
		inner: &dst.CompositeLit{Decs: CompositeDecs},
	}
}

func (structure *Structure) RemoveDecorations() *Structure {
	structure.inner.Decs = dst.CompositeLitDecorations{}
	return structure
}

func (structure *Structure) SetDecorations(decs dst.CompositeLitDecorations) *Structure {
	structure.inner.Decs = decs
	return structure
}

func (structure *Structure) AppendExpr(name string, field dst.Expr) *Structure {
	return structure.Append(&dst.KeyValueExpr{
		Decs:  KVDecs,
		Key:   Identifier(name),
		Value: field,
	})
}

func (structure *Structure) AppendField(name string, item interface{}) *Structure {
	return structure.Append(KeyValue(name, item))
}

func (structure *Structure) Append(kv *dst.KeyValueExpr) *Structure {
	structure.inner.Elts = append(structure.inner.Elts, kv)
	return structure
}

func (structure *Structure) Done() *dst.CompositeLit {
	return structure.inner
}

func (structure *Structure) AddressOf() *dst.UnaryExpr {
	return &dst.UnaryExpr{
		Op: token.AND,
		X:  structure.inner,
	}
}

func (structure *Structure) Build() dst.Expr {
	return structure.Done()
}

// Construct returns a Structure built from the fields placed inside of
// Anonymous
func (anonymous Anonymous) Construct() *Structure {
	structure := AnonymousStruct()
	for key, value := range anonymous {
		structure.AppendField(key, value)
	}
	return structure
}

// Build returns a dst.Expr constructed from the fields placed inside of
// Anonymous
func (anonymous Anonymous) Build() dst.Expr {
	return anonymous.Construct().Build()
}

// KeyValue returns a dst.KeyValueExpr based on the give name, and the value
// provided.
//
// Due to how rune is an alias for Int32, we do not support BasicLit of kind
// token.Char
func KeyValue(name string, item interface{}) *dst.KeyValueExpr {
	return &dst.KeyValueExpr{
		Decs:  KVDecs,
		Key:   Identifier(name),
		Value: Item(item),
	}
}

// KeyValues returns the key value expressions needed to initialize any
// map-like item, such as an anonymous struct, or as the value needed to
// initialize some explicit type as the sub-node.
func KeyValues(fields map[string]interface{}) *dst.CompositeLit {
	expressions := []dst.Expr{}
	for key, value := range fields {
		expressions = append(expressions, KeyValue(key, value))
	}
	return &dst.CompositeLit{
		Decs: CompositeDecs,
		Elts: expressions,
	}
}
