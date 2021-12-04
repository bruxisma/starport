package list

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dave/dst"
	"github.com/gobuffalo/genny"
	"github.com/tendermint/starport/starport/pkg/gocode"
	"github.com/tendermint/starport/starport/pkg/placeholder"
	"github.com/tendermint/starport/starport/templates/typed"
	"github.com/tendermint/starport/starport/templates/typed/list/mutate"
)

func moduleSimulationModify(replacer placeholder.Replacer, opts *typed.Options) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "module_simulation.go")
		file, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		sequence := mutate.GoSequence{}

		tree, err := sequence.Apply(file, opts)
		if err != nil {
			return reportModifyError(path, err)
		}

		fn, err := gocode.FindFunction(tree, "GenerateGenesisState")
		if err != nil {
			return err
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

		buffer, err := gocode.Write(tree)
		if err != nil {
			return reportModifyError(path, err)
		}

		io.Copy(os.Stdout, buffer)
		return errors.New("STOP")

		//		content := replacer.Replace(file.String(), typed.PlaceholderSimappGenesisState, replacementGs)
		//
		//		content = typed.ModuleSimulationMsgModify(
		//			replacer,
		//			content,
		//			opts.ModuleName,
		//			opts.TypeName,
		//			"Create", "Update", "Delete",
		//		)
		//
		// newFile := genny.NewFile(path, buffer)
		// return r.File(newFile)
	}
}
