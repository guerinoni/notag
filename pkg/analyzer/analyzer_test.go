package analyzer

import (
	"reflect"
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
	{
		c := Setting{
			Pkg: pkgDenyMap{
				"globally": "db",
			},
		}

		td := analysistest.TestData()
		a := NewAnalyzerWithConfig(c)

		analysistest.Run(t, td, a, "bugs")
	}
}

func TestExtractTagsFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"`xml:\"name\" json:\"name\"`", []string{"json", "xml"}},
		{`json:"name"            xml:"Name"`, []string{"json", "xml"}},
		{`json:"name" xml:"Name"`, []string{"json", "xml"}},
		{`db:"name" xml:"Name"`, []string{"db", "xml"}},
		{`json:"name"`, []string{"json"}},
		{``, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := extractTagsFromString(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("splitTagsString(%q) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}
