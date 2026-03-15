# Agent Guide for go-builder-gen

## Project Overview

`go-builder-gen` is a CLI tool that generates Go builder pattern code from struct definitions. It parses Go source files using the AST and generates builder types with fluent setter methods.

## Essential Commands

```bash
# Build the project
go build -v ./...

# Run tests
go test -v ./...

# Install as CLI tool
go install github.com/brunolueders/go-builder-gen@latest
```

## Project Structure

```
/home/bruno/projects/go-builder-gen/
├── main.go           # Entry point, CLI argument handling, file loading
├── generate.go       # Core generation logic, AST traversal, template execution
├── generate_test.go # Unit tests for generate.go
├── main_test.go      # Unit tests for main.go
├── helpers_test.go   # Test utilities
├── template.gotext   # Embedded Go template for builder code
├── go.mod            # Go 1.26
└── README.md         # Usage documentation
```

## Code Patterns

### Naming Conventions
- **Private identifiers**: Unexported (lowercase first letter)
- **Internal types**: Prefixed with underscore (e.g., `_structTypeFinder`, `_fieldData`, `_fieldOptions`)
- **Test helper types**: Prefixed with underscore (e.g., `_testDescription`, `_structField`)
- **Template data struct**: Prefixed with underscore (e.g., `_builderTemplateData`)

### Code Organization
- **main.go**: File loading, CLI argument parsing, output file generation
- **generate.go**: AST walking, field extraction, template execution
- **Tests**: Co-located with implementation (`*_test.go`)

### Key Implementation Details

1. **AST Walking**: Uses `ast.Walk()` with custom visitor (`_structTypeFinder`) to locate struct definitions

2. **Type Extraction**: Uses `go/printer.Fprint()` to format type expressions (replaced handrolled formatting - see commit 627cfbc)

3. **Template**: Embedded via `//go:embed` directive in `generate.go:148-149`

4. **Field Options**: Parsed from `builder` struct tag (currently only supports `ignore`)

5. **Reserved Words**: Map in `generate.go:119-128` for handling Go keywords/predefined types in unexported names

### Testing Patterns
- Table-driven tests with descriptive `t.Run()` names
- Test descriptions use `//GIVEN //WHEN //THEN` style comments
- Helper functions in `helpers_test.go` for creating AST nodes

## Gotchas

1. **Input Format**: CLI expects `FILE:STRUCT_NAME` format (e.g., `go-builder-gen /path/user.go:User`)

2. **Unexported Fields**: Struct fields must be exported (capitalized) to be included in the builder

3. **Reserved Word Handling**: If an unexported name conflicts with a Go keyword or built-in type, it's prefixed with underscore (e.g., `Type` → `_type`)

4. **Acronym Handling**: `CamelToSnakeCase()` preserves acronyms (e.g., `URL` stays as `url`, not `u_r_l`)

5. **Template Type Annotation**: The `template.gotext` file uses a Go type annotation comment (`gotype`) for IDE/tooling support - this is not runtime code

6. **File Output**: Generated files are named `{snake_case_struct_name}_builder.go` in the same directory as the source file

## Dependencies

- `github.com/stretchr/testify/assert` - Test assertions
