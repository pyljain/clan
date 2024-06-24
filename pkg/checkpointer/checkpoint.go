package checkpointer

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
