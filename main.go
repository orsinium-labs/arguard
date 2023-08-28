package main

import (
	"github.com/orsinium-labs/gosafe/arguard"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(arguard.NewAnalyzer())
}
