package arguard

import (
	"errors"
	"go/ast"
	"go/types"

	"github.com/orsinium-labs/gosafe/contracts"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

func NewAnalyzer(
	config Config,
	contractsAnalyzer *analysis.Analyzer,
) *analysis.Analyzer {
	a := analyzer{config, contractsAnalyzer}
	return &analysis.Analyzer{
		Name:     "gosafe",
		Doc:      "finds code that will fail",
		Run:      a.run,
		Requires: []*analysis.Analyzer{contractsAnalyzer},
		Flags:    *config.flagSet(),
	}
}

type analyzer struct {
	config    Config
	contracts *analysis.Analyzer
}

// run is the entry point for the analyzer
func (a analyzer) run(pass *analysis.Pass) (any, error) {
	rawFuncs, ok := pass.ResultOf[a.contracts]
	if !ok {
		return nil, errors.New("contracts analyzer is required but was not run")
	}
	funcs := rawFuncs.(contracts.Result)
	// analyze every file
	for _, file := range pass.Files {
		fa := fileAnalyzer{
			config: a.config,
			funcs:  funcs,
			pass:   pass,
			file:   file,
		}
		fa.analyze()
	}
	return nil, nil
}

type fileAnalyzer struct {
	config Config
	funcs  contracts.Result
	pass   *analysis.Pass
	file   *ast.File
}

func (fa *fileAnalyzer) analyze() {
	ast.Inspect(fa.file, func(node ast.Node) bool {
		fa.inspect(node)
		return true
	})
}

func (fa *fileAnalyzer) inspect(node ast.Node) {
	// resolve the call target type
	nCall, ok := node.(*ast.CallExpr)
	if !ok {
		return
	}
	obj, ok := typeutil.Callee(fa.pass.TypesInfo, nCall).(*types.Func)
	if !ok { // anonymous function or something
		return
	}

	// get function contracts
	fn, ok := fa.funcs[obj]
	if !ok { // function doesn't have contracts
		return
	}

	// validate contracts
	vars := fn.MapArgs(nCall.Args)
	contract, err := fn.Validate(vars)
	if err != nil {
		if fa.config.ReportErrors {
			fa.pass.Reportf(node.Pos(), "error executing contracts: %v", err)
		}
		return
	}
	if contract != nil {
		fa.pass.Reportf(node.Pos(), "contract violated (%s): %s",
			contract.Condition, contract.Message)
	}
}
