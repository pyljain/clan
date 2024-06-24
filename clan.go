package clan

import (
	"clan/pkg/checkpointer"
	"encoding/json"
	"errors"
	"fmt"
)

var NodeNotFoundErr = errors.New("node not found")

type ClanGraph[T any] struct {
	Nodes       []*Node[T]
	State       *T
	CurrentNode *Node[T]
}

type Node[T any] struct {
	Name     string
	NextNode func(*T) (string, error)
	NodeFn   func(*T) (*T, error)
}

func NewClanGraph[T any](initialState *T) *ClanGraph[T] {
	return &ClanGraph[T]{
		State: initialState,
	}
}

type ExecuteOptions struct {
	TraversalDepth int
	WorkflowID     string
	Checkpointer   checkpointer.Checkpointer
}

func (cg *ClanGraph[T]) Execute(options ExecuteOptions) (*T, error) {

	// Add end node to the graph
	cg.AddNode("End", func(t *T) (*T, error) {
		return t, nil
	})

	if options.TraversalDepth == 0 {
		options.TraversalDepth = 100
	}

	var currentDepth int

	if options.Checkpointer != nil {
		// Restore from checkpoint if exists
		if options.WorkflowID == "" {
			return nil, fmt.Errorf("in order to checkpoint, you must pass a WorkflowID")
		}

		existingCheckpoint, err := options.Checkpointer.GetLastCheckpoint(options.WorkflowID)
		if err == nil && existingCheckpoint != nil {
			cg.CurrentNode = cg.findNodebyName(existingCheckpoint.NodeName)
			var existingState T
			err = json.Unmarshal([]byte(existingCheckpoint.State), &existingState)
			if err != nil {
				return nil, err
			}
			cg.State = &existingState
			currentDepth = existingCheckpoint.CurrentDepth
		}
	}

	for {
		var err error
		err = cg.checkpoint(options, currentDepth)
		if err != nil {
			return nil, err
		}

		if cg.CurrentNode.Name == "End" {
			break
		}

		cg.State, err = cg.CurrentNode.NodeFn(cg.State)
		if err != nil {
			return nil, err
		}
		nextNode, err := cg.CurrentNode.NextNode(cg.State)
		if err != nil {
			return nil, err
		}

		cg.CurrentNode = cg.findNodebyName(nextNode)
		if cg.CurrentNode == nil {
			return nil, fmt.Errorf("No node found with the given name %s", nextNode)
		}

		currentDepth += 1
		if currentDepth > options.TraversalDepth {
			return nil, fmt.Errorf("Traversal depth exceeded %d", currentDepth)
		}
	}

	return cg.State, nil
}

func (cg *ClanGraph[T]) AddNode(nodeName string, nodeFn func(*T) (*T, error)) {
	cg.Nodes = append(cg.Nodes, &Node[T]{
		Name:   nodeName,
		NodeFn: nodeFn,
	})
}

func (cg *ClanGraph[T]) AddConditionalEdge(nodeName string, fn func(*T) (string, error)) error {
	n := cg.findNodebyName(nodeName)
	if n == nil {
		return NodeNotFoundErr
	}
	n.NextNode = fn
	return nil
}

func (cg *ClanGraph[T]) AddEdge(nodeFromName, nodeToName string) error {
	return cg.AddConditionalEdge(nodeFromName, func(state *T) (string, error) {
		return nodeToName, nil
	})
}

func (cg *ClanGraph[T]) findNodebyName(name string) *Node[T] {
	for _, nd := range cg.Nodes {
		if nd.Name == name {
			return nd
		}
	}

	return nil
}

func (cg *ClanGraph[T]) checkpoint(options ExecuteOptions, currentDepth int) error {
	if options.Checkpointer != nil {
		stateJSON, err := json.Marshal(cg.State)
		if err != nil {
			return err
		}

		err = options.Checkpointer.Checkpoint(options.WorkflowID, checkpointer.Checkpoint{
			NodeName:     cg.CurrentNode.Name,
			State:        string(stateJSON),
			CurrentDepth: currentDepth,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (cg *ClanGraph[T]) SetStartNode(nodeName string) error {
	n := cg.findNodebyName(nodeName)
	if n == nil {
		return NodeNotFoundErr
	}

	cg.CurrentNode = n

	return nil
}
