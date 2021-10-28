package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"strings"
	"unicode"

	"github.com/pkg/errors"
)

func CamelToSnakeCase(str string) string {
	runes := []rune(str)
	isUpperOrNumber := make([]bool, 0, len(runes))
	for _, char := range runes {
		isUpperOrNumber = append(isUpperOrNumber, unicode.IsUpper(char) || unicode.IsNumber(char))
	}

	var snakeCase strings.Builder
	for i := 0; i < len(runes); i++ {
		if i > 0 && isUpperOrNumber[i] && unicode.IsLower(runes[i-1]) ||
			i > 1 && i < len(runes)-1 && isUpperOrNumber[i] && isUpperOrNumber[i-1] && !isUpperOrNumber[i+1] {
			snakeCase.WriteRune('_')
		}
		snakeCase.WriteRune(unicode.ToLower(runes[i]))
	}
	return snakeCase.String()
}

type InputFile struct {
	Name   string
	Target string
	File   *ast.File
}

func (file InputFile) OutputName() string {
	dir := path.Dir(file.Name)
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}
	return dir + CamelToSnakeCase(file.Target) + "_builder.go"
}

func loadFiles(inputs []string) ([]InputFile, error) {
	files := make([]InputFile, 0, len(inputs))
	fileSet := token.NewFileSet()
	for _, input := range inputs {
		parts := strings.SplitN(input, ":", 2)
		if len(parts) != 2 {
			return nil, errors.Errorf("'%s' is not in the format FILE:TARGET_STRUCT", input)
		}

		file, err := parser.ParseFile(fileSet, parts[0], nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse file '%s'", parts[0])
		}

		files = append(files, InputFile{
			Name:   parts[0],
			Target: parts[1],
			File:   file,
		})
	}
	return files, nil
}

func exitWithError(err error) {
	fmt.Printf("ERROR %v\n", err)
	os.Exit(1)
}

func main() {
	files, err := loadFiles(os.Args[1:])
	if err != nil {
		exitWithError(errors.Wrap(err, "failed to load files"))
	}

	for _, file := range files {
		generatedCode, err := generate(file.Target, file.File)
		if err != nil {
			exitWithError(errors.Wrapf(err, "failed to generate builder code for '%s:%s'", file.Name, file.Target))
		}

		fmt.Printf("Generating file '%s'...", file.OutputName())
		if err := os.WriteFile(file.OutputName(), generatedCode, 0666); err != nil {
			exitWithError(errors.Wrapf(err, "failed to write builder code to '%s'", file.OutputName()))
		}
	}
}
