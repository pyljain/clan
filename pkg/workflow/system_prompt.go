package workflow

import (
	"bytes"
	"text/template"
)

func generateSystemPrompt(sp string, workflowDef *WorkflowDefinition) (string, error) {
	tmpl := template.New("systemPrompt.tmpl")
	parsedTemplate, err := tmpl.Parse(sp)
	if err != nil {
		return "", err
	}

	buff := bytes.NewBuffer([]byte{})
	err = parsedTemplate.Execute(buff, workflowDef)
	if err != nil {
		return "", err
	}

	return buff.String(), nil
}
