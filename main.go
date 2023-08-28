package main

import (
	"github.com/orsinium-labs/gosafe/arguard"
	"github.com/orsinium-labs/gosafe/contracts"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	cConfig := contracts.NewConfig()
	cAnalyzer := contracts.NewAnalyzer(cConfig)
	aConfig := arguard.NewConfig()
	aAnalyzer := arguard.NewAnalyzer(aConfig, &cAnalyzer)
	singlechecker.Main(aAnalyzer)
}
