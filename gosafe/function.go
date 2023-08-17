package gosafe

import (
	"fmt"
	"go/ast"
)

type Function struct {
	Recv      string     `json:"recv"`
	Name      string     `json:"name"`
	Args      []string   `json:"args"`
	Contracts []Contract `json:"contracts"`
}

func FunctionFromAST(node ast.Decl) *Function {
	nFunc, ok := node.(*ast.FuncDecl)
	if !ok {
		return nil
	}
	if nFunc.Body == nil {
		return nil
	}

	contracts := make([]Contract, 0)
	for _, stmt := range nFunc.Body.List {
		contract := ContractFromAST(stmt)
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
		Recv:      funcRecvName(nFunc),
		Name:      nFunc.Name.String(),
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

func (fn Function) Validate(vars map[string]string) (bool, error) {
	for i, c := range fn.Contracts {
		valid, err := c.Validate(vars)
		if err != nil {
			return false, fmt.Errorf("contract #%d: %v", i+1, err)
		}
		if !valid {
			return false, nil
		}
	}
	return true, nil
}

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
