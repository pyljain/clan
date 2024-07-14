package tools

import (
	"clan/pkg/llm"
)

type nextAgentSelector struct{}

func NewNextAgentSelector() Tool {
	return &nextAgentSelector{}
}

func (nas *nextAgentSelector) Name() string {
	return "NextAgentSelector"
}

func (nas *nextAgentSelector) Schema() llm.Tool {
	return llm.Tool{
		Name:        nas.Name(),
		Description: "Call this function to handover to the next agent. Please specify an agent name when you want to handover to a specific agent, else pass an empty string.",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"summary": map[string]interface{}{
					"type":        "string",
					"description": "Summary of the what was done to complete the task and key highlights.",
				},
				"next_agent": map[string]interface{}{
					"type":        "string",
					"description": "Exact name of the next agent to handover to. Pass an empty string if you are not sure or have no specific agent to handover to.",
				},
			},
		},
	}
}

func (nas *nextAgentSelector) Execute(params map[string]interface{}) (string, error) {
	taskSummary := params["summary"].(string)
	return string(taskSummary), nil
}
