# notag

A tiny Go linter that ensures specific struct tags are not used globally or in designated packages.

## Why Use notag?

Your project, your rules! 

But seriously, it's incredibly useful for maintaining clean architecture. For example, when you have:
- **Controller/API layer** → JSON tags belong here for request/response structs
- **Business logic layer** → No JSON tags needed, keeps internal structs clean
- **Data layer** → Different serialization needs

This prevents confusion and enforces architectural boundaries, saving debugging time and maintaining code clarity.

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
│         Business Logic Layer           │
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

```bash
go install github.com/guerinoni/notag@latest
```

## Usage

### Package-specific Tag Restrictions
```bash
# Warn if JSON tags are found in the "service" package
notag -denied-pkg service:json ./...

# Multiple packages with different restrictions
notag -denied-pkg service:json -denied-pkg repository:xml ./...
```

### Package Path-based Restrictions
```bash
# Deny XML tags in the "repository" package
notag -denied-pkg-path github.com/guerinoni/notag/analyzer:xml ./...
```

### Global Tag Restrictions
```bash
# Deny validate and xml tags globally across all packages
notag -denied validate,xml ./...

# Combine global and package-specific rules
notag -denied validate -denied-pkg service:json ./...
```

### All combination of restrictions
```bash
# This denys:
# globally db tags,
# in the "service" package: json tags,
# in the "github.com/org/be/internal/controllers" package: xml tags
notag --denied db --denied-pkg service:json --denied-pkg-path github.com/org/be/internal/controllers:xml ./...
```

## Benefits

Instead of littering your internal structs with `json:"-"` tags, `notag` encourages you to:
- Keep API request/response structs in dedicated packages
- Maintain clean internal domain models
- Change serialization formats without touching business logic
- Have explicit control over what gets exposed in your API

## Features

- [x] Global tag restrictions
- [x] Package-specific tag restrictions
- [x] Combine global + pkg specific directive
- [x] Multiple tag support
- [x] Package path-based restrictions
- [ ] Find in nested structs (coming soon)

