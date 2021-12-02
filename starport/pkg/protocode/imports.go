package protocode

import (
	"errors"
	"fmt"

	"github.com/emicklei/proto"
)

func AppendImportf(tree *File, format string, args ...interface{}) (*File, error) {
	return AppendImport(tree, fmt.Sprintf(format, args...))
}

// AppendImport returns a modified protocode.File where the given filename is
// appended to the list of imports.
func AppendImport(tree *File, filename string) (*File, error) {
	node, err := FindImport(tree, filename)

	if err != nil && !errors.Is(err, ErrImportNotFound) {
		return nil, err
	} else if node != nil {
		return tree, nil
	}

	for idx := len(tree.Elements) - 1; idx != 0; idx-- {
		if _, ok := tree.Elements[idx].(*proto.Import); ok {
			idx++
			importNode := createImportNode(tree, filename)
			tree.Elements = append(tree.Elements[:idx+1], tree.Elements[idx:]...)
			tree.Elements[idx] = importNode
			break
		}
	}
	return tree, nil
}

func createImportNode(parent proto.Visitee, filename string) *proto.Import {
	return &proto.Import{Parent: parent, Filename: filename}
}
