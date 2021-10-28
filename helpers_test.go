package main

import "go/ast"

func structTypeWithFields(fields map[string]ast.Expr) *ast.StructType {
	fieldList := make([]*ast.Field, 0, len(fields))
	for name, typ := range fields {
		fieldList = append(fieldList, &ast.Field{
			Names: []*ast.Ident{{Name: name}},
			Type:  typ,
		})
	}

	return &ast.StructType{
		Fields: &ast.FieldList{List: fieldList},
	}
}
