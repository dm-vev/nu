# Task Execution Package

The task execution package provides functionality for executing tasks synchronously and asynchronously, including API calls and Temporal workflows.

## Features

- Execute tasks synchronously and asynchronously
- Built-in retry mechanism with configurable retry policies
- API client for making HTTP requests
- Temporal workflow integration
- Task cancellation and status tracking
- Task adapter pattern for integrating with agent-specific models

## Usage

### Basic Task Execution

```go
// Create a task executor
executor := task.NewExecutor()

// Register a task
executor.RegisterTask("hello", func(ctx context.Context, params interface{}) (interface{}, error) {
    name, ok := params.(string)
    if !ok {
        name = "World"
    }
    return fmt.Sprintf("Hello, %s!", name), nil
})

// Execute the task synchronously
result, err := executor.ExecuteSync(context.Background(), "hello", "John", nil)
if err != nil {
    fmt.Printf("Error: %v\n", err)
} else {
    fmt.Printf("Result: %v\n", result.Data)
}

// Execute the task asynchronously
resultChan, err := executor.ExecuteAsync(context.Background(), "hello", "Jane", nil)
if err != nil {
    fmt.Printf("Error: %v\n", err)
} else {
    result := <-resultChan
    fmt.Printf("Result: %v\n", result.Data)
}
```

### API Task Execution

```go
// Create an API client
apiClient := service.NewAPIClient("https://api.example.com", 10*time.Second)

// Register an API task
apiTask := service.NewAPITask(apiClient)
executor.RegisterTask("get_data", apiTask.Task(service.APIRequest{
    Method: "GET",
    Path:   "/data",
    Query:  map[string]string{"limit": "10"},
}))

// Execute the API task with retry policy
timeout := 5 * time.Second
retryPolicy := &contracts.RetryPolicy{
    MaxRetries:        3,
    InitialBackoff:    100 * time.Millisecond,
    MaxBackoff:        1 * time.Second,
    BackoffMultiplier: 2.0,
}

result, err := executor.ExecuteSync(context.Background(), "get_data", nil, &contracts.TaskOptions{
    Timeout:     &timeout,
    RetryPolicy: retryPolicy,
})
```

### Using the Task Adapter Pattern

The task adapter pattern allows you to use your own agent-specific models while still leveraging the SDK's task management functionality. This pattern separates the concerns of the SDK from your agent-specific implementations.

#### Default Implementation

The SDK provides a default implementation of the task models and adapter that you can use directly:

```go
package main

import (
    "context"
    "fmt"

    "nu/internal/task/service"
    "nu/internal/telemetry"
)

func main() {
    ctx := context.Background()
    logger := telemetry.NewLogger()

    taskService, err := service.NewAgentTaskService(logger)
    if err != nil {
        panic(err)
    }

    // Create a task using the default models
    newTask, err := taskService.CreateTask(
        ctx, "Service Deployment", "Deploy a new service", "user123", nil,
    )

    if err != nil {
        panic(err)
    }

    fmt.Printf("Created task: %s\n", newTask.ID)
}
```

#### Custom Implementation

Alternatively, you can create your own models and adapter:

```go
// Define your agent-specific task models
type MyTask struct {
    ID          string
    Name        string
    Status      string
    CreatedAt   time.Time
    CompletedAt *time.Time
}

type MyCreateRequest struct {
    Name   string
    UserID string
}

type MyApprovalRequest struct {
    Approved bool
    Comment  string
}

type MyTaskUpdate struct {
    Type   string
    ID     string
    Status string
}

// Implement the Adapter interface
type MyTaskAdapter struct {
    logger telemetry.Logger
}

// Create a new adapter
func NewMyTaskAdapter(logger telemetry.Logger) service.Adapter[MyTask, MyCreateRequest, MyApprovalRequest, MyTaskUpdate] {
    return &MyTaskAdapter{
        logger: logger,
    }
}

// Implement conversion methods
func (a *MyTaskAdapter) ConvertCreateRequest(req MyCreateRequest) task.CreateTaskRequest {
    return task.CreateTaskRequest{
        Description: req.Name,
        UserID:      req.UserID,
        Metadata:    make(map[string]interface{}),
    }
}

func (a *MyTaskAdapter) ConvertApproveRequest(req MyApprovalRequest) task.ApproveTaskPlanRequest {
    return task.ApproveTaskPlanRequest{
        Approved: req.Approved,
        Feedback: req.Comment,
    }
}

func (a *MyTaskAdapter) ConvertTaskUpdates(updates []MyTaskUpdate) []task.Update {
    sdkUpdates := make([]task.Update, len(updates))
    for i, update := range updates {
        sdkUpdates[i] = task.Update{
            Type:   update.Type,
            StepID: update.ID,
            Status: update.Status,
        }
    }
    return sdkUpdates
}

func (a *MyTaskAdapter) ConvertTask(sdkTask *task.Task) MyTask {
    if sdkTask == nil {
        return MyTask{}
    }

    return MyTask{
        ID:          sdkTask.ID,
        Name:        sdkTask.Description,
        Status:      string(sdkTask.Status),
        CreatedAt:   sdkTask.CreatedAt,
        CompletedAt: sdkTask.CompletedAt,
    }
}

func (a *MyTaskAdapter) ConvertTasks(sdkTasks []*task.Task) []MyTask {
    tasks := make([]MyTask, len(sdkTasks))
    for i, sdkTask := range sdkTasks {
        tasks[i] = a.ConvertTask(sdkTask)
    }
    return tasks
}

// Use the adapter with a task service
type MyTaskService struct {
    sdkService service.Service
    adapter    service.Adapter[MyTask, MyCreateRequest, MyApprovalRequest, MyTaskUpdate]
}

func NewMyTaskService(sdkService service.Service, adapter service.Adapter[MyTask, MyCreateRequest, MyApprovalRequest, MyTaskUpdate]) *MyTaskService {
    return &MyTaskService{
        sdkService: sdkService,
        adapter:    adapter,
    }
}

// Create a task using your own models
func (s *MyTaskService) CreateTask(ctx context.Context, req MyCreateRequest) (MyTask, error) {
    // Convert to SDK request
    sdkReq := s.adapter.ConvertCreateRequest(req)

    // Create task using SDK service
    sdkTask, err := s.sdkService.CreateTask(ctx, sdkReq)
    if err != nil {
        return MyTask{}, err
    }

    // Convert back to your model
    return s.adapter.ConvertTask(sdkTask), nil
}
```

### Using the Agent Task Service

The SDK provides an `AgentTaskService` that you can use to work with your agent-specific models. It wraps the adapter service and provides a simpler interface:

```go
// Create the agent task service
taskService := service.NewAgentTaskService(
    ctx,
    logger,
    sdkTaskService,
    myAdapter,
)

// Use the service with your agent-specific models
myTask, err := taskService.CreateTask(ctx, myCreateRequest)
if err != nil {
    // Handle error
}

// Get a task
myTask, err = taskService.GetTask(ctx, "task-123")
if err != nil {
    // Handle error
}

// List tasks for a user
myTasks, err := taskService.ListTasks(ctx, "user-456")
if err != nil {
    // Handle error
}

// Approve a task plan
myTask, err = taskService.ApproveTaskPlan(ctx, "task-123", myApproveRequest)
if err != nil {
    // Handle error
}

// Update a task
myTask, err = taskService.UpdateTask(ctx, "task-123", "conversation-789", myTaskUpdates)
if err != nil {
    // Handle error
}

// Add a log entry to a task
err = taskService.AddTaskLog(ctx, "task-123", "Starting deployment", "info")
if err != nil {
    // Handle error
}
```

### Temporal Workflow Execution

```go
// Create a Temporal client
temporalClient := task.NewTemporalClient(task.TemporalConfig{
    HostPort:                 "localhost:7233",
    Namespace:                "default",
    TaskQueue:                "example",
    WorkflowIDPrefix:         "example-",
    WorkflowExecutionTimeout: 10 * time.Minute,
    WorkflowRunTimeout:       5 * time.Minute,
    WorkflowTaskTimeout:      10 * time.Second,
})

// Register a Temporal workflow task
executor.RegisterTask("example_workflow", task.TemporalWorkflowTask(temporalClient, "ExampleWorkflow"))

// Execute the Temporal workflow task
result, err := executor.ExecuteSync(context.Background(), "example_workflow", map[string]interface{}{
    "input": "example input",
}, nil)
```

## Task Options

You can configure task execution with the following options:

```go
options := &contracts.TaskOptions{
    // Timeout specifies the maximum duration for task execution
    Timeout: &timeout,

    // RetryPolicy specifies the retry policy for the task
    RetryPolicy: &contracts.RetryPolicy{
        MaxRetries:        3,
        InitialBackoff:    100 * time.Millisecond,
        MaxBackoff:        1 * time.Second,
        BackoffMultiplier: 2.0,
    },

    // Metadata contains additional information for the task execution
    Metadata: map[string]interface{}{
        "purpose": "example",
    },
}
```

## Task Result

The task result contains the following information:

```go
type TaskResult struct {
    // Data contains the result data
    Data interface{}

    // Error contains any error that occurred during task execution
    Error error

    // Metadata contains additional information about the task execution
    Metadata map[string]interface{}
}
```

# Task Package

The task package provides comprehensive task management functionality for agents. It includes models, interfaces, and services for creating, retrieving, and managing tasks.

## Concepts

A task represents a unit of work that an agent needs to perform. Tasks have a lifecycle, starting from creation, through planning, execution, and finally completion. Each task can have multiple steps, which are executed sequentially.

## Core Components

### Task Model

The Task struct represents a task in the system:

```go
type Task struct {
	ID             string                 // Unique identifier
	Description    string                 // Task description
	Status         Status                 // Current status (pending, planning, awaiting_approval, executing, completed, failed)
	Title          string                 // Task title
	TaskKind       string                 // Type of task
	ConversationID string                 // Associated conversation ID
	Plan           *Plan                  // Execution plan
	Steps          []Step                 // Task steps
	CreatedAt      time.Time              // Creation timestamp
	UpdatedAt      time.Time              // Last update timestamp
	StartedAt      *time.Time             // When execution started
	CompletedAt    *time.Time             // When execution completed
	UserID         string                 // Owner user ID
	Logs           []LogEntry             // Activity logs
	Requirements   interface{}            // Task requirements
	Feedback       string                 // User feedback
	Metadata       map[string]interface{} // Additional metadata
}
```

### Task Service

The `Service` interface defines methods for interacting with tasks:

```go
type Service interface {
	CreateTask(ctx context.Context, req CreateTaskRequest) (*Task, error)
	GetTask(ctx context.Context, taskID string) (*Task, error)
	ListTasks(ctx context.Context, filter Filter) ([]*Task, error)
	ApproveTaskPlan(ctx context.Context, taskID string, req ApproveTaskPlanRequest) (*Task, error)
	UpdateTask(ctx context.Context, taskID string, updates []Update) (*Task, error)
	AddTaskLog(ctx context.Context, taskID string, message string, level string) error
}
```

## Task Adapter Pattern

The task package supports the adapter pattern to allow agents to work with their own domain-specific task models while leveraging the SDK's task management capabilities.

### File Organization

Canonical/core and legacy task models, executors, planners, and shared
contracts/options live in `internal/task`. API support, in-memory/core services,
adapters, and compatibility conversion live in `internal/task/service`.
Workflow models and execution live in `internal/task/workflow`; LLM/code/workflow
orchestrators, handoffs, registries, and routers live in
`internal/task/orchestration`. The root does not import or re-export its children.

Canonical models use the `Core` prefix, API transport types use `API`, workflow
types use `Workflow`, and orchestration/handoff types use `Orchestrator` or
`Handoff`. The legacy agent-task model keeps `Task`, `Plan`, `Step`, and
`Status`.

### Adapter Interface

The `Adapter` is a generic interface that allows you to define custom conversion methods between SDK and agent-specific models:

```go
type Adapter[AgentTask any, AgentCreateRequest any, AgentApprovalRequest any, AgentTaskUpdate any] interface {
	// ToSDK conversions (Agent -> SDK)
	ConvertCreateRequest(req AgentCreateRequest) CreateTaskRequest
	ConvertApproveRequest(req AgentApprovalRequest) ApproveTaskPlanRequest
	ConvertTaskUpdates(updates []AgentTaskUpdate) []Update

	// FromSDK conversions (SDK -> Agent)
	ConvertTask(sdkTask *Task) AgentTask
	ConvertTasks(sdkTasks []*Task) []AgentTask
}
```
