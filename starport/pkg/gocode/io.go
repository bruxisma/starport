package gocode

import (
	"bytes"
	"io"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func Write(tree *dst.File) (io.Reader, error) {
	buffer := &bytes.Buffer{}
	if err := decorator.Fprint(buffer, tree); err != nil {
		return nil, err
	}
	return buffer, nil
}
