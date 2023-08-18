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

var Analyzer = analysis.Analyzer{
	Name:       "contracts",
	Doc:        "extract conditions that function arguments must satisfy",
	Run:        run,
	ResultType: reflect.TypeOf((Result)(nil)),
	// Flags:      flagSet,
}

// run is the entry point for the analyzer
func run(pass *analysis.Pass) (any, error) {
	facts := make(Result)
	// analyze the current package
	exportFacts(facts, pass.TypesInfo, pass.Files)

	// analyze all imported packages
	for _, file := range pass.Files {
		for _, nImport := range file.Imports {
			importPath := getImportPath(nImport)
			if importPath == "" {
				continue
			}
			pkg, err := loadPackageInfo(importPath)
			if err != nil {
				pass.Reportf(nImport.Pos(), "load package info: %v", err)
			}
			exportFacts(facts, pkg.TypesInfo, pkg.Syntax)
		}
	}
	return facts, nil
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
	if !ok || fdecl.Body == nil {
		return
	}
	obj, ok := info.Defs[fdecl.Name].(*types.Func)
	if !ok {
		return
	}
	_, exists := facts[obj]
	if exists {
		return
	}

	fact := FunctionFromAST(decl)
	if fact == nil {
		return
	}
	facts[obj] = fact
	// fmt.Printf("exported fact for %s\n", obj.FullName())
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
