package arguard

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/types"

	"github.com/orsinium-labs/gosafe/contracts"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
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
	rawFuncs, ok := pass.ResultOf[&contracts.Analyzer]
	if !ok {
		return nil, errors.New("contracts analyzer is required but was not run")
	}
	funcs := rawFuncs.(contracts.Result)
	// analyze every file
	for _, f := range pass.Files {
		fa := fileAnalyzer{funcs: funcs, pass: pass, file: f}
		fa.analyze()
	}
	return nil, nil
}

type fileAnalyzer struct {
	funcs contracts.Result
	pass  *analysis.Pass
	file  *ast.File
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
	obj, ok := typeutil.Callee(fa.pass.TypesInfo, nCall).(*types.Func)
	if !ok { // anonymous function or something
		return nil
	}

	// get function contracts
	fn, ok := fa.funcs[obj]
	if !ok { // function doesn't have contracts
		return nil
	}

	// validate contracts
	vars := fn.MapArgs(nCall.Args)
	contract, err := fn.Validate(vars)
	if err != nil {
		return fmt.Errorf("validate contracts: %v", err)
	}
	if contract != nil {
		fa.pass.Reportf(node.Pos(), "contract violated (%s): %s",
			contract.Condition, contract.Message)
	}
	return nil
}
