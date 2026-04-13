package main

import (
	_ "embed"
	"strings"
	"text/template"
	"unicode"
)

//go:embed http.tpl
var httpTemplate string

// parsedTemplate is the parsed Go template for generating Gin HTTP code.
var parsedTemplate = template.Must(
	template.New("http").Funcs(template.FuncMap{
		"runtimePkg":  runtimeImportPath,
		"exportField": exportField,
		"trimPrefix":  strings.TrimPrefix,
	}).Parse(httpTemplate),
)

// runtimeImportPath returns the Go import path for the runtime package.
func runtimeImportPath() string {
	return "github.com/fztcjjl/quix/internal/protoc-gen-quix-gin/runtime"
}

// exportField converts a proto field name (snake_case) to Go exported name (PascalCase).
// e.g., "user_id" → "UserId", "user" → "User"
func exportField(name string) string {
	parts := strings.Split(name, ".")
	for i, p := range parts {
		parts[i] = toPascalCase(p)
	}
	return strings.Join(parts, ".")
}

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
			b.WriteRune(r)
		}
	}
	return b.String()
}

// RouteData holds information about a single HTTP route binding.
type RouteData struct {
	Method      string // GET, POST, PUT, DELETE, PATCH
	Path        string // Gin-style path with :var
	HandlerName string // _Service_MethodN_HTTP_Handler
	HasBody     bool   // Whether the route has a request body
	BodyField   string // Body field name ("*" means entire message, "" means no body)
	PathVars    []string
}

// MethodData holds information about a proto method.
type MethodData struct {
	GoName       string      // e.g., SayHello
	InputType    string      // e.g., *HelloRequest
	OutputType   string      // e.g., *HelloReply (full Go type reference)
	IsVoidReturn bool        // true if output is google.protobuf.Empty
	Routes       []RouteData // One or more routes (additional_bindings produce multiple)
}

// ServiceData holds information about a proto service.
type ServiceData struct {
	GoName           string       // e.g., Greeter
	InterfaceName    string       // e.g., GreeterHTTPService
	RegisterFuncName string       // e.g., RegisterGreeterHTTPService
	Methods          []MethodData // All methods in the service
}

// FileData holds all data needed to generate a single _gin.go file.
type FileData struct {
	PackageName  string        // Go package name
	SourcePath   string        // Proto file path
	ExtraImports []string      // Additional import paths needed (e.g., emptypb)
	Services     []ServiceData // All services in the file
}
