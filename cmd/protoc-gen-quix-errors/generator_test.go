package main

import (
	"bytes"
	"flag"
	"os"
	"testing"

	errdesc "github.com/fztcjjl/quix/proto/errdesc"
	"github.com/google/go-cmp/cmp"
	protogen "google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

var updateGolden = flag.Bool("update", false, "update golden file")

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"USER_NOT_FOUND", "UserNotFound"},
		{"USER_ALREADY_EXISTS", "UserAlreadyExists"},
		{"ORDER_PAYMENT_FAILED", "OrderPaymentFailed"},
		{"INVALID_INPUT", "InvalidInput"},
		{"REGULAR_ENUM_UNKNOWN", "RegularEnumUnknown"},
	}
	for _, tt := range tests {
		got := toPascalCase(tt.input)
		if got != tt.want {
			t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"UserError", "user_error"},
		{"OrderError", "order_error"},
		{"GreeterService", "greeter_service"},
	}
	for _, tt := range tests {
		got := toSnakeCase(tt.input)
		if got != tt.want {
			t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEnumPrefix(t *testing.T) {
	tests := []struct {
		enumName string
		want     string
	}{
		{"UserError", "USER_ERROR_"},
		{"OrderError", "ORDER_ERROR_"},
		{"Greeter", "GREETER_"},
	}
	for _, tt := range tests {
		got := enumPrefix(tt.enumName)
		if got != tt.want {
			t.Errorf("enumPrefix(%q) = %q, want %q", tt.enumName, got, tt.want)
		}
	}
}

// TestGenerate uses protogen.Options.New to construct a Plugin from a
// CodeGeneratorRequest without needing the protoc binary. It runs the
// generator and compares the output against the golden file.
func TestGenerate(t *testing.T) {
	protoFiles := []*descriptorpb.FileDescriptorProto{
		protodesc.ToFileDescriptorProto(descriptorpb.File_google_protobuf_descriptor_proto),
		protodesc.ToFileDescriptorProto(errdesc.File_errdesc_proto),
		buildTestProtoFile(),
	}

	req := &pluginpb.CodeGeneratorRequest{
		Parameter:      proto.String("Mgoogle/protobuf/descriptor.proto=google.golang.org/protobuf/types/descriptorpb,Merrdesc/errdesc.proto=github.com/fztcjjl/quix/proto/errdesc"),
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
		generateFiles(plugin, file)
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

func makeEnumValueOpts(httpStatus int32, message string) *descriptorpb.EnumValueOptions {
	opts := &descriptorpb.EnumValueOptions{}
	proto.SetExtension(opts, errdesc.E_HttpStatus, httpStatus)
	if message != "" {
		proto.SetExtension(opts, errdesc.E_ErrorMessage, message)
	}
	return opts
}

// buildTestProtoFile constructs a FileDescriptorProto matching testdata/test.proto.
func buildTestProtoFile() *descriptorpb.FileDescriptorProto {
	return &descriptorpb.FileDescriptorProto{
		Name:    proto.String("testdata/test.proto"),
		Package: proto.String("test"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("github.com/fztcjjl/quix/cmd/protoc-gen-quix-errors/testdata;testdata"),
		},
		Dependency: []string{"errdesc.proto"},
		EnumType: []*descriptorpb.EnumDescriptorProto{
			{
				Name: proto.String("UserError"),
				Value: []*descriptorpb.EnumValueDescriptorProto{
					{Name: proto.String("USER_ERROR_UNSPECIFIED"), Number: proto.Int32(0), Options: makeEnumValueOpts(400, "未知错误")},
					{Name: proto.String("USER_ERROR_NOT_FOUND"), Number: proto.Int32(1), Options: makeEnumValueOpts(404, "用户不存在")},
					{Name: proto.String("USER_ERROR_ALREADY_EXISTS"), Number: proto.Int32(2), Options: makeEnumValueOpts(409, "用户已存在")},
					{Name: proto.String("USER_ERROR_PERMISSION_DENIED"), Number: proto.Int32(3), Options: makeEnumValueOpts(403, "没有权限")},
					{Name: proto.String("USER_ERROR_INVALID_INPUT"), Number: proto.Int32(4), Options: makeEnumValueOpts(400, "参数验证失败")},
				},
			},
			// RegularEnum has no annotations — should produce no output
			{
				Name: proto.String("RegularEnum"),
				Value: []*descriptorpb.EnumValueDescriptorProto{
					{Name: proto.String("REGULAR_ENUM_UNKNOWN"), Number: proto.Int32(0)},
					{Name: proto.String("REGULAR_ENUM_VALUE"), Number: proto.Int32(1)},
				},
			},
			// OrderError has no error_message — should use enum value name
			{
				Name: proto.String("OrderError"),
				Value: []*descriptorpb.EnumValueDescriptorProto{
					{Name: proto.String("ORDER_ERROR_UNSPECIFIED"), Number: proto.Int32(0), Options: makeEnumValueOpts(400, "")},
					{Name: proto.String("ORDER_ERROR_NOT_FOUND"), Number: proto.Int32(1), Options: makeEnumValueOpts(404, "")},
					{Name: proto.String("ORDER_ERROR_PAYMENT_FAILED"), Number: proto.Int32(2), Options: makeEnumValueOpts(402, "")},
				},
			},
		},
	}
}
