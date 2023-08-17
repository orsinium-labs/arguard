package gosafe

import "go/ast"

type Function struct {
	Recv      string
	Name      string
	LineNo    int
	Contracts []Contract
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
	return &Function{
		Recv:      funcRecvName(nFunc),
		Name:      nFunc.Name.String(),
		Contracts: contracts,
	}
}
