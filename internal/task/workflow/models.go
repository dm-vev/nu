package workflow

// Workflow represents a sequence of steps to be executed.
type Workflow struct {
	Name        string
	Tasks       []WorkflowTask
	FinalTaskID string
}

// Task represents a single task in a workflow (renamed from Step)
type WorkflowTask struct {
	ID           string
	AgentID      string
	Dependencies []string
	// Other fields as needed
}

// Step represents a single step in a workflow
type WorkflowStep struct {
	WorkflowAgent       *WorkflowAgent
	Name                string
	Description         string
	Input               string
	Output              string
	NextStep            string
	HandoffInstructions string
}

// Agent represents an agent that can perform steps in a workflow
type WorkflowAgent struct {
	ID           string
	SystemPrompt string
}

// NewWorkflow creates a workflow with the given name.
func NewWorkflow(name string) *Workflow {
	return &Workflow{
		Name:  name,
		Tasks: []WorkflowTask{},
	}
}

// NewAgent creates a new agent with the given ID and system prompt
func NewWorkflowAgent(id string, systemPrompt string) *WorkflowAgent {
	return &WorkflowAgent{
		ID:           id,
		SystemPrompt: systemPrompt,
	}
}

// AddStep adds a step to the workflow and converts it to a task
func (w *Workflow) AddStep(step WorkflowStep) {
	task := WorkflowTask{
		ID:      step.Name,
		AgentID: step.WorkflowAgent.ID,
	}

	// If this step has a next step, add it as a dependency for the next task
	if step.NextStep != "" {
		// Find or create the next task
		nextTaskExists := false
		for i, t := range w.Tasks {
			if t.ID == step.NextStep {
				w.Tasks[i].Dependencies = append(w.Tasks[i].Dependencies, step.Name)
				nextTaskExists = true
				break
			}
		}

		if !nextTaskExists {
			// Create a placeholder for the next task
			nextTask := WorkflowTask{
				ID:           step.NextStep,
				Dependencies: []string{step.Name},
			}
			w.Tasks = append(w.Tasks, nextTask)
		}
	} else {
		// If this step has no next step, it's the final step
		w.FinalTaskID = step.Name
	}

	w.Tasks = append(w.Tasks, task)
}
