package protocode

import (
	"fmt"

	"github.com/emicklei/proto"
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
	return node, nil
}
