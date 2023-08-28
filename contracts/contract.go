package contracts

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/traefik/yaegi/interp"
)

type Contract struct {
	Pos       token.Pos // contract position, used for positioning debug messages
	Condition string    // valid Go-syntax expression which if true, the contract is violated
	Message   string    // error message to show on contract failure
}

func contractFromAST(node ast.Node) *Contract {
	nIf, ok := node.(*ast.IfStmt)
	if !ok {
		return nil
	}
	cond, err := extractCondition(nIf.Cond)
	if err != nil {
		return nil
	}
	msg := extractMessage(nIf.Body)
	if msg == "" {
		msg = "pre-condition failed"
	}
	return &Contract{node.Pos(), cond, msg}
}

// vlaidate returns false if the contract is violated.
func (c Contract) validate(interpreter *interp.Interpreter) (bool, error) {
	res, err := interpreter.Eval(c.Condition)
	if err != nil {
		return false, fmt.Errorf("evaluate condition: %v", err)
	}
	condOk, isBool := res.Interface().(bool)
	if !isBool {
		return false, fmt.Errorf("condition result: expected bool, actual %s", res.Kind())
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
