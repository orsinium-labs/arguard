package gosafe

import (
	"fmt"
	"go/ast"
)

type Packages struct {
	pkgs map[string]Package
}

func (ps *Packages) Load(node ast.Node) error {
	importPath := getImportPath(node)
	if importPath == "" {
		return nil
	}
	_, contains := ps.pkgs[importPath]
	if contains {
		return nil
	}
	pkg, err := Extract(importPath)
	if err != nil {
		return fmt.Errorf("extract contracts from %s: %v", importPath, err)
	}
	ps.pkgs[pkg.Name] = *pkg
	return nil
}
