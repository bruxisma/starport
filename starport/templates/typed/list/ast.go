package list

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/dave/dst"
	"github.com/tendermint/starport/starport/pkg/gocode"
	"github.com/tendermint/starport/starport/templates/ast"
	"github.com/tendermint/starport/starport/templates/typed"
)

var (
	structType       = reflect.TypeOf((*dst.StructType)(nil))
	compositeLitType = reflect.TypeOf((*dst.CompositeLit)(nil))
)

func genesisTypesInsertDefaultGenesisList(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	// Find the DefaultGenesis function in the tree
	fn, err := ast.FindFunction(tree, "DefaultGenesis")
	if err != nil {
		return nil, err
	}
	literal := fn.Body.
		List[0].(*dst.ReturnStmt).
		Results[0].(*dst.UnaryExpr).
		X.(*dst.CompositeLit)
	literal.Elts = append(genesisTypesCreateDefaultGenesisList(opts), literal.Elts...)
	return tree, nil
}

func genesisTypesInsertValidateCheck(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := ast.FindFunction(tree, "Validate")
	if err != nil {
		return nil, err
	}
	statements, err := genesisTypesCreateValidateCheck(opts)
	if err != nil {
		return nil, err
	}
	idx := len(fn.Body.List) - 1
	statements = append(statements, fn.Body.List[idx])
	fn.Body.List = append(fn.Body.List[:idx], statements...)
	return tree, nil
}

func genesisModuleInsertInit(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := ast.FindFunction(tree, "InitGenesis")
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
	fn, err := ast.FindFunction(tree, "ExportGenesis")
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
	fn, err := ast.FindFunction(tree, "TestGenesis")
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
	fn, err := ast.FindFunction(tree, "TestGenesis")
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
	fn, err := ast.FindFunction(tree, "TestGenesisState_Validate")
	if err != nil {
		return nil, err
	}

	var state *dst.CompositeLit

	gocode.Apply(fn, func(cursor *gocode.Cursor) bool {
		kv, ok := cursor.Node().(*dst.KeyValueExpr)
		if !ok || reflect.TypeOf(cursor.Parent()) != compositeLitType {
			return true
		}
		if name, ok := gocode.KeyAsIdentifier(kv); !ok || name != "desc" {
			return true
		}
		if basic, ok := gocode.ValueAsBasicLiteral(kv); !ok || basic.Value != strconv.Quote("valid genesis state") {
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
	fn, err := ast.FindFunction(tree, "TestGenesisState_Validate")
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
	fn, err := ast.FindFunction(tree, "TestGenesisState_Validate")
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

func genesisTypesCreateValidateCheck(opts *typed.Options) ([]dst.Stmt, error) {
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

	return []dst.Stmt{
		idMapAssign,
		countAssign,
		rangeFor,
	}, nil
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

// genesisTypesCreateDefaultGenesisList returns an AST Node equivalent to Name:
// &Type{}
func genesisTypesCreateDefaultGenesisList(opts *typed.Options) []dst.Expr {
	node := &dst.KeyValueExpr{
		Decs:  dst.KeyValueExprDecorations{NodeDecs: dst.NodeDecs{Before: 1}},
		Key:   gocode.Name("%sList", opts.TypeName.UpperCamel),
		Value: gocode.SliceOf(opts.TypeName.UpperCamel).Node(),
	}
	return []dst.Expr{node}
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
