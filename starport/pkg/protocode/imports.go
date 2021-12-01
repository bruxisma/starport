package protocode

import (
	"fmt"

	"github.com/emicklei/proto"
)

func AppendImportf(tree *File, format string, args ...interface{}) (*File, error) {
	return AppendImport(tree, fmt.Sprintf(format, args...))
}

// AppendImport returns a modified protocode.File where the given filename is
// appended to the list of imports.
//
// NOTE: We do not currently use the organized file format, so quite a bit of
// manual work is involved, and to make the function easier to read, we
// "search" from the back.
func AppendImport(tree *File, filename string) (*File, error) {
	for idx := len(tree.Elements) - 1; idx != 0; idx-- {
		if _, ok := tree.Elements[idx].(*proto.Import); ok {
			idx++
			importNode := createImportNode(tree, filename)
			tree.Elements = append(tree.Elements[:idx+1], tree.Elements[idx:]...)
			tree.Elements[idx] = importNode
			// tree.Elements = append(append(tree.Elements[:idx+1], importNode), tree.Elements[idx:]...)
			break
		}
	}
	return tree, nil
}

func createImportNode(parent proto.Visitee, filename string) *proto.Import {
	return &proto.Import{Parent: parent, Filename: filename}
}
