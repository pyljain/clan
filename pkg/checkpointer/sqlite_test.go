package checkpointer

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSQLiteInsertCheckpoint(t *testing.T) {
	testWorkflowId := "sample"
	cp := Checkpoint{
		NodeName:     "Programmer",
		State:        "{\"node\": \"Programmer\", \"task\": \"Write a Python program to print hello world\"}",
		CurrentDepth: 1,
	}

	filepath := "./test_database.db"
	db, err := NewSQLite(filepath)
	require.NoError(t, err)
	defer func() {
		err = os.Remove(filepath)
		require.NoError(t, err)
	}()

	err = db.Checkpoint(testWorkflowId, cp)
	require.NoError(t, err)
}

func TestSQLiteGetLastCheckpoint(t *testing.T) {
	testWorkflowId := "sample"
	cp1 := Checkpoint{
		NodeName:     "Programmer",
		State:        "{\"node\": \"Programmer\", \"task\": \"Write a Python program to print hello world\"}",
		CurrentDepth: 1,
	}

	cp2 := Checkpoint{
		NodeName:     "Reviewer",
		State:        "{\"node\": \"Programmer\", \"task\": \"Could you create unit tests?\"}",
		CurrentDepth: 2,
	}

	filepath := "./test_database.db"
	db, err := NewSQLite(filepath)
	require.NoError(t, err)
	defer func() {
		err = os.Remove(filepath)
		require.NoError(t, err)
	}()

	err = db.Checkpoint(testWorkflowId, cp1)
	require.NoError(t, err)

	err = db.Checkpoint(testWorkflowId, cp2)
	require.NoError(t, err)

	cp, err := db.GetLastCheckpoint("sample")
	require.NoError(t, err)
	require.Equal(t, "Reviewer", cp.NodeName)
	require.Equal(t, 2, cp.CurrentDepth)

}

func TestSQLiteListAllCheckpoints(t *testing.T) {
	testWorkflowId := "sample"
	cp1 := Checkpoint{
		NodeName:     "Programmer",
		State:        "{\"node\": \"Programmer\", \"task\": \"Write a Python program to print hello world\"}",
		CurrentDepth: 1,
	}

	cp2 := Checkpoint{
		NodeName:     "Reviewer",
		State:        "{\"node\": \"Programmer\", \"task\": \"Could you create unit tests?\"}",
		CurrentDepth: 2,
	}

	filepath := "./test_database.db"
	db, err := NewSQLite(filepath)
	require.NoError(t, err)
	defer func() {
		err = os.Remove(filepath)
		require.NoError(t, err)
	}()

	err = db.Checkpoint(testWorkflowId, cp1)
	require.NoError(t, err)

	err = db.Checkpoint(testWorkflowId, cp2)
	require.NoError(t, err)

	checkpoints, err := db.ListAll(testWorkflowId)
	require.NoError(t, err)
	require.Equal(t, 2, len(checkpoints))
	require.Equal(t, "Programmer", checkpoints[0].NodeName)
	require.Equal(t, "Reviewer", checkpoints[1].NodeName)
	require.Equal(t, 2, checkpoints[1].CurrentDepth)

}
