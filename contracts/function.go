package contracts

import (
	"fmt"
	"go/ast"
)

type Function struct {
	Args      []string
	Contracts []Contract
}

func (*Function) AFact() {}

func functionFromAST(nFunc *ast.FuncDecl) *Function {
	if nFunc.Body == nil {
		return nil
	}
	contracts := make([]Contract, 0)
	for _, stmt := range nFunc.Body.List {
		contract := contractFromAST(stmt)
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
//   - (nil, SomeError): one or more contracts failed, the first failure is returned.
//   - (contract, nil): a contract is violated, that contract is returned.
func (fn Function) Validate(vars map[string]string) (*Contract, error) {
	var firstErr error = nil
	for _, c := range fn.Contracts {
		valid, err := c.validate(vars)
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
		argName := fieldToName(nField)
		if argName != "" {
			res = append(res, argName)
		}
	}
	return res
}

func fieldToName(nField *ast.Field) string {
	if nField.Names == nil {
		return ""
	}
	if len(nField.Names) != 1 {
		return ""
	}
	nIdent := nField.Names[0]
	return nIdent.Name
}
