package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/fztcjjl/quix/examples/proto-demo/gen/errors"
	pb "github.com/fztcjjl/quix/examples/proto-demo/gen/task/v1"
)

// TaskService implements pb.TaskServiceHTTPService.
type TaskService struct {
	mu    sync.RWMutex
	tasks map[string]*pb.Task
	next  atomic.Int64
}

// NewTaskService creates a new TaskService.
func NewTaskService() *TaskService {
	return &TaskService{
		tasks: make(map[string]*pb.Task),
	}
}

func (s *TaskService) CreateTask(_ context.Context, req *pb.CreateTaskRequest) (*pb.Task, error) {
	if req.GetTitle() == "" {
		return nil, errors.TaskTitleRequired()
	}

	id := fmt.Sprintf("%d", s.next.Add(1))
	task := &pb.Task{
		Id:     id,
		Title:  req.GetTitle(),
		Status: pb.TaskStatus_TASK_STATUS_TODO,
	}

	s.mu.Lock()
	s.tasks[id] = task
	s.mu.Unlock()

	return task, nil
}

func (s *TaskService) GetTask(_ context.Context, req *pb.GetTaskRequest) (*pb.Task, error) {
	s.mu.RLock()
	task, ok := s.tasks[req.GetTaskId()]
	s.mu.RUnlock()

	if !ok {
		return nil, errors.TaskNotFound()
	}
	return task, nil
}

func (s *TaskService) ListTasks(_ context.Context, _ *pb.ListTasksRequest) (*pb.TaskList, error) {
	s.mu.RLock()
	tasks := make([]*pb.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		tasks = append(tasks, t)
	}
	s.mu.RUnlock()

	return &pb.TaskList{Tasks: tasks}, nil
}

func (s *TaskService) DeleteTask(_ context.Context, req *pb.DeleteTaskRequest) (*pb.Task, error) {
	s.mu.Lock()
	task, ok := s.tasks[req.GetTaskId()]
	if !ok {
		s.mu.Unlock()
		return nil, errors.TaskNotFound()
	}
	if task.GetStatus() == pb.TaskStatus_TASK_STATUS_DONE {
		s.mu.Unlock()
		return nil, errors.TaskAlreadyDone()
	}
	delete(s.tasks, req.GetTaskId())
	s.mu.Unlock()

	return task, nil
}
