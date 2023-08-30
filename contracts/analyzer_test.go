package contracts_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/orsinium-labs/gosafe/contracts"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAll(t *testing.T) {
	t.Parallel()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	testdata := filepath.Join(wd, "testdata")
	config := contracts.NewConfig()
	config.ReportContracts = true
	config.FollowImports = false
	analyzer := contracts.NewAnalyzer(config)
	analysistest.Run(t, testdata, &analyzer, "p")
}

func TestImports(t *testing.T) {
	t.Parallel()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	testdata := filepath.Join(wd, "testdata")
	config := contracts.NewConfig()
	analyzer := contracts.NewAnalyzer(config)
	analysistest.Run(t, testdata, &analyzer, "i")
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
			config := contracts.NewConfig()
			config.FollowImports = false
			analyzer := contracts.NewAnalyzer(config)
			analysistest.Run(t, testdata, &analyzer, pkgName)
		})
	}
}
