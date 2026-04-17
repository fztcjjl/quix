package server

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/form/v4"
)

var formDecoder = form.NewDecoder()

func init() {
	// Proto-generated structs only have json tags, not uri/form tags.
	// SetTagName("json") allows the form decoder to decode using json tag names.
	formDecoder.SetTagName("json")
}

// ShouldBindUri decodes path parameters into req using the form decoder.
// This works with proto-generated structs that have json tags.
func (c *Context) ShouldBindUri(req any) error {
	values := make(map[string][]string)
	for _, p := range c.Params {
		values[p.Key] = []string{p.Value}
	}
	return formDecoder.Decode(req, values)
}

// ShouldBindQuery decodes URL query parameters into req using the form decoder.
// This works with proto-generated structs that have json tags.
func (c *Context) ShouldBindQuery(req any) error {
	return formDecoder.Decode(req, c.Request.URL.Query())
}

// ShouldBindJSON binds the request body as JSON into req.
func (c *Context) ShouldBindJSON(req any) error {
	return c.Context.ShouldBindJSON(req)
}

// ShouldBindUriConflictCheck binds path parameters into req after checking
// for conflicts with values already set by JSON body binding.
// If a path variable has a corresponding field that was already set by the body
// with a different value, an error is returned. Otherwise, URI values are bound
// (overwriting zero-value or matching body values).
func (c *Context) ShouldBindUriConflictCheck(req any, pathVars []string) error {
	rv := reflect.ValueOf(req)
	if rv.Kind() != reflect.Pointer {
		return nil
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return nil
	}

	for _, p := range c.Params {
		fieldIdx := findFieldByJSONTag(rv.Type(), p.Key)
		if fieldIdx == nil {
			continue
		}
		fv := rv.FieldByIndex(fieldIdx)
		if !fv.IsValid() || fv.IsZero() {
			continue // body didn't set this field, no conflict
		}
		// Body set this field — compare with URI value
		if err := compareFieldValue(fv, p.Value); err != nil {
			return fmt.Errorf("path param %q=%q conflicts with body field value %v",
				p.Key, p.Value, fv.Interface())
		}
	}

	// No conflicts — bind URI values (overwrites matching or zero-value fields)
	values := make(map[string][]string)
	for _, p := range c.Params {
		values[p.Key] = []string{p.Value}
	}
	return formDecoder.Decode(req, values)
}

// findFieldByJSONTag returns the field index for the struct field whose json tag
// matches the given name, or nil if not found.
func findFieldByJSONTag(typ reflect.Type, name string) []int {
	for i := range typ.NumField() {
		f := typ.Field(i)
		tag := f.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		// json tag can be "name,omitempty" — take the first part
		if idx := strings.Index(tag, ","); idx != -1 {
			tag = tag[:idx]
		}
		if tag == name {
			return []int{i}
		}
	}
	return nil
}

// compareFieldValue compares a reflect.Value (set by JSON body) with a URI string value.
// Returns nil if they match, or an error if they differ.
func compareFieldValue(fv reflect.Value, uriVal string) error {
	switch fv.Kind() {
	case reflect.String:
		if fv.String() != uriVal {
			return fmt.Errorf("string mismatch: body=%q, uri=%q", fv.String(), uriVal)
		}
	case reflect.Int, reflect.Int32, reflect.Int64:
		iv, err := strconv.ParseInt(uriVal, 10, 64)
		if err == nil && fv.Int() != iv {
			return fmt.Errorf("int mismatch: body=%v, uri=%s", fv.Int(), uriVal)
		}
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		uv, err := strconv.ParseUint(uriVal, 10, 64)
		if err == nil && fv.Uint() != uv {
			return fmt.Errorf("uint mismatch: body=%v, uri=%s", fv.Uint(), uriVal)
		}
	}
	return nil
}
