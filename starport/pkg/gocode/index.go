package gocode

import (
	"fmt"

	"github.com/dave/dst"
)

type IndexExpression struct {
	inner *dst.IndexExpr
}

func IndexInto(format string, args ...interface{}) *IndexExpression {
	return Index(Identifier(fmt.Sprintf(format, args...)))
}

func Index(expr dst.Expr) *IndexExpression {
	return &IndexExpression{
		inner: &dst.IndexExpr{X: expr},
	}
}

func (ie *IndexExpression) WithIdentifier(name string) *dst.IndexExpr {
	return ie.With(Identifier(name))
}

func (ie *IndexExpression) With(expr dst.Expr) *dst.IndexExpr {
	ie.inner.Index = expr
	return ie.inner
}
