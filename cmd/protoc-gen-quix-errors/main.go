package main

import (
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

var pathsSourceRelative bool

// ParamFunc is the primary way protogen parses parameters. However, some
// environments (e.g., certain buf or protoc versions) pass parameters as a
// raw string without invoking ParamFunc. The fallback in Run() ensures
// paths=source_relative is always detected.

func main() {
	protogen.Options{
		ParamFunc: func(name, value string) error {
			if name == "paths" && value == "source_relative" {
				pathsSourceRelative = true
			}
			return nil
		},
	}.Run(func(plugin *protogen.Plugin) error {
		plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		// Fallback: also check the parameter string directly
		if !pathsSourceRelative {
			param := plugin.Request.GetParameter()
			if param != "" {
				for _, p := range strings.Split(param, ",") {
					p = strings.TrimSpace(p)
					if p == "paths=source_relative" {
						pathsSourceRelative = true
						break
					}
				}
			}
		}
		for _, file := range plugin.Files {
			if !file.Generate {
				continue
			}
			generateFiles(plugin, file)
		}
		return nil
	})
}

// genFilename returns the output filename for a generated file.
// When paths=source_relative is set, the file is placed in the same
// directory as the source proto (using GeneratedFilenamePrefix's directory).
// Otherwise, it uses the Go import path.
func genFilename(file *protogen.File, filename string) string {
	if pathsSourceRelative {
		dir := filepath.Dir(file.GeneratedFilenamePrefix)
		return filepath.Join(dir, filename)
	}
	return filepath.Join(string(file.GoImportPath), filename)
}
