package main

import (
	"github.com/orsinium-labs/gosafe/arguard"
	"github.com/orsinium-labs/gosafe/contracts"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	contractsConfig := contracts.NewConfig()
	contractsAnalyzer := contracts.NewAnalyzer(contractsConfig)
	arguardAnalyzer := arguard.NewAnalyzer(&contractsAnalyzer)
	singlechecker.Main(arguardAnalyzer)
}
