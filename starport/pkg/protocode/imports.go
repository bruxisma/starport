package protocode

import (
	"fmt"

	"github.com/emicklei/proto"
)

func AppendImportf(tree *File, format string, args ...interface{}) (*File, error) {
	return AppendImport(tree, fmt.Sprintf(format, args...))
}

func AppendImport(tree *File, filename string) (*File, error) {
	of := NewOrganizedFile(tree)
	of.Imports = append(of.Imports, createImportNode(tree, filename))
	return of.AsFile(), nil
}

func PrependImport(tree *File, filename string) (*File, error) {
	nodes := []*proto.Import{createImportNode(tree, filename)}
	of := NewOrganizedFile(tree)
	of.Imports = append(nodes, of.Imports...)
	return of.AsFile(), nil
}

func createImportNode(parent proto.Visitee, filename string) *proto.Import {
	return &proto.Import{Parent: parent, Filename: filename}
}
