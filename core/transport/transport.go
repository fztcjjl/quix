package transport

import "context"

// Server is the interface for all transport servers (HTTP, gRPC, etc.).
type Server interface {
	Start() error
	Stop(ctx context.Context) error
}
