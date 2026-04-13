package runtime

import (
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
