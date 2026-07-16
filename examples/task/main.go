package main

import (
	"context"
	"fmt"
	"log"

	"nu/internal/task"
	"nu/internal/task/service"
	"nu/internal/telemetry"
)

func main() {
	ctx := context.Background()
	tasks := service.NewService(telemetry.NewLogger())
	created, err := tasks.CreateTask(ctx, task.CoreCreateTaskRequest{
		Name: "Prepare release notes", Description: "Summarize changes and verify examples.",
	})
	if err != nil {
		log.Fatal(err)
	}
	createdTask := created.(*task.CoreTask)

	planner := &task.MockAIPlanner{}
	executor := task.NewExecutor()
	executor.RegisterTask("plan", func(ctx context.Context, params interface{}) (interface{}, error) {
		return planner.GeneratePlan(ctx, params.(*task.CoreTask))
	})

	result, err := executor.ExecuteSync(ctx, "plan", createdTask, nil)
	if err == nil {
		err = result.Error
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Data)
}
