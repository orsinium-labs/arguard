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
	cAnalyzer := contracts.NewAnalyzer(cConfig)
	aConfig := arguard.NewConfig()
	aAnalyzer := arguard.NewAnalyzer(aConfig, &cAnalyzer)

	testdata := filepath.Join(wd, "testdata")
	analysistest.Run(t, testdata, aAnalyzer, "p")
}
