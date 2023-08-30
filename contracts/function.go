package contracts

import (
	"fmt"
	"go/ast"
	"go/types"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type Function struct {
	Args      []string
	Contracts []Contract
}

func (*Function) AFact() {}

func functionFromAST(nFunc *ast.FuncDecl, info *types.Info) *Function {
	if nFunc.Body == nil {
		return nil
	}
	contracts := make([]Contract, 0)
	for _, stmt := range nFunc.Body.List {
		contract := contractFromAST(stmt, info)
		if contract == nil {
			break
		}
		contracts = append(contracts, *contract)
	}
	if len(contracts) == 0 {
		return nil
	}
	args := getFuncArgs(nFunc)
	if len(args) == 0 {
		return nil
	}
	return &Function{
		Contracts: contracts,
		Args:      args,
	}
}

func (fn Function) MapArgs(exprs []ast.Expr) map[string]string {
	if len(fn.Args) != len(exprs) {
		return nil
	}
	res := make(map[string]string)
	for i, arg := range fn.Args {
		expr := exprs[i]
		nLit, ok := expr.(*ast.BasicLit)
		if !ok {
			continue
		}
		res[arg] = nLit.Value
	}
	return res
}

// Validate chackes all contracts for a function using the given function arguments.
//
// If a contract is violated, that contract is returned.
//
// Possible return values:
//
//   - (contract, nil): a contract is violated, that contract is returned.
//   - (nil, SomeError): one or more contracts failed, the first failure is returned.
//
// If a contract failed and another one succeeded but violated,
// the error is nil and the violated contract is returned.
// In other words, don't worry that we cannot execute a contract
// if we have a meaningful error to show for another contract.
// That allows the analyzer to safely ignore contract errors.
func (fn Function) Validate(vars map[string]string) (*Contract, error) {
	// prepare interpreter
	interpreter := interp.New(interp.Options{})
	err := interpreter.Use(stdlib.Symbols)
	if err != nil {
		return nil, fmt.Errorf("use stdlib: %v", err)
	}
	interpreter.ImportUsed()
	for name, val := range vars {
		expr := fmt.Sprintf("%s := %s", name, val)
		_, err = interpreter.Eval(expr)
		if err != nil {
			return nil, fmt.Errorf("set value for %s: %v", name, err)
		}
	}

	// check all contracts
	var firstErr error = nil
	for _, c := range fn.Contracts {
		if !c.allDefined(vars) {
			continue
		}
		valid, err := c.validate(interpreter)
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("run `%s`: %v", c.Condition, err)
			}
			continue
		}
		if !valid {
			return &c, nil
		}
	}
	return nil, firstErr
}

// getFuncArgs returns argument names for the given function declaration.
func getFuncArgs(nFunc *ast.FuncDecl) []string {
	if nFunc.Type == nil {
		return nil
	}
	if nFunc.Type.Params == nil {
		return nil
	}
	res := make([]string, 0)
	for _, nField := range nFunc.Type.Params.List {
		for _, nIdent := range nField.Names {
			res = append(res, nIdent.Name)
		}
	}
	return res
}
