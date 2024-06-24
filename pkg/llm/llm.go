package llm

type LLM interface {
	Generate([]Message) (Message, error)
}

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type Content struct {
	Text        string                 `json:"text,omitempty"`
	ContentType string                 `json:"type,omitempty"`
	Id          string                 `json:"id,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Input       map[string]interface{} `json:"input,omitempty"`
	Content     string                 `json:"content,omitempty"`
	ToolUseId   string                 `json:"tool_use_id,omitempty"`
}

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"input_schema"`
}
