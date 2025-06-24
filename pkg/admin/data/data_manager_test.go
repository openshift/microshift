package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/sets"
)

func Test_checkDirectoryContents(t *testing.T) {
	testData := []struct {
		name    string
		input   sets.Set[string]
		isValid bool
	}{
		{
			name:    "Perfect backup",
			input:   sets.New[string]("certs", "etcd", "kubelet-plugins", "resources", "version"),
			isValid: true,
		},
		{
			name:    "Backup with some extra files - still valid",
			input:   sets.New[string]("certs", "etcd", "kubelet-plugins", "resources", "version", "extra", "random", "dirs"),
			isValid: true,
		},
		{
			name:    "version file is missing",
			input:   sets.New[string]("certs", "etcd", "kubelet-plugins", "resources"),
			isValid: false,
		},
		{
			name:    "certs dir is missing",
			input:   sets.New[string]("etcd", "kubelet-plugins", "resources", "version"),
			isValid: false,
		},
		{
			name:    "None of the directories match",
			input:   sets.New[string]("1", "2", "3", "4"),
			isValid: false,
		},
	}

	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			err := checkDirectoryContents(td.input)
			if td.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
