package gocode

import (
	"fmt"

	"github.com/dave/dst"
)

type IndexExpression struct {
	inner *dst.IndexExpr
}

func IndexInto(name string, fields ...string) *IndexExpression {
	return Index(Identifier(name, fields...))
}

func IndexIntof(format string, args ...interface{}) *IndexExpression {
	return IndexInto(fmt.Sprintf(format, args...))
}

func Index(expr dst.Expr) *IndexExpression {
	return &IndexExpression{
		inner: &dst.IndexExpr{X: expr},
	}
}

func (ie *IndexExpression) WithIdentifier(name string, fields ...string) *dst.IndexExpr {
	return ie.With(Identifier(name, fields...))
}

func (ie *IndexExpression) With(expr dst.Expr) *dst.IndexExpr {
	ie.inner.Index = expr
	return ie.inner
}
