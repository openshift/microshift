package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddToNoProxyEnv(t *testing.T) {
	clearNoProxy()
	AddToNoProxyEnv(".svc", "10.40.0.0/16")

	assert.Equal(t, ".svc,10.40.0.0/16", os.Getenv("NO_PROXY"), "NO_PROXY has unexpected value")
	assert.Equal(t, "", os.Getenv("no_proxy"), "no_proxy expected to be empty")
	clearNoProxy()
}

func clearNoProxy() {
	os.Setenv("NO_PROXY", "")
	os.Setenv("no_proxy", "")
}

func TestAddToNoProxyEnv_with_contents(t *testing.T) {
	os.Setenv("NO_PROXY", "my.host.local")
	os.Setenv("no_proxy", "")
	AddToNoProxyEnv(".svc", "10.40.0.0/16")

	assert.Equal(t, ".svc,10.40.0.0/16,my.host.local", os.Getenv("NO_PROXY"), "NO_PROXY has unexpected value")
	assert.Equal(t, "", os.Getenv("no_proxy"), "no_proxy expected to be empty")
	clearNoProxy()
}

func TestAddToNoProxyEnv_with_dups(t *testing.T) {
	os.Setenv("NO_PROXY", "my.host.local")
	os.Setenv("no_proxy", "my.host.local")
	AddToNoProxyEnv(".svc", "10.40.0.0/16")

	assert.Equal(t, ".svc,10.40.0.0/16,my.host.local", os.Getenv("NO_PROXY"), "NO_PROXY has unexpected value")
	assert.Equal(t, "", os.Getenv("no_proxy"), "no_proxy expected to be empty")
	clearNoProxy()
}

func TestAddToNoProxyEnv_with_both(t *testing.T) {
	os.Setenv("NO_PROXY", "my.host.local")
	os.Setenv("no_proxy", "another.host.local")
	AddToNoProxyEnv(".svc", "10.40.0.0/16")

	assert.Equal(t, ".svc,10.40.0.0/16,another.host.local,my.host.local", os.Getenv("NO_PROXY"), "NO_PROXY has unexpected value")
	assert.Equal(t, "", os.Getenv("no_proxy"), "no_proxy expected to be empty")
	clearNoProxy()
}
