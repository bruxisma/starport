package gocode

import (
	"go/token"
	"reflect"

	"github.com/dave/dst"
)

type Structure struct {
	inner *dst.CompositeLit
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

// KeyValue returns a dst.KeyValueExpr based on the give name, and the value
// provided.
//
// Due to how rune is an alias for Int32, we do not support BasicLit of kind
// token.Char
func KeyValue(name string, item interface{}) *dst.KeyValueExpr {
	var expr dst.Expr
	switch value := reflect.Indirect(reflect.ValueOf(item)); value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		expr = BasicInt(value.Int())
	case reflect.String:
		expr = BasicString(value.String())
	case reflect.Bool:
		expr = Name("%t", value.Bool())
	default:
		panic("Not Yet Implemented")
	}
	return &dst.KeyValueExpr{
		Decs:  KVDecs,
		Key:   Identifier(name),
		Value: expr,
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
