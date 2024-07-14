package clan

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type emptyState struct{}

func TestStreaming(t *testing.T) {
	sc := make(chan interface{})

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
	require.NoError(t, err)

	go func() {
		_, err = graph.Execute(eo)
		require.NoError(t, err)
	}()

	results := []string{}
	for ss := range sc {
		update := ss.(StreamState[emptyState])
		results = append(results, update.NodeName)
	}

	require.Equal(t, []string{"One", "Two", "End"}, results)

}
