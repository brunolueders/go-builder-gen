package main

import (
	"bytes"
	_ "embed"
	"go/ast"
	"go/printer"
	"go/token"
	"reflect"
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
	Name    string
	Type    string
	Options _fieldOptions
}

type _fieldOptions struct {
	Ignore bool
}

func parseFieldOptions(tag string) (_fieldOptions, error) {
	var options _fieldOptions
	for _, token := range strings.Split(tag, ",") {
		switch strings.TrimSpace(token) {
		case "":
			continue
		case "ignore":
			options.Ignore = true
		default:
			return _fieldOptions{}, errors.Errorf("invalid field option '%s'", strings.TrimSpace(token))
		}
	}
	return options, nil
}

func getTypeName(fieldType ast.Expr, fileSet *token.FileSet) (string, error) {
	var buffer bytes.Buffer
	if err := printer.Fprint(&buffer, fileSet, fieldType); err != nil {
		return "", errors.Wrap(err, "failed to format type expression")
	}
	return buffer.String(), nil
}

func getOptions(tag *ast.BasicLit) (_fieldOptions, error) {
	if tag == nil {
		return _fieldOptions{}, nil
	}
	return parseFieldOptions(reflect.StructTag(tag.Value).Get("builder"))
}

func extractFieldData(structType *ast.StructType, fileSet *token.FileSet) ([]_fieldData, error) {
	if structType == nil {
		return nil, nil
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

		options, err := getOptions(field.Tag)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse tags of field %s", name)
		}
		if options.Ignore {
			continue
		}

		typeName, err := getTypeName(field.Type, fileSet)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to determine type of field %s", name)
		}

		fieldData = append(fieldData, _fieldData{
			Name: name.Name,
			Type: typeName,
		})
	}
	return fieldData, nil
}

var reservedWords = map[string]struct{}{
	"any": {}, "append": {}, "bool": {}, "break": {}, "byte": {}, "cap": {}, "case": {}, "chan": {}, "close": {},
	"comparable": {}, "complex": {}, "complex128": {}, "complex64": {}, "const": {}, "continue": {}, "copy": {},
	"default": {}, "defer": {}, "delete": {}, "else": {}, "error": {}, "fallthrough": {}, "false": {}, "float32": {},
	"float64": {}, "for": {}, "func": {}, "go": {}, "goto": {}, "if": {}, "imag": {}, "import": {}, "int": {},
	"int16": {}, "int32": {}, "int64": {}, "int8": {}, "interface": {}, "iota": {}, "len": {}, "make": {}, "map": {},
	"new": {}, "nil": {}, "package": {}, "panic": {}, "print": {}, "println": {}, "range": {}, "real": {},
	"recover": {}, "return": {}, "rune": {}, "select": {}, "string": {}, "switch": {}, "true": {}, "type": {},
	"uint": {}, "uint16": {}, "uint32": {}, "uint64": {}, "uint8": {}, "uintptr": {}, "var": {},
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

func generate(target string, file *InputFile) ([]byte, error) {
	// Find target struct definition
	structTypeFinder := newStructTypeFinder(target)
	ast.Walk(structTypeFinder, file.File)
	if structTypeFinder.StructType == nil {
		return nil, errors.Errorf("could not find definition of struct %s", target)
	}

	fields, err := extractFieldData(structTypeFinder.StructType, file.FileSet)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to extract fields of struct %s", target)
	}

	// Parse builder template
	builderTemplate, err := template.New("builder").
		Funcs(template.FuncMap{"unexported": unexported}).
		Parse(builderTemplateString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse builder template")
	}

	// Generate code
	templateData := _builderTemplateData{
		Package: file.File.Name.String(),
		Target:  target,
		Fields:  fields,
	}

	generatedCode := bytes.NewBuffer(nil)
	if err = builderTemplate.Execute(generatedCode, &templateData); err != nil {
		return nil, errors.Wrap(err, "failed to generate builder code")
	}
	return generatedCode.Bytes(), nil
}
