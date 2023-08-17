package gosafe

import (
	"flag"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func NewAnalyzer() *analysis.Analyzer {
	var flagSet flag.FlagSet
	return &analysis.Analyzer{
		Name:  "gosafe",
		Doc:   "finds code that will fail",
		Run:   run,
		Flags: flagSet,
	}
}

func run(pass *analysis.Pass) (any, error) {
	pkgs := NewPackages()
	for _, f := range pass.Files {
		fa := fileAnalyzer{pkgs: pkgs, pass: pass, file: f}
		fa.analyze()
	}
	return nil, nil
}

type fileAnalyzer struct {
	pkgs *Packages
	pass *analysis.Pass
	file *ast.File
}

func (fa *fileAnalyzer) analyze() {
	ast.Inspect(fa.file, func(node ast.Node) bool {
		err := fa.pkgs.LoadImport(node)
		if err != nil {
			fa.pass.Reportf(node.Pos(), "cannot parse dependency package: %v", err)
			return true
		}
		return true
	})
}

// func (fa *fileAnalyzer) inspect(node ast.Node) bool {
// 	nCall, ok := node.(*ast.CallExpr)
// 	if !ok {
// 		return true
// 	}
// 	nIdent, ok := nCall.Fun.(*ast.Ident)
// 	if !ok {
// 		return true
// 	}
// 	obj, ok := fa.pass.TypesInfo.Uses[nIdent]
// 	if !ok {
// 		obj = fa.pass.TypesInfo.Defs[nIdent]
// 	}

// 	return true
// }

// func (i *Inspector) report() {
// 	for name, count := range i.counts {
// 		if count != 0 {
// 			i.pass.Reportf(0, "%s: %d", name, count)
// 		}
// 	}
// }
