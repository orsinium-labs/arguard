package gosafe

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"

	"github.com/orsinium-labs/gosafe/contracts"
	"golang.org/x/tools/go/analysis"
)

func NewAnalyzer() *analysis.Analyzer {
	var flagSet flag.FlagSet
	return &analysis.Analyzer{
		Name:  "gosafe",
		Doc:   "finds code that will fail",
		Run:   run,
		Flags: flagSet,
		Requires: []*analysis.Analyzer{
			&contracts.Analyzer,
		},
	}
}

// run is the entry point for the analyzer
func run(pass *analysis.Pass) (any, error) {
	contractsResult, ok := pass.ResultOf[&contracts.Analyzer]
	if !ok {
		return nil, errors.New("contracts analyzer is required but was not run")
	}
	pkgs := contractsResult.(*contracts.Packages)
	// analyze every file
	for _, f := range pass.Files {
		fa := fileAnalyzer{pkgs: pkgs, pass: pass, file: f}
		fa.analyze()
	}
	return nil, nil
}

type fileAnalyzer struct {
	pkgs *contracts.Packages
	pass *analysis.Pass
	file *ast.File
}

func (fa *fileAnalyzer) analyze() {

	ast.Inspect(fa.file, func(node ast.Node) bool {
		err := fa.inspect(node)
		if err != nil {
			fa.pass.Reportf(node.Pos(), err.Error())
		}
		return true
	})
}

func (fa *fileAnalyzer) inspect(node ast.Node) error {
	// resolve the call target type
	nCall, ok := node.(*ast.CallExpr)
	if !ok {
		return nil
	}
	nIdent, ok := nCall.Fun.(*ast.Ident)
	if !ok {
		return nil
	}
	obj, ok := fa.pass.TypesInfo.Uses[nIdent]
	if !ok {
		obj = fa.pass.TypesInfo.Defs[nIdent]
	}

	// get the package
	p := obj.Pkg()
	if p == nil { // built-ins
		return nil
	}
	pkgName := contracts.PackageName(p.Name())
	pkg := fa.pkgs.Get(pkgName)
	if pkg == nil {
		return fmt.Errorf("package %s not loaded", pkgName)
	}

	// get function contracts
	fName := obj.Name()
	fn := pkg.Function("", fName)
	if fn == nil { // function doesn't have contracts
		return nil
	}

	// validate contracts
	vars := fn.MapArgs(nCall.Args)
	valid, err := fn.Validate(vars)
	if err != nil {
		return fmt.Errorf("validate contracts: %v", err)
	}
	if !valid {
		fa.pass.Reportf(node.Pos(), "contract violated")
	}
	return nil
}
