package clan

import (
	"clan/pkg/checkpointer"
	"clan/pkg/llm"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckpointing(t *testing.T) {
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

		// log.Printf("Response from LLM %+v", result)

		return ws, nil
	})

	graph.AddEdge("Programmer", "Reviewer")
	graph.AddEdge("Reviewer", "End")

	err := graph.SetStartNode("Programmer")
	require.NoError(t, err)

	dbPath := "./test_database.db"
	cp, err := checkpointer.NewSQLite(dbPath)
	require.NoError(t, err)
	defer func() {
		err = os.Remove(dbPath)
		require.NoError(t, err)
	}()

	_, err = graph.Execute(ExecuteOptions{
		TraversalDepth: 3,
		WorkflowID:     "sample",
		Checkpointer:   cp,
	})
	require.NoError(t, err)
}

type resumeExample struct {
	reviewerRun bool
}

func TestResume(t *testing.T) {
	initialState := resumeExample{}

	hasRunBefore := false
	graph := NewClanGraph(&initialState)
	graph.AddNode("Programmer", func(ws *resumeExample) (*resumeExample, error) {
		return ws, nil
	})

	graph.AddConditionalEdge("Programmer", func(wsp *resumeExample) (string, error) {
		if !hasRunBefore {
			hasRunBefore = true
			return "Pause", nil
		}

		return "Reviewer", nil
	})

	graph.AddNode("Reviewer", func(ws *resumeExample) (*resumeExample, error) {
		ws.reviewerRun = true
		return ws, nil
	})

	graph.AddEdge("Reviewer", "End")

	err := graph.SetStartNode("Programmer")
	require.NoError(t, err)

	dbPath := "./test_database.db"
	cp, err := checkpointer.NewSQLite(dbPath)
	require.NoError(t, err)
	defer func() {
		err = os.Remove(dbPath)
		require.NoError(t, err)
	}()

	state, err := graph.Execute(ExecuteOptions{
		TraversalDepth: 3,
		WorkflowID:     "sample",
		Checkpointer:   cp,
	})
	require.NoError(t, err)
	require.False(t, state.reviewerRun)

	state, err = graph.Execute(ExecuteOptions{
		TraversalDepth: 3,
		WorkflowID:     "sample",
		Checkpointer:   cp,
	})
	require.NoError(t, err)
	require.True(t, state.reviewerRun)
}
