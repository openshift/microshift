package ostree

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testWriter struct {
	writeFail bool
	lenFail   bool
	data      []byte
}

func (w *testWriter) Write(p []byte) (int, error) {
	if w.writeFail {
		return 0, fmt.Errorf("failed to write")
	}

	if w.lenFail {
		return len(p) / 2, nil
	}

	w.data = append([]byte{}, p...)
	return len(p), nil
}

func Test_PersistOnDisk_success(t *testing.T) {
	nb := preRunAction{Action: actionBackup, OstreeID: "rhel-1234.0"}
	expectedData := `{"action":"backup","ostree":"rhel-1234.0"}`

	writer := &testWriter{writeFail: false, lenFail: false}
	getFileWriter = func() (io.Writer, error) {
		return writer, nil
	}

	err := nb.Persist()
	assert.NoError(t, err)
	assert.Equal(t, expectedData, string(writer.data))
}

func Test_PersistOnDisk_FailGettingWriter(t *testing.T) {
	nb := preRunAction{Action: actionBackup, OstreeID: "rhel-1234.0"}

	getFileWriter = func() (io.Writer, error) {
		return nil, fmt.Errorf("failed to open file")
	}

	err := nb.Persist()
	assert.Error(t, err)
}

func Test_PersistOnDisk_FailWriting(t *testing.T) {
	nb := preRunAction{Action: actionBackup, OstreeID: "rhel-1234.0"}

	writers := []*testWriter{
		{writeFail: true, lenFail: false},
		{writeFail: false, lenFail: true},
	}

	for _, writer := range writers {
		getFileWriter = func() (io.Writer, error) {
			return writer, nil
		}
		err := nb.Persist()
		assert.Error(t, err)
	}
}

func Test_preRunActionFromDisk(t *testing.T) {
	r := strings.NewReader(`{"action":"backup","ostree":"rhel-1234.0"}`)
	expectedNB := &preRunAction{Action: actionBackup, OstreeID: "rhel-1234.0"}

	getFileReader = func() (io.Reader, error) {
		return r, nil
	}
	fileExists = func() (bool, error) { return true, nil }

	nb, err := preRunActionFromDisk()
	assert.NoError(t, err)
	assert.Equal(t, expectedNB, nb)
}
