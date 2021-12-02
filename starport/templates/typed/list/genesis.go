package list

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gobuffalo/genny"
	"github.com/tendermint/starport/starport/pkg/gocode"
	"github.com/tendermint/starport/starport/pkg/placeholder"
	"github.com/tendermint/starport/starport/pkg/protocode"
	"github.com/tendermint/starport/starport/templates/typed"
	"github.com/tendermint/starport/starport/templates/typed/list/mutate"
)

func reportModifyError(path string, err error) error {
	return fmt.Errorf("modifying %q errored with %w", path, err)
}

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
			return fmt.Errorf("protocode parse failure: %w", err)
		}

		of := protocode.NewOrganizedFile(tree)

		// Ensure gogoproto/gogo.proto is the *first* import in the list
		if idx := of.IndexOfImport("gogoproto/gogo.proto"); idx > 0 {
			of.RemoveImportAt(idx)
		}
		of.PrependImport("gogoproto/gogo.proto")
		of.AppendImportf("%s/%s.proto", opts.ModuleName, opts.TypeName.Snake)

		of, err = mutate.GenesisProtoGenesisState(of, opts)
		if err != nil {
			return reportModifyError(path, err)
		}

		buffer, err := protocode.Write(of.AsFile())
		if err != nil {
			return reportModifyError(path, err)
		}

		io.Copy(os.Stdout, buffer)
		return errors.New("STOP")

		newFile := genny.NewFile(path, buffer)
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
		sequence := mutate.GoSequence{
			mutate.GenesisTypesDefaultGenesisReturnValue,
			mutate.GenesisTypesValidateStatements,
		}
		tree, err := sequence.Apply(file, opts)
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
		sequence := mutate.GoSequence{
			mutate.GenesisModuleInsertInit,
			mutate.GenesisModuleInsertExport,
		}
		tree, err := sequence.Apply(file, opts)
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
		sequence := mutate.GoSequence{
			mutate.GenesisTestsInsertList,
			mutate.GenesisTestsInsertComparison,
		}
		tree, err := sequence.Apply(file, opts)
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
		sequence := mutate.GoSequence{
			mutate.GenesisTestsInsertValidGenesisState,
			mutate.GenesisTestsInsertDuplicatedState,
			mutate.GenesisTestsInsertInvalidCount,
		}
		tree, err := sequence.Apply(file, opts)
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
