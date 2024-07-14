package main

import (
	"clan/pkg/clan"
	"clan/pkg/workflow"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"gopkg.in/yaml.v3"
)

func main() {
	workflowPath := os.Args[1]
	if workflowPath == "" {
		fmt.Fprintf(os.Stderr, "Please pass a path for your Clan workflow definition\n")
		os.Exit(-1)
	}

	def, err := parseWorkflow(workflowPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse your workflow definition: %s\n", err)
		os.Exit(-1)
	}

	sChan, err := workflow.Execute(def)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to execute your workflow: %s\n", err)
		os.Exit(-1)
	}

	for element := range sChan {
		res := element.(clan.StreamState[workflow.WorkflowState])
		if res.State.CurrentAgent == "" {
			continue
		}
		agentColor := color.New(color.Bold).SprintFunc()
		fmt.Printf(color.BlueString("AGENT NAME: ")+agentColor(" %s\n"), res.State.CurrentAgent)
		history := res.State.AgentHistory[res.State.CurrentAgent]
		if len(history) > 0 {
			// fmt.Printf("LAST MESSAGE FROM AGENT HISTORY IS:  %+v\n", history[len(history)-1])
			for _, c := range history[len(history)-1].Content {
				switch c.ContentType {
				case "text":
					fmt.Printf("%s\n", c.Text)
				case "tool_use":
					fmt.Printf(color.YellowString("CALLING TOOL: %s\n"), c.Name)
				case "tool_result":
					fmt.Printf(color.CyanString("TOOL RESULT: \n%s\n", c.Content))
				}
			}
		}
		fmt.Println(agentColor("Plan"))

		headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
		columnFmt := color.New(color.FgYellow).SprintfFunc()
		tbl := table.New("Name", "Description", "Owner", "Status")
		tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
		for _, task := range res.State.Plan {
			tbl.AddRow(task.Name, task.Description, task.Owner, task.Status)
		}
		tbl.Print()
	}
}

func parseWorkflow(workflowPath string) (*workflow.WorkflowDefinition, error) {
	workflowBytes, err := os.ReadFile(workflowPath)
	if err != nil {
		return nil, err
	}

	wd := workflow.WorkflowDefinition{}
	err = yaml.Unmarshal(workflowBytes, &wd)
	if err != nil {
		return nil, err
	}

	return &wd, nil
}
