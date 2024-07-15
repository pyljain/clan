package workflow

import (
	"clan/pkg/checkpointer"
	"clan/pkg/clan"
	"clan/pkg/llm"
	"clan/pkg/planning"
	"clan/pkg/tools"
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

var NoAgentsDefinedErr = errors.New("no agents defined")

func Execute(definition *WorkflowDefinition, workflowID string) (chan interface{}, error) {
	streamChannel := make(chan interface{})
	ws := WorkflowState{
		AgentHistory: make(map[string][]llm.Message),
	}

	if len(definition.Agents) == 0 {
		return nil, NoAgentsDefinedErr
	}

	graph := clan.NewClanGraph(&ws)
	for _, agent := range definition.Agents {
		sysPrompt, err := generateSystemPrompt(agent.SystemPrompt, definition)
		if err != nil {
			return nil, err
		}

		ws.AgentHistory[agent.Name] = []llm.Message{
			{
				Role: "system",
				Content: []llm.Content{
					{
						Text:        sysPrompt,
						ContentType: "text",
					},
				},
			},
			{
				Role: "user",
				Content: []llm.Content{
					{
						Text:        definition.Goal,
						ContentType: "text",
					},
				},
			},
		}

		// Get list of tools from yaml and send definition to Anthropic

		llmTools := []llm.Tool{}
		for _, agentTool := range agent.AvailableTools {
			toolFound := false
			for _, toolRef := range tools.AllTools(definition.Tools) {
				if agentTool == toolRef.Name() {
					toolFound = true
					llmTools = append(llmTools, toolRef.Schema())
					break
				}
			}
			if !toolFound {
				return nil, fmt.Errorf("invalid tool %s", agentTool)
			}
		}

		model := llm.NewAnthropic(&llm.AnthropicOptions{Model: agent.Model, Tools: llmTools})

		graph.AddNode(agent.Name, func(ws *WorkflowState) (*WorkflowState, error) {
			if len(ws.Summaries) > 0 && agent.Name != ws.CurrentAgent {
				summaryBytes, err := json.Marshal(ws.Summaries)
				if err != nil {
					return nil, err
				}
				agentHistory := ws.AgentHistory[agent.Name]
				latestRoleEntry := agentHistory[len(agentHistory)-1].Role
				if latestRoleEntry == "user" {
					if agentHistory[len(agentHistory)-1].Content[0].ContentType == "tool_result" {
						currentText := agentHistory[len(agentHistory)-1].Content[0].Content
						currentText = fmt.Sprintf("%s\nThe summaries for the tasks completed by other agents who have worked on this so far are: %s", currentText, string(summaryBytes))
						agentHistory[len(agentHistory)-1].Content[0].Content = currentText
					} else {
						currentText := agentHistory[len(agentHistory)-1].Content[0].Text
						currentText = fmt.Sprintf("%s\nThe summaries for the tasks completed by other agents who have worked on this so far are: %s", currentText, string(summaryBytes))
						agentHistory[len(agentHistory)-1].Content[0].Text = currentText
					}
				} else {
					ws.AgentHistory[agent.Name] = append(ws.AgentHistory[agent.Name], llm.Message{
						Role: "user",
						Content: []llm.Content{
							{
								ContentType: "text",
								Text:        fmt.Sprintf("The summaries for the tasks completed by other agents who have worked on this so far are: %s", string(summaryBytes)),
							},
						},
					})
				}

			}
			ws.CurrentAgent = agent.Name

			log.Printf("HISTORY SENT TO THE MODEL IS %+v\n", ws.AgentHistory[agent.Name])

			resp, err := model.Generate(ws.AgentHistory[agent.Name])
			if err != nil {
				return nil, err
			}
			// log.Printf("Response from LLM %+v", resp)

			ws.AgentHistory[agent.Name] = append(ws.AgentHistory[agent.Name], resp...)
			return ws, nil
		})

		toolsNodeName := fmt.Sprintf("%s_tools", agent.Name)

		graph.AddNode(toolsNodeName, func(ws *WorkflowState) (*WorkflowState, error) {

			ws.completionMarkerCalled = false
			ws.toolInvoked = false

			// Identify which tool was called
			agentsHistory := ws.AgentHistory[agent.Name]
			lastItemFromHistory := agentsHistory[len(agentsHistory)-1]
			// log.Printf("agentsHistory is %+v", agentsHistory)
			// log.Printf("lastItemFromHistory is %s", lastItemFromHistory)
			agentDidNotCallAnyTool := true
			for _, contentNode := range lastItemFromHistory.Content {
				if contentNode.ContentType == "tool_use" {
					agentDidNotCallAnyTool = false
					toolFound := false
					for _, t := range tools.AllTools(definition.Tools) {
						if t.Name() == contentNode.Name {
							log.Printf("Tool called %s", t.Name())
							// Call tool function
							result, err := t.Execute(contentNode.Input)
							if err != nil {
								return nil, err
							}

							if t.Name() == "NextAgentSelector" {
								ws.completionMarkerCalled = true
								// Update state with summary
								ws.Summaries = append(ws.Summaries, Summary{
									AgentName: agent.Name,
									Summary:   contentNode.Input["summary"].(string),
								})

								requestedNextAgent := contentNode.Input["next_agent"]
								ws.RequestedNextAgent = requestedNextAgent.(string)
							}

							if t.Name() == "PlanCreator" {
								pc := t.(*planning.CreatePlan)
								ws.Plan = pc.CurrentPlan
							}

							if t.Name() == "PlanUpdater" {
								// Find the task
								for i := range ws.Plan {
									name := contentNode.Input["taskName"].(string)
									if name == ws.Plan[i].Name {
										// Update the status of the task
										status := contentNode.Input["status"].(string)
										ws.Plan[i].Status = status
									}
								}
							}

							if t.Name() == "GetPlan" {
								planBytes, err := json.Marshal(ws.Plan)
								if err != nil {

									return nil, err
								}

								result = string(planBytes)
							}

							// Write result to history
							ws.AgentHistory[agent.Name] = append(ws.AgentHistory[agent.Name], llm.Message{
								Role: "user",
								Content: []llm.Content{
									{
										Content:     result,
										ContentType: "tool_result",
										ToolUseId:   contentNode.Id,
									},
								},
							})
							toolFound = true
							ws.toolInvoked = true
						}
					}

					if !toolFound {
						return nil, fmt.Errorf(
							"invalid tool call by LLM %s", contentNode.Name)
					}
				}
			}

			if agentDidNotCallAnyTool {
				ws.AgentHistory[agent.Name] = append(ws.AgentHistory[agent.Name], llm.Message{
					Role: "user",
					Content: []llm.Content{
						{
							Content:     "What do you want to do next? Please call the `NextAgentSelector` tool after you have completed the users task",
							ContentType: "text",
						},
					},
				})
			}

			return ws, nil
		})

		err = graph.AddEdge(agent.Name, toolsNodeName)
		if err != nil {
			return nil, err
		}

		err = graph.AddConditionalEdge(toolsNodeName, func(ws *WorkflowState) (string, error) {
			// If goal complete tool was called then go to next node
			if ws.completionMarkerCalled {

				if ws.RequestedNextAgent != "" {
					return ws.RequestedNextAgent, nil
				} else if agent.NextAgent != "" {
					return agent.NextAgent, nil
				}

				return executeNextAgentFn(agent.NextAgentFunction, ws)
				// return agent.NextAgent, nil
			}

			// If no tool called then go back to LLM saying what is next
			return agent.Name, nil
		})
	}

	graph.SetStartNode(definition.StartAgent)
	go func() {
		traversalDepth := 100
		if definition.TraversalDepth > 0 {
			traversalDepth = definition.TraversalDepth
		}

		var checkpointProvider checkpointer.Checkpointer
		var err error
		if definition.Checkpoint != nil {
			checkpointProvider, err = checkpointer.NewCheckpointerWithName(definition.Checkpoint.Type, definition.Checkpoint.ConnectionString)
			if err != nil {
				log.Printf("Could not setup checkpointer %s", err)
			}
		}

		_, err = graph.Execute(clan.ExecuteOptions{
			StreamChannel:  streamChannel,
			TraversalDepth: traversalDepth,
			Checkpointer:   checkpointProvider,
			WorkflowID:     workflowID,
		})
		if err != nil {
			log.Printf("Error occured during execution %s", err)
		}
	}()

	// log.Printf("State is %+v", state)
	return streamChannel, nil
}

type WorkflowState struct {
	AgentHistory           map[string][]llm.Message
	Summaries              []Summary
	CurrentAgent           string
	Plan                   []planning.Task
	toolInvoked            bool
	completionMarkerCalled bool
	RequestedNextAgent     string
}

type Summary struct {
	AgentName string `json:"agent_name"`
	Summary   string `json:"summary"`
}
