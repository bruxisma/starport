package mutate

import (
	"fmt"
	"go/token"

	"github.com/dave/dst"
	"github.com/tendermint/starport/starport/pkg/gocode"
	"github.com/tendermint/starport/starport/templates/typed"
)

// SimulationInsertGenesisState mutates the GenesisState value with new list and count values
func SimulationInsertGenesisState(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	// TODO: move into a mutate Sequence
	fn, err := gocode.FindFunction(tree, "GenerateGenesisState")
	if err != nil {
		return nil, err
	}

	gocode.Apply(fn, func(cursor *gocode.Cursor) bool {
		if composite, ok := cursor.Node().(*dst.CompositeLit); !ok || composite.Type == nil {
			return true
		} else if name, ok := composite.Type.(*dst.SelectorExpr); !ok {
			return true
		} else if name.Sel.Name != "GenesisState" {
			return true
		}
		composite := cursor.Node().(*dst.CompositeLit)

		signer := opts.MsgSigner.UpperCamel
		list := gocode.KeyValue(
			fmt.Sprintf("%sList", opts.TypeName.UpperCamel),
			gocode.SliceOf("types", opts.TypeName.UpperCamel).Extend(
				gocode.Anonymous{
					"Id":   0,
					signer: gocode.Call("sample", "AccAddress"),
				},
				gocode.Anonymous{
					"Id":   1,
					signer: gocode.Call("sample", "AccAddress"),
				}))
		count := gocode.KeyValue(fmt.Sprintf("%sCount", opts.TypeName.UpperCamel), 2)

		composite.Elts = append(composite.Elts, list, count)

		return false
	})

	return tree, nil
}

// SimulationInsertConstOpWeightMsg mutates the AST by inserting const global values
func SimulationInsertConstOpWeightMsg(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	// Find the first function, the insert ConstDecls before it using its cursor

	gocode.Apply(tree, func(cursor *gocode.Cursor) bool {
		decl, ok := cursor.Node().(*dst.GenDecl)
		if !ok || decl.Tok != token.CONST {
			return true
		}
		for _, action := range []string{"Create", "Update", "Delete"} {
			target := fmt.Sprintf("%s%s", action, opts.TypeName.UpperCamel)
			opName := fmt.Sprintf("opWeightMsg%s", target)
			defaultName := fmt.Sprintf("defaultWeightMsg%s", target)
			// FIXME: There is a missing comment that needs to be attached
			defaultWeight := gocode.TypedGlobal(defaultName, "int", 100)
			defaultWeight.Decs.Start.Prepend("\n// TODO: Determine the simulation weight value")
			decl.Specs = append(decl.Specs,
				gocode.Global(opName, "op_weight_msg_create_chain"),
				defaultWeight,
			)
		}
		return false
	})
	return tree, nil
}

func SimulationInsertWeightedOperations(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := gocode.FindFunction(tree, "WeightedOperations")
	if err != nil {
		return nil, err
	}
	gocode.Apply(fn, func(cursor *gocode.Cursor) bool {
		if _, ok := cursor.Node().(*dst.ReturnStmt); !ok {
			return true
		}

		for _, action := range []string{"Create", "Update", "Delete"} {
			name := fmt.Sprintf("%s%s", action, opts.TypeName.UpperCamel)
			weightMsg := fmt.Sprintf("weightMsg%s", name)

			randFunc := gocode.Func().
				Parameter(gocode.Name("_"), gocode.Identifier("rand", "Rand")).
				Do(func(block *gocode.Block) {
					block.Assign(gocode.Name(weightMsg)).To(gocode.Name("defaultWeightMsg%s", name))
				})
			getOrGenerate := gocode.Call("simState", "AppParams", "GetOrGenerate").
				WithArgument("simState", "Cdc").
				WithArgumentf("opWeightMsg%s", name).
				WithParameters(gocode.AddressOf(weightMsg)).
				WithParameters(gocode.Nil()).
				WithParameters(randFunc.Build()).
				AsStatement()

			simulateMsg := gocode.Call(fmt.Sprintf("%ssimulation", opts.ModuleName), fmt.Sprintf("SimulateMsg%s", name)).
				WithArgument("am", "accountKeeper").
				WithArgument("am", "bankKeeper").
				WithArgument("am", "keeper").
				Build()

			newWeightOperation := gocode.Call("simulation", "NewWeightOperation").
				WithArgument(weightMsg).
				WithParameters(simulateMsg).
				Node()
			newWeightOperation.Decs.Before = dst.NewLine

			operations := gocode.AssignVariable("operations").
				To(gocode.Call("append").
					WithArgument("operations").
					WithParameters(newWeightOperation).
					Build())

			cursor.InsertBefore(&dst.DeclStmt{Decl: gocode.UninitializedVar(weightMsg, "int")})
			cursor.InsertBefore(getOrGenerate)
			cursor.InsertBefore(operations)
		}
		return false
	})
	return tree, nil
}
