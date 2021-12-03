package typed

import (
	"fmt"
	"go/token"

	"github.com/dave/dst"
	"github.com/tendermint/starport/starport/pkg/gocode"
)

func findGenericDeclaration(tree *dst.File, kind token.Token) (*dst.GenDecl, error) {
	for _, decl := range tree.Decls {
		node, ok := decl.(*dst.GenDecl)
		if ok && node.Tok == kind {
			return node, nil
		}
	}
	return nil, fmt.Errorf("could not locate generic declaration %v", kind)
}

// MutateImport will add the given path (or name, path combination) to the
// golang AST
func MutateImport(tree *dst.File, values ...string) (*dst.File, error) {
	var value string
	var name string

	switch len(values) {
	case 1:
		value = values[0]
	case 2:
		name, value = values[0], values[1]
	default:
		return nil, fmt.Errorf("MutateImport takes 1 or 2 string arguments. Received %d", len(values))
	}

	imports := []*dst.ImportSpec{}
	gocode.Apply(tree, func(cursor *gocode.Cursor) bool {
		if decl, ok := cursor.Node().(*dst.GenDecl); ok && decl.Tok == token.IMPORT {
			for _, spec := range decl.Specs {
				imports = append(imports, spec.(*dst.ImportSpec))
			}
			cursor.Delete()
			return false
		}
		return true
	})

	imports = append(imports, &dst.ImportSpec{
		Name: gocode.Name(name),
		Path: gocode.BasicString(value),
	})

	specs := []dst.Spec{}
	for _, spec := range imports {
		specs = append(specs, spec)
	}

	decl := &dst.GenDecl{
		Tok:   token.IMPORT,
		Specs: specs,
	}

	tree.Decls = append([]dst.Decl{decl}, tree.Decls...)
	tree.Imports = imports

	return tree, nil
}
