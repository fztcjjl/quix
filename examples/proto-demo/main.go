package main

import (
	quix "github.com/fztcjjl/quix"
	pb "github.com/fztcjjl/quix/examples/proto-demo/gen/task/v1"
	"github.com/fztcjjl/quix/examples/proto-demo/service"
)

func main() {
	app := quix.New()

	svc := service.NewTaskService()
	pb.RegisterTaskServiceHTTPService(app.Group(""), svc)

	// curl -X POST http://localhost:8080/api/v1/tasks -d '{"title":"Buy milk"}'
	// curl http://localhost:8080/api/v1/tasks/1
	// curl -X DELETE http://localhost:8080/api/v1/tasks/1
	app.Run()
}
