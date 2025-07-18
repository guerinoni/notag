package analyzer

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// NewAnalyzer creates a new instance of the analyzer with default configuration.
func NewAnalyzer() *analysis.Analyzer {
	var r runner

	a := &analysis.Analyzer{
		Name:     "notag",
		Doc:      "warns about specific tags used, or in a specific pkg",
		Run:      r.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}

	a.Flags.StringVar(&r.c.GlobalTagsDenied, "denied-tags", "", "comma-separated list of tags that are not allowed globally")

	return a
}

// newAnalyzerWithConfig creates a new analyzer with the provided configuration.
// This is useful for testing purposes, allowing you to pass a specific configuration.
func newAnalyzerWithConfig(c config) *analysis.Analyzer {
	var r runner
	r.c = c

	a := &analysis.Analyzer{
		Name:     "notag",
		Doc:      "warns about specific tags used, or in a specific pkg",
		Run:      r.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}

	return a
}

type config struct {
	GlobalTagsDenied string
}

type runner struct {
	c    config
	pass *analysis.Pass
}

func (r *runner) run(pass *analysis.Pass) (interface{}, error) {
	r.pass = pass

	inspector, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	inspector.Preorder(r.filters(), func(node ast.Node) {
		switch node := node.(type) {
		case *ast.StructType:
			fieldAffected, tagFailed, found := containsTags(splitTags(r.c.GlobalTagsDenied), node)
			if found {
				pass.Reportf(node.Pos(), "field '%s' contains denied tags: '%v'", fieldAffected, tagFailed)
			}
		default:
			fmt.Println("Found a node of type:", fmt.Sprintf("%T", node))
		}
	})

	return nil, nil
}

func (r *runner) filters() []ast.Node {
	return []ast.Node{&ast.StructType{}}
}

func splitTags(tags string) []string {
	if tags == "" {
		return nil
	}

	var result []string
	for _, tag := range strings.Split(tags, ",") {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			result = append(result, tag)
		}
	}

	return result
}

func containsTags(tags []string, n *ast.StructType) (string, string, bool) {
	if len(tags) == 0 {
		return "", "", false
	}

	for _, field := range n.Fields.List {
		tagsFailed := []string{}
		for _, denied := range tags {
			if strings.Contains(field.Tag.Value, denied) {
				tagsFailed = append(tagsFailed, denied)
			}
		}

		if len(tagsFailed) > 0 {
			return field.Names[0].Name, strings.Join(tagsFailed, ","), true
		}
	}

	return "", "", false
}
