package main

import (
	"fmt"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func generateFile(plugin *protogen.Plugin, file *protogen.File) {
	data := FileData{
		PackageName: string(file.GoPackageName),
		SourcePath:  string(file.Desc.Path()),
	}

	extraImports := make(map[string]bool)

	for _, svc := range file.Services {
		svcData := ServiceData{
			GoName:           svc.GoName,
			InterfaceName:    svc.GoName + "HTTPService",
			RegisterFuncName: "Register" + svc.GoName + "HTTPService",
		}

		for _, method := range svc.Methods {
			outputIdent := method.Output.GoIdent
			outputPkg := string(outputIdent.GoImportPath)
			isVoid := outputIdent.GoName == "Empty" &&
				(outputPkg == "google.golang.org/protobuf/types/known/emptypb" ||
					strings.HasSuffix(outputPkg, "/emptypb"))

			// Determine output type string for generated code
			var outputType string
			switch {
			case isVoid:
				outputType = "" // No rsp variable needed
			case outputIdent.GoImportPath == file.GoImportPath:
				// Same package — just use the type name
				outputType = "*" + outputIdent.GoName
			default:
				// Different package — need full qualified name
				outputType = "*" + outputIdent.GoName
				extraImports[outputPkg] = true
			}

			methodData := MethodData{
				GoName:       method.GoName,
				InputType:    "*" + method.Input.GoIdent.GoName,
				OutputType:   outputType,
				IsVoidReturn: isVoid,
			}

			rule, ok := getHTTPRule(method)
			if !ok {
				plugin.Error(fmt.Errorf("method %s.%s has no google.api.http annotation", svc.GoName, method.GoName))
				continue
			}

			// Extract all bindings (primary + additional)
			bindings := extractAllBindings(rule)
			for i, binding := range bindings {
				route := RouteData{
					Method:      binding.Method,
					Path:        ConvertPath(binding.Path),
					HandlerName: fmt.Sprintf("_%s_%s%d_HTTP_Handler", svc.GoName, method.GoName, i),
					HasBody:     binding.Body != "",
					BodyField:   binding.Body,
					PathVars:    ExtractPathVars(binding.Path),
				}
				methodData.Routes = append(methodData.Routes, route)
			}

			svcData.Methods = append(svcData.Methods, methodData)
		}

		data.Services = append(data.Services, svcData)
	}

	if len(data.Services) == 0 {
		return
	}

	// Collect extra imports
	for imp := range extraImports {
		data.ExtraImports = append(data.ExtraImports, imp)
	}

	filename := genFilename(file, "_gin.go")

	g := plugin.NewGeneratedFile(filename, file.GoImportPath)

	if err := parsedTemplate.Execute(g, data); err != nil {
		plugin.Error(fmt.Errorf("template execution failed: %v", err))
	}
}

// httpBinding represents a single HTTP route binding.
type httpBinding struct {
	Method string // GET, POST, PUT, DELETE, PATCH
	Path   string // Proto-style path with {var}
	Body   string // Body field name
}

// getHTTPRule extracts the google.api.http annotation from a method.
func getHTTPRule(method *protogen.Method) (*annotations.HttpRule, bool) {
	ext := proto.GetExtension(method.Desc.Options(), annotations.E_Http)
	rule, ok := ext.(*annotations.HttpRule)
	return rule, ok
}

// extractAllBindings extracts the primary rule plus all additional_bindings.
func extractAllBindings(rule *annotations.HttpRule) []httpBinding {
	var bindings []httpBinding
	bindings = append(bindings, httpRuleToBinding(rule))
	for _, ab := range rule.GetAdditionalBindings() {
		bindings = append(bindings, httpRuleToBinding(ab))
	}
	return bindings
}

// httpRuleToBinding converts a single HttpRule to an httpBinding.
func httpRuleToBinding(rule *annotations.HttpRule) httpBinding {
	b := httpBinding{
		Body: rule.GetBody(),
	}

	switch {
	case rule.GetGet() != "":
		b.Method = "GET"
		b.Path = rule.GetGet()
	case rule.GetPost() != "":
		b.Method = "POST"
		b.Path = rule.GetPost()
	case rule.GetPut() != "":
		b.Method = "PUT"
		b.Path = rule.GetPut()
	case rule.GetDelete() != "":
		b.Method = "DELETE"
		b.Path = rule.GetDelete()
	case rule.GetPatch() != "":
		b.Method = "PATCH"
		b.Path = rule.GetPatch()
	case rule.GetCustom() != nil:
		b.Method = strings.ToUpper(rule.GetCustom().GetKind())
		b.Path = rule.GetCustom().GetPath()
	}

	return b
}

// init is used to ensure emptypb is available (referenced by type check).
var _ = emptypb.Empty{}
