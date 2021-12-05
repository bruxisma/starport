package gocode

import "github.com/dave/dst"

// WhenBuilder is used to declare an initializing IfStatement
type WhenBuilder struct {
	condition *IfStatement
	assign    *Assignment
}

func WhenAssigning(item string, items ...string) *WhenBuilder {
	assign := AssignVariable(item)
	return whenCreateTargets(assign, items...)
}

func WhenDefining(item string, items ...string) *WhenBuilder {
	define := DefineVariable(item)
	return whenCreateTargets(define, items...)
}

func whenCreateTargets(assignment *Assignment, items ...string) *WhenBuilder {
	for _, item := range items {
		assignment.addTarget(item)
	}
	return &WhenBuilder{
		condition: &IfStatement{inner: &dst.IfStmt{}},
		assign:    assignment,
	}
}

func (when *WhenBuilder) To(items ...interface{}) *WhenBuilder {
	when.assign.To(items...)
	return when
}

func (when *WhenBuilder) IfVar(name string) *IfStatement {
	return when.If(Identifier(name))
}

func (when *WhenBuilder) If(expr dst.Expr) *IfStatement {
	when.condition.inner.Init = when.assign.inner
	when.condition.inner.Cond = expr
	return &IfStatement{inner: when.condition.inner}
}
