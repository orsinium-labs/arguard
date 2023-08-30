package contracts

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"
)

type Result map[*types.Func]*Function

func NewAnalyzer(config Config) *analysis.Analyzer {
	analyzer := analyzer{&config}
	return &analysis.Analyzer{
		Name:       "contracts",
		Doc:        "extracts conditions that function arguments must satisfy",
		Run:        analyzer.run,
		ResultType: reflect.TypeOf((Result)(nil)),
		Flags:      *config.flagSet(),
	}
}

type analyzer struct {
	config *Config
}

// run is the entry point for the analyzer.
func (a analyzer) run(pass *analysis.Pass) (any, error) {
	facts := make(Result)

	// analyze the current package
	exportFacts(facts, pass.TypesInfo, pass.Files)

	// analyze all imported packages
	if a.config.FollowImports {
		analyzeImports(facts, pass)
	}

	// if in debug mode, report all detected contracts
	if a.config.ReportContracts {
		for _, fInfo := range facts {
			for _, c := range fInfo.Contracts {
				pass.Reportf(c.Pos, "contract: %s", c.Message)
			}
		}
	}

	return facts, nil
}

func analyzeImports(facts Result, pass *analysis.Pass) {
	analyzedPackages := make(map[string]struct{})
	for _, file := range pass.Files {
		for _, nImport := range file.Imports {
			importPath := getImportPath(nImport)
			if importPath == "" { // shouldn't happen for any valid AST
				continue
			}
			_, analyzed := analyzedPackages[importPath]
			if analyzed { // already analyzed, skip
				continue
			}
			analyzedPackages[importPath] = struct{}{}
			pkg, err := loadPackageInfo(importPath)
			if err != nil {
				pass.Reportf(nImport.Pos(), "load package info: %v", err)
			}
			exportFacts(facts, pkg.TypesInfo, pkg.Syntax)
		}
	}
}

func loadPackageInfo(pkgName string) (*packages.Package, error) {
	loadMode := (packages.NeedName |
		packages.NeedSyntax |
		packages.NeedDeps |
		packages.NeedTypes |
		packages.NeedTypesInfo)
	cfg := &packages.Config{Mode: loadMode}
	pkgs, err := packages.Load(cfg, string(pkgName))
	if err != nil {
		return nil, fmt.Errorf("load package: %v", err)
	}
	if len(pkgs) == 0 {
		return nil, errors.New("no packages loaded")
	}
	if len(pkgs) > 1 {
		return nil, fmt.Errorf("loaded %d packages, expected 1", len(pkgs))
	}
	return pkgs[0], nil
}

func exportFacts(facts Result, info *types.Info, files []*ast.File) {
	for _, file := range files {
		for _, decl := range file.Decls {
			exportFact(facts, info, decl)
		}
	}
}

func exportFact(facts Result, info *types.Info, decl ast.Decl) {
	fdecl, ok := decl.(*ast.FuncDecl)
	if !ok || fdecl.Body == nil { // not a func declaration or func without a body
		return
	}
	obj, ok := info.Defs[fdecl.Name].(*types.Func)
	if !ok {
		return
	}
	_, exists := facts[obj]
	if exists { // already analyzed
		return
	}

	fact := functionFromAST(fdecl, info)
	if fact == nil {
		return
	}
	facts[obj] = fact
}

func getImportPath(nImport *ast.ImportSpec) string {
	if nImport.Path == nil {
		return ""
	}
	if nImport.Path.Kind != token.STRING {
		return ""
	}
	path := nImport.Path.Value
	path, _ = strings.CutPrefix(path, "\"")
	path, _ = strings.CutSuffix(path, "\"")
	return path
}
