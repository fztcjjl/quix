package main

import (
	"bytes"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/genproto/googleapis/api/annotations"
	protogen "google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/pluginpb"
)

var updateGolden = flag.Bool("update", false, "update golden file")

// TestPathUtils tests the ConvertPath and ExtractPathVars functions.
func TestConvertPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/v1/users/{user_id}", "/v1/users/:user_id"},
		{"/v1/users/{user_id}/books/{book_id}", "/v1/users/:user_id/books/:book_id"},
		{"/hello/{name}", "/hello/:name"},
		{"/users", "/users"},
		{"/v1/{book.id}", "/v1/:book.id"},
	}
	for _, tt := range tests {
		got := ConvertPath(tt.input)
		if got != tt.want {
			t.Errorf("ConvertPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestExtractPathVars(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"/v1/users/{user_id}", []string{"user_id"}},
		{"/v1/users/{user_id}/books/{book_id}", []string{"user_id", "book_id"}},
		{"/hello/{name}", []string{"name"}},
		{"/users", []string{}},
	}
	for _, tt := range tests {
		got := ExtractPathVars(tt.input)
		if !cmp.Equal(got, tt.want) {
			t.Errorf("ExtractPathVars(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// TestGenerate uses protogen.Options.New to construct a Plugin from a
// CodeGeneratorRequest without needing the protoc binary. It runs the
// generator and compares the output against the golden file.
func TestGenerate(t *testing.T) {
	protoFiles := []*descriptorpb.FileDescriptorProto{
		protodesc.ToFileDescriptorProto(descriptorpb.File_google_protobuf_descriptor_proto),
		protodesc.ToFileDescriptorProto(annotations.File_google_api_http_proto),
		protodesc.ToFileDescriptorProto(annotations.File_google_api_annotations_proto),
		protodesc.ToFileDescriptorProto(emptypb.File_google_protobuf_empty_proto),
		buildTestProtoFile(),
	}

	req := &pluginpb.CodeGeneratorRequest{
		Parameter:      proto.String("Mgoogle/protobuf/descriptor.proto=google.golang.org/protobuf/types/descriptorpb,Mgoogle/protobuf/empty.proto=google.golang.org/protobuf/types/known/emptypb"),
		ProtoFile:      protoFiles,
		FileToGenerate: []string{"testdata/test.proto"},
	}

	plugin, err := (&protogen.Options{}).New(req)
	if err != nil {
		t.Fatalf("protogen.Options.New() failed: %v", err)
	}

	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}
		generateFile(plugin, file)
	}

	resp := plugin.Response()
	if resp.GetError() != "" {
		t.Fatalf("plugin error: %s", resp.GetError())
	}
	if len(resp.File) == 0 {
		t.Fatal("plugin produced no output files")
	}

	got := []byte(resp.File[0].GetContent())

	goldenFile := "testdata/golden.golden"
	if *updateGolden {
		if err := os.WriteFile(goldenFile, got, 0600); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		t.Logf("golden file updated: %s", goldenFile)
		return
	}

	want, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	if !bytes.Equal(got, want) {
		t.Errorf("generated code mismatch (-want +got):\n%s", cmp.Diff(string(want), string(got)))
	}
}

func optionalField() *descriptorpb.FieldDescriptorProto_Label {
	return descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum()
}
func stringType() *descriptorpb.FieldDescriptorProto_Type {
	return descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()
}
func int32Type() *descriptorpb.FieldDescriptorProto_Type {
	return descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum()
}
func messageType() *descriptorpb.FieldDescriptorProto_Type {
	return descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()
}

// buildTestProtoFile constructs a FileDescriptorProto matching testdata/test.proto.
func buildTestProtoFile() *descriptorpb.FileDescriptorProto {
	sayHelloOpts := &descriptorpb.MethodOptions{}
	proto.SetExtension(sayHelloOpts, annotations.E_Http, &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Get{
			Get: "/hello/{name}",
		},
	})

	createUserOpts := &descriptorpb.MethodOptions{}
	proto.SetExtension(createUserOpts, annotations.E_Http, &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Post{
			Post: "/users",
		},
		Body: "*",
	})

	updateUserOpts := &descriptorpb.MethodOptions{}
	proto.SetExtension(updateUserOpts, annotations.E_Http, &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Put{
			Put: "/users/{user_id}",
		},
		Body: "user",
	})

	deleteUserOpts := &descriptorpb.MethodOptions{}
	proto.SetExtension(deleteUserOpts, annotations.E_Http, &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Delete{
			Delete: "/users/{user_id}",
		},
	})

	searchUsersOpts := &descriptorpb.MethodOptions{}
	proto.SetExtension(searchUsersOpts, annotations.E_Http, &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Get{
			Get: "/users/search",
		},
		AdditionalBindings: []*annotations.HttpRule{
			{
				Pattern: &annotations.HttpRule_Post{
					Post: "/users/search",
				},
				Body: "*",
			},
		},
	})

	addItemOpts := &descriptorpb.MethodOptions{}
	proto.SetExtension(addItemOpts, annotations.E_Http, &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Post{
			Post: "/users/{user_id}/items",
		},
		Body: "*",
	})

	return &descriptorpb.FileDescriptorProto{
		Name:    proto.String("testdata/test.proto"),
		Package: proto.String("test"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("github.com/fztcjjl/quix/cmd/protoc-gen-quix-gin/testdata;testdata"),
		},
		Dependency: []string{
			"google/api/annotations.proto",
			"google/protobuf/empty.proto",
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("HelloRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("name"), Number: proto.Int32(1), Type: stringType(), Label: optionalField()},
				},
			},
			{
				Name: proto.String("HelloReply"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("message"), Number: proto.Int32(1), Type: stringType(), Label: optionalField()},
				},
			},
			{
				Name: proto.String("CreateUserRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("name"), Number: proto.Int32(1), Type: stringType(), Label: optionalField()},
					{Name: proto.String("email"), Number: proto.Int32(2), Type: stringType(), Label: optionalField()},
				},
			},
			{
				Name: proto.String("UpdateUserRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("user_id"), Number: proto.Int32(1), Type: stringType(), Label: optionalField()},
					{Name: proto.String("user"), Number: proto.Int32(2), TypeName: proto.String(".test.UserUpdate"), Type: messageType(), Label: optionalField()},
				},
			},
			{
				Name: proto.String("UserUpdate"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("name"), Number: proto.Int32(1), Type: stringType(), Label: optionalField()},
					{Name: proto.String("email"), Number: proto.Int32(2), Type: stringType(), Label: optionalField()},
				},
			},
			{
				Name: proto.String("UserResponse"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("id"), Number: proto.Int32(1), Type: stringType(), Label: optionalField()},
					{Name: proto.String("name"), Number: proto.Int32(2), Type: stringType(), Label: optionalField()},
					{Name: proto.String("email"), Number: proto.Int32(3), Type: stringType(), Label: optionalField()},
				},
			},
			{
				Name: proto.String("DeleteUserRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("user_id"), Number: proto.Int32(1), Type: stringType(), Label: optionalField()},
				},
			},
			{
				Name: proto.String("SearchUsersRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("query"), Number: proto.Int32(1), Type: stringType(), Label: optionalField()},
					{Name: proto.String("page_size"), Number: proto.Int32(2), Type: int32Type(), Label: optionalField()},
				},
			},
			{
				Name: proto.String("UserListResponse"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("users"), Number: proto.Int32(1), TypeName: proto.String(".test.UserResponse"), Type: messageType(), Label: descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum()},
					{Name: proto.String("total"), Number: proto.Int32(2), Type: int32Type(), Label: optionalField()},
				},
			},
			{
				Name: proto.String("AddItemRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("title"), Number: proto.Int32(1), Type: stringType(), Label: optionalField()},
					{Name: proto.String("quantity"), Number: proto.Int32(2), Type: int32Type(), Label: optionalField()},
				},
			},
			{
				Name: proto.String("ItemResponse"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("id"), Number: proto.Int32(1), Type: stringType(), Label: optionalField()},
					{Name: proto.String("title"), Number: proto.Int32(2), Type: stringType(), Label: optionalField()},
				},
			},
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name: proto.String("Greeter"),
				Method: []*descriptorpb.MethodDescriptorProto{
					{
						Name:       proto.String("SayHello"),
						InputType:  proto.String(".test.HelloRequest"),
						OutputType: proto.String(".test.HelloReply"),
						Options:    sayHelloOpts,
					},
					{
						Name:       proto.String("CreateUser"),
						InputType:  proto.String(".test.CreateUserRequest"),
						OutputType: proto.String(".test.UserResponse"),
						Options:    createUserOpts,
					},
					{
						Name:       proto.String("UpdateUser"),
						InputType:  proto.String(".test.UpdateUserRequest"),
						OutputType: proto.String(".test.UserResponse"),
						Options:    updateUserOpts,
					},
					{
						Name:       proto.String("DeleteUser"),
						InputType:  proto.String(".test.DeleteUserRequest"),
						OutputType: proto.String(".google.protobuf.Empty"),
						Options:    deleteUserOpts,
					},
					{
						Name:       proto.String("SearchUsers"),
						InputType:  proto.String(".test.SearchUsersRequest"),
						OutputType: proto.String(".test.UserListResponse"),
						Options:    searchUsersOpts,
					},
					{
						Name:       proto.String("AddItemToUser"),
						InputType:  proto.String(".test.AddItemRequest"),
						OutputType: proto.String(".test.ItemResponse"),
						Options:    addItemOpts,
					},
				},
			},
		},
	}
}

// helper to run generateFile and capture plugin error + stderr output
func runGenerate(t *testing.T, protoFile *descriptorpb.FileDescriptorProto) (*pluginpb.CodeGeneratorResponse, string) {
	t.Helper()

	protoFiles := []*descriptorpb.FileDescriptorProto{
		protodesc.ToFileDescriptorProto(descriptorpb.File_google_protobuf_descriptor_proto),
		protodesc.ToFileDescriptorProto(annotations.File_google_api_http_proto),
		protodesc.ToFileDescriptorProto(annotations.File_google_api_annotations_proto),
		protodesc.ToFileDescriptorProto(emptypb.File_google_protobuf_empty_proto),
		protoFile,
	}

	req := &pluginpb.CodeGeneratorRequest{
		Parameter:      proto.String("Mgoogle/protobuf/descriptor.proto=google.golang.org/protobuf/types/descriptorpb,Mgoogle/protobuf/empty.proto=google.golang.org/protobuf/types/known/emptypb"),
		ProtoFile:      protoFiles,
		FileToGenerate: []string{protoFile.GetName()},
	}

	// Capture stderr for warnings
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	plugin, err := (&protogen.Options{}).New(req)
	if err != nil {
		t.Fatalf("protogen.Options.New() failed: %v", err)
	}

	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}
		generateFile(plugin, file)
	}

	w.Close()
	os.Stderr = oldStderr

	var stderrOutput string
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	if n > 0 {
		stderrOutput = string(buf[:n])
	}

	return plugin.Response(), stderrOutput
}

func TestGenerate_GetWithBody_Error(t *testing.T) {
	opts := &descriptorpb.MethodOptions{}
	proto.SetExtension(opts, annotations.E_Http, &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Get{
			Get: "/hello/{name}",
		},
		Body: "*",
	})

	protoFile := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("test_error.proto"),
		Package: proto.String("test"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("test/test;test"),
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("HelloRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("name"), Number: proto.Int32(1), Type: stringType(), Label: optionalField()},
				},
			},
			{
				Name: proto.String("HelloReply"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("message"), Number: proto.Int32(1), Type: stringType(), Label: optionalField()},
				},
			},
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name: proto.String("Greeter"),
				Method: []*descriptorpb.MethodDescriptorProto{
					{
						Name:       proto.String("SayHello"),
						InputType:  proto.String(".test.HelloRequest"),
						OutputType: proto.String(".test.HelloReply"),
						Options:    opts,
					},
				},
			},
		},
	}

	resp, _ := runGenerate(t, protoFile)
	if resp.GetError() == "" {
		t.Fatal("expected plugin error for GET with body, got none")
	}
	if !strings.Contains(resp.GetError(), "GET must not have a body") {
		t.Errorf("error message should mention GET body violation, got: %s", resp.GetError())
	}
}

func TestGenerate_DeleteWithBody_Error(t *testing.T) {
	opts := &descriptorpb.MethodOptions{}
	proto.SetExtension(opts, annotations.E_Http, &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Delete{
			Delete: "/users/{user_id}",
		},
		Body: "*",
	})

	protoFile := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("test_delete_body.proto"),
		Package: proto.String("test"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("test/test;test"),
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("DeleteUserRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("user_id"), Number: proto.Int32(1), Type: stringType(), Label: optionalField()},
				},
			},
			{
				Name: proto.String("Empty"),
			},
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name: proto.String("Greeter"),
				Method: []*descriptorpb.MethodDescriptorProto{
					{
						Name:       proto.String("DeleteUser"),
						InputType:  proto.String(".test.DeleteUserRequest"),
						OutputType: proto.String(".test.Empty"),
						Options:    opts,
					},
				},
			},
		},
	}

	resp, _ := runGenerate(t, protoFile)
	if resp.GetError() == "" {
		t.Fatal("expected plugin error for DELETE with body, got none")
	}
	if !strings.Contains(resp.GetError(), "DELETE must not have a body") {
		t.Errorf("error message should mention DELETE body violation, got: %s", resp.GetError())
	}
}

func TestGenerate_BodyStarPathVarSameName_ConflictCheck(t *testing.T) {
	opts := &descriptorpb.MethodOptions{}
	proto.SetExtension(opts, annotations.E_Http, &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Post{
			Post: "/agents/{id}",
		},
		Body: "*",
	})

	protoFile := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("test_conflict.proto"),
		Package: proto.String("test"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("test/test;test"),
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("AgentRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("id"), Number: proto.Int32(1), Type: stringType(), Label: optionalField()},
					{Name: proto.String("name"), Number: proto.Int32(2), Type: stringType(), Label: optionalField()},
				},
			},
			{
				Name: proto.String("AgentResponse"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("id"), Number: proto.Int32(1), Type: stringType(), Label: optionalField()},
				},
			},
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name: proto.String("Agent"),
				Method: []*descriptorpb.MethodDescriptorProto{
					{
						Name:       proto.String("CreateAgent"),
						InputType:  proto.String(".test.AgentRequest"),
						OutputType: proto.String(".test.AgentResponse"),
						Options:    opts,
					},
				},
			},
		},
	}

	resp, _ := runGenerate(t, protoFile)
	if resp.GetError() != "" {
		t.Fatalf("expected no plugin error, got: %s", resp.GetError())
	}
	if len(resp.File) == 0 {
		t.Fatal("expected generated output file, got none")
	}
	// Should generate ShouldBindUriConflictCheck since path var "id" matches body field "id"
	generated := resp.File[0].GetContent()
	if !strings.Contains(generated, "ShouldBindUriConflictCheck") {
		t.Errorf("expected ShouldBindUriConflictCheck in generated code, got: %s", generated)
	}
	if strings.Contains(generated, "ShouldBindUri(") {
		t.Errorf("should NOT use plain ShouldBindUri when path var conflicts with body field, got: %s", generated)
	}
}

func TestGenerate_BodyStarPathVarDifferentName_PlainUri(t *testing.T) {
	opts := &descriptorpb.MethodOptions{}
	proto.SetExtension(opts, annotations.E_Http, &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Post{
			Post: "/users/{user_id}/items",
		},
		Body: "*",
	})

	protoFile := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("test_no_conflict.proto"),
		Package: proto.String("test"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("test/test;test"),
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("CreateItemRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("title"), Number: proto.Int32(1), Type: stringType(), Label: optionalField()},
				},
			},
			{
				Name: proto.String("ItemResponse"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("id"), Number: proto.Int32(1), Type: stringType(), Label: optionalField()},
				},
			},
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name: proto.String("Item"),
				Method: []*descriptorpb.MethodDescriptorProto{
					{
						Name:       proto.String("CreateItem"),
						InputType:  proto.String(".test.CreateItemRequest"),
						OutputType: proto.String(".test.ItemResponse"),
						Options:    opts,
					},
				},
			},
		},
	}

	resp, _ := runGenerate(t, protoFile)
	if resp.GetError() != "" {
		t.Fatalf("expected no plugin error, got: %s", resp.GetError())
	}
	if len(resp.File) == 0 {
		t.Fatal("expected generated output file, got none")
	}
	// Should generate plain ShouldBindUri since path var "user_id" doesn't match any field
	generated := resp.File[0].GetContent()
	if !strings.Contains(generated, "ShouldBindUri(") {
		t.Errorf("expected ShouldBindUri in generated code, got: %s", generated)
	}
	if strings.Contains(generated, "ShouldBindUriConflictCheck") {
		t.Errorf("should NOT use ShouldBindUriConflictCheck when path var doesn't match body field, got: %s", generated)
	}
}
