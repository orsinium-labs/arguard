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
	if nFunc.Body == nil { // should be unreachable, the caller also checks that
		return nil
	}

	args := getFuncArgs(nFunc)
	if len(args) == 0 { // functions without arguments can't have pre-conditions
		return nil
	}

	contracts := make([]Contract, 0)
	for _, stmt := range nFunc.Body.List {
		contract, err := contractFromAST(stmt, info)
		if err != nil {
			// We assume that contracts go before any other code in the function.
			// If we don't do that, the function might modify the argument value
			// or break early before the contract. So, the contract that we check
			// might be actually unreachable or different in the runtime.
			break
		}
		contracts = append(contracts, *contract)
	}
	if len(contracts) == 0 { // we're not interested in functions without contracts
		return nil
	}
	return &Function{args, contracts}
}

// MapArgs converts list of expressions to strings and maps them to function argument names.
func (fn Function) MapArgs(exprs []ast.Expr, info *types.Info) map[string]string {
	if len(fn.Args) != len(exprs) {
		return nil
	}
	res := make(map[string]string)
	for i, arg := range fn.Args {
		expr := exprs[i]
		strExpr, names, err := expr2string(expr, info)
		if len(names) != 0 { // argument definition must not have any unbound variables
			continue
		}
		if err != nil {
			continue
		}
		res[arg] = strExpr
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
