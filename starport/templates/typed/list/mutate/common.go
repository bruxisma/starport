package mutate

import (
	"io"
	"reflect"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/tendermint/starport/starport/pkg/protocode"
	"github.com/tendermint/starport/starport/templates/typed"
)

type (
	// ProtoASTModifier is used to apply mutations to a protocode.File (a protobuf AST root)
	ProtoASTModifier func(*protocode.File, *typed.Options) (*protocode.File, error)
	// GoASTModifier is used to apply mutations to a dst.File (a golang AST root)
	GoASTModifier func(*dst.File, *typed.Options) (*dst.File, error)
	// ProtoSequence is a Slice of ProtoASTModifier
	ProtoSequence []ProtoASTModifier
	// GoSequence is a Slice of GoASTModifier
	GoSequence []GoASTModifier
)

var (
	structType       = reflect.TypeOf((*dst.StructType)(nil))
	compositeLitType = reflect.TypeOf((*dst.CompositeLit)(nil))
)

// Apply executes each mutator on the provided root AST node and Options value
func (sequence GoSequence) Apply(source io.Reader, opts *typed.Options) (*dst.File, error) {
	tree, err := decorator.Parse(source)
	if err != nil {
		return nil, err
	}
	for _, modifier := range sequence {
		if tree, err = modifier(tree, opts); err != nil {
			return nil, err
		}
	}
	return tree, nil
}
