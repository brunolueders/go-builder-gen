package main

import (
	"bytes"
	_ "embed"
	"go/ast"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

type _structTypeFinder struct {
	name       string
	StructType *ast.StructType
}

func newStructTypeFinder(name string) *_structTypeFinder {
	return &_structTypeFinder{
		name: name,
	}
}

func (finder *_structTypeFinder) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return finder
	}

	if typeSpec, ok := node.(*ast.TypeSpec); !ok || typeSpec.Name.String() != finder.name {
		return finder
	} else if structType, ok := typeSpec.Type.(*ast.StructType); !ok {
		return finder
	} else {
		finder.StructType = structType
		return nil
	}
}

type _fieldData struct {
	Name string
	Type string
}

func getTypeName(fieldType ast.Expr) string {
	switch fieldType.(type) {
	case *ast.Ident:
		return fieldType.(*ast.Ident).Name
	case *ast.ArrayType:
		return "[]" + getTypeName(fieldType.(*ast.ArrayType).Elt)
	case *ast.StarExpr:
		return "*" + getTypeName(fieldType.(*ast.StarExpr).X)
	case *ast.MapType:
		mapType := fieldType.(*ast.MapType)
		return "map[" + getTypeName(mapType.Key) + "]" + getTypeName(mapType.Value)
	case *ast.ChanType:
		chanType := fieldType.(*ast.ChanType)
		var keyword string
		if chanType.Dir == ast.RECV {
			keyword = "<-chan"
		} else {
			keyword = "chan<-"
		}
		return keyword + " " + getTypeName(chanType.Value)
	}
	return ""
}

func extractFieldData(structType *ast.StructType) []_fieldData {
	if structType == nil {
		return nil
	}

	var fieldData []_fieldData
	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 || field.Names[0] == nil {
			continue
		}

		name := field.Names[0]
		if !name.IsExported() {
			continue
		}

		typeName := getTypeName(field.Type)
		if typeName == "" {
			continue
		}

		fieldData = append(fieldData, _fieldData{
			Name: name.Name,
			Type: typeName,
		})
	}
	return fieldData
}

var reservedWords = map[string]struct{}{
	"append": {}, "bool": {}, "break": {}, "byte": {}, "cap": {}, "case": {}, "chan": {}, "close": {}, "complex": {},
	"complex128": {}, "complex64": {}, "const": {}, "continue": {}, "copy": {}, "default": {}, "defer": {}, "delete": {},
	"else": {}, "error": {}, "fallthrough": {}, "false": {}, "float32": {}, "float64": {}, "for": {}, "func": {}, "go": {},
	"goto": {}, "if": {}, "imag": {}, "import": {}, "int": {}, "int16": {}, "int32": {}, "int64": {}, "int8": {}, "interface": {},
	"iota": {}, "len": {}, "make": {}, "map": {}, "new": {}, "nil": {}, "package": {}, "panic": {}, "print": {}, "println": {},
	"range": {}, "real": {}, "recover": {}, "return": {}, "rune": {}, "select": {}, "string": {}, "switch": {}, "true": {},
	"type": {}, "uint": {}, "uint16": {}, "uint32": {}, "uint64": {}, "uint8": {}, "uintptr": {}, "var": {},
}

func unexported(identifier string) string {
	if identifier == "" {
		return identifier
	}

	var result string
	if identifier == strings.ToUpper(identifier) {
		result = strings.ToLower(identifier)
	} else {
		result = strings.ToLower(identifier[:1]) + identifier[1:]
	}

	if _, reserved := reservedWords[result]; reserved || result == identifier {
		result = "_" + result
	}
	return result
}

//go:embed template.gotext
var builderTemplateString string

type _builderTemplateData struct {
	Package string
	Target  string
	Fields  []_fieldData
}

func generate(target string, file *ast.File) ([]byte, error) {
	// Find target struct definition
	structTypeFinder := newStructTypeFinder(target)
	ast.Walk(structTypeFinder, file)
	if structTypeFinder.StructType == nil {
		return nil, errors.Errorf("could not find definition of struct %s", target)
	}

	fields := extractFieldData(structTypeFinder.StructType)

	// Parse builder template
	builderTemplate, err := template.New("builder").
		Funcs(template.FuncMap{"unexported": unexported}).
		Parse(builderTemplateString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse builder template")
	}

	// Generate code
	templateData := _builderTemplateData{
		Package: file.Name.String(),
		Target:  target,
		Fields:  fields,
	}

	generatedCode := bytes.NewBuffer(nil)
	if err = builderTemplate.Execute(generatedCode, &templateData); err != nil {
		return nil, errors.Wrap(err, "failed to generate builder code")
	}
	return generatedCode.Bytes(), nil
}
