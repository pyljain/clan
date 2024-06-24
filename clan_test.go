package clan

import (
	"clan/pkg/llm"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClan(t *testing.T) {
	initialState := WorkflowStateProgrammer{
		ProgrammerMessages: []llm.Message{
			{
				Role: "user",
				Content: []llm.Content{
					{
						ContentType: "text",
						Text:        "Write a short Python program to print the first 10 prime numbers?",
					},
				},
			},
		},
		ReviewerMessages: []llm.Message{
			{
				Role: "system",
				Content: []llm.Content{
					{
						ContentType: "text",
						Text:        "You are an experienced programmer that has lots of experience providing feedback for code. Could you please review the following code?",
					},
				},
			},
		},
	}

	model := llm.NewAnthropic(&llm.AnthropicOptions{})

	graph := NewClanGraph(&initialState)
	graph.AddNode("Programmer", func(ws *WorkflowStateProgrammer) (*WorkflowStateProgrammer, error) {
		result, err := model.Generate(ws.ProgrammerMessages)
		if err != nil {
			return nil, err
		}

		ws.ProgrammerMessages = append(ws.ProgrammerMessages, result...)
		return ws, nil
	})

	graph.AddNode("Reviewer", func(ws *WorkflowStateProgrammer) (*WorkflowStateProgrammer, error) {
		lastMessage := ws.ProgrammerMessages[len(ws.ProgrammerMessages)-1]
		lastMessage.Role = "user"
		ws.ReviewerMessages = append(ws.ReviewerMessages, lastMessage)
		result, err := model.Generate(ws.ReviewerMessages)
		if err != nil {
			return nil, err
		}

		ws.ReviewerMessages = append(ws.ReviewerMessages, result...)

		log.Printf("Response from LLM %+v", result)

		return ws, nil
	})

	graph.AddEdge("Programmer", "Reviewer")
	graph.AddEdge("Reviewer", "End")

	err := graph.SetStartNode("Programmer")
	require.NoError(t, err)

	_, err = graph.Execute(ExecuteOptions{})
	require.NoError(t, err)
}

func TestClanTools(t *testing.T) {
	initialState := WorkflowStateWeather{
		WeatherManMessages: []llm.Message{
			{
				Role: "user",
				Content: []llm.Content{
					{
						ContentType: "text",
						Text:        "Can you tell me about the weather in London today?",
					},
				},
			},
		},
	}

	model := llm.NewAnthropic(&llm.AnthropicOptions{
		Tools: []llm.Tool{
			{
				Name:        "GetWeather",
				Description: "Fetches the weather for a location",
				Schema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "The city and state, e.g. San Francisco, CA",
						},
					},
				},
			},
		},
	})

	graph := NewClanGraph(&initialState)
	graph.AddNode("WeatherMan", func(ws *WorkflowStateWeather) (*WorkflowStateWeather, error) {
		result, err := model.Generate(ws.WeatherManMessages)
		if err != nil {
			return nil, err
		}

		ws.WeatherManMessages = append(ws.WeatherManMessages, result...)
		return ws, nil
	})

	graph.AddConditionalEdge("WeatherMan", func(ws *WorkflowStateWeather) (string, error) {
		// Are any tool use nodes present then goto Tools else goto End
		toolUsePresent := false
		for _, contentNode := range ws.WeatherManMessages[len(ws.WeatherManMessages)-1].Content {
			if contentNode.ContentType == "tool_use" {
				toolUsePresent = true
				break
			}
		}
		if toolUsePresent {
			return "Tools", nil
		}

		return "End", nil
	})

	graph.AddNode("Tools", func(ws *WorkflowStateWeather) (*WorkflowStateWeather, error) {
		for _, contentNode := range ws.WeatherManMessages[len(ws.WeatherManMessages)-1].Content {
			if contentNode.ContentType == "tool_use" {
				if contentNode.Name == "GetWeather" {
					ws.WeatherManMessages = append(ws.WeatherManMessages, llm.Message{
						Role: "user",
						Content: []llm.Content{
							{
								Content:     "The weather in London today is sunny with a high of 20 degrees",
								ContentType: "tool_result",
								ToolUseId:   contentNode.Id,
							},
						},
					})
				}
			}
		}

		return ws, nil
	})

	graph.AddEdge("Tools", "WeatherMan")

	err := graph.SetStartNode("WeatherMan")
	require.NoError(t, err)

	_, err = graph.Execute(ExecuteOptions{})
	require.NoError(t, err)
}

func TestNodeNotFoundError(t *testing.T) {
	initialState := WorkflowStateProgrammer{
		ProgrammerMessages: []llm.Message{
			{
				Role: "user",
				Content: []llm.Content{
					{
						ContentType: "text",
						Text:        "Write a short Python program to print the first 10 prime numbers?",
					},
				},
			},
		},
	}

	model := llm.NewAnthropic(&llm.AnthropicOptions{})

	graph := NewClanGraph(&initialState)
	graph.AddNode("Programmer", func(ws *WorkflowStateProgrammer) (*WorkflowStateProgrammer, error) {
		result, err := model.Generate(ws.ProgrammerMessages)
		if err != nil {
			return nil, err
		}

		ws.ProgrammerMessages = append(ws.ProgrammerMessages, result...)
		return ws, nil
	})

	graph.AddConditionalEdge("Programmer", func(wsp *WorkflowStateProgrammer) (string, error) {
		return "TestNonExistentEdge", nil
	})

	err := graph.SetStartNode("Programmer")
	require.NoError(t, err)

	_, err = graph.Execute(ExecuteOptions{})
	require.Error(t, err)

}

func TestTraversalDepth(t *testing.T) {
	initialState := WorkflowStateProgrammer{
		ProgrammerMessages: []llm.Message{},
	}

	graph := NewClanGraph(&initialState)
	graph.AddNode("Programmer", func(ws *WorkflowStateProgrammer) (*WorkflowStateProgrammer, error) {
		ws.ProgrammerMessages = append(ws.ProgrammerMessages, llm.Message{
			Role: "user",
			Content: []llm.Content{{
				ContentType: "text",
				Text:        "hello",
			}},
		})
		return ws, nil
	})

	graph.AddConditionalEdge("Programmer", func(wsp *WorkflowStateProgrammer) (string, error) {
		return "Programmer", nil
	})

	err := graph.SetStartNode("Programmer")
	require.NoError(t, err)

	_, err = graph.Execute(ExecuteOptions{
		TraversalDepth: 5,
	})
	require.Error(t, err)
}

type WorkflowStateProgrammer struct {
	ProgrammerMessages []llm.Message
	ReviewerMessages   []llm.Message
}

type WorkflowStateWeather struct {
	WeatherManMessages []llm.Message
}
