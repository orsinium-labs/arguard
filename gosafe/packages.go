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
func (ps *Packages) LoadImport(nImport *ast.ImportSpec) error {
	importPath := getImportPath(nImport)
	if importPath == "" {
		return nil
	}
	return ps.LoadName(importPath)
}

// LoadImport loads the package by its full name (import path).
func (ps *Packages) LoadName(pkgName PackageName) error {
	_, contains := ps.pkgs[pkgName]
	if contains {
		return nil
	}
	pkg, err := PackageFromName(pkgName)
	if err != nil {
		return fmt.Errorf("extract contracts from %s: %v", pkgName, err)
	}
	ps.pkgs[pkg.Name] = *pkg
	return nil
}

func (ps *Packages) Add(name PackageName, pkg Package) {
	ps.pkgs[pkg.Name] = pkg
}

func (ps *Packages) Get(name PackageName) *Package {
	pkg, ok := ps.pkgs[name]
	if !ok {
		return nil
	}
	return &pkg
}

func getImportPath(nImport *ast.ImportSpec) PackageName {
	if nImport.Path == nil {
		return ""
	}
	if nImport.Path.Kind != token.STRING {
		return ""
	}
	path := nImport.Path.Value
	path, _ = strings.CutPrefix(path, "\"")
	path, _ = strings.CutSuffix(path, "\"")
	return PackageName(path)
}
