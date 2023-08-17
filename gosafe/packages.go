package gosafe

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

type Packages struct {
	pkgs map[PackageName]Package
}

func NewPackages() *Packages {
	return &Packages{
		pkgs: make(map[PackageName]Package),
	}
}

// LoadImport loads the package from the given import statement node.
func (ps *Packages) LoadImport(node ast.Node) error {
	importPath := getImportPath(node)
	if importPath == "" {
		return nil
	}
	return ps.LoadName(PackageName(importPath))
}

// LoadImport loads the package by its full name (import path).
func (ps *Packages) LoadName(pkgName PackageName) error {
	_, contains := ps.pkgs[pkgName]
	if contains {
		return nil
	}
	pkg, err := Extract(pkgName)
	if err != nil {
		return fmt.Errorf("extract contracts from %s: %v", pkgName, err)
	}
	ps.pkgs[pkg.Name] = *pkg
	return nil
}

func getImportPath(node ast.Node) string {
	nImport, ok := node.(*ast.ImportSpec)
	if !ok {
		return ""
	}
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
