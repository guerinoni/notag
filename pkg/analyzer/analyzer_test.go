package analyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestConfigEmpty(t *testing.T) {
	c := config{}
	td := analysistest.TestData()
	a := NewAnalyzerWithConfig(c)

	analysistest.Run(t, td, a, "noconfig")
}

func TestConfigGloballyDenied(t *testing.T) {
	c := config{GlobalTagsDenied: "json,xml"}
	td := analysistest.TestData()
	a := NewAnalyzerWithConfig(c)

	analysistest.Run(t, td, a, "globally")
}

func TestConfigSpecificPkg(t *testing.T) {
	{
		c := config{
			Pkg: pkgDenyMap{
				"globally": "json,xml",
			},
		}

		td := analysistest.TestData()
		a := NewAnalyzerWithConfig(c)

		analysistest.Run(t, td, a, "globally")
	}
	{
		c := config{
			Pkg: pkgDenyMap{
				"globally": "json,xml",
			},
		}

		td := analysistest.TestData()
		a := NewAnalyzerWithConfig(c)

		analysistest.Run(t, td, a, "noconfig")
	}
	{
		c := config{
			Pkg: pkgDenyMap{
				"globally": "json,xml",
			},
		}

		td := analysistest.TestData()
		a := NewAnalyzerWithConfig(c)

		analysistest.Run(t, td, a, "tags")
	}
}
