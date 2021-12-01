package gocode

import (
	"errors"
	"fmt"

	"github.com/dave/dst"
)

var ErrMissingFunction = errors.New("could not locate function")

func FindFunction(tree *dst.File, name string) (*dst.FuncDecl, error) {
	for _, decl := range tree.Decls {
		if fn, ok := decl.(*dst.FuncDecl); ok && fn.Name.Name == name {
			return fn, nil
		}
	}
	return nil, fmt.Errorf("gocode find function: %w '%s'", ErrMissingFunction, name)
}
