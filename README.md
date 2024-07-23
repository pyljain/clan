# Clan

Clan is a framework to develop agent teams using flow engineering. These agent teams can be expressed declaritively in a YAML manifest. 
Just as you would with human teams, you could assemble teams and give them access to one or more tools, define the team's vision and
express conditional interaactions in the form of Starlark functions. 

## APIs

Clan offers two API; a higher level API that can be used to outline Clan workflows as manifests. This is the recommended starting point for teams building GenAI teams (tribes). Clan also exposes a low level interface that you can use to express Clan workflows as graphs that you construct using 
the exposed methods.

### Workflow API

Clan offers a higher level API that is used when you express the workflow definition using a declarative manifest.

### Low level Clan APIs

Use this to construct Clan workflow graphs using the methods exposed.

## Features

### Checkpointing

Long running workflows, or workflows that are inherently variable given real time agentic interactions, benefit from the ability to persist progress (state) so that they can recover and restart from checkpoints. Clan expresses agents and tools as nodes. It checkpoints progress at handovers in between nodes.

### Graceful exits

You can declaratively decide how many handovers should your Clan workflow autononously support. This is extremely useful when agentic workflows run into recursive error loops and need human intervention. 

Add traversal depth in the manifest as

```sh
traversal_depth: 50
```

### Routing using Starlark functions


### Tools

### Custom tools

### Streaming


## Getting started


