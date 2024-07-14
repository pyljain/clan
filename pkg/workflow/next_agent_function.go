package workflow

import (
	"log"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

func executeNextAgentFn(funcDef string, state *WorkflowState) (string, error) {
	// Execute Starlark program in a file.
	thread := &starlark.Thread{Name: "my thread"}
	globals, err := starlark.ExecFileOptions(&syntax.FileOptions{}, thread, "func.star", funcDef, nil)
	if err != nil {
		log.Printf("Unable to execute the Starlark function. Err is %s", err)
		return "", err
	}

	nextAgentfn := globals["next_agent"]
	slState := convertStateToStarlarkDict(state)
	res, err := starlark.Call(thread, nextAgentfn, starlark.Tuple{
		slState,
	}, nil)

	if err != nil {
		log.Printf("Unable to run the Starlark function. Err is %s", err)
		return "", err
	}

	resultString := res.(starlark.String).GoString()

	return resultString, nil
}

func convertStateToStarlarkDict(ws *WorkflowState) *starlark.Dict {
	res := starlark.NewDict(3)
	res.SetKey(starlark.String("CurrentAgent"), starlark.String(ws.CurrentAgent))
	// res.SetKey(starlark.String("AgentHistory"), starlark.NewList(ws.AgentHistory[ws.CurrentAgent]))

	var summaries []starlark.Value
	for _, summary := range ws.Summaries {
		starlarkSummary := starlark.NewDict(2)
		starlarkSummary.SetKey(starlark.String("agentName"), starlark.String(summary.AgentName))
		starlarkSummary.SetKey(starlark.String("summary"), starlark.String(summary.Summary))
		summaries = append(summaries, starlarkSummary)
	}
	res.SetKey(starlark.String("Summaries"), starlark.NewList(summaries))

	// Convert agent history to Starlark dta structures
	starlarkAgentHistory := starlark.NewDict(len(ws.AgentHistory))
	for agentName, llmMessages := range ws.AgentHistory {
		var starlarkLlmMessages []starlark.Value
		for _, llmMessage := range llmMessages {
			starlarkLlmMessage := starlark.NewDict(2)
			starlarkLlmMessage.SetKey(starlark.String("Role"), starlark.String(llmMessage.Role))

			var starlarkContent []starlark.Value
			for _, c := range llmMessage.Content {
				starlarkC := starlark.NewDict(7)
				starlarkC.SetKey(starlark.String("Text"), starlark.String(c.Text))
				starlarkC.SetKey(starlark.String("Content"), starlark.String(c.Content))
				starlarkC.SetKey(starlark.String("ContentType"), starlark.String(c.ContentType))
				starlarkC.SetKey(starlark.String("Id"), starlark.String(c.Id))
				starlarkC.SetKey(starlark.String("Name"), starlark.String(c.Name))
				starlarkC.SetKey(starlark.String("ToolUseId"), starlark.String(c.ToolUseId))
				starlarkContent = append(starlarkContent, starlarkC)
			}

			starlarkLlmMessage.SetKey(starlark.String("Content"), starlark.NewList(starlarkContent))
			starlarkLlmMessages = append(starlarkLlmMessages, starlarkLlmMessage)
		}
		starlarkAgentHistory.SetKey(starlark.String(agentName), starlark.NewList(starlarkLlmMessages))
	}

	res.SetKey(starlark.String("AgentHistory"), starlarkAgentHistory)

	return res
}
