package main

import (
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

var pathsSourceRelative bool

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
			// protogen may split "paths=source_relative" into separate calls
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
			generateFile(plugin, file)
		}
		return nil
	})
}

// genFilename returns the output filename for a generated file.
// When paths=source_relative is set, the GeneratedFilenamePrefix already
// reflects the source-relative path, so we just append the suffix.
// Otherwise, we use the Go import path.
func genFilename(file *protogen.File, suffix string) string {
	if pathsSourceRelative {
		// GeneratedFilenamePrefix already has the source-relative path
		return file.GeneratedFilenamePrefix + suffix
	}
	return filepath.Join(string(file.GoImportPath), file.GeneratedFilenamePrefix+suffix)
}
