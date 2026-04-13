package service

import (
	"context"
	"fmt"

	pb "github.com/fztcjjl/quix/examples/proto-api/gen/greeter"
)

type GreeterService struct{}

func NewGreeterService() *GreeterService {
	return &GreeterService{}
}

func (s *GreeterService) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{
		Message: fmt.Sprintf("Hello, %s!", req.GetName()),
	}, nil
}
