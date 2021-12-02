package list

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/emicklei/proto"
	"github.com/gobuffalo/genny"
	"github.com/tendermint/starport/starport/pkg/gocode"
	"github.com/tendermint/starport/starport/pkg/placeholder"
	"github.com/tendermint/starport/starport/pkg/protocode"
	"github.com/tendermint/starport/starport/templates/typed"
)

func genesisModify(replacer placeholder.Replacer, opts *typed.Options, g *genny.Generator) {
	g.RunFn(genesisProtoModify(replacer, opts))
	g.RunFn(genesisTypesModify(replacer, opts))
	g.RunFn(genesisModuleModify(replacer, opts))
	g.RunFn(genesisTestsModify(replacer, opts))
	g.RunFn(genesisTypesTestsModify(replacer, opts))
}

func genesisProtoModify(replacer placeholder.Replacer, opts *typed.Options) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "proto", opts.ModuleName, "genesis.proto")
		file, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		tree, err := protocode.Parse(file, path)
		if err != nil {
			return fmt.Errorf("proto filaure: %w", err)
		}

		tree, err = protocode.AppendImportf(tree, "%s/%s.proto", opts.ModuleName, opts.TypeName.Snake)
		if err != nil {
			return err
		}

		// TODO: Add GogoProtoImport path to *front* of imports
		// Add gogo.proto
		// replacementGogoImport := typed.EnsureGogoProtoImported(path, typed.PlaceholderGenesisProtoImport)
		// content = replacer.Replace(content, typed.PlaceholderGenesisProtoImport, replacementGogoImport)

		message, err := protocode.FindMessage(tree, "GenesisState")
		if err != nil {
			return nil
		}

		// determine highest field number
		var seq int
		for _, item := range message.Elements {
			if field, ok := item.(*proto.NormalField); ok {
				seq = field.Sequence
			}
			fmt.Printf("%[1]T with %#[1]v\n", item)
		}
		seq++
		message.Elements = append(message.Elements,
			&proto.NormalField{
				Field: &proto.Field{
					Name:     fmt.Sprintf("%sList", opts.TypeName.LowerCamel),
					Type:     opts.TypeName.UpperCamel,
					Sequence: seq,
					Options: []*proto.Option{
						{
							Name:     "(gogoproto.nullable)",
							Constant: proto.Literal{Source: "false"},
						},
					},
				},
				Repeated: true,
			},
			&proto.NormalField{
				Field: &proto.Field{
					Name:     fmt.Sprintf("%sCount", opts.TypeName.LowerCamel),
					Type:     "uint64",
					Sequence: seq + 1,
				},
			})

		buffer := &bytes.Buffer{}
		if err = protocode.Fprint(buffer, tree); err != nil {
			return nil
		}

		newFile := genny.NewFileS(path, buffer.String())
		return r.File(newFile)
	}
}

func genesisTypesModify(replacer placeholder.Replacer, opts *typed.Options) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "types/genesis.go")
		file, err := r.Disk.Find(path)
		if err != nil {
			return err
		}
		mutations := mutators{
			mutateTypesDefaultGenesisReturnValue,
			mutateTypesValidateStatements,
		}
		tree, err := mutations.Apply(file, opts)
		if err != nil {
			return reportModifyError(path, err)
		}
		if tree, err = typed.MutateImport(tree, "fmt"); err != nil {
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

func genesisModuleModify(replacer placeholder.Replacer, opts *typed.Options) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "genesis.go")
		file, err := r.Disk.Find(path)
		if err != nil {
			return err
		}
		mutations := mutators{
			genesisModuleInsertInit,
			genesisModuleInsertExport,
		}
		tree, err := mutations.Apply(file, opts)
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

func genesisTestsModify(replacer placeholder.Replacer, opts *typed.Options) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "genesis_test.go")
		file, err := r.Disk.Find(path)
		if err != nil {
			return err
		}
		mutations := mutators{
			genesisTestsInsertList,
			genesisTestsInsertComparison,
		}
		tree, err := mutations.Apply(file, opts)
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

func genesisTypesTestsModify(replacer placeholder.Replacer, opts *typed.Options) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "types/genesis_test.go")
		file, err := r.Disk.Find(path)
		if err != nil {
			return err
		}
		mutations := mutators{
			genesisTestsInsertValidGenesisState,
			genesisTestsInsertDuplicatedState,
			genesisTestsInsertInvalidCount,
		}
		tree, err := mutations.Apply(file, opts)
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
