package analyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestConfigEmpty(t *testing.T) {
	c := Setting{}
	td := analysistest.TestData()
	a := NewAnalyzerWithConfig(c)

	analysistest.Run(t, td, a, "noconfig")
}

func TestConfigGloballyDenied(t *testing.T) {
	c := Setting{GlobalTagsDenied: "json,xml"}
	td := analysistest.TestData()
	a := NewAnalyzerWithConfig(c)

	analysistest.Run(t, td, a, "globally")
}

func TestConfigSpecificPkg(t *testing.T) {
	{
		c := Setting{
			Pkg: pkgDenyMap{
				"globally": "json,xml",
			},
		}

		td := analysistest.TestData()
		a := NewAnalyzerWithConfig(c)

		analysistest.Run(t, td, a, "globally")
	}
	{
		c := Setting{
			Pkg: pkgDenyMap{
				"globally": "json,xml",
			},
		}

		td := analysistest.TestData()
		a := NewAnalyzerWithConfig(c)

		analysistest.Run(t, td, a, "noconfig")
	}
	{
		c := Setting{
			Pkg: pkgDenyMap{
				"globally": "json,xml",
			},
		}

		td := analysistest.TestData()
		a := NewAnalyzerWithConfig(c)

		analysistest.Run(t, td, a, "tags")
	}
}
