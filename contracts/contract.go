package contracts

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"strings"

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
	cond, err := extractCondition(nIf.Cond)
	if err != nil {
		return nil
	}
	msg := extractMessage(nIf.Body)
	if msg == "" {
		msg = "pre-condition failed"
	}
	return &Contract{node, cond, msg}
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

// extractCondition converts the given AST expression into a valid Go syntax string.
//
// Returns an error for unsupported or not safe to execute expressions.
func extractCondition(expr ast.Expr) (string, error) {
	switch v := expr.(type) {
	case *ast.BinaryExpr:
		left, err := extractCondition(v.X)
		if err != nil {
			return "", err
		}
		right, err := extractCondition(v.Y)
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

func extractMessage(nBody *ast.BlockStmt) string {
	if nBody.List == nil {
		return ""
	}
	if len(nBody.List) != 1 {
		return ""
	}
	nStmt := nBody.List[0]

	// check if it's panic
	nExpr, ok := nStmt.(*ast.ExprStmt)
	if ok {
		return extractMessageFromPanic(nExpr)
	}

	return ""
}

func extractMessageFromPanic(nExpr *ast.ExprStmt) string {
	// check if the expression is a "panic"
	if nExpr.X == nil {
		return ""
	}
	nCall, ok := nExpr.X.(*ast.CallExpr)
	if !ok {
		return ""
	}
	if nCall.Fun == nil {
		return ""
	}
	nIdent, ok := nCall.Fun.(*ast.Ident)
	if !ok {
		return ""
	}
	if nIdent.Name != "panic" {
		return ""
	}
	if nCall.Args == nil {
		return ""
	}
	if len(nCall.Args) != 1 {
		return ""
	}

	// extract the error message
	nArg := nCall.Args[0]
	nLit, ok := nArg.(*ast.BasicLit)
	if ok {
		if nLit.Kind == token.STRING {
			return strings.Trim(nLit.Value, `"`)
		}
		return nLit.Value
	}

	return ""
}
