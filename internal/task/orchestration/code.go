package orchestration

import (
	"context"
	"fmt"
	"sync"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	// TaskPending indicates the task is pending
	TaskPending TaskStatus = "pending"

	// TaskRunning indicates the task is running
	TaskRunning TaskStatus = "running"

	// TaskCompleted indicates the task is completed
	TaskCompleted TaskStatus = "completed"

	// TaskFailed indicates the task failed
	TaskFailed TaskStatus = "failed"
)

// Task represents a task to be executed by an agent
type Task struct {
	// ID is the unique identifier for the task
	ID string

	// AgentID is the ID of the agent to execute the task
	AgentID string

	// Input is the input to provide to the agent
	Input string

	// Dependencies are the IDs of tasks that must complete before this one
	Dependencies []string

	// Status is the current status of the task
	Status TaskStatus

	// Result is the result of the task
	Result string

	// Error is any error that occurred during execution
	Error error
}

// Workflow represents a workflow of tasks
type Workflow struct {
	// Tasks is the list of tasks in the workflow
	Tasks []*Task

	// Results is a map of task IDs to results
	Results map[string]string

	// Errors is a map of task IDs to errors
	Errors map[string]error

	// FinalTaskID is the ID of the task that produces the final result
	FinalTaskID string
}

// NewWorkflow creates a new workflow
func NewWorkflow() *Workflow {
	return &Workflow{
		Tasks:   make([]*Task, 0),
		Results: make(map[string]string),
		Errors:  make(map[string]error),
	}
}

// AddTask adds a task to the workflow
func (w *Workflow) AddTask(id string, agentID string, input string, dependencies []string) {
	task := &Task{
		ID:           id,
		AgentID:      agentID,
		Input:        input,
		Dependencies: dependencies,
		Status:       TaskPending,
	}

	w.Tasks = append(w.Tasks, task)
}

// SetFinalTask sets the final task
func (w *Workflow) SetFinalTask(id string) {
	w.FinalTaskID = id
}

// CodeOrchestrator orchestrates agents using code-defined workflows
type Code struct {
	registry *AgentRegistry
}

// NewCodeOrchestrator creates a new code orchestrator
func NewCode(registry *AgentRegistry) *Code {
	return &Code{
		registry: registry,
	}
}

// ExecuteWorkflow executes a workflow
func (o *Code) ExecuteWorkflow(ctx context.Context, workflow *Workflow) (string, error) {
	// Create a wait group to wait for all tasks
	var wg sync.WaitGroup

	// Create a channel to signal task completion
	taskCompletionCh := make(chan string)

	// Create a map to track completed tasks
	completedTasks := make(map[string]bool)
	var completedTasksMu sync.Mutex

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start a goroutine to monitor task completion
	go func() {
		for {
			select {
			case taskID := <-taskCompletionCh:
				// Mark task as completed
				completedTasksMu.Lock()
				completedTasks[taskID] = true
				completedTasksMu.Unlock()

				// Check if all tasks are completed
				allCompleted := true
				for _, task := range workflow.Tasks {
					if task.Status != TaskCompleted && task.Status != TaskFailed {
						allCompleted = false
						break
					}
				}

				if allCompleted {
					// All tasks are completed, cancel the context
					cancel()
					return
				}

				// Check if any tasks can now be executed
				for _, task := range workflow.Tasks {
					if task.Status == TaskPending {
						// Check if all dependencies are completed
						allDepsCompleted := true
						for _, depID := range task.Dependencies {
							if !completedTasks[depID] {
								allDepsCompleted = false
								break
							}
						}

						if allDepsCompleted {
							// All dependencies are completed, execute the task
							wg.Add(1)
							go o.executeTask(ctx, task, workflow, &wg, taskCompletionCh)
						}
					}
				}
			case <-ctx.Done():
				// Context is cancelled, exit
				return
			}
		}
	}()

	// Start tasks with no dependencies
	for _, task := range workflow.Tasks {
		if len(task.Dependencies) == 0 {
			wg.Add(1)
			go o.executeTask(ctx, task, workflow, &wg, taskCompletionCh)
		}
	}

	// Wait for all tasks to complete
	wg.Wait()

	// Check if the final task completed successfully
	if workflow.FinalTaskID != "" {
		if err, ok := workflow.Errors[workflow.FinalTaskID]; ok {
			return "", fmt.Errorf("final task failed: %w", err)
		}

		if result, ok := workflow.Results[workflow.FinalTaskID]; ok {
			return result, nil
		}

		return "", fmt.Errorf("final task result not found")
	}

	// No final task specified, return an empty string
	return "", nil
}

// executeTask executes a task
func (o *Code) executeTask(ctx context.Context, task *Task, workflow *Workflow, wg *sync.WaitGroup, completionCh chan<- string) {
	defer wg.Done()

	// Update task status
	task.Status = TaskRunning

	// Get the agent
	agent, ok := o.registry.Get(task.AgentID)
	if !ok {
		task.Status = TaskFailed
		task.Error = fmt.Errorf("agent not found: %s", task.AgentID)
		workflow.Errors[task.ID] = task.Error
		completionCh <- task.ID
		return
	}

	// Prepare input with results from dependencies
	input := task.Input
	for _, depID := range task.Dependencies {
		if result, ok := workflow.Results[depID]; ok {
			input = fmt.Sprintf("%s\n\nResult from %s: %s", input, depID, result)
		}
	}

	// Execute the agent
	result, err := agent.Run(ctx, input)
	if err != nil {
		task.Status = TaskFailed
		task.Error = fmt.Errorf("agent execution failed: %w", err)
		workflow.Errors[task.ID] = task.Error
		completionCh <- task.ID
		return
	}

	// Update task status and result
	task.Status = TaskCompleted
	task.Result = result
	workflow.Results[task.ID] = result

	// Signal task completion
	completionCh <- task.ID
}
