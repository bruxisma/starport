package mutate

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/dave/dst"
	"github.com/emicklei/proto"
	"github.com/tendermint/starport/starport/pkg/gocode"
	"github.com/tendermint/starport/starport/pkg/protocode"
	"github.com/tendermint/starport/starport/templates/typed"
)

func GenesisProtoGenesisState(tree *protocode.File, opts *typed.Options) (*protocode.File, error) {
	message, err := tree.FindMessage("GenesisState")
	if err != nil {
		return nil, err
	}

	list := proto.Field{
		Name: fmt.Sprintf("%sList", opts.TypeName.LowerCamel),
		Type: opts.TypeName.UpperCamel,
		Options: []*proto.Option{
			{Name: "(gogoproto.nullable)", Constant: protocode.False()},
		},
	}

	message.AppendRepeatedField(list)
	message.Append(proto.Field{
		Name: fmt.Sprintf("%sCount", opts.TypeName.LowerCamel),
		Type: "uint64",
	})

	return tree, nil
}

func GenesisTypesDefaultGenesisReturnValue(tree *dst.File, opts *typed.Options) (*dst.File, error) {
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

func GenesisTypesValidateStatements(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := gocode.FindMethod(tree, "GenesisState.Validate")
	if err != nil {
		return nil, err
	}
	gocode.Apply(fn, func(cursor *gocode.Cursor) bool {
		if _, ok := cursor.Node().(*dst.ReturnStmt); !ok {
			return true
		}
		for _, stmt := range genesisTypesCreateValidateCheckNodes(opts) {
			cursor.InsertBefore(stmt)
		}
		return false
	})
	return tree, nil
}

func GenesisModuleInsertInit(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := gocode.FindFunction(tree, "InitGenesis")
	if err != nil {
		return nil, err
	}
	fn.Body.List = append(genesisModuleCreateInit(opts), fn.Body.List...)
	return tree, nil
}

func GenesisModuleInsertExport(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := gocode.FindFunction(tree, "ExportGenesis")
	if err != nil {
		return nil, err
	}
	gocode.Apply(fn, func(cursor *gocode.Cursor) bool {
		if _, ok := cursor.Node().(*dst.ReturnStmt); !ok {
			return true
		}
		for _, statement := range genesisModuleCreateExport(opts) {
			cursor.InsertBefore(statement)
		}
		return false
	})
	return tree, nil
}

func GenesisTestsInsertList(tree *dst.File, opts *typed.Options) (*dst.File, error) {
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

func GenesisTestsInsertComparison(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := gocode.FindFunction(tree, "TestGenesis")
	if err != nil {
		return nil, err
	}

	fn.Body.List = append(fn.Body.List, genesisTestsCreateComparison(opts)...)
	return tree, nil
}

// GenesisTestsInsertValidGenesisState inserts valid genesis state values for
// testing.
//
// This function is quite complicated, as it inserts several values deep into
// the state.
func GenesisTestsInsertValidGenesisState(tree *dst.File, opts *typed.Options) (*dst.File, error) {
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

func GenesisTestsInsertInvalidGenesisState(tree *dst.File, opts *typed.Options) (*dst.File, error) {
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
			composite.Elts = append(
				composite.Elts,
				genesisTestsCreateDuplicatedState(opts).Build(),
				genesisTestsCreateInvalidCount(opts).Build(),
			)
			return false
		}
		return true
	})

	return tree, nil
}

// genesisTypesCreateValidateCheckNodes is a fairly complex set of statements
// built for inserting into the Validate function
func genesisTypesCreateValidateCheckNodes(opts *typed.Options) []dst.Stmt {
	lower := opts.TypeName.LowerCamel

	idMapAssign := gocode.Definef("%sIdMap", lower).
		To(gocode.MakeMapOf("uint64").WithIndexOf("bool"))
	countAssign := gocode.Definef("%sCount", lower).
		To(gocode.Callf("gs.Get%sCount", opts.TypeName.UpperCamel))

	list := fmt.Sprintf("%sList", opts.TypeName.UpperCamel)
	rangeFor := gocode.ForEachItem("elem").In("gs.%s", list).Do(func(ctx *gocode.Block) {
		index := gocode.IndexInto("%sIdMap", lower).WithIdentifier("elem.Id")
		ctx.WhenDefining("_", "ok").To(index).IfVar("ok").IsTrue().Then(func(ctx *gocode.Block) {
			ctx.Returns(gocode.Errorf("duplicated id for %s", lower))
		})
		ctx.IfVar("elem.Id").IsGreaterOrEqualToVarf("%sCount", lower).Then(func(ctx *gocode.Block) {
			countComparisonString := fmt.Sprintf("%s id should be lower or equal than the last id", lower)
			ctx.Returns(gocode.Errorf(countComparisonString))
		})
		ctx.AssignIndex(gocode.Name("%sIdMap", lower), gocode.Identifier("elem.Id")).
			To(gocode.True())
	}).Done()

	return []dst.Stmt{idMapAssign, countAssign, rangeFor}
}

func genesisModuleCreateInit(opts *typed.Options) []dst.Stmt {
	typename := opts.TypeName.UpperCamel
	variable := opts.TypeName.LowerCamel
	forLoop := gocode.ForEachItem("elem").In("genState. %sList", typename).
		Do(func(block *gocode.Block) {
			block.Callf("k.Set%s", typename).WithVars("ctx", "elem")
		}).PrependComment("Set all the %s", variable).Done()
	set := gocode.Callf("k.Set%sCount", typename).
		WithArgument("ctx").
		WithArgumentf("genState.%sCount", typename).
		PrependComment("Set %s count", variable).
		AsStatement()

	return []dst.Stmt{forLoop, set}
}

func genesisModuleCreateExport(opts *typed.Options) []dst.Stmt {
	typename := opts.TypeName.UpperCamel
	return []dst.Stmt{
		gocode.Assignf("genesis.%sList", typename).
			To(gocode.Callf("k.GetAll%s", typename).WithArgument("ctx")),
		gocode.Assignf("genesis.%sCount", typename).
			To(gocode.Callf("k.Get%sCount", typename).WithArgument("ctx")),
	}
}

func genesisTestsCreateLists(opts *typed.Options) []dst.Expr {
	list := gocode.KeyValue(
		fmt.Sprintf("%sList", opts.TypeName.UpperCamel),
		gocode.Slicef("types.%s", opts.TypeName.UpperCamel).Extend(
			gocode.Anonymous{"Id": 0},
			gocode.Anonymous{"Id": 0},
		),
	)
	count := gocode.KeyValue(fmt.Sprintf("%sCount", opts.TypeName.UpperCamel), 2)
	return []dst.Expr{list, count}
}

func genesisTestsCreateComparison(opts *typed.Options) []dst.Stmt {
	count := fmt.Sprintf("%sCount", opts.TypeName.UpperCamel)
	list := fmt.Sprintf("%sList", opts.TypeName.UpperCamel)
	compareList := gocode.Call("require.ElementsMatch").
		WithArgument("t").
		WithArgumentf("genesisState.%s", list).
		WithArgumentf("got.%s", list).
		AsStatement()
	compareCount := gocode.Call("require.Equal").
		WithArgument("t").
		WithArgumentf("genesisState.%s", count).
		WithArgumentf("got.%s", count).
		AsStatement()
	return []dst.Stmt{compareList, compareCount}
}

func genesisTestsCreateValidGenesisState(opts *typed.Options) []dst.Expr {
	return []dst.Expr{
		gocode.KeyValue(
			fmt.Sprintf("%sList", opts.TypeName.UpperCamel),
			gocode.Slicef("types.%s", opts.TypeName.UpperCamel).Extend(
				gocode.Anonymous{"Id": 0},
				gocode.Anonymous{"Id": 1},
			)),
		gocode.KeyValue(fmt.Sprintf("%sCount", opts.TypeName.UpperCamel), 2),
	}
}

func genesisTestsCreateDuplicatedState(opts *typed.Options) gocode.Builder {
	typename := opts.TypeName.UpperCamel
	return gocode.Anonymous{
		"desc":  fmt.Sprintf("duplicated %s", opts.TypeName.LowerCamel),
		"valid": false,
		"genState": gocode.Struct("types.GenesisState").Extend(map[string]interface{}{
			fmt.Sprintf("%sList", typename): gocode.Slicef("types.%s", typename).Extend(
				gocode.Anonymous{"Id": 0},
				gocode.Anonymous{"Id": 0},
			),
		}).AddressOf(),
	}
}

func genesisTestsCreateInvalidCount(opts *typed.Options) gocode.Builder {
	typename := opts.TypeName.UpperCamel
	variable := opts.TypeName.LowerCamel
	return gocode.Anonymous{
		"desc":  fmt.Sprintf("invalid %s count", variable),
		"valid": false,
		"genState": gocode.Struct("types.GensisState").Extend(map[string]interface{}{
			fmt.Sprintf("%sCount", typename): 0,
			fmt.Sprintf("%sList", typename): gocode.Slicef("types.%s", typename).Extend(
				gocode.Anonymous{"Id": 1},
			),
		}).AddressOf(),
	}
}
