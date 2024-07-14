package planning

import (
	"clan/pkg/llm"
)

type CreatePlan struct {
	CurrentPlan []Task
}

func NewCreatePlan() *CreatePlan {
	return &CreatePlan{}
}

func (cp *CreatePlan) Name() string {
	return "PlanCreator"
}

func (cp *CreatePlan) Schema() llm.Tool {
	return llm.Tool{
		Name:        cp.Name(),
		Description: "Use this tool to create a plan for a mission",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tasks": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name": map[string]interface{}{
								"type":        "string",
								"description": "Short name of the task",
							},
							"description": map[string]interface{}{
								"type":        "string",
								"description": "Description of the task",
							},
							"owner": map[string]interface{}{
								"type":        "string",
								"description": "Name of the agent who should complete the task",
							},
						},
						"required": []string{"name", "description", "owner"},
					},
				},
			},
		},
	}
}

func (cp *CreatePlan) Execute(input map[string]interface{}) (string, error) {
	cp.CurrentPlan = []Task{}

	for _, t := range input["tasks"].([]interface{}) {
		tt := t.(map[string]interface{})
		task := Task{
			Name:        tt["name"].(string),
			Description: tt["description"].(string),
			Owner:       tt["owner"].(string),
			Status:      "Not Started",
		}
		cp.CurrentPlan = append(cp.CurrentPlan, task)
	}

	return "Tasks updated", nil
}

// 	Execute(map[string]interface{}) (string, error)
