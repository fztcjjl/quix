package main

import (
	_ "embed"
	"strings"
	"text/template"
	"unicode"
)

//go:embed errors.tpl
var errorsTemplate string

// parsedTemplate is the parsed Go template for generating error code.
var parsedTemplate = template.Must(
	template.New("errors").Funcs(template.FuncMap{
		"toPascalCase": toPascalCase,
		"toSnakeCase":  toSnakeCase,
		"apperrorsPkg": apperrorsImportPath,
	}).Parse(errorsTemplate),
)

// apperrorsImportPath returns the Go import path for the apperrors package.
func apperrorsImportPath() string {
	return "github.com/fztcjjl/quix/core/errors"
}

// toPascalCase converts UPPER_SNAKE_CASE to PascalCase.
// e.g., "USER_NOT_FOUND" → "UserNotFound"
func toPascalCase(s string) string {
	var b strings.Builder
	nextUpper := true
	for _, r := range s {
		if r == '_' {
			nextUpper = true
			continue
		}
		if nextUpper {
			b.WriteRune(unicode.ToUpper(r))
			nextUpper = false
		} else {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

// toSnakeCase converts PascalCase to snake_case.
// e.g., "UserError" → "user_error"
func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			b.WriteRune('_')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// enumPrefix converts an enum name (PascalCase) to the expected UPPER_SNAKE_CASE prefix.
// e.g., "UserError" → "USER_ERROR_"
func enumPrefix(enumName string) string {
	return strings.ToUpper(toSnakeCase(enumName)) + "_"
}

// EnumValueData holds data for a single proto enum value.
type EnumValueData struct {
	ProtoName     string // e.g., USER_NOT_FOUND
	ConstName     string // e.g., UserNotFoundCode
	FuncName      string // e.g., UserNotFound
	FuncNameWD    string // e.g., UserNotFoundWithDetails
	Code          string // e.g., USER_NOT_FOUND
	Message       string // e.g., 用户不存在
	HTTPStatus    int    // e.g., 404
	IsUnspecified bool   // true for the zero value
}

// EnumData holds data for a single proto enum.
type EnumData struct {
	ProtoName string // e.g., UserError
	FileName  string // e.g., user_error_errors.go
	Values    []EnumValueData
}

// FileData holds all data needed to generate a single _errors.go file.
type FileData struct {
	PackageName string   // Go package name
	SourcePath  string   // Proto file path
	Enum        EnumData // The single enum for this file
}
