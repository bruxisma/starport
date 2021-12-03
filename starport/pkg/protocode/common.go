package protocode

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/emicklei/proto"
	"github.com/emicklei/proto-contrib/pkg/protofmt"
)

var ErrMissingPackage = errors.New("'package' statement missing")

func Comment(text ...string) *proto.Comment {
	return &proto.Comment{Lines: text}
}

func Commentf(format string, args ...interface{}) *proto.Comment {
	return Comment(fmt.Sprintf(format, args...))
}

func False() proto.Literal {
	return proto.Literal{Source: "false"}
}

func String(format string, args ...interface{}) proto.Literal {
	return proto.Literal{
		IsString: true,
		Source:   fmt.Sprintf(format, args...),
	}
}

func Parse(reader io.Reader, filename string) (*File, error) {
	parser := proto.NewParser(reader)
	parser.Filename(filename)
	p, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	return NewFile(p), nil
}

func Write(tree *File) (io.Reader, error) {
	buffer := &bytes.Buffer{}
	if err := Fprint(buffer, tree); err != nil {
		return nil, err
	}
	return buffer, nil
}

func Fprint(w io.Writer, root *File) error {
	formatter := protofmt.NewFormatter(w, "\t")
	formatter.Format(root.Proto())
	return nil
}
