package checkpointer

import "fmt"

type Checkpointer interface {
	Checkpoint(workflowID string, checkpoint Checkpoint) error
	GetLastCheckpoint(workflowID string) (*Checkpoint, error)
	ListAll(workflowID string) ([]Checkpoint, error)
}

type Checkpoint struct {
	NodeName     string
	State        string
	CurrentDepth int
}

func NewCheckpointerWithName(checkpointerType string, connectionString string) (Checkpointer, error) {
	switch checkpointerType {
	case "sqlite3":
		return NewSQLite(connectionString)
	default:
		return nil, fmt.Errorf("Invalid checkpointer %s", checkpointerType)
	}
}
