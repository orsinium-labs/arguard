package contracts

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"strings"

	"github.com/traefik/yaegi/interp"
)

type Contract struct {
	Pos       token.Pos // contract position, used for positioning debug messages
	Condition string    // valid Go-syntax expression which if true, the contract is violated
	Names     []string  // unbound variables used by the condition
	Message   string    // error message to show on contract failure
}

// contractFromAST returns a contract if the given AST node looks like one.
//
// The returned error explains why the node cannot be converted into a contract.
// It might be not an if statement, not use input args, be unsafe to statically execute
// and lots of other reasons. Most of the real code isn't a contract.
func contractFromAST(node ast.Node, info *types.Info) (*Contract, error) {
	nIf, ok := node.(*ast.IfStmt)
	if !ok {
		return nil, errors.New("not an if statement")
	}
	cond, names, err := expr2string(nIf.Cond, info)
	if err != nil {
		return nil, fmt.Errorf("extract condition: %v", err)
	}
	if len(names) == 0 {
		return nil, errors.New("condition is static (uses no variables)")
	}
	msg, isError := extractMessage(nIf.Body, info)
	if !isError {
		return nil, errors.New("body doesn't look like a contract")
	}
	if msg == "" {
		msg = "should be false: " + cond
	}
	return &Contract{node.Pos(), cond, names, msg}, nil
}

// allDefined checks if vars define all unbound variables needed to execute the contract.
func (c Contract) allDefined(vars map[string]string) bool {
	for _, name := range c.Names {
		_, defined := vars[name]
		if !defined {
			return false
		}
	}
	return true
}

// validate returns false if the contract is violated.
func (c Contract) validate(interpreter *interp.Interpreter) (bool, error) {
	res, err := safeEval(interpreter, c.Condition)
	if err != nil {
		return false, fmt.Errorf("evaluate condition: %v", err)
	}
	condOk, isBool := res.Interface().(bool)
	if !isBool {
		return false, fmt.Errorf("condition result: expected bool, actual %s", res.Kind())
	}
	return !condOk, nil
}

// expr2string converts the given AST expression into a valid Go syntax string.
//
// Returns an error for unsupported or not safe to execute expressions.
func expr2string(expr ast.Expr, info *types.Info) (string, []string, error) {
	switch v := expr.(type) {
	case *ast.BinaryExpr:
		lExpr, lNames, err := expr2string(v.X, info)
		if err != nil {
			return "", nil, err
		}
		rExpr, rNames, err := expr2string(v.Y, info)
		if err != nil {
			return "", nil, err
		}
		cond := fmt.Sprintf("%s %s %s", lExpr, v.Op.String(), rExpr)
		names := append(lNames, rNames...)
		return cond, names, nil
	case *ast.BasicLit:
		return v.Value, nil, nil
	case *ast.Ident:
		folded := foldConstant(v, info)
		if folded != "" {
			return folded, nil, nil
		}
		return v.Name, []string{v.Name}, nil
	default:
		return "", nil, fmt.Errorf("unsupported node: %v", expr)
	}
}

func foldConstant(nIdent *ast.Ident, info *types.Info) string {
	constType, ok := info.Types[nIdent]
	if !ok {
		return ""
	}
	if constType.Value == nil {
		return ""
	}
	return constType.Value.ExactString()
}

// extractMessage extracts error message for the contract.
//
// The first result value is the extraccted error message
// which might be empty if it cannot be extracted.
//
// The second result value tells if the given code block
// is an error of some kind typical for a contract.
// A contract must either panic or return an error as one of the return values.
func extractMessage(nBody *ast.BlockStmt, info *types.Info) (string, bool) {
	if nBody.List == nil {
		return "", false
	}
	if len(nBody.List) != 1 {
		return "", false
	}
	nStmt := nBody.List[0]

	// check if it's panic
	nExpr, ok := nStmt.(*ast.ExprStmt)
	if ok {
		return extractMessageFromPanic(nExpr)
	}

	nRet, ok := nStmt.(*ast.ReturnStmt)
	if ok {
		return extractMessageFromReturn(nRet, info)
	}
	return "", false
}

func extractMessageFromPanic(nExpr *ast.ExprStmt) (string, bool) {
	// check if the expression is a "panic"
	if nExpr.X == nil {
		return "", false
	}
	nCall, ok := nExpr.X.(*ast.CallExpr)
	if !ok {
		return "", false
	}
	if nCall.Fun == nil {
		return "", false
	}
	nIdent, ok := nCall.Fun.(*ast.Ident)
	if !ok {
		return "", false
	}
	if nIdent.Name != "panic" {
		return "", false
	}
	if nCall.Args == nil {
		return "", false
	}
	if len(nCall.Args) != 1 {
		return "", false
	}

	// extract the error message
	nArg := nCall.Args[0]
	nLit, ok := nArg.(*ast.BasicLit)
	if ok {
		if nLit.Kind == token.STRING {
			return strings.Trim(nLit.Value, `"`), true
		}
		return nLit.Value, true
	}
	return "", true
}

func extractMessageFromReturn(nRet *ast.ReturnStmt, info *types.Info) (string, bool) {
	if nRet.Results == nil {
		return "", false
	}
	for _, nExpr := range nRet.Results {
		exprType, ok := info.Types[nExpr]
		if !ok {
			continue
		}
		if exprType.Type.String() == "error" {
			// TODO: get the error message from fmt.Errorf and alike
			return "", true
		}
	}
	return "", false
}

// safeEval evals the expression using the interpreter and catches panics.
func safeEval(i *interp.Interpreter, expr string) (res reflect.Value, err error) {
	defer func() {
		failure := recover()
		if failure != nil {
			err = fmt.Errorf("panic: %v", failure)
		}
	}()
	return i.Eval(expr)
}
