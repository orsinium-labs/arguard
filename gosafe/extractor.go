package gosafe

import (
	"errors"
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/packages"
)

type Package struct {
	funcs []Function
}

func Extract(pkgName string) (*Package, error) {
	loadMode := (packages.NeedName |
		packages.NeedSyntax |
		packages.NeedDeps |
		packages.NeedTypes |
		packages.NeedTypesInfo)
	cfg := &packages.Config{Mode: loadMode}
	pkgs, err := packages.Load(cfg, pkgName)
	if err != nil {
		return nil, fmt.Errorf("load package: %v", err)
	}
	if len(pkgs) == 0 {
		return nil, errors.New("no packages loaded")
	}
	if len(pkgs) > 1 {
		return nil, fmt.Errorf("loaded %d packages, expected 1", len(pkgs))
	}
	pkg := pkgs[0]
	funcs := make([]Function, 0)
	for _, astFile := range pkg.Syntax {
		funcs = append(funcs, extractFile(astFile)...)
	}
	return &Package{funcs: funcs}, nil
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
