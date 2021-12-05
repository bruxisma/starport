package list

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/andybalholm/cascadia"
	"github.com/dave/dst"
	"github.com/gobuffalo/genny"
	"github.com/tendermint/starport/starport/pkg/gocode"
	"github.com/tendermint/starport/starport/pkg/protocode"
	"github.com/tendermint/starport/starport/pkg/xgenny"
	"github.com/tendermint/starport/starport/templates/typed"
	"github.com/tendermint/starport/starport/templates/typed/list/mutate"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	//go:embed stargate/component/* stargate/component/**/*
	fsStargateComponent embed.FS

	//go:embed stargate/messages/* stargate/messages/**/*
	fsStargateMessages embed.FS
)

// NewStargate returns the generator to scaffold a new type in a Stargate module
func NewStargate(opts *typed.Options) (*genny.Generator, error) {
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
		g.RunFn(typesCodecModify(opts))
		g.RunFn(clientCliTxModify(opts))
		g.RunFn(moduleSimulationModify(opts))

		// Messages template
		if err := typed.Box(messagesTemplate, opts, g); err != nil {
			return nil, err
		}
	}

	g.RunFn(frontendSrcStoreAppModify(opts))

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
		if err != nil {
			return err
		}

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
		if err != nil {
			return err
		}

		newFile := genny.NewFile(path, buffer)
		return r.File(newFile)
	}
}

func typesCodecModify(opts *typed.Options) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "types/codec.go")
		file, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		sequences := mutate.GoSequence{
			mutate.StargateTypesCodecRegisterCodec,
			mutate.StargateTypesCodecRegisterInterfaces,
		}

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

// Unfortunately due to limitations with net/html, we cannot use the same
// behavior we've used everywhere else, and instead must *give up* and then
// place the SpType node in as a RawNode 😢 This also seriously limits how the
// text ends up looking as well. For now, we hard code things, but this needs
// work unfortunately
func frontendSrcStoreAppModify(opts *typed.Options) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "vue/src/views/Types.vue")
		file, err := r.Disk.Find(path)
		// Skip modification if the app doesn't contain front-end
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		} else if err != nil {
			return err
		}

		selector := cascadia.MustCompile(".container")

		nodes, err := html.ParseFragment(file, &html.Node{
			Type:     html.ElementNode,
			Data:     "body",
			DataAtom: atom.Body,
		})
		if err != nil {
			return err
		}

		if len(nodes) == 0 {
			return fmt.Errorf(`could not locate any valid vue tags in %q`, path)
		}

		div := cascadia.Query(nodes[0], selector)
		if div == nil {
			return fmt.Errorf(`could not locate <div class="container"> in %q`, path)
		}
		modulePath := strings.Join(
			[]string{opts.OwnerName, opts.AppName, opts.ModuleName},
			".")
		moduleType := opts.TypeName.UpperCamel
		div.AppendChild(&html.Node{
			Type: html.RawNode,
			Data: fmt.Sprintf("  <SpType modulePath=%q moduleType=%q />\n    ", modulePath, moduleType),
		})

		buffer := &bytes.Buffer{}
		for _, node := range nodes {
			html.Render(buffer, node)
		}

		newFile := genny.NewFile(path, buffer)
		return r.File(newFile)
	}
}
