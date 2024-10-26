package autorecovery

import (
	"fmt"
	"testing"
	"testing/fstest"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/stretchr/testify/assert"
)

func Test_getStateFS(t *testing.T) {
	lastBackup := "20241001111100_4.18.0"
	storageFS := fstest.MapFS{
		stateFilename: &fstest.MapFile{
			Data: []byte(fmt.Sprintf(`{"LastBackup": "%s"}`, lastBackup))},
	}

	state, err := getStateFS(storageFS)
	assert.NoError(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, data.BackupName(lastBackup), state.LastBackup)
}

func Test_getStateFS_NoState(t *testing.T) {
	storageFS := fstest.MapFS{}

	state, err := getStateFS(storageFS)
	assert.NoError(t, err)
	assert.Nil(t, state)
}

func Test_state_MarshalJSON(t *testing.T) {
	{
		s := state{LastBackup: "asd"}
		data, err := s.Serialize()
		assert.NoError(t, err)
		assert.Equal(t, `{"LastBackup":"asd"}`, string(data))
	}

	{
		s := &state{LastBackup: "asd"}
		data, err := s.Serialize()
		assert.NoError(t, err)
		assert.Equal(t, `{"LastBackup":"asd"}`, string(data))
	}
}
