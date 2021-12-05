package gocode

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/dave/dst"
)

// ErrMissingFunction is returned by gocode.FindFunction when the function
// cannot be located
var ErrMissingFunction = errors.New("could not locate function")

// ErrMissingMethod is returned by gocode.FindMethod when the method cannot be
// located
var ErrMissingMethod = errors.New("could not locate method")

// ErrInvalidMethodName is returned by gocode.FindMethod when the name provided
// is missing a '.'
var ErrInvalidMethodName = errors.New("expected a type and method name")

// ErrInvalidMethodReceiver is returned by gocode.FindMethod when the name
// provided results in a missing component. This is usually due to an extra `.`
// inserted.
var ErrInvalidMethodReceiver = errors.New("invalid method receiver")

// FindFunction looks for the FuncDecl with the provided name
//
// NOTE: FindFunction *does not* return functions with a Receiver. For that,
// use FindMethod
func FindFunction(tree *dst.File, name string) (*dst.FuncDecl, error) {
	for _, decl := range tree.Decls {
		if fn, ok := decl.(*dst.FuncDecl); ok && fn.Name.Name == name && fn.Recv == nil {
			return fn, nil
		}
	}
	return nil, fmt.Errorf("gocode find function: %w '%s'", ErrMissingFunction, name)
}

// FindMethod looks for the first FuncDecl whose receiver *type* matchs the one provided
//
// NOTE: Because of how golang works, this function checks for receivers of
// both pointer and non-pointer Identifiers. This function's logic is MUCH more
// complicated than FindFunction due to the AST layout provided by golang.
func FindMethod(tree *dst.File, method string) (*dst.FuncDecl, error) {
	components := strings.Split(method, ".")
	length := len(components)
	if length < 2 {
		return nil, fmt.Errorf("%w but received %q", ErrInvalidMethodName, method)
	}
	for _, component := range components {
		if component == "" {
			return nil, fmt.Errorf("%w %q", ErrInvalidMethodReceiver, method)
		}
	}
	name := components[len(components)-1]
	receiver := strings.Join(components[:1], ".")

	for _, decl := range tree.Decls {
		fn, ok := decl.(*dst.FuncDecl)
		if !ok || fn.Name.Name != name || fn.Recv == nil {
			continue
		}
		// Remove points from the type name
		typename := fn.Recv.List[0].Type
		if star, ok := typename.(*dst.StarExpr); ok {
			typename = star.X
		}

		switch value := typename.(type) {
		case *dst.Ident:
			if value.Name == receiver {
				return fn, nil
			}
		case *dst.SelectorExpr:
			components := sort.StringSlice{}
			for {
				components = append(components, value.Sel.Name)
				if x, ok := value.X.(*dst.SelectorExpr); ok {
					value = x
					continue
				}
				components = append(components, value.X.(*dst.Ident).Name)
				break
			}
			if strings.Join(sort.Reverse(components).(sort.StringSlice), ".") == receiver {
				return fn, nil
			}
		}
	}
	return nil, fmt.Errorf("gocode find method: %w %q for receiver %q", ErrMissingFunction, name, receiver)
}
