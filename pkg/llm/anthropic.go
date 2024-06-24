package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type Anthropic struct {
	apiKey  string
	baseURL string
	model   string
	tools   []Tool
}

type AnthropicOptions struct {
	APIKey  string
	BaseURL string
	Model   string
	Tools   []Tool
}

func NewAnthropic(options *AnthropicOptions) *Anthropic {

	if options.APIKey == "" {
		options.APIKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	if options.BaseURL == "" {
		options.BaseURL = "https://api.anthropic.com/v1"
	}

	if options.Model == "" {
		options.Model = "claude-3-haiku-20240307"
	}

	if options.Tools == nil {
		options.Tools = []Tool{}
	}

	return &Anthropic{
		apiKey:  options.APIKey,
		baseURL: options.BaseURL,
		model:   options.Model,
		tools:   options.Tools,
	}
}

func (a *Anthropic) Generate(messages []Message) ([]Message, error) {

	systemMessage := ""
	var cleansedMessages []Message
	for _, sm := range messages {
		if sm.Role == "system" {
			systemMessage += sm.Content[0].Text
			continue
		}

		cleansedMessages = append(cleansedMessages, sm)
	}

	rb := anthropicReqBody{
		MaxTokens: 4096,
		Model:     a.model,
		Messages:  cleansedMessages,
		System:    systemMessage,
		Tools:     a.tools,
	}

	rbBytes, err := json.Marshal(rb)
	if err != nil {
		return nil, err
	}

	bufferedReq := bytes.NewBuffer(rbBytes)
	log.Printf("Request to Anthropic is: %+v", string(rbBytes))

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/messages", a.baseURL), bufferedReq)
	if err != nil {
		return nil, err
	}

	req.Header.Add("x-api-key", a.apiKey)
	req.Header.Add("anthropic-version", "2023-06-01")
	req.Header.Add("content-type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("Response from Anthropic %+v", string(respBytes))

	ar := anthropicResponse{}
	err = json.Unmarshal(respBytes, &ar)
	if err != nil {
		return nil, err
	}

	var result []Message
	result = append(result, Message{
		Role:    "assistant",
		Content: ar.Content,
	})

	return result, nil
}

type anthropicReqBody struct {
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
	Model     string    `json:"model"`
	System    string    `json:"system"`
	Tools     []Tool    `json:"tools"`
}

type anthropicResponse struct {
	Content []Content      `json:"content"`
	Id      string         `json:"id"`
	Model   string         `json:"model"`
	Usage   map[string]int `json:"usage"`
}

// type anthropicResponseContent struct {
// 	Text        string                 `json:"text"`
// 	ContentType string                 `json:"type"`
// 	Id          string                 `json:"id"`
// 	Name        string                 `json:"name"`
// 	Input       map[string]interface{} `json:"input"`
// }
