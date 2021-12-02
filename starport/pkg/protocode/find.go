package protocode

import (
	"errors"
	"fmt"

	"github.com/emicklei/proto"
)

var (
	ErrImportNotFound  = errors.New("could not find import")
	ErrMessageNotFound = errors.New("could not find message")
)

func FindMessagef(tree *File, format string, args ...interface{}) (*proto.Message, error) {
	return FindMessage(tree, fmt.Sprintf(format, args...))
}

func FindMessage(tree *File, name string) (*proto.Message, error) {
	var node *proto.Message
	proto.Walk(tree, proto.WithMessage(func(message *proto.Message) {
		if message.Name == name {
			node = message
		}
	}))
	if node == nil {
		return nil, fmt.Errorf("%w: %q", ErrMessageNotFound)
	}
	return node, nil
}

func FindImport(tree *File, filename string) (*proto.Import, error) {
	var node *proto.Import
	proto.Walk(tree, proto.WithImport(func(importNode *proto.Import) {
		if importNode.Filename == filename {
			node = importNode
		}
	}))
	if node == nil {
		return nil, fmt.Errorf("%w: %q", ErrImportNotFound, filename)
	}
	return node, nil
}
