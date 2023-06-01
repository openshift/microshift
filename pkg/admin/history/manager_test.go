package history

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/openshift/microshift/pkg/admin/system"
	"github.com/stretchr/testify/assert"
)

type fakeStorage struct {
	history *History
	loadErr error
	saveErr error
}

func (fs *fakeStorage) Load() (*History, error) {
	err := fs.loadErr
	fs.loadErr = nil

	return fs.history, err
}

func (fs *fakeStorage) Save(history *History) error {
	err := fs.saveErr
	fs.saveErr = nil

	fs.history = history
	return err
}

type testBoot struct {
	info   system.Boot
	health Health
}

func Test_SimulateRealScenario(t *testing.T) {
	testBoots := []testBoot{
		{
			info: system.Boot{
				ID:           "boot-1",
				BootTime:     time.Date(2023, 06, 01, 0, 0, 0, 0, time.UTC),
				DeploymentID: "deploy-1",
			},
			health: Healthy,
		},
		{
			info: system.Boot{
				ID:           "boot-2",
				BootTime:     time.Date(2023, 06, 02, 0, 0, 0, 0, time.UTC),
				DeploymentID: "deploy-1",
			},
			health: Healthy,
		},
		{
			info: system.Boot{
				ID:           "boot-3",
				BootTime:     time.Date(2023, 06, 03, 0, 0, 0, 0, time.UTC),
				DeploymentID: "deploy-1",
			},
			health: Unhealthy,
		},
		{
			info: system.Boot{
				ID:           "boot-4",
				BootTime:     time.Date(2023, 06, 04, 0, 0, 0, 0, time.UTC),
				DeploymentID: "deploy-1",
			},
			health: Healthy,
		},
	}

	checkHistory := func(t *testing.T, dhm HistoryManager, expectedBoots []testBoot) {
		history, err := dhm.Get()
		assert.NoError(t, err)

		bootsNum := len(expectedBoots)
		assert.Len(t, history.Boots, bootsNum)
		// access items of history.Boots with `bootsNum-idx-1` because it is sorted (most recent first)
		// and thus in reverse order compared to expectedBoots
		for idx, eb := range expectedBoots {
			assert.Equal(t, eb.info.ID, history.Boots[bootsNum-idx-1].ID)
			assert.Equal(t, eb.info.BootTime, history.Boots[bootsNum-idx-1].BootTime)
			assert.Equal(t, eb.info.DeploymentID, history.Boots[bootsNum-idx-1].DeploymentID)
			assert.Equal(t, eb.health, history.Boots[bootsNum-idx-1].Health)
		}
	}

	stor := &fakeStorage{loadErr: ErrNoHistory, history: &History{}}
	dhm := NewHistoryManager(stor)

	for idx, boot := range testBoots {
		if idx > 0 {
			history, err := dhm.Get()
			boot, found := history.GetBootByID(testBoots[idx-1].info.ID)
			assert.NoError(t, err)
			assert.True(t, found)
			assert.Equal(t, testBoots[idx-1].health, boot.Health)
		}

		assert.NoError(t, dhm.Update(boot.info, BootInfo{Health: boot.health}))

		history, err := dhm.Get()
		boot, found := history.GetBootByID(boot.info.ID)
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, testBoots[idx].health, boot.Health)

		checkHistory(t, dhm, testBoots[:idx+1])
	}
}

func Test_HistoryFileExistsButIsEmpty(t *testing.T) {
	input := []string{"{}", `{"Deployments":[]}`}

	for _, in := range input {
		hist := &History{}
		bin := bytes.NewBufferString(in)
		err := json.Unmarshal(bin.Bytes(), hist)
		assert.NoError(t, err)

		currentBootInfo := system.Boot{
			ID:           "boot-1",
			BootTime:     time.Date(2023, 06, 03, 0, 0, 0, 0, time.UTC),
			DeploymentID: "0",
		}

		dhm := NewHistoryManager(&fakeStorage{loadErr: nil, history: hist})
		err = dhm.Update(currentBootInfo, BootInfo{Health: Healthy})
		assert.NoError(t, err)

		history, err := dhm.Get()
		assert.NoError(t, err)
		assert.Len(t, history.Boots, 1)
		assert.Equal(t, currentBootInfo.ID, history.Boots[0].ID)
		assert.Equal(t, currentBootInfo.DeploymentID, history.Boots[0].DeploymentID)
		assert.Equal(t, Healthy, history.Boots[0].Health)
	}
}

func Test_HistoryFileDoesNotExist(t *testing.T) {
	stor := &fakeStorage{loadErr: ErrNoHistory, history: nil}

	currentBootInfo := system.Boot{
		ID:           "boot-1",
		BootTime:     time.Date(2023, 06, 03, 0, 0, 0, 0, time.UTC),
		DeploymentID: "0",
	}

	dhm := NewHistoryManager(stor)
	err := dhm.Update(currentBootInfo, BootInfo{Health: Healthy})
	assert.NoError(t, err)

	history, err := dhm.Get()
	assert.NoError(t, err)
	assert.Len(t, history.Boots, 1)
	assert.Equal(t, currentBootInfo.ID, history.Boots[0].ID)
	assert.Equal(t, currentBootInfo.DeploymentID, history.Boots[0].DeploymentID)
	assert.Equal(t, Healthy, history.Boots[0].Health)

	// Update health of the boot
	err = dhm.Update(currentBootInfo, BootInfo{Health: Unhealthy})
	assert.NoError(t, err)
	history, err = dhm.Get()
	assert.NoError(t, err)
	assert.Len(t, history.Boots, 1)
	assert.Equal(t, currentBootInfo.ID, history.Boots[0].ID)
	assert.Equal(t, currentBootInfo.DeploymentID, history.Boots[0].DeploymentID)
	assert.Equal(t, Unhealthy, history.Boots[0].Health)
}

func Test_ProblemReadingHistoryFile(t *testing.T) {
	loadErr := fmt.Errorf("no permissions")
	stor := &fakeStorage{loadErr: loadErr, history: nil}

	currentBootInfo := system.Boot{
		ID:           "boot-1",
		BootTime:     time.Date(2023, 06, 03, 0, 0, 0, 0, time.UTC),
		DeploymentID: "0",
	}

	dhm := NewHistoryManager(stor)
	err := dhm.Update(currentBootInfo, BootInfo{Health: Healthy})
	assert.ErrorIs(t, err, loadErr)
}
