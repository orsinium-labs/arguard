package main

import (
	"github.com/orsinium-labs/arguard/arguard"
	"github.com/orsinium-labs/arguard/contracts"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	cConfig := contracts.NewConfig()
	cAnalyzer := contracts.NewAnalyzer(cConfig)
	aConfig := arguard.NewConfig()
	aAnalyzer := arguard.NewAnalyzer(aConfig, cAnalyzer)
	multichecker.Main(aAnalyzer, cAnalyzer)
}
