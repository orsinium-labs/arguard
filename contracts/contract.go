package contracts

import (
	"errors"
	"fmt"
	"go/ast"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type Contract struct {
	node      ast.Node
	Condition string
	Message   string
}

func contractFromAST(node ast.Node) *Contract {
	nIf, ok := node.(*ast.IfStmt)
	if !ok {
		return nil
	}
	// nIf.Body
	cond, err := stringify(nIf.Cond)
	if err != nil {
		return nil
	}
	return &Contract{node, cond, "pre-condition failed"}
}

// vlaidate returns false if the contract is violated.
func (c Contract) validate(vars map[string]string) (bool, error) {
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
	res, err := i.Eval(c.Condition)
	if err != nil {
		return false, fmt.Errorf("evaluate condition: %v", err)
	}
	condOk, isBool := res.Interface().(bool)
	if !isBool {
		return false, errors.New("condition result is not bool")
	}
	return !condOk, nil
}

// Convert the given AST expression into a valid Go syntax string.
//
// Returns an error for unsupported or not safe to execute expressions.
func stringify(expr ast.Expr) (string, error) {
	switch v := expr.(type) {
	case *ast.BinaryExpr:
		left, err := stringify(v.X)
		if err != nil {
			return "", err
		}
		right, err := stringify(v.Y)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s %s %s", left, v.Op.String(), right), nil
	case *ast.BasicLit:
		return v.Value, nil
	case *ast.Ident:
		return v.Name, nil
	default:
		return "", fmt.Errorf("unsupported node: %v", expr)
	}
}
