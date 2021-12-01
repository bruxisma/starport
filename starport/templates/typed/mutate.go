package typed

import (
	"fmt"
	"go/token"
	"strconv"

	"github.com/dave/dst"
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

	node, err := findGenericDeclaration(tree, token.IMPORT)
	if err != nil {
		return nil, err
	}

	for _, specs := range node.Specs {
		imp, ok := specs.(*dst.ImportSpec)
		if !ok {
			continue
		}
		if imp.Path.Value == strconv.Quote(value) {
			return tree, nil
		}
	}
	node.Specs = append(node.Specs, &dst.ImportSpec{
		Name: dst.NewIdent(name),
		Path: &dst.BasicLit{
			Kind:  token.STRING,
			Value: strconv.Quote(value),
		},
	})
	return tree, nil
}
