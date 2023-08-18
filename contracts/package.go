package contracts

import (
	"errors"
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/packages"
)

type PackageName string

type Package struct {
	Name  PackageName
	funcs []Function
}

func PackageFromName(pkgName PackageName) (*Package, error) {
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
	return PackageFromFiles(pkgName, pkgs[0].Syntax)
}

func PackageFromFiles(pkgName PackageName, files []*ast.File) (*Package, error) {
	funcs := make([]Function, 0)
	for _, astFile := range files {
		funcs = append(funcs, extractFile(astFile)...)
	}
	return &Package{Name: pkgName, funcs: funcs}, nil
}

func extractFile(astFile *ast.File) []Function {
	funcs := make([]Function, 0)
	for _, node := range astFile.Decls {
		fn := FunctionFromAST(node)
		if fn != nil {
			funcs = append(funcs, *fn)
		}
	}
	return funcs
}

func funcRecvName(nFunc *ast.FuncDecl) string {
	if nFunc.Recv == nil {
		return ""
	}
	if len(nFunc.Recv.List) != 1 {
		return ""
	}
	names := nFunc.Recv.List[0].Names
	if len(names) != 1 {
		return ""
	}
	return names[0].Name
}

func (p Package) Function(recv string, name string) *Function {
	for _, fn := range p.funcs {
		if fn.Recv == recv && fn.Name == name {
			return &fn
		}
	}
	return nil
}
