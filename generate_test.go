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
	var structType = structTypeWithFields(map[string]ast.Expr{"Foo": ast.NewIdent("int"), "Bar": ast.NewIdent("string")})
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
	var structType = structTypeWithFields(map[string]ast.Expr{"Foo": ast.NewIdent("int"), "Bar": ast.NewIdent("string")})
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
			StructType: structTypeWithFields(map[string]ast.Expr{
				"foo": ast.NewIdent("int"),
				"Bar": ast.NewIdent("string"),
			}),
			Expected:    []_fieldData{{Name: "Bar", Type: "string"}},
			Description: "Should only include exported fields",
		},
		{
			StructType: structTypeWithFields(map[string]ast.Expr{
				"SomeArray":   &ast.ArrayType{Elt: ast.NewIdent("float")},
				"SomeChannel": &ast.ChanType{Dir: ast.SEND, Value: ast.NewIdent("int")},
				"SomeMap":     &ast.MapType{Key: ast.NewIdent("string"), Value: ast.NewIdent("int")},
				"SomePointer": &ast.StarExpr{X: ast.NewIdent("User")},
			}),
			Expected: []_fieldData{
				{Name: "SomeArray", Type: "[]float"},
				{Name: "SomeChannel", Type: "chan<- int"},
				{Name: "SomeMap", Type: "map[string]int"},
				{Name: "SomePointer", Type: "*User"},
			},
			Description: "Should include arrays, channels, maps, and pointers",
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.Description, func(t *testing.T) {
			assert.ElementsMatch(t, test.Expected, extractFieldData(test.StructType))
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
