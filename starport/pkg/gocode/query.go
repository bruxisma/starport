package gocode

import "github.com/dave/dst"

func KeyAsIdentifier(kv *dst.KeyValueExpr) (string, bool) {
	ident, ok := kv.Key.(*dst.Ident)
	if ok {
		return ident.Name, ok
	}
	return "", ok
}

func ValueAsBasicLiteral(kv *dst.KeyValueExpr) (*dst.BasicLit, bool) {
	switch value := kv.Value.(type) {
	case *dst.UnaryExpr:
		literal, ok := value.X.(*dst.BasicLit)
		return literal, ok
	case *dst.BasicLit:
		return value, true
	default:
		return nil, false
	}
}

func ValueAsCompositeLiteral(kv *dst.KeyValueExpr) (*dst.CompositeLit, bool) {
	switch value := kv.Value.(type) {
	case *dst.UnaryExpr:
		literal, ok := value.X.(*dst.CompositeLit)
		return literal, ok
	case *dst.CompositeLit:
		return value, true
	default:
		return nil, false
	}
}
