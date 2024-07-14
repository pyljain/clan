package tools

import (
	"clan/pkg/llm"
	"os"
	"path"
)

type reader struct{}

func NewReader() Tool {
	return &reader{}
}

func (r *reader) Name() string {
	return "Reader"
}

func (r *reader) Schema() llm.Tool {
	return llm.Tool{
		Name:        r.Name(),
		Description: "Read files in the filesystem for the filepath passed in as an argument",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"filepath": map[string]interface{}{
					"type":        "string",
					"description": "Location of the file in the filesystem",
				},
			},
		},
	}
}

func (r *reader) Execute(params map[string]interface{}) (string, error) {
	fp := params["filepath"].(string)
	fp = path.Join("./workspace", fp)
	fileBytes, err := os.ReadFile(fp)
	if err != nil {
		return "", err
	}

	return string(fileBytes), nil
}
