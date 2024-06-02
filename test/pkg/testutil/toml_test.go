package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetTOMLFieldValue(t *testing.T) {
	data := `name = "rhel-9.2-microshift-source"
version = "0.0.1"
modules = []
groups = []
distro = "rhel-92"
`

	name, err := GetTOMLFieldValue(data, "name")
	assert.NoError(t, err)
	assert.Equal(t, "rhel-9.2-microshift-source", name)

	distro, err := GetTOMLFieldValue(data, "distro")
	assert.NoError(t, err)
	assert.Equal(t, "rhel-92", distro)

	unknown, err := GetTOMLFieldValue(data, "unknown")
	assert.NoError(t, err)
	assert.Equal(t, "", unknown)
}

func Test_GetTOMLFieldValue_WrongAmountOfFields(t *testing.T) {
	data := `name = "rhel-9.2-microshift-source"
version = "0.0.1"
modules = []
groups = []
distro = "rhel"-92"
`

	_, err := GetTOMLFieldValue(data, "distro")
	assert.Error(t, err)
}
