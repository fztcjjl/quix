package main

import (
	"log"

	"github.com/fztcjjl/quix"
	pb "github.com/fztcjjl/quix/examples/proto-api/gen/greeter"
	"github.com/fztcjjl/quix/examples/proto-api/service"
)

func main() {
	app := quix.New()

	greeterSvc := service.NewGreeterService()
	pb.RegisterGreeterHTTPService(app.Group("/api"), greeterSvc)

	log.Println("Server listening on :8080")
	app.Run()
}
