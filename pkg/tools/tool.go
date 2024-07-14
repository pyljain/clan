package tools

import (
	"clan/pkg/llm"
	"clan/pkg/planning"
)

type Tool interface {
	Name() string
	Schema() llm.Tool
	Execute(map[string]interface{}) (string, error)
}

type StarlarkTool struct {
	Name        string                  `yaml:"name"`
	Description string                  `yaml:"description"`
	Parameters  []StarlarkToolParameter `yaml:"parameters"`
	Function    string                  `yaml:"function"`
}

type StarlarkToolParameter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
}

// var AllTools = []Tool{NewReader(), NewWriter(), NewRunner(), NewNextAgentSelector(), planning.NewCreatePlan(), planning.NewUpdatePlan(), planning.NewGetPlan()}

func AllTools(starlarkToolDefs []StarlarkTool) []Tool {
	baseTools := []Tool{NewReader(), NewWriter(), NewRunner(), NewNextAgentSelector(), planning.NewCreatePlan(), planning.NewUpdatePlan(), planning.NewGetPlan()}
	for _, def := range starlarkToolDefs {
		baseTools = append(baseTools, NewStarlarkHandler(&def))
	}

	return baseTools
}
