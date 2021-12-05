package gocode

import (
	"fmt"
	"go/token"

	"github.com/dave/dst"
)

// Block is used when 'scoped' callbacks are required by users to group
// statements together (e.g., inside of an if statement block or a for range
// loop.
//
// NOTE: Not all possible statements and constructs are available on Blocks at
// the moment. Only the ones currently required
type Block struct {
	inner *dst.BlockStmt
}

// Assign returns an Assignment that will be exported as part of the Block's
// body
func (block *Block) Assign(expr dst.Expr, exprs ...dst.Expr) *Assignment {
	assignment := Assign(expr, exprs...)
	block.Append(assignment.inner)
	return assignment
}

// Callf calls Call with the constructed format string
func (block *Block) Callf(format string, args ...interface{}) *FunctionCall {
	return block.Call(fmt.Sprintf(format, args...))
}

// Call appends a function call (as a statement) to the Block
func (block *Block) Call(name string) *FunctionCall {
	call := Call(name)
	block.Append(call.AsStatement())
	return call
}

// WhenDefining returns a WhenBuilder for declaring the sometimes optional
// initialization statement that is attached to an if statement.
//
// This function is used when defining new variables.
func (block *Block) WhenDefining(item string, items ...string) *WhenBuilder {
	when := WhenDefining(item, items...)
	block.Append(when.condition.inner)
	return when
}

// WhenAssigning returns a WhenBuilder for declaring the sometimes optional
// initialization statement that is attached to an if statement.
//
// This function is used when assigning to predeclared variables.
func (block *Block) WhenAssigning(item string, items ...string) *WhenBuilder {
	when := WhenAssigning(item, items...)
	block.Append(when.condition.inner)
	return when
}

// AssignIndex returns an Assignment that be exported as part of the Block's
// body by calling gocode.AssignIndex
func (block *Block) AssignIndex(collection dst.Expr, index dst.Expr) *Assignment {
	assignment := AssignIndex(collection, index)
	block.Append(assignment.inner)
	return assignment
}

// IfVar returns an IfStatement attached to the Block's body, by calling IfVar
func (block *Block) IfVar(name string) *IfStatement {
	stmt := IfVar(name)
	block.Append(stmt.inner)
	return stmt
}

// Continue appends a continue statement to the Block
func (block *Block) Continue() {
	block.Append(&dst.BranchStmt{Tok: token.CONTINUE})
}

// Break appends a break statement to the Block
func (block *Block) Break() {
	block.Append(&dst.BranchStmt{Tok: token.BREAK})
}

// Return appends an empty return statement to the Block
func (block *Block) Return() {
	block.Append(&dst.ReturnStmt{})
}

// Returns appends a return statement with the provided expressions to the Block
func (block *Block) Returns(items ...interface{}) {
	values := []dst.Expr{}
	for _, item := range items {
		values = append(values, Item(item))
	}
	block.Append(&dst.ReturnStmt{Results: values})
}

// Append will append the provided statements to the block directly
//
// NOTE: Unlike other functions, Append does not return the block itself.
func (block *Block) Append(stmts ...dst.Stmt) {
	block.inner.List = append(block.inner.List, stmts...)
}
