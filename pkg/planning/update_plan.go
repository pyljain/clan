package planning

import "clan/pkg/llm"

type UpdatePlan struct {
	CurrentPlan []Task
}

func NewUpdatePlan() *UpdatePlan {
	return &UpdatePlan{}
}

func (cp *UpdatePlan) Name() string {
	return "PlanUpdater"
}

func (cp *UpdatePlan) Schema() llm.Tool {
	return llm.Tool{
		Name:        cp.Name(),
		Description: "Use this tool to update a task in the plan",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"taskName": map[string]interface{}{
					"type":        "string",
					"description": "Exact name of the task to update",
				},
				"status": map[string]interface{}{
					"type":        "string",
					"description": "Status to update - can be In Progress, Cancelled, Completed, Error",
				},
			},
		},
	}
}

func (cp *UpdatePlan) Execute(input map[string]interface{}) (string, error) {
	return "Tasks updated", nil
}

// 	Execute(map[string]interface{}) (string, error)
