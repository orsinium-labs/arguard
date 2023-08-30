package arguard_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/orsinium-labs/gosafe/arguard"
	"github.com/orsinium-labs/gosafe/contracts"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAll(t *testing.T) {
	t.Parallel()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	cConfig := contracts.NewConfig()
	cConfig.ReportContracts = true
	cAnalyzer := contracts.NewAnalyzer(cConfig)
	aConfig := arguard.NewConfig()
	aAnalyzer := arguard.NewAnalyzer(aConfig, cAnalyzer)

	testdata := filepath.Join(wd, "testdata")
	analysistest.Run(t, testdata, aAnalyzer, "p")
}

// Run the linter on random stdlib packages and see if it explodes.
func TestSmoke(t *testing.T) {
	t.Parallel()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	testdata := filepath.Join(wd, "testdata")

	packages := []string{
		"sync",
		"flag",
		"os",
		"fmt",
		"go/ast",
	}
	for _, pkgName := range packages {
		pkgName := pkgName
		t.Run(pkgName, func(t *testing.T) {
			t.Parallel()
			cConfig := contracts.NewConfig()
			cConfig.FollowImports = false
			cAnalyzer := contracts.NewAnalyzer(cConfig)
			aConfig := arguard.NewConfig()
			aConfig.ReportErrors = true
			aAnalyzer := arguard.NewAnalyzer(aConfig, cAnalyzer)
			analysistest.Run(t, testdata, aAnalyzer, pkgName)
		})
	}
}
