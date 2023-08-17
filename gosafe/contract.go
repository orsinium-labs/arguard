package gosafe

import (
	"errors"
	"fmt"
	"go/ast"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type Contract struct {
	Condition ast.Expr
	Message   string
}

func ContractFromAST(node ast.Node) *Contract {
	nIf, ok := node.(*ast.IfStmt)
	if !ok {
		return nil
	}
	// nIf.Body
	return &Contract{nIf.Cond, "pre-condition failed"}
}

func (c Contract) Validate(vars map[string]string) (bool, error) {
	i := interp.New(interp.Options{})
	err := i.Use(stdlib.Symbols)
	if err != nil {
		return false, fmt.Errorf("use stdlib: %v", err)
	}
	i.ImportUsed()
	for name, val := range vars {
		expr := fmt.Sprintf("%s := %s", name, val)
		_, err = i.Eval(expr)
		if err != nil {
			return false, fmt.Errorf("set value for %s: %v", name, err)
		}
	}
	prog, err := i.CompileAST(c.Condition)
	if err != nil {
		return false, fmt.Errorf("compile condition: %v", err)
	}
	res, err := i.Execute(prog)
	if err != nil {
		return false, fmt.Errorf("execute condition: %v", err)
	}
	condOk, isBool := res.Interface().(bool)
	if !isBool {
		return false, errors.New("condition result is not bool")
	}
	return condOk, nil
}
