package contracts

import (
	"fmt"
	"reflect"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = analysis.Analyzer{
	Name:       "contracts",
	Doc:        "extract conditions that function arguments must satisfy",
	Run:        run,
	ResultType: reflect.TypeOf(&Packages{}),
	// Flags:      flagSet,
}

// run is the entry point for the analyzer
func run(pass *analysis.Pass) (any, error) {
	// create packages registry and load analyzed package into it
	pkgs := NewPackages()
	name := PackageName(pass.Pkg.Name())
	pkg, err := PackageFromFiles(name, pass.Files)
	if err != nil {
		return nil, fmt.Errorf("read contracts from the package: %v", err)
	}
	pkgs.Add(name, *pkg)

	// analyze every file
	for _, file := range pass.Files {
		for _, node := range file.Imports {
			err := pkgs.LoadImport(node)
			if err != nil {
				pass.Reportf(node.Pos(), "cannot parse dependency package: %v", err)
			}
		}
	}
	return pkgs, nil
}
