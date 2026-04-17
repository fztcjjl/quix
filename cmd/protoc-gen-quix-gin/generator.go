package main

import (
	"fmt"
	"slices"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
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
				// Validate body field exists on the input message
				if binding.Body != "" && binding.Body != "*" {
					found := false
					for _, field := range method.Input.Fields {
						if field.Desc.Name() == protoreflect.Name(binding.Body) {
							found = true
							break
						}
					}
					if !found {
						plugin.Error(fmt.Errorf("method %s.%s: body field %q not found in %s",
							svc.GoName, method.GoName, binding.Body, method.Input.GoIdent.GoName))
						return
					}
				}

				// GET and DELETE must not have a body (HTTP semantic violation)
				if (binding.Method == "GET" || binding.Method == "DELETE") && binding.Body != "" {
					bindErrorf(plugin, "method %s.%s: %s must not have a body",
						svc.GoName, method.GoName, binding.Method)
					return
				}

				// Warn when body "*" and path variables share names with input message fields
				if binding.Body == "*" {
					pathVars := ExtractPathVars(binding.Path)
					for _, pv := range pathVars {
						for _, field := range method.Input.Fields {
							if string(field.Desc.Name()) == pv {
								bindWarnf("method %s.%s: path variable %q has same name as body field, potential conflict; "+
									"consider using body: \"<field>\" to specify the body field explicitly",
									svc.GoName, method.GoName, pv)
								break
							}
						}
					}
				}

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

	// Collect extra imports (sorted for deterministic output)
	for imp := range extraImports {
		data.ExtraImports = append(data.ExtraImports, imp)
	}
	slices.Sort(data.ExtraImports)

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
