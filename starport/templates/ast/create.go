package ast

import (
	"github.com/dave/dst"
)

// TODO: Move *Walker types to traverse.go

type CompositeWalker struct {
	compare func(composite *dst.CompositeLit) bool
}

type KeyValueWalker struct {
	compare func(kv *dst.KeyValueExpr) bool
}

func Walk(node dst.Node, visitor dst.Visitor) {
	dst.Walk(visitor, node)
}

func WithCompositeFinder(compare func(*dst.CompositeLit) bool) *CompositeWalker {
	return &CompositeWalker{compare: compare}
}

func WithKeyValueFinder(compare func(*dst.KeyValueExpr) bool) *KeyValueWalker {
	return &KeyValueWalker{compare: compare}
}

func (walker *CompositeWalker) Visit(node dst.Node) dst.Visitor {
	// For some reason node *might* be nil from the ast API. This is a decision
	// beyond our control. Don't feel like checking if we can do a type insertion
	// on nil node
	if node == nil {
		return walker
	}
	composite, ok := node.(*dst.CompositeLit)
	if !ok || !walker.compare(composite) {
		return walker
	}
	return nil
}

func (walker *KeyValueWalker) Visit(node dst.Node) dst.Visitor {
	if node == nil {
		return walker
	}
	kv, ok := node.(*dst.KeyValueExpr)
	if !ok || !walker.compare(kv) {
		return walker
	}
	return nil
}
