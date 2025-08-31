// Package analyzer provides a static analysis tool to warn about the usage of specific struct tags in Go code.
package analyzer

import (
	"fmt"
	"go/ast"
	"slices"
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

	r.setting.Pkg = make(pkgDenyMap)
	r.setting.PkgPath = make(pkgDenyMap)

	a.Flags.StringVar(&r.setting.GlobalTagsDenied, "denied", "", "comma-separated list of tags that are not allowed globally")
	a.Flags.Var(&r.setting.Pkg, "denied-pkg", "Per-package denied tags, format: pkg:tag1,tag2")
	a.Flags.Var(&r.setting.PkgPath, "denied-pkg-path", "Per-package path denied tags, format: pkg_path:tag1,tag2")

	return a
}

// NewAnalyzerWithConfig creates a new analyzer with the provided configuration.
// This is useful for testing purposes, allowing you to pass a specific configuration.
func NewAnalyzerWithConfig(c Setting) *analysis.Analyzer {
	var r runner

	r.setting = c

	a := &analysis.Analyzer{
		Name:     "notag",
		Doc:      "warns about specific tags used, or in a specific pkg",
		Run:      r.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}

	return a
}

type pkgDenyMap map[string]string

func (p *pkgDenyMap) String() string {
	if p == nil {
		return ""
	}

	result := make([]string, 0, len(*p))

	for pkg, tags := range *p {
		tags = strings.TrimSpace(tags)
		if tags == "" {
			continue
		}

		result = append(result, fmt.Sprintf("%s:%s", pkg, tags))
	}

	return strings.Join(result, ",")
}

func (p *pkgDenyMap) Set(value string) error {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid format for denied-pkg: %s, expected pkg:tag1,tag2", value)
	}

	pkg := strings.TrimSpace(parts[0])
	tags := parts[1]

	(*p)[pkg] = tags

	return nil
}

type Setting struct {
	GlobalTagsDenied string
	// Pkg is a map where the key is the package name and the value is a comma-separated list of denied tags.
	Pkg pkgDenyMap
	// PkgPath is the map using the full path of the package as the key.
	PkgPath pkgDenyMap
}

type runner struct {
	setting Setting
	pass    *analysis.Pass
}

func (r *runner) run(pass *analysis.Pass) (any, error) {
	r.pass = pass

	if r.setting.GlobalTagsDenied == "" && len(r.setting.Pkg) == 0 && len(r.setting.PkgPath) == 0 {
		return nil, nil
	}

	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	tagsToCheck := splitTags(r.setting.GlobalTagsDenied)

	pkgName := pass.Pkg.Name()

	if tags, found := r.setting.Pkg[pkgName]; found {
		tagsToCheck = append(tagsToCheck, splitTags(tags)...)
	}

	pkgPath := pass.Pkg.Path()

	if tags, found := r.setting.PkgPath[pkgPath]; found {
		tagsToCheck = append(tagsToCheck, splitTags(tags)...)
	}

	insp.Preorder(r.filters(), func(node ast.Node) {
		switch node := node.(type) {
		case *ast.StructType:
			fieldAffected, tagFailed, found := containsTags(tagsToCheck, node)
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
		if field.Tag == nil {
			continue
		}

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
