package main

import (
	"github.com/orsinium-labs/gosafe/gosafe"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(gosafe.NewAnalyzer())
}
