package protocode

import (
	"bytes"
	"errors"
	"io"

	"github.com/emicklei/proto"
	"github.com/emicklei/proto-contrib/pkg/protofmt"
)

type File = proto.Proto

var ErrMissingPackage = errors.New("'package' statement missing")

func Parse(reader io.Reader, filename string) (*File, error) {
	parser := proto.NewParser(reader)
	parser.Filename(filename)
	return parser.Parse()
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
	formatter.Format(root)
	return nil
}
