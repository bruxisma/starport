package ast

import (
	"fmt"

	"github.com/dave/dst"
)

// FindFunction returns the FuncDecl with the name provided.
func FindFunction(tree *dst.File, name string) (*dst.FuncDecl, error) {
	for _, decl := range tree.Decls {
		fn, ok := decl.(*dst.FuncDecl)
		if ok && fn.Name.Name == name {
			return fn, nil
		}
	}
	return nil, fmt.Errorf("%w '%s'", ErrMissingFunction, name)
}
