package gocode

import (
	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"
)

// Cursor is an alias for dstutil.Cursor
type Cursor = dstutil.Cursor

// Apply is a small wrapper around dstutil.Apply
func Apply(root dst.Node, pre dstutil.ApplyFunc) dst.Node {
	return dstutil.Apply(root, pre, nil)
}
