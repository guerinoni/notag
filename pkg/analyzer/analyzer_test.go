package analyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestConfigEmpty(t *testing.T) {
	c := config{}
	td := analysistest.TestData()
	a := newAnalyzerWithConfig(c)

	analysistest.Run(t, td, a, "noconfig")
}

func TestConfigGloballyDenied(t *testing.T) {
	c := config{GlobalTagsDenied: "json,xml"}
	td := analysistest.TestData()
	a := newAnalyzerWithConfig(c)

	analysistest.Run(t, td, a, "globally")
}
