package main

import (
	"fmt"
	"go/ast"
	"reflect"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func Test_structTypeFinder_Visit_ShouldFindTargetStructType(t *testing.T) {
	// GIVEN a struct declaration node
	const structName = "TestStruct"
	var structType = structTypeWithFields([]_structField{
		{Name: "Foo", Type: ast.NewIdent("int")},
		{Name: "Bar", Type: ast.NewIdent("string")},
	})
	node := &ast.TypeSpec{
		Name: ast.NewIdent(structName),
		Type: structType,
	}

	// AND a finder that is looking for the struct
	finder := newStructTypeFinder(structName)

	// WHEN visiting the declaration node
	result := finder.Visit(node)

	// THEN stop searching
	assert.Nil(t, result)
	assert.Equal(t, structType, finder.StructType)
}

func Test_structTypeFinder_Visit_ShouldSkipStructTypeWithWrongName(t *testing.T) {
	// GIVEN a struct declaration node
	var structType = structTypeWithFields([]_structField{
		{Name: "Foo", Type: ast.NewIdent("int")},
		{Name: "Bar", Type: ast.NewIdent("string")},
	})
	node := &ast.TypeSpec{
		Name: ast.NewIdent("NotTargetStruct"),
		Type: structType,
	}

	// AND a finder that is looking for the struct
	finder := newStructTypeFinder("TargetStruct")

	// WHEN visiting the declaration node
	result := finder.Visit(node)

	// THEN continue searching
	assert.Equal(t, finder, result)
}

func Test_structTypeFinder_Visit_ShouldSkipIrrelevantNodes(t *testing.T) {
	nodes := []ast.Node{
		nil, &ast.GoStmt{}, &ast.Field{}, &ast.CompositeLit{}, &ast.BadDecl{}, &ast.KeyValueExpr{}, &ast.Ident{}, &ast.IndexExpr{},
		&ast.IncDecStmt{}, &ast.SwitchStmt{}, &ast.SelectStmt{}, &ast.CaseClause{}, &ast.ReturnStmt{}, &ast.ArrayType{},
		&ast.InterfaceType{}, &ast.Ellipsis{}, &ast.UnaryExpr{}, &ast.Package{}, &ast.SendStmt{}, &ast.EmptyStmt{}, &ast.BlockStmt{},
		&ast.MapType{}, &ast.BranchStmt{}, &ast.Comment{}, &ast.BasicLit{}, &ast.AssignStmt{}, &ast.FuncDecl{}, &ast.SliceExpr{},
		&ast.DeclStmt{}, &ast.File{}, &ast.BadExpr{}, &ast.ValueSpec{}, &ast.CommClause{}, &ast.SelectorExpr{}, &ast.TypeSwitchStmt{},
		&ast.GenDecl{}, &ast.FuncLit{}, &ast.StarExpr{}, &ast.TypeAssertExpr{}, &ast.BinaryExpr{}, &ast.CommentGroup{}, &ast.RangeStmt{},
		&ast.IfStmt{}, &ast.CallExpr{}, &ast.BadStmt{}, &ast.FieldList{}, &ast.ForStmt{}, &ast.LabeledStmt{}, &ast.DeferStmt{},
		&ast.ExprStmt{}, &ast.ChanType{}, &ast.ImportSpec{}, &ast.ParenExpr{}, &ast.FuncType{}, &ast.StructType{},
	}

	for _, node := range nodes {
		t.Run(fmt.Sprintf("should skip %s nodes", reflect.ValueOf(node).String()), func(t *testing.T) {
			finder := newStructTypeFinder("TestStruct")

			assert.Equal(t, finder, finder.Visit(node))
		})
	}
}

func Test_parseFieldOptions(t *testing.T) {
	type _testDescription struct {
		Tag         string
		Expected    _fieldOptions
		Description string
	}

	tests := []_testDescription{
		{
			Tag:         "",
			Expected:    _fieldOptions{},
			Description: "Empty options should be valid",
		},
		{
			Tag:         "  \t ",
			Expected:    _fieldOptions{},
			Description: "Should ignore whitespace",
		},
		{
			Tag:         "ignore",
			Expected:    _fieldOptions{Ignore: true},
			Description: "Should recognise 'ignore' option",
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.Description, func(t *testing.T) {
			options, err := parseFieldOptions(test.Tag)

			assert.Nil(t, err)
			assert.Equal(t, test.Expected, options)
		})
	}
}

func Test_parseFieldOptions_UnknownOption(t *testing.T) {
	_, err := parseFieldOptions("ignore,foobar")

	assert.EqualError(t, err, "invalid field option 'foobar'")
}

func Test_extractFieldData(t *testing.T) {
	type _testDescription struct {
		StructType  *ast.StructType
		Expected    []_fieldData
		Description string
	}

	tests := []_testDescription{
		{
			StructType:  nil,
			Expected:    nil,
			Description: "Should not fail if struct type is nil",
		},
		{
			StructType: structTypeWithFields([]_structField{
				{Name: "foo", Type: ast.NewIdent("int")},
				{Name: "Bar", Type: ast.NewIdent("string")},
			}),
			Expected:    []_fieldData{{Name: "Bar", Type: "string"}},
			Description: "Should only include exported fields",
		},
		{
			StructType: structTypeWithFields([]_structField{
				{Name: "Omit", Type: ast.NewIdent("bool"), Tag: `builder:"ignore"`},
				{Name: "Include", Type: ast.NewIdent("int")},
			}),
			Expected:    []_fieldData{{Name: "Include", Type: "int"}},
			Description: "Should not include ignored fields",
		},
		{
			StructType: structTypeWithFields([]_structField{
				{Name: "SomeArray", Type: &ast.ArrayType{Elt: ast.NewIdent("float")}},
				{Name: "SomeSendChannel", Type: &ast.ChanType{Dir: ast.SEND, Value: ast.NewIdent("int")}},
				{Name: "SomeRecvChannel", Type: &ast.ChanType{Dir: ast.RECV, Value: ast.NewIdent("string")}},
				{Name: "SomeMap", Type: &ast.MapType{Key: ast.NewIdent("string"), Value: ast.NewIdent("int")}},
				{Name: "SomePointer", Type: &ast.StarExpr{X: ast.NewIdent("User")}},
			}),
			Expected: []_fieldData{
				{Name: "SomeArray", Type: "[]float"},
				{Name: "SomeSendChannel", Type: "chan<- int"},
				{Name: "SomeRecvChannel", Type: "<-chan string"},
				{Name: "SomeMap", Type: "map[string]int"},
				{Name: "SomePointer", Type: "*User"},
			},
			Description: "Should include arrays, channels, maps, and pointers",
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.Description, func(t *testing.T) {
			fields, err := extractFieldData(test.StructType)

			assert.Nil(t, err)
			assert.ElementsMatch(t, test.Expected, fields)
		})
	}
}

func Test_unexported(t *testing.T) {
	type _testDescription struct {
		Identifier  string
		Expected    string
		Description string
	}

	tests := []_testDescription{
		{
			Identifier:  "",
			Expected:    "",
			Description: "Should play nicely with empty strings",
		},
		{
			Identifier:  "SomethingExported",
			Expected:    "somethingExported",
			Description: "Exported identifiers",
		},
		{
			Identifier:  "alreadyUnexported",
			Expected:    "_alreadyUnexported",
			Description: "Identifiers that are already unexported should be prefixed",
		},
		{
			Identifier:  "Type",
			Expected:    "_type",
			Description: "If the unexported name would be a keyword, it should be prefixed",
		},
		{
			Identifier:  "Uint64",
			Expected:    "_uint64",
			Description: "If the unexported name would be a pre-defined type, it should be prefixed",
		},
		{
			Identifier:  "UUID",
			Expected:    "uuid",
			Description: "All upper-case identifiers should be converted to all lower-case",
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.Description, func(t *testing.T) {
			assert.Equal(t, test.Expected, unexported(test.Identifier))
		})
	}
}

func Test_builderTemplate_ShouldBeValid(t *testing.T) {
	_, err := template.New("builder").
		Funcs(template.FuncMap{"unexported": unexported}).
		Parse(builderTemplateString)
	assert.Nil(t, err)
}
