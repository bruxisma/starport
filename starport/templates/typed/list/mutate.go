package list

import (
	"fmt"
	"io"
	"reflect"
	"strconv"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/tendermint/starport/starport/pkg/gocode"
	"github.com/tendermint/starport/starport/templates/typed"
)

type (
	mutator  func(*dst.File, *typed.Options) (*dst.File, error)
	mutators []mutator
)

func reportModifyError(path string, err error) error {
	return fmt.Errorf("modifying %q errored with %w", path, err)
}

// Apply executes each mutator on the provided root AST node and Options value
func (fns mutators) Apply(source io.Reader, opts *typed.Options) (*dst.File, error) {
	tree, err := decorator.Parse(source)
	if err != nil {
		return nil, err
	}
	for _, fn := range fns {
		if tree, err = fn(tree, opts); err != nil {
			return nil, err
		}
	}
	return tree, nil
}

var (
	structType       = reflect.TypeOf((*dst.StructType)(nil))
	compositeLitType = reflect.TypeOf((*dst.CompositeLit)(nil))
)

func mutateTypesDefaultGenesisReturnValue(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := gocode.FindFunction(tree, "DefaultGenesis")
	if err != nil {
		return nil, err
	}
	gocode.Apply(fn, func(cursor *gocode.Cursor) bool {
		var composite *dst.CompositeLit
		if ret, ok := cursor.Node().(*dst.ReturnStmt); !ok {
			return true
		} else if unary, ok := ret.Results[0].(*dst.UnaryExpr); !ok {
			return true
		} else if composite, ok = unary.X.(*dst.CompositeLit); !ok {
			return true
		}
		composite.Elts = append(composite.Elts, &dst.KeyValueExpr{
			Key:   gocode.Name("%sList", opts.TypeName.UpperCamel),
			Value: gocode.SliceOf(opts.TypeName.UpperCamel).Node(),
		})
		return false
	})
	return tree, nil
}

func mutateTypesValidateStatements(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := gocode.FindFunction(tree, "Validate")
	if err != nil {
		return nil, err
	}
	gocode.Apply(fn, func(cursor *gocode.Cursor) bool {
		if _, ok := cursor.Node().(*dst.ReturnStmt); !ok {
			return true
		}
		for _, stmt := range mutateTypesCreateValidateCheckNodes(opts) {
			cursor.InsertBefore(stmt)
		}
		return false
	})
	return tree, nil
}

func genesisModuleInsertInit(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := gocode.FindFunction(tree, "InitGenesis")
	if err != nil {
		return nil, err
	}
	statements, err := genesisModuleCreateInit(opts)
	if err != nil {
		return nil, err
	}
	if len(statements) > 0 {
		fn.Body.List = append(statements, fn.Body.List...)
	}
	return tree, nil
}

func genesisModuleInsertExport(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := gocode.FindFunction(tree, "ExportGenesis")
	if err != nil {
		return nil, err
	}
	statements, err := genesisModuleCreateExport(opts)
	if err != nil {
		return nil, err
	}
	if len(statements) > 0 {
		// We want to preserve the 'return' statement
		idx := len(fn.Body.List) - 1
		back := fn.Body.List[idx]
		front := fn.Body.List[:idx]
		fn.Body.List = append(append(front, statements...), back)
	}
	return tree, nil
}

func genesisTestsInsertList(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := gocode.FindFunction(tree, "TestGenesis")
	if err != nil {
		return nil, err
	}

	literal := fn.Body.
		List[0].(*dst.AssignStmt).
		Rhs[0].(*dst.CompositeLit)
	literal.Elts = append(literal.Elts, genesisTestsCreateLists(opts)...)

	return tree, nil
}

func genesisTestsInsertComparison(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := gocode.FindFunction(tree, "TestGenesis")
	if err != nil {
		return nil, err
	}

	statements, err := genesisTestsCreateComparison(opts)
	if err != nil {
		return nil, err
	}

	fn.Body.List = append(fn.Body.List, statements...)
	return tree, nil
}

func genesisTestsInsertValidGenesisState(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := gocode.FindFunction(tree, "TestGenesisState_Validate")
	if err != nil {
		return nil, err
	}

	var state *dst.CompositeLit

	gocode.Apply(fn, func(cursor *gocode.Cursor) bool {
		if reflect.TypeOf(cursor.Parent()) != compositeLitType {
			return true
		} else if kv, ok := cursor.Node().(*dst.KeyValueExpr); !ok {
			return true
		} else if name, ok := gocode.KeyAsIdentifier(kv); !ok || name != "desc" {
			return true
		} else if basic, ok := gocode.ValueAsBasicLiteral(kv); !ok {
			return true
		} else if basic.Value != strconv.Quote("valid genesis state") {
			return true
		}
		state = cursor.Parent().(*dst.CompositeLit)
		return false
	})

	if state == nil {
		return nil, fmt.Errorf("unable to find composite literal containing 'valid genesis state'")
	}

	gocode.Apply(state, func(cursor *gocode.Cursor) bool {
		var (
			kv *dst.KeyValueExpr
			ok bool
		)
		if kv, ok = cursor.Node().(*dst.KeyValueExpr); !ok {
			return true
		}
		if name, ok := gocode.KeyAsIdentifier(kv); !ok || name != "genState" {
			return true
		}
		if composite, ok := gocode.ValueAsCompositeLiteral(kv); ok {
			composite.Elts = append(composite.Elts, genesisTestsCreateValidGenesisState(opts)...)
			return false
		}
		return true
	})

	return tree, nil
}

func genesisTestsInsertDuplicatedState(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := gocode.FindFunction(tree, "TestGenesisState_Validate")
	if err != nil {
		return nil, err
	}

	gocode.Apply(fn, func(cursor *gocode.Cursor) bool {
		arrayType, ok := cursor.Node().(*dst.ArrayType)
		if !ok || reflect.TypeOf(arrayType.Elt) != structType {
			return true
		}
		if composite, ok := cursor.Parent().(*dst.CompositeLit); ok {
			composite.Elts = append(composite.Elts, genesisTestsCreateDuplicatedState(opts)...)
			return false
		}
		return true
	})

	return tree, nil
}

func genesisTestsInsertInvalidCount(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := gocode.FindFunction(tree, "TestGenesisState_Validate")
	if err != nil {
		return nil, err
	}
	gocode.Apply(fn, func(cursor *gocode.Cursor) bool {
		arrayType, ok := cursor.Node().(*dst.ArrayType)
		if !ok || reflect.TypeOf(arrayType.Elt) != structType {
			return true
		}
		composite, ok := cursor.Parent().(*dst.CompositeLit)
		if !ok {
			return true
		}
		composite.Elts = append(composite.Elts, genesisTestsCreateInvalidCount(opts)...)
		return false
	})
	return tree, nil
}

// mutateTypesCreateValidateCheckNodes is a fairly complex set of statements
// built for inserting into the Validate function
func mutateTypesCreateValidateCheckNodes(opts *typed.Options) []dst.Stmt {
	duplicateIdString := fmt.Sprintf("duplicated id for %s", opts.TypeName.LowerCamel)
	countComparisonString := fmt.Sprintf(
		"%s id should be lower or equal than the last id",
		opts.TypeName.LowerCamel)

	idMapAssign := gocode.Assignf("%sIdMap", opts.TypeName.LowerCamel).
		To(gocode.MakeMapOf("uint64").WithIndexOf("bool"))
	countAssign := gocode.Definef("%sCount", opts.TypeName.LowerCamel).
		To(gocode.Call("gs", fmt.Sprintf("Get%sCount", opts.TypeName.UpperCamel)).Node())

	list := fmt.Sprintf("%sList", opts.TypeName.UpperCamel)
	rangeFor := gocode.ForEachItem("elem").In("gs", list).Do(func(ctx *gocode.Block) {
		index := gocode.IndexIntof("%sIdMap", opts.TypeName.LowerCamel).
			WithIdentifier("elem", "Id")
		ctx.WhenDefining("_", "ok").To(index).IfVar("ok").IsTrue().Then(func(ctx *gocode.Block) {
			ctx.Returns(gocode.Call("fmt", "Errorf").WithString(duplicateIdString).Node())
		})
		ctx.IfVar("elem", "id").IsGreaterOrEqualToVarf("%sCount", opts.TypeName.LowerCamel).Then(func(ctx *gocode.Block) {
			ctx.Returns(gocode.Call("fmt", "Errorf").WithString(countComparisonString).Node())
		})
		ctx.AssignIndex(
			gocode.Name("%sIdMap", opts.TypeName.LowerCamel),
			gocode.Identifier("elem", "Id"),
		).
			To(gocode.True())
	}).Done()

	return []dst.Stmt{idMapAssign, countAssign, rangeFor}
}

func genesisModuleCreateInit(opts *typed.Options) ([]dst.Stmt, error) {
	forLoop := gocode.ForEachItem("elem").
		PrependComment("Set all the %s", opts.TypeName.LowerCamel).
		In("genState", fmt.Sprintf("%sList", opts.TypeName.UpperCamel)).
		Do(func(block *gocode.Block) {
			block.Call("k", fmt.Sprintf("Set%s", opts.TypeName.UpperCamel)).
				WithArgument("ctx").
				WithArgument("elem")
		}).
		Done()
	set := gocode.Call("k", fmt.Sprintf("Set%sCount", opts.TypeName.UpperCamel)).
		WithArgument("ctx").
		WithArgument("genState", fmt.Sprintf("%sCount", opts.TypeName.UpperCamel)).
		PrependComment("Set %s count", opts.TypeName.LowerCamel).
		AsStatement()

	return []dst.Stmt{forLoop, set}, nil
}

func genesisModuleCreateExport(opts *typed.Options) ([]dst.Stmt, error) {
	typename := opts.TypeName.UpperCamel
	statements := []dst.Stmt{
		gocode.AssignVariable("genesis", fmt.Sprintf("%sList", typename)).
			To(gocode.Call("k", fmt.Sprintf("GetAll%s", typename)).WithArgument("ctx").Node()),
		gocode.AssignVariable("genesis", fmt.Sprintf("%sCount", typename)).
			To(gocode.Call("k", fmt.Sprintf("Get%sCount", typename)).WithArgument("ctx").Node()),
	}
	return statements, nil
}

func genesisTestsCreateLists(opts *typed.Options) []dst.Expr {
	list := &dst.KeyValueExpr{
		Decs: gocode.KVDecs,
		Key:  gocode.Name("%sList", opts.TypeName.UpperCamel),
		Value: gocode.SliceOf("types", opts.TypeName.UpperCamel).
			AppendExpr(gocode.KeyValues(map[string]interface{}{"Id": 0})).
			AppendExpr(gocode.KeyValues(map[string]interface{}{"Id": 0})).
			Node(),
	}
	count := gocode.KeyValue(fmt.Sprintf("%sCount", opts.TypeName.UpperCamel), 2)
	return []dst.Expr{list, count}
}

func genesisTestsCreateComparison(opts *typed.Options) ([]dst.Stmt, error) {
	count := fmt.Sprintf("%sCount", opts.TypeName.UpperCamel)
	list := fmt.Sprintf("%sList", opts.TypeName.UpperCamel)
	compareList := gocode.Call("require", "ElementsMatch").
		WithArgument("t").
		WithArgument("genesisState", list).
		WithArgument("got", list).
		AsStatement()
	compareCount := gocode.Call("require", "Equal").
		WithArgument("t").
		WithArgument("genesisState", count).
		WithArgument("got", count).
		AsStatement()
	return []dst.Stmt{compareList, compareCount}, nil
}

func genesisTestsCreateValidGenesisState(opts *typed.Options) []dst.Expr {
	return []dst.Expr{
		&dst.KeyValueExpr{
			Key: gocode.Name("%sList", opts.TypeName.UpperCamel),
			Value: gocode.SliceOf("types", opts.TypeName.UpperCamel).
				AppendExpr(
					gocode.AnonymousStruct().AppendField("Id", 0).Done(),
					gocode.AnonymousStruct().AppendField("Id", 1).Done(),
				).
				Node(),
		},
		gocode.KeyValue(fmt.Sprintf("%sCount", opts.TypeName.UpperCamel), 2),
	}
}

func genesisTestsCreateDuplicatedState(opts *typed.Options) []dst.Expr {
	return []dst.Expr{
		gocode.AnonymousStruct().
			AppendField("desc", fmt.Sprintf("duplicated %s", opts.TypeName.LowerCamel)).
			AppendField("valid", false).
			AppendExpr("genState",
				gocode.Struct("types", "GenesisState").
					AppendExpr(
						fmt.Sprintf("%sList", opts.TypeName.UpperCamel),
						gocode.SliceOf("types", opts.TypeName.UpperCamel).
							AppendExpr(
								gocode.AnonymousStruct().AppendField("Id", 0).Done(),
								gocode.AnonymousStruct().AppendField("Id", 0).Done(),
							).Node(),
					).AddressOf(),
			).Done(),
	}
}

func genesisTestsCreateInvalidCount(opts *typed.Options) []dst.Expr {
	return []dst.Expr{
		gocode.AnonymousStruct().
			AppendField("desc", fmt.Sprintf("duplicated %s", opts.TypeName.LowerCamel)).
			AppendField("valid", false).
			AppendExpr(
				"genState",
				gocode.Struct("types", "GenesisState").
					AppendField(fmt.Sprintf("%sCount", opts.TypeName.UpperCamel), 0).
					AppendExpr(
						fmt.Sprintf("%sList", opts.TypeName.UpperCamel),
						gocode.SliceOf("types", opts.TypeName.UpperCamel).
							AppendExpr(gocode.AnonymousStruct().AppendField("Id", 1).Done()).
							Node(),
					).AddressOf(),
			).Done(),
	}
}
