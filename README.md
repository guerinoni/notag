# notag

A tiny Go linter that flags struct tags you don't want used globally, in a specific package by name, or in a specific package by import path.

`notag` is built on top of `golang.org/x/tools/go/analysis`, so it plugs into anything that speaks the standard analyzer protocol (e.g. `singlechecker`, custom drivers).

## Why use notag?

Different layers of an application have different serialization needs. JSON tags belong in the layer that talks to clients; database tags belong in the persistence layer; the domain layer in between should generally not carry either. `notag` lets you encode that boundary as a lint rule instead of relying on review discipline.

```
┌─────────────────────────────────────────┐
│            API/Controller Layer         │
│  ┌─────────────────────────────────────┐│
│  │ type UserRequest struct {           ││  ✅ JSON tags OK
│  │     Name  string `json:"name"`      ││
│  │     Email string `json:"email"`     ││
│  │ }                                   ││
│  └─────────────────────────────────────┘│
└─────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────┐
│         Business Logic Layer            │
│  ┌─────────────────────────────────────┐│
│  │ type User struct {                  ││  ❌ notag catches this!
│  │     Name  string `json:"name"`      ││
│  │     Email string                    ││
│  │ }                                   ││
│  └─────────────────────────────────────┘│
└─────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────┐
│              Data Layer                 │
│  ┌─────────────────────────────────────┐│
│  │ type UserEntity struct {            ││  ✅ DB tags OK
│  │     ID    int    `db:"id"`          ││
│  │     Name  string `db:"name"`        ││
│  │     Email string `db:"email"`       ││
│  │ }                                   ││
│  └─────────────────────────────────────┘│
└─────────────────────────────────────────┘
```

## Installation

```zsh
go install github.com/guerinoni/notag@latest
```

## How it works

For each Go package the analyzer visits, `notag`:

1. Computes the set of denied tag keys by combining `-denied` (global) with any `-denied-pkg` entry whose key matches the package **name** and any `-denied-pkg-path` entry whose key matches the package **import path**.
2. Walks every `struct` literal, including nested anonymous structs and embedded fields.
3. Extracts the tag keys from each field's struct tag using a `reflect.StructTag`-style parser
4. Reports each field whose tag declares any denied key.

The diagnostic is positioned on the offending field, not on the enclosing struct, so editors and CI annotations point at the exact line:

```
internal/domain/user.go:7:2: field 'Name' contains denied tags: 'json'
```

## Usage

### Global denial

```zsh
# Deny `validate` and `xml` everywhere
notag -denied validate,xml ./...
```

### By package name

The key matches `pass.Pkg.Name()` — the bare package name as it appears in the `package` clause.

```zsh
# In the "service" package, deny `json`
notag -denied-pkg service:json ./...

# Multiple flag uses are appended (different keys are independent)
notag -denied-pkg service:json -denied-pkg repository:xml ./...

# Repeating the same key also appends
notag -denied-pkg service:json -denied-pkg service:xml ./...
# equivalent to:
notag -denied-pkg service:json,xml ./...
```

### By package import path

The key matches `pass.Pkg.Path()` — the full import path. Use this when several packages in the project share a name.

```zsh
notag -denied-pkg-path github.com/guerinoni/notag/internal/domain:json,xml ./...
```

### Combining sources

All three sources are merged for each package the analyzer visits, deduplicated, then matched.

```zsh
notag \
  -denied db \
  -denied-pkg service:json \
  -denied-pkg-path github.com/org/be/internal/controllers:xml \
  ./...
```

## Example

Given:

```go
package domain

type User struct {
    Name  string `json:"name"`
    Email string `xml:"email"`
}
```

Running:

```zsh
notag -denied json,xml ./...
```

produces:

```
domain/user.go:4:2: field 'Name' contains denied tags: 'json'
domain/user.go:5:2: field 'Email' contains denied tags: 'xml'
```

Multiple denied keys on the same field are reported once per field, joined by comma:

```
field 'Name' contains denied tags: 'json,xml'
```

Fields declared with multiple names share their tag in Go, so each identifier is reported separately:

```go
type T struct {
    A, B string `json:"shared"`
}
```

```
t.go:2:2: field 'A' contains denied tags: 'json'
t.go:2:2: field 'B' contains denied tags: 'json'
```

## Programmatic use

If you embed `notag` in a custom analyzer driver, build the analyzer with explicit configuration instead of CLI flags:

```go
import (
    "github.com/guerinoni/notag/pkg/analyzer"
    "golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
    a := analyzer.NewAnalyzerWithConfig(analyzer.Setting{
        GlobalTagsDenied: "validate",
        Pkg: analyzer.PkgDenyMap{
            "service": []string{"json"},
        },
        PkgPath: analyzer.PkgDenyMap{
            "github.com/org/be/internal/domain": []string{"json", "xml"},
        },
    })
    singlechecker.Main(a)
}
```

## Features

- [x] Global tag restrictions
- [x] Package-specific tag restrictions (by name)
- [x] Package path-based restrictions
- [x] Combine global + per-package directives
- [x] Multiple tag support, deduped across sources
- [x] Embedded fields and nested anonymous structs
- [x] Multi-name fields (`A, B string \`json:"x"\``) reported per identifier
- [x] Robust struct-tag parsing (tab separators, escaped quotes)
