package gocode

import (
	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"
)

type Cursor = dstutil.Cursor

func Apply(root dst.Node, pre dstutil.ApplyFunc) dst.Node {
	return dstutil.Apply(root, pre, nil)
}
