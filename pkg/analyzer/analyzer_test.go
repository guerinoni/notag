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
			Pkg: PkgDenyMap{
				"globally": []string{"json", "xml"},
			},
		}

		td := analysistest.TestData()
		a := NewAnalyzerWithConfig(c)

		analysistest.Run(t, td, a, "globally")
	}
	{
		c := Setting{
			Pkg: PkgDenyMap{
				"globally": []string{"json", "xml"},
			},
		}

		td := analysistest.TestData()
		a := NewAnalyzerWithConfig(c)

		analysistest.Run(t, td, a, "noconfig")
	}
	{
		c := Setting{
			Pkg: PkgDenyMap{
				"globally": []string{"json", "xml"},
			},
		}

		td := analysistest.TestData()
		a := NewAnalyzerWithConfig(c)

		analysistest.Run(t, td, a, "tags")
	}
	{
		c := Setting{
			Pkg: PkgDenyMap{
				"globally": []string{"db"},
			},
		}

		td := analysistest.TestData()
		a := NewAnalyzerWithConfig(c)

		analysistest.Run(t, td, a, "bugs")
	}
}

func TestPkgDenyMapSet(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		m := PkgDenyMap{}
		if err := m.Set("svc:json,xml"); err != nil {
			t.Fatalf("unexpected err: %v", err)
		}

		want := []string{"json", "xml"}
		if !reflect.DeepEqual(m["svc"], want) {
			t.Errorf("svc = %v; want %v", m["svc"], want)
		}
	})

	t.Run("repeat appends", func(t *testing.T) {
		m := PkgDenyMap{}
		_ = m.Set("svc:json")
		_ = m.Set("svc:xml")

		want := []string{"json", "xml"}
		if !reflect.DeepEqual(m["svc"], want) {
			t.Errorf("svc = %v; want %v", m["svc"], want)
		}
	})

	t.Run("path with colon preserved", func(t *testing.T) {
		m := PkgDenyMap{}
		if err := m.Set("github.com/x/y:json"); err != nil {
			t.Fatalf("unexpected err: %v", err)
		}

		want := []string{"json"}
		if !reflect.DeepEqual(m["github.com/x/y"], want) {
			t.Errorf("got %v; want %v", m["github.com/x/y"], want)
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		m := PkgDenyMap{}
		if err := m.Set("bogus"); err == nil {
			t.Error("expected err for missing colon")
		}
	})
}

func TestPkgDenyMapStringDeterministic(t *testing.T) {
	m := PkgDenyMap{
		"b": []string{"json"},
		"a": []string{"xml"},
		"c": []string{"db"},
	}

	got := m.String()
	want := "a:xml;b:json;c:db"

	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestDedup(t *testing.T) {
	got := dedup([]string{"a", "b", "a", "c", "b"})
	want := []string{"a", "b", "c"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v; want %v", got, want)
	}
}

func TestConfigPkgPath(t *testing.T) {
	c := Setting{
		PkgPath: PkgDenyMap{
			"globally": []string{"json", "xml"},
		},
	}

	td := analysistest.TestData()
	a := NewAnalyzerWithConfig(c)

	analysistest.Run(t, td, a, "globally")
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
		{"json:\"name\"\txml:\"Name\"", []string{"json", "xml"}},
		{`json:"a\"b" xml:"name"`, []string{"json", "xml"}},
		{`"json:\"name\" xml:\"Name\""`, []string{"json", "xml"}},
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
