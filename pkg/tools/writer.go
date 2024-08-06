package tools

import (
	"clan/pkg/llm"
	"os"
	"path"
)

type writer struct{}

func NewWriter() Tool {
	return &writer{}
}

func (w *writer) Name() string {
	return "Writer"
}

func (w *writer) Schema() llm.Tool {
	return llm.Tool{
		Name:        w.Name(),
		Description: "Write files in the filesystem including modifying existing files",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"filepath": map[string]interface{}{
					"type":        "string",
					"description": "Location of the file in the filesystem",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Content to write into the file",
				},
			},
		},
	}
}

func (r *writer) Execute(params map[string]interface{}) (string, error) {
	fp := params["filepath"].(string)
	fp = path.Join("./workspace", fp)
	fileContent := params["content"].(string)

	err := os.WriteFile(fp, []byte(fileContent), os.ModePerm)
	if err != nil {
		return "", err
	}

	return "File written successfully", nil
}
