package planning

import "clan/pkg/llm"

type GetPlan struct {
}

func (plan *GetPlan) Name() string {
	return "GetPlan"
}

func NewGetPlan() *GetPlan {
	return &GetPlan{}
}

func (plan *GetPlan) Schema() llm.Tool {
	return llm.Tool{
		Name:        plan.Name(),
		Description: "Use this tool to get the generated plan for the misison",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"fullPlan": map[string]interface{}{
					"type":        "boolean",
					"description": "Should retrieve full plan",
				},
			},
		},
	}
}

func (plan *GetPlan) Execute(map[string]interface{}) (string, error) {
	return "", nil
}

// Name() string
// Schema() llm.Tool
// Execute(map[string]interface{}) (string, error)
// }
