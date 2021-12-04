package list

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/gobuffalo/genny"
	"github.com/tendermint/starport/starport/pkg/gocode"
	"github.com/tendermint/starport/starport/templates/typed"
	"github.com/tendermint/starport/starport/templates/typed/list/mutate"
)

func moduleSimulationModify(opts *typed.Options) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "module_simulation.go")
		file, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		sequence := mutate.GoSequence{
			mutate.SimulationInsertGenesisState,
			mutate.SimulationInsertConstOpWeightMsg,
			mutate.SimulationInsertWeightedOperations,
		}

		tree, err := sequence.Apply(file, opts)
		if err != nil {
			return reportModifyError(path, err)
		}

		buffer, err := gocode.Write(tree)
		if err != nil {
			return reportModifyError(path, err)
		}

		io.Copy(os.Stdout, buffer)
		return errors.New("STOP")

		newFile := genny.NewFile(path, buffer)
		return r.File(newFile)
	}
}
