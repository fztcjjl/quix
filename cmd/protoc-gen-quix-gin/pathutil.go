package main

import (
	"regexp"
)

var pathVarRe = regexp.MustCompile(`\{([^}]+)\}`)

// ConvertPath converts proto path template to Gin style.
// "/v1/users/{user_id}" → "/v1/users/:user_id"
func ConvertPath(protoPath string) string {
	return pathVarRe.ReplaceAllString(protoPath, ":$1")
}

// ExtractPathVars extracts variable names from a proto path template.
// "/v1/users/{user_id}/{book.id}" → ["user_id", "book.id"]
func ExtractPathVars(protoPath string) []string {
	matches := pathVarRe.FindAllStringSubmatch(protoPath, -1)
	vars := make([]string, 0, len(matches))
	for _, m := range matches {
		vars = append(vars, m[1])
	}
	return vars
}
