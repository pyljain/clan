package tools

import (
	"clan/pkg/llm"
	"os/exec"
	"strings"
)

type runner struct{}

func NewRunner() Tool {
	return &runner{}
}

func (r *runner) Name() string {
	return "CommandRunner"
}

func (r *runner) Schema() llm.Tool {
	return llm.Tool{
		Name:        r.Name(),
		Description: "Run commands such as listing files, installing software, running a program",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "Command to run",
				},
			},
		},
	}
}

func (r *runner) Execute(params map[string]interface{}) (string, error) {
	cmd := params["command"].(string)
	cmdArgs := strings.Split(cmd, " ")
	command := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	command.Dir = "./workspace"

	output, err := command.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
