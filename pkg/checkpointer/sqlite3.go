package checkpointer

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type sqlite struct {
	db *sql.DB
}

func NewSQLite(filepath string) (*sqlite, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

	return &sqlite{
		db: db,
	}, nil
}

func (s *sqlite) setup() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS checkpoints (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			workflow_id TEXT NOT NULL,
			node_name TEXT NOT NULL,
			state TEXT NOT NULL,
			depth INTEGER NOT NULL
		);`,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *sqlite) Checkpoint(workflowID string, checkpoint Checkpoint) error {
	err := s.setup()
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		INSERT INTO checkpoints (timestamp, workflow_id, node_name, state, depth)
		VALUES ($1, $2, $3, $4, $5);`,
		time.Now(),
		workflowID,
		checkpoint.NodeName,
		checkpoint.State,
		checkpoint.CurrentDepth,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *sqlite) GetLastCheckpoint(workflowID string) (*Checkpoint, error) {
	err := s.setup()
	if err != nil {
		return nil, err
	}

	row := s.db.QueryRow(`
		SELECT node_name, state, depth FROM 
		checkpoints WHERE workflow_id = $1
		ORDER BY timestamp DESC LIMIT 1
	`, workflowID)

	if row.Err() != nil {
		return nil, err
	}

	var cp Checkpoint
	err = row.Scan(&cp.NodeName, &cp.State, &cp.CurrentDepth)
	if err != nil {
		return nil, err
	}

	return &cp, nil
}

func (s *sqlite) ListAll(workflowID string) ([]Checkpoint, error) {
	err := s.setup()
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(`
		SELECT node_name, state, depth FROM 
		checkpoints WHERE workflow_id = $1
		ORDER BY timestamp
	`, workflowID)
	if err != nil {
		return nil, err
	}

	var result []Checkpoint
	for rows.Next() {
		var cp Checkpoint
		err := rows.Scan(&cp.NodeName, &cp.State, &cp.CurrentDepth)
		if err != nil {
			return nil, err
		}
		result = append(result, cp)
	}
	return result, nil
}
