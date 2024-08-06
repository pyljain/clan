# Clan

Clan is a framework to develop agent teams using flow engineering. These agent teams can be expressed declaritively in a YAML manifest. 
Just as you would with human teams, you could assemble teams and give them access to one or more tools, define the team's vision and
express conditional interaactions in the form of Starlark functions. 

## APIs

Clan offers two API; a higher level API that can be used to outline Clan workflows as manifests. This is the recommended starting point for teams building GenAI teams (tribes). Clan also exposes a low level interface that you can use to express Clan workflows as graphs that you construct using 
the exposed methods.

### Workflow API

Clan offers a higher level API that is used when you express the workflow definition using a declarative manifest. 

An example manifest is as follows:

```yaml
name: Research
description: "Search"
type: Workflow Definition
goal: "Your task is to write an essay on the current Wimbledon champion."
start_agent: Researcher
checkpoint:
  type: sqlite3
  connection_string: "./wimbledon.db"
agents:
- name: Researcher
  purpose: "Find information and write an essay"
  system_prompt: |
    Your job is to write a detailed essay on the topic the user specifies.

    Please write the essay in markdown format to a file called `wimbledon_essay.md`. Please call the NextAgentSelector tool after you have completed your task.
  model: claude-3-5-sonnet-20240620
  temperature: 0.2
  next_agent: End
  available_tools: 
  - SerperSearch
  - Writer
  - NextAgentSelector

tools:
- name: SerperSearch
  description: "Google search for a term"
  parameters:
  - name: term
    description: "The term to search for"
    type: string
  function: |
    def serpersearch(term):
      result = post('https://google.serper.dev/search', {
        "X-API-KEY": getEnv("SERPER_KEY"),
        "Content-Type": "application/json"
      }, '{"q": "' + term + '"}')
      return result
```

You specify a list of agents and the agent to start the workflow with. Each agent contains a system prompt, model, a reference to the next agent to call in the chain and a list of tools that the agents can call. The tools can either be built in or you can define tools using the StarLark language and provide our in-build functions such as `getEnv("SERPER_KEY")`

### Low level Clan APIs

Use this to construct Clan workflow graphs using the exposed methods. The low level API is a generic graph-like API that lets you define nodes for each of the steps in the workflow. It then also lets you define the conditions to transition from one node to another.

```go
graph := NewClanGraph(&initialState)
graph.AddNode("Programmer", func(ws *WorkflowStateProgrammer) (*WorkflowStateProgrammer, error) {
    // Make some changes to the state
    return ws, nil
})

graph.AddNode("Reviewer", func(ws *WorkflowStateProgrammer) (*WorkflowStateProgrammer, error) {
    // Make some changes to the state
    return ws, nil
})

// Define your transistions
graph.AddEdge("Programmer", "Reviewer")
graph.AddEdge("Reviewer", "End") // End is a special node type that indicates that the workflow can end

// Set the start node
err := graph.SetStartNode("Programmer")

// Run the graph
_, err = graph.Execute(ExecuteOptions{})
```

## Features

### Checkpointing

Long running workflows, or workflows that are inherently variable given real time agentic interactions, benefit from the ability to persist progress (state) so that they can recover and restart from checkpoints. Clan expresses agents and tools as nodes. It checkpoints progress at handovers in between nodes.

To enable checkpointing, in the Clan manifest add

```sh
checkpoint:
  type: sqlite3
  connection_string: [CONNECTION-STRING-TO-YOUR-DATABASE]
```

### Graceful exits

You can declaratively decide how many handovers should your Clan workflow autonomously support. This is extremely useful when agentic workflows run into recursive error loops and need human intervention. 

Add traversal depth in the manifest as

```sh
traversal_depth: 50
```

### Routing using Starlark functions

When expressing your Clan workflow in the manifest you can declaratively select the next agent in the flow by setting a value for `next_agent`. However, you 
may sometimes want to conditionally determine who the next agent should be in a workflow based on outcomes from a previous step or tool runs. Clan lets you express such conditionality using Starlark functions.

You can express this as follows in the manifest:
```sh
next_agent_function: |
    def choose_agent(api_response):
    status_code = api_response['status_code']

    if status_code > 399:
        return 'Planner'
    elif status_code == 200:
        return 'Reviewer'
    else:
        return 'Unknown'

```


### Tools

Clan offers some standard tools for use, namely the following:

  - `Reader` to read files
  - `Writer` to write files
  - `CommandRunner` to execute commands on the machine Clan is running on
  - `NextAgentSelector` an inbuilt function that is invoked to select the next agent
  - `PlanUpdater` an inbuilt function that is invoked to update the plan created for workflow execution
  - `GetPlan` an inbuilt function that is invoked by agents to fetch the current plan

### Custom tools

### Streaming

This feature allows the fetching of latest state as the Clan workflow executes. This makes it possible for clients, such as UIs and CLIs, to render state as it transpires. Inherently, Clan uses a channel to manage state. At this time, it is streaming is enabled by default and can only be disabled when using the low level API if you choose to do so.

```go
sc := make(chan interface{}) // Create a channel to stream state into

initialState := emptyState{}
eo := ExecuteOptions{WorkflowID: "sample", StreamChannel: sc} 

graph := NewClanGraph(&initialState)
graph.AddNode("One", func(ws *emptyState) (*emptyState, error) {
    return &emptyState{}, nil
})

graph.AddNode("Two", func(ws *emptyState) (*emptyState, error) {
    return &emptyState{}, nil
})

graph.AddEdge("One", "Two")
graph.AddEdge("Two", "End")
err := graph.SetStartNode("One")

go func() {
    _, err = graph.Execute(eo) // Execute the graph
    require.NoError(t, err)
}()

results := []string{}
for ss := range sc { // Get streaming results from the channel
    update := ss.(StreamState[emptyState])
    results = append(results, update.NodeName)
}
```


## Getting started

To get started with Clan

1. Get the project by running

```sh
git clone https://github.com/pyljain/clan.git
```

2. Build

```sh
go build -o clan
```

3. Create your Clan workflow definition & set environment variables

4. Run

```sh
./clan ./samples/[YOUR-CLAN-MANIFEST].yaml  
```


