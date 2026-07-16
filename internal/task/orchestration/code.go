package orchestration

import (
	"context"
	"fmt"
	"sync"
)

// TaskStatus represents the status of a task
type OrchestratorTaskStatus string

const (
	// TaskPending indicates the task is pending
	OrchestratorTaskPending OrchestratorTaskStatus = "pending"

	// TaskRunning indicates the task is running
	OrchestratorTaskRunning OrchestratorTaskStatus = "running"

	// TaskCompleted indicates the task is completed
	OrchestratorTaskCompleted OrchestratorTaskStatus = "completed"

	// TaskFailed indicates the task failed
	OrchestratorTaskFailed OrchestratorTaskStatus = "failed"
)

// Task represents a task to be executed by an agent
type OrchestratorTask struct {
	// ID is the unique identifier for the task
	ID string

	// AgentID is the ID of the agent to execute the task
	AgentID string

	// Input is the input to provide to the agent
	Input string

	// Dependencies are the IDs of tasks that must complete before this one
	Dependencies []string

	// Status is the current status of the task
	Status OrchestratorTaskStatus

	// Result is the result of the task
	Result string

	// Error is any error that occurred during execution
	Error error
}

// Workflow represents a workflow of tasks
type OrchestratorWorkflow struct {
	// Tasks is the list of tasks in the workflow
	Tasks []*OrchestratorTask

	// Results is a map of task IDs to results
	Results map[string]string

	// Errors is a map of task IDs to errors
	Errors map[string]error

	// FinalTaskID is the ID of the task that produces the final result
	FinalTaskID string
}

// NewWorkflow creates a new workflow
func NewOrchestratorWorkflow() *OrchestratorWorkflow {
	return &OrchestratorWorkflow{
		Tasks:   make([]*OrchestratorTask, 0),
		Results: make(map[string]string),
		Errors:  make(map[string]error),
	}
}

// AddTask adds a task to the workflow
func (w *OrchestratorWorkflow) AddTask(id string, agentID string, input string, dependencies []string) {
	task := &OrchestratorTask{
		ID:           id,
		AgentID:      agentID,
		Input:        input,
		Dependencies: dependencies,
		Status:       OrchestratorTaskPending,
	}

	w.Tasks = append(w.Tasks, task)
}

// SetFinalTask sets the final task
func (w *OrchestratorWorkflow) SetFinalTask(id string) {
	w.FinalTaskID = id
}

// CodeOrchestrator orchestrates agents using code-defined workflows
type OrchestratorCode struct {
	registry *OrchestratorAgentRegistry
}

// NewCodeOrchestrator creates a new code orchestrator
func NewOrchestratorCode(registry *OrchestratorAgentRegistry) *OrchestratorCode {
	return &OrchestratorCode{
		registry: registry,
	}
}

// ExecuteWorkflow executes a workflow
func (o *OrchestratorCode) ExecuteWorkflow(ctx context.Context, workflow *OrchestratorWorkflow) (string, error) {
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
					if task.Status != OrchestratorTaskCompleted && task.Status != OrchestratorTaskFailed {
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
					if task.Status == OrchestratorTaskPending {
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
func (o *OrchestratorCode) executeTask(ctx context.Context, task *OrchestratorTask, workflow *OrchestratorWorkflow, wg *sync.WaitGroup, completionCh chan<- string) {
	defer wg.Done()

	// Update task status
	task.Status = OrchestratorTaskRunning

	// Get the agent
	agent, ok := o.registry.Get(task.AgentID)
	if !ok {
		task.Status = OrchestratorTaskFailed
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
		task.Status = OrchestratorTaskFailed
		task.Error = fmt.Errorf("agent execution failed: %w", err)
		workflow.Errors[task.ID] = task.Error
		completionCh <- task.ID
		return
	}

	// Update task status and result
	task.Status = OrchestratorTaskCompleted
	task.Result = result
	workflow.Results[task.ID] = result

	// Signal task completion
	completionCh <- task.ID
}
