package transport

import (
	"context"
	"testing"
)

type mockServer struct{}

func (m *mockServer) Start() error                   { return nil }
func (m *mockServer) Stop(ctx context.Context) error { return nil }

var _ Server = (*mockServer)(nil)

func TestMockServerSatisfiesInterface(t *testing.T) {
	var s Server = &mockServer{}
	if err := s.Start(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if err := s.Stop(context.TODO()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}
