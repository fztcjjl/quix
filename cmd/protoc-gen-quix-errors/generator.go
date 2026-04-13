package main

import (
	"fmt"
	"slices"
	"strings"

	errdesc "github.com/fztcjjl/quix/proto/errdesc"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

// generateFiles iterates over all enums in a proto file and generates
// an _errors.go file for each enum that has http_status annotations.
func generateFiles(plugin *protogen.Plugin, file *protogen.File) {
	for _, enum := range file.Enums {
		values := extractEnumValues(enum)
		if len(values) == 0 {
			continue
		}

		data := FileData{
			PackageName: string(file.GoPackageName),
			SourcePath:  string(file.Desc.Path()),
			Enum: EnumData{
				ProtoName: enum.GoIdent.GoName,
				FileName:  toSnakeCase(enum.GoIdent.GoName) + "_errors.go",
				Values:    values,
			},
		}

		filename := genFilename(file, data.Enum.FileName)
		g := plugin.NewGeneratedFile(filename, file.GoImportPath)
		if err := parsedTemplate.Execute(g, data); err != nil {
			plugin.Error(fmt.Errorf("template execution failed for %s: %v", data.Enum.ProtoName, err))
		}
	}
}

// extractEnumValues reads http_status and error_message extensions from
// enum values and returns the extracted data. Returns nil if no values
// have the http_status annotation.
func extractEnumValues(enum *protogen.Enum) []EnumValueData {
	var values []EnumValueData
	prefix := enumPrefix(enum.GoIdent.GoName)

	for _, val := range enum.Values {
		opts := val.Desc.Options()
		if opts == nil {
			continue
		}

		httpStatus, ok := proto.GetExtension(opts, errdesc.E_HttpStatus).(int32)
		if !ok {
			continue
		}

		message := ""
		if rawMsg, ok := proto.GetExtension(opts, errdesc.E_ErrorMessage).(string); ok {
			message = rawMsg
		}
		if message == "" {
			message = string(val.Desc.Name())
		}

		// Derive function name by stripping enum prefix and converting to PascalCase.
		valueName := string(val.Desc.Name())
		stripped := strings.TrimPrefix(valueName, prefix)
		if stripped == valueName {
			stripped = valueName
		}

		funcName := toPascalCase(stripped)

		// If enum name ends with "Error", prepend the stripped enum name prefix
		// to avoid collisions (e.g., NotFound from UserError, NotFound from OrderError).
		enumGoName := enum.GoIdent.GoName
		if strings.HasSuffix(enumGoName, "Error") {
			strippedEnumName, _ := strings.CutSuffix(enumGoName, "Error")
			funcName = strippedEnumName + funcName
		}

		values = append(values, EnumValueData{
			ProtoName:     valueName,
			ConstName:     toPascalCase(valueName) + "Code",
			FuncName:      funcName,
			FuncNameWD:    funcName + "WithDetails",
			Code:          valueName,
			Message:       message,
			HTTPStatus:    int(httpStatus),
			IsUnspecified: val.Desc.Number() == 0,
		})
	}

	slices.SortFunc(values, func(a, b EnumValueData) int {
		if a.IsUnspecified != b.IsUnspecified {
			// Non-UNSPECIFIED first
			if a.IsUnspecified {
				return 1
			}
			return -1
		}
		return a.HTTPStatus - b.HTTPStatus
	})

	return values
}
