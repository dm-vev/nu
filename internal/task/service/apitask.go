package service

import (
	"context"
	"encoding/json"
	"fmt"

	. "nu/internal/task"
)

// APITask provides a way to execute tasks over an API.
type APITask struct {
	client *APIClient
}

// NewAPITask creates a new task API with the given client.
func NewAPITask(client *APIClient) *APITask {
	return &APITask{
		client: client,
	}
}

// Task returns a Func that executes a task via API
func (a *APITask) Task(request APIRequest) Func {
	return func(ctx context.Context, params interface{}) (interface{}, error) {
		// If params are provided, update the request body
		if params != nil {
			request.Body = params
		}

		// Execute the request
		resp, err := a.client.Do(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("failed to execute API task: %w", err)
		}

		// Check response status
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("API task failed with status %d: %s", resp.StatusCode, string(resp.Body))
		}

		return resp.Body, nil
	}
}

// RegisterWithExecutor registers API tasks with an executor
func (a *APITask) RegisterWithExecutor(exec *Executor, taskName string, request APIRequest) {
	exec.RegisterTask(taskName, a.Task(request))
}

// ExecuteTask executes a task via the API
func (a *APITask) ExecuteTask(ctx context.Context, taskID string) error {
	// Construct the request to execute a task
	req := APIRequest{
		Method: "POST",
		Path:   fmt.Sprintf("/tasks/%s/execute", taskID),
	}

	resp, err := a.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute task via API: %w", err)
	}

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("task execution failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	return nil
}

// GetTaskStatus gets the status of a task
func (a *APITask) GetTaskStatus(ctx context.Context, taskID string) (CoreStatus, error) {
	// Construct the request to get a task
	req := APIRequest{
		Method: "GET",
		Path:   fmt.Sprintf("/tasks/%s", taskID),
	}

	resp, err := a.client.Do(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to get task status via API: %w", err)
	}

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("get task status failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	// Parse the response to get the task status
	var taskResponse struct {
		Status CoreStatus `json:"status"`
	}
	if err := json.Unmarshal(resp.Body, &taskResponse); err != nil {
		return "", fmt.Errorf("failed to parse task status response: %w", err)
	}

	return taskResponse.Status, nil
}
