package main

import (
	"fmt"
	"os"

	protogen "google.golang.org/protobuf/compiler/protogen"
)

const (
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorReset  = "\033[0m"
)

// bindErrorf reports an error via plugin.Error (displayed in red by protoc/buf)
// and terminates code generation.
func bindErrorf(plugin *protogen.Plugin, format string, args ...any) {
	plugin.Error(fmt.Errorf(format, args...))
}

// bindWarnf prints a warning to stderr in yellow and continues code generation.
// protoc/buf passes stderr through to the user.
func bindWarnf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, colorYellow+"warning: "+format+colorReset+"\n", args...)
}
