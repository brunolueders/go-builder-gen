package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_camelToSnakeCase(t *testing.T) {
	type _testDescription struct {
		Input       string
		Expected    string
		Description string
	}

	tests := []_testDescription{
		{
			Input:       "",
			Expected:    "",
			Description: "Should play nicely with empty inputs",
		},
		{
			Input:       "UpperCamelCase",
			Expected:    "upper_camel_case",
			Description: "Should convert upper-camel-case strings",
		},
		{
			Input:       "lowerCamelCase",
			Expected:    "lower_camel_case",
			Description: "Should convert lower-camel-case strings",
		},
		{
			Input:       "AbC",
			Expected:    "ab_c",
			Description: "Edge-case: String ends with an upper-case character",
		},
		{
			Input:       "URL",
			Expected:    "url",
			Description: "Should not split up acronyms",
		},
		{
			Input:       "AssignedIDNumber",
			Expected:    "assigned_id_number",
			Description: "Should handle acronyms inside strings correctly",
		},
		{
			Input:       "Hel10Yall",
			Expected:    "hel_10_yall",
			Description: "Should split around numbers",
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.Description, func(t *testing.T) {
			assert.Equal(t, test.Expected, CamelToSnakeCase(test.Input))
		})
	}
}
