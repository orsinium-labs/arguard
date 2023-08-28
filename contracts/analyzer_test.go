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
