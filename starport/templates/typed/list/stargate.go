package list

import (
	"embed"
	"errors"
	"fmt"
	"go/token"
	"io"
	"os"
	"path/filepath"

	"github.com/dave/dst"
	"github.com/gobuffalo/genny"
	"github.com/tendermint/starport/starport/pkg/gocode"
	"github.com/tendermint/starport/starport/pkg/placeholder"
	"github.com/tendermint/starport/starport/pkg/protocode"
	"github.com/tendermint/starport/starport/pkg/xgenny"
	"github.com/tendermint/starport/starport/templates/typed"
	"github.com/tendermint/starport/starport/templates/typed/list/mutate"
)

var (
	//go:embed stargate/component/* stargate/component/**/*
	fsStargateComponent embed.FS

	//go:embed stargate/messages/* stargate/messages/**/*
	fsStargateMessages embed.FS
)

// NewStargate returns the generator to scaffold a new type in a Stargate module
func NewStargate(replacer placeholder.Replacer, opts *typed.Options) (*genny.Generator, error) {
	var (
		g = genny.New()

		messagesTemplate = xgenny.NewEmbedWalker(
			fsStargateMessages,
			"stargate/messages/",
			opts.AppPath,
		)
		componentTemplate = xgenny.NewEmbedWalker(
			fsStargateComponent,
			"stargate/component/",
			opts.AppPath,
		)
	)

	g.RunFn(protoQueryModify(opts))
	g.RunFn(moduleGRPCGatewayModify(opts))
	g.RunFn(typesKeyModify(opts))
	g.RunFn(clientCliQueryModify(opts))

	// Genesis modifications
	genesisModify(opts, g)

	if !opts.NoMessage {
		// Modifications for new messages
		g.RunFn(handlerModify(opts))
		g.RunFn(protoTxModify(opts))
		g.RunFn(typesCodecModify(replacer, opts))
		g.RunFn(clientCliTxModify(opts))
		g.RunFn(moduleSimulationModify(replacer, opts))

		// Messages template
		if err := typed.Box(messagesTemplate, opts, g); err != nil {
			return nil, err
		}
	}

	g.RunFn(frontendSrcStoreAppModify(replacer, opts))

	return g, typed.Box(componentTemplate, opts, g)
}

func handlerModify(opts *typed.Options) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "handler.go")
		file, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		sequences := mutate.GoSequence{
			mutate.StargateHandlerInsertMsgServer,
			mutate.StargateHandlerInsertCases,
		}

		tree, err := sequences.Apply(file, opts)
		if err != nil {
			return reportModifyError(path, err)
		}
		buffer, err := gocode.Write(tree)
		if err != nil {
			return reportModifyError(path, err)
		}

		newFile := genny.NewFile(path, buffer)
		return r.File(newFile)
	}
}

func protoTxModify(opts *typed.Options) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "proto", opts.ModuleName, "tx.proto")
		file, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		tree, err := protocode.Parse(file, path)
		if err != nil {
			return fmt.Errorf("protocode parse failure: %w", err)
		}

		tree, err = mutate.StargateProtoTx(tree, opts)
		if err != nil {
			return err
		}

		buffer, err := protocode.Write(tree)

		newFile := genny.NewFile(path, buffer)
		return r.File(newFile)
	}
}

func protoQueryModify(opts *typed.Options) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "proto", opts.ModuleName, "query.proto")
		file, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		tree, err := protocode.Parse(file, path)
		if err != nil {
			return fmt.Errorf("protocode parse failure: %w", err)
		}

		tree, err = mutate.StargateProtoQuery(tree, opts)
		if err != nil {
			return err
		}

		buffer, err := protocode.Write(tree)
		if err != nil {
			return err
		}

		newFile := genny.NewFile(path, buffer)
		return r.File(newFile)
	}
}

func moduleGRPCGatewayModify(opts *typed.Options) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "module.go")
		file, err := r.Disk.Find(path)
		if err != nil {
			return err
		}
		sequence := mutate.GoSequence{
			mutate.StargateInsertGRPCGateway,
		}
		tree, err := sequence.Apply(file, opts)
		if err != nil {
			return reportModifyError(path, err)
		}
		if tree, err = typed.MutateImport(tree, "context"); err != nil {
			return reportModifyError(path, err)
		}
		buffer, err := gocode.Write(tree)
		if err != nil {
			return reportModifyError(path, err)
		}

		newFile := genny.NewFile(path, buffer)
		return r.File(newFile)
	}
}

func typesKeyModify(opts *typed.Options) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "types/keys.go")
		file, err := r.Disk.Find(path)
		if err != nil {
			return err
		}
		sequence := mutate.GoSequence{}
		tree, err := sequence.Apply(file, opts)
		if err != nil {
			return err
		}

		// TODO: Check if we need to make one spec per assignment
		key := &dst.ValueSpec{
			Names:  []*dst.Ident{gocode.Name("%sKey", opts.TypeName.UpperCamel)},
			Values: []dst.Expr{gocode.BasicStringf("%s-value-", opts.TypeName.UpperCamel)},
		}
		count := &dst.ValueSpec{
			Names:  []*dst.Ident{gocode.Name("%sCountKey", opts.TypeName.UpperCamel)},
			Values: []dst.Expr{gocode.BasicStringf("%s-count-", opts.TypeName.UpperCamel)},
		}

		tree.Decls = append(tree.Decls, &dst.GenDecl{
			Tok:   token.CONST,
			Specs: []dst.Spec{key, count},
		})
		buffer, err := gocode.Write(tree)

		newFile := genny.NewFile(path, buffer)
		return r.File(newFile)
	}
}

func typesCodecModify(replacer placeholder.Replacer, opts *typed.Options) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "types/codec.go")
		file, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		sequences := mutate.GoSequence{}

		tree, err := sequences.Apply(file, opts)
		if err != nil {
			return err
		}

		tree, err = typed.MutateImport(tree, "sdk", "github.com/cosmos/cosmos-sdk/types")
		if err != nil {
			return err
		}

		buffer, err := gocode.Write(tree)
		if err != nil {
			return err
		}

		//		// Concrete
		//		templateConcrete := `cdc.RegisterConcrete(&MsgCreate%[2]v{}, "%[3]v/Create%[2]v", nil)
		//cdc.RegisterConcrete(&MsgUpdate%[2]v{}, "%[3]v/Update%[2]v", nil)
		//cdc.RegisterConcrete(&MsgDelete%[2]v{}, "%[3]v/Delete%[2]v", nil)
		//%[1]v`
		//		replacementConcrete := fmt.Sprintf(templateConcrete, typed.Placeholder2, opts.TypeName.UpperCamel, opts.ModuleName)
		//		content := replacer.Replace(content, typed.Placeholder2, replacementConcrete)
		//
		//		// Interface
		//		templateInterface := `registry.RegisterImplementations((*sdk.Msg)(nil),
		//	&MsgCreate%[2]v{},
		//	&MsgUpdate%[2]v{},
		//	&MsgDelete%[2]v{},
		//)
		//%[1]v`
		//		replacementInterface := fmt.Sprintf(templateInterface, typed.Placeholder3, opts.TypeName.UpperCamel)
		//		content = replacer.Replace(content, typed.Placeholder3, replacementInterface)

		io.Copy(os.Stdout, buffer)

		return errors.New("STOP")

		newFile := genny.NewFile(path, buffer)
		return r.File(newFile)
	}
}

func clientCliTxModify(opts *typed.Options) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "client/cli/tx.go")
		file, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		sequences := mutate.GoSequence{
			mutate.StargateClientCliTxInsertCommands,
		}
		tree, err := sequences.Apply(file, opts)
		if err != nil {
			return reportModifyError(path, err)
		}
		buffer, err := gocode.Write(tree)
		if err != nil {
			return reportModifyError(path, err)
		}

		newFile := genny.NewFile(path, buffer)
		return r.File(newFile)
	}
}

func clientCliQueryModify(opts *typed.Options) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "client/cli/query.go")
		file, err := r.Disk.Find(path)
		if err != nil {
			return err
		}
		sequences := mutate.GoSequence{
			mutate.StargateClientCliQueryInserCommands,
		}
		tree, err := sequences.Apply(file, opts)
		if err != nil {
			return reportModifyError(path, err)
		}
		buffer, err := gocode.Write(tree)
		if err != nil {
			return reportModifyError(path, err)
		}
		newFile := genny.NewFile(path, buffer)
		return r.File(newFile)
	}
}

func frontendSrcStoreAppModify(replacer placeholder.Replacer, opts *typed.Options) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "vue/src/views/Types.vue")
		f, err := r.Disk.Find(path)
		if os.IsNotExist(err) {
			// Skip modification if the app doesn't contain front-end
			return nil
		}
		if err != nil {
			return err
		}
		replacement := fmt.Sprintf(`%[1]v
		<SpType modulePath="%[2]v.%[3]v.%[4]v" moduleType="%[5]v"  />`,
			typed.Placeholder4,
			opts.OwnerName,
			opts.AppName,
			opts.ModuleName,
			opts.TypeName.UpperCamel,
		)
		content := replacer.Replace(f.String(), typed.Placeholder4, replacement)
		newFile := genny.NewFileS(path, content)
		return r.File(newFile)
	}
}
