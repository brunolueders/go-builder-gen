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
)

func CamelToSnakeCase(str string) string {
	runes := []rune(str)
	isUpperOrNumber := make([]bool, 0, len(runes))
	for _, char := range runes {
		isUpperOrNumber = append(isUpperOrNumber, unicode.IsUpper(char) || unicode.IsNumber(char))
	}

	var snakeCase strings.Builder
	for i := 0; i < len(runes); i++ {
		if isUpperOrNumber[i] && (i > 0 && unicode.IsLower(runes[i-1]) ||
			i > 1 && i < len(runes)-1 && isUpperOrNumber[i-1] && !isUpperOrNumber[i+1]) {
			snakeCase.WriteRune('_')
		}
		snakeCase.WriteRune(unicode.ToLower(runes[i]))
	}
	return snakeCase.String()
}

type InputFile struct {
	Name    string
	Target  string
	File    *ast.File
	FileSet *token.FileSet
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
			return nil, fmt.Errorf("'%s' is not in the format FILE:TARGET_STRUCT", input)
		}

		file, err := parser.ParseFile(fileSet, parts[0], nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, fmt.Errorf("failed to parse file '%s': %w", parts[0], err)
		}

		files = append(files, InputFile{
			Name:    parts[0],
			Target:  parts[1],
			File:    file,
			FileSet: fileSet,
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
		exitWithError(fmt.Errorf("failed to load files: %w", err))
	}

	for _, file := range files {
		generatedCode, err := generate(file.Target, &file)
		if err != nil {
			exitWithError(fmt.Errorf("failed to generate builder code for '%s:%s': %w", file.Name, file.Target, err))
		}

		fmt.Printf("Generating '%s'...\n", file.OutputName())
		if err := os.WriteFile(file.OutputName(), generatedCode, 0666); err != nil {
			exitWithError(fmt.Errorf("failed to write builder code to '%s': %w", file.OutputName(), err))
		}
	}
}
