// Package analyzer provides a static analysis tool to warn about the usage of specific struct tags in Go code.
package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// NewAnalyzer creates a new instance of the analyzer with default configuration.
func NewAnalyzer() *analysis.Analyzer {
	r := &runner{
		setting: Setting{
			Pkg:     make(PkgDenyMap),
			PkgPath: make(PkgDenyMap),
		},
	}

	a := buildAnalyzer(r)

	a.Flags.StringVar(&r.setting.GlobalTagsDenied, "denied", "", "comma-separated list of tags that are not allowed globally")
	a.Flags.Var(&r.setting.Pkg, "denied-pkg", "Per-package denied tags, format: pkg:tag1,tag2")
	a.Flags.Var(&r.setting.PkgPath, "denied-pkg-path", "Per-package path denied tags, format: pkg_path:tag1,tag2")
	a.Flags.BoolVar(&r.setting.IncludeGenerated, "include-generated", false, "also analyze files marked as generated (default: skip)")

	return a
}

// NewAnalyzerWithConfig creates a new analyzer with the provided configuration.
// This is useful for testing purposes, allowing you to pass a specific configuration.
func NewAnalyzerWithConfig(c Setting) *analysis.Analyzer {
	return buildAnalyzer(&runner{setting: c})
}

func buildAnalyzer(r *runner) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "notag",
		Doc:      "warns about specific tags used, or in a specific pkg",
		Run:      r.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}

// PkgDenyMap maps a package identifier (name or import path) to its denied tag keys.
type PkgDenyMap map[string][]string

// String renders the map in the deterministic flag form `pkg:tag1,tag2;pkg2:tag3`.
func (p *PkgDenyMap) String() string {
	if p == nil {
		return ""
	}

	keys := make([]string, 0, len(*p))

	for pkg := range *p {
		keys = append(keys, pkg)
	}

	slices.Sort(keys)

	result := make([]string, 0, len(keys))

	for _, pkg := range keys {
		tags := (*p)[pkg]
		if len(tags) == 0 {
			continue
		}

		result = append(result, fmt.Sprintf("%s:%s", pkg, strings.Join(tags, ",")))
	}

	return strings.Join(result, ";")
}

// Set parses a `pkg:tag1,tag2` flag value. Repeated calls for the same key append.
func (p *PkgDenyMap) Set(value string) error {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format for denied-pkg: %s, expected pkg:tag1,tag2", value)
	}

	pkg := strings.TrimSpace(parts[0])

	tags := splitTags(parts[1])
	if len(tags) == 0 {
		return nil
	}

	(*p)[pkg] = append((*p)[pkg], tags...)

	return nil
}

type Setting struct {
	GlobalTagsDenied string
	// Pkg is keyed by package name (pass.Pkg.Name()).
	Pkg PkgDenyMap
	// PkgPath is keyed by full import path (pass.Pkg.Path()).
	PkgPath PkgDenyMap
	// IncludeGenerated, when true, analyzes files marked as generated.
	// Default (false) skips them.
	IncludeGenerated bool
}

type runner struct {
	setting Setting
}

func (r *runner) run(pass *analysis.Pass) (any, error) {
	tagsToCheck := r.tagsForPass(pass)
	if len(tagsToCheck) == 0 {
		return nil, nil
	}

	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector) //nolint:forcetypeassert // guaranteed by Requires

	var generated map[*token.File]bool
	if !r.setting.IncludeGenerated {
		generated = generatedFiles(pass)
	}

	insp.Preorder([]ast.Node{&ast.StructType{}}, func(node ast.Node) {
		if generated[pass.Fset.File(node.Pos())] {
			return
		}

		inspectStruct(pass, tagsToCheck, node)
	})

	return nil, nil
}

func generatedFiles(pass *analysis.Pass) map[*token.File]bool {
	out := map[*token.File]bool{}

	for _, f := range pass.Files {
		if !ast.IsGenerated(f) {
			continue
		}

		if tf := pass.Fset.File(f.Pos()); tf != nil {
			out[tf] = true
		}
	}

	return out
}

func (r *runner) tagsForPass(pass *analysis.Pass) []string {
	if r.setting.GlobalTagsDenied == "" && len(r.setting.Pkg) == 0 && len(r.setting.PkgPath) == 0 {
		return nil
	}

	tags := splitTags(r.setting.GlobalTagsDenied)

	if extra, found := r.setting.Pkg[pass.Pkg.Name()]; found {
		tags = append(tags, extra...)
	}

	if extra, found := r.setting.PkgPath[pass.Pkg.Path()]; found {
		tags = append(tags, extra...)
	}

	return dedup(tags)
}

func inspectStruct(pass *analysis.Pass, tagsToCheck []string, node ast.Node) {
	st, ok := node.(*ast.StructType)
	if !ok || st.Fields == nil {
		return
	}

	for _, field := range st.Fields.List {
		if field.Tag == nil {
			continue
		}

		failed := matchedTags(tagsToCheck, field.Tag.Value)
		if len(failed) == 0 {
			continue
		}

		for _, name := range fieldNames(field) {
			pass.Reportf(field.Pos(), "field '%s' contains denied tags: '%s'", name, strings.Join(failed, ","))
		}
	}
}

func fieldNames(field *ast.Field) []string {
	if len(field.Names) > 0 {
		out := make([]string, len(field.Names))
		for i, n := range field.Names {
			out[i] = n.Name
		}

		return out
	}

	if id, ok := field.Type.(*ast.Ident); ok {
		return []string{id.Name}
	}

	return []string{"<embedded>"}
}

func matchedTags(denied []string, tagStr string) []string {
	tags := extractTagsFromString(tagStr)
	if len(tags) == 0 {
		return nil
	}

	var out []string

	for _, d := range denied {
		if slices.Contains(tags, d) {
			out = append(out, d)
		}
	}

	return out
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

func dedup(in []string) []string {
	if len(in) <= 1 {
		return in
	}

	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))

	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}

		seen[v] = struct{}{}

		out = append(out, v)
	}

	return out
}

// extractTagsFromString returns the tag keys present in a struct tag string.
// It mirrors the parsing rules of reflect.StructTag, handling tab/space separators
// and quoted values that may contain whitespace or escaped quotes.
// Example: `json:"name"` -> [json].
// Example: `json:"name,omitempty" xml:"Name"` -> [json, xml].
func extractTagsFromString(s string) []string {
	if unq, err := strconv.Unquote(s); err == nil {
		s = unq
	} else {
		s = strings.Trim(s, "`")
	}

	if s == "" {
		return []string{}
	}

	var keys []string

	for s != "" {
		s = trimTagSpace(s)
		if s == "" {
			break
		}

		key, rest, ok := nextTagKey(s)
		if !ok {
			break
		}

		rest, ok = skipTagValue(rest)
		if !ok {
			break
		}

		keys = append(keys, key)
		s = rest
	}

	slices.Sort(keys)

	if keys == nil {
		return []string{}
	}

	return keys
}

func trimTagSpace(s string) string {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}

	return s[i:]
}

func nextTagKey(s string) (string, string, bool) {
	i := 0
	for i < len(s) && s[i] > ' ' && s[i] != ':' && s[i] != '"' && s[i] != 0x7f {
		i++
	}

	if i == 0 || i >= len(s) || s[i] != ':' {
		return "", s, false
	}

	return s[:i], s[i+1:], true
}

func skipTagValue(s string) (string, bool) {
	if s == "" || s[0] != '"' {
		return s, false
	}

	i := 1
	for i < len(s) && s[i] != '"' {
		if s[i] == '\\' {
			i++
		}

		i++
	}

	if i >= len(s) {
		return s, false
	}

	return s[i+1:], true
}
