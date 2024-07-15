package workflow

import "clan/pkg/tools"

type WorkflowDefinition struct {
	Name           string                `yaml:"name"`
	Type           string                `yaml:"type"`
	Goal           string                `yaml:"goal"`
	Description    string                `yaml:"description"`
	StartAgent     string                `yaml:"start_agent"`
	Agents         []AgentDefinition     `yaml:"agents"`
	Tools          []tools.StarlarkTool  `yaml:"tools"`
	TraversalDepth int                   `yaml:"traversal_depth"`
	Checkpoint     *CheckpointDefinition `yaml:"checkpoint"`
}

type AgentDefinition struct {
	Name              string   `yaml:"name"`
	SystemPrompt      string   `yaml:"system_prompt"`
	Purpose           string   `yaml:"purpose"`
	Temperature       float32  `yaml:"temperature"`
	Model             string   `yaml:"model"`
	NextAgent         string   `yaml:"next_agent"`
	NextAgentFunction string   `yaml:"next_agent_function"`
	AvailableTools    []string `yaml:"available_tools"`
}

type CheckpointDefinition struct {
	Type             string `yaml:"type"`
	ConnectionString string `yaml:"connection_string"`
}
