package main

import (
	"go/ast"
)

type _structField struct {
	Name string
	Type ast.Expr
	Tag  string
}

func (field _structField) tagLiteral() *ast.BasicLit {
	if field.Tag == "" {
		return nil
	}
	return &ast.BasicLit{Value: field.Tag}
}

func structTypeWithFields(fields []_structField) *ast.StructType {
	fieldList := make([]*ast.Field, 0, len(fields))
	for _, field := range fields {
		fieldList = append(fieldList, &ast.Field{
			Names: []*ast.Ident{{Name: field.Name}},
			Type:  field.Type,
			Tag:   field.tagLiteral(),
		})
	}

	return &ast.StructType{
		Fields: &ast.FieldList{List: fieldList},
	}
}
