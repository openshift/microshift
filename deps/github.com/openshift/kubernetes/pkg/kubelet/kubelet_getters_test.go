/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubelet

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
)

func TestKubeletDirs(t *testing.T) {
	testKubelet := newTestKubelet(t, false /* controllerAttachDetachEnabled */)
	defer testKubelet.Cleanup()
	kubelet := testKubelet.kubelet
	root := kubelet.rootDirectory

	var exp, got string

	got = kubelet.getPodsDir()
	exp = filepath.Join(root, "pods")
	assert.Equal(t, exp, got)

	got = kubelet.getPodLogsDir()
	assert.Equal(t, kubelet.podLogsDirectory, got)

	got = kubelet.getPluginsDir()
	exp = filepath.Join(root, "plugins")
	assert.Equal(t, exp, got)

	got = kubelet.getPluginsRegistrationDir()
	exp = filepath.Join(root, "plugins_registry")
	assert.Equal(t, exp, got)

	got = kubelet.getPluginDir("foobar")
	exp = filepath.Join(root, "plugins/foobar")
	assert.Equal(t, exp, got)

	got = kubelet.getPodDir("abc123")
	exp = filepath.Join(root, "pods/abc123")
	assert.Equal(t, exp, got)

	got = kubelet.getPodVolumesDir("abc123")
	exp = filepath.Join(root, "pods/abc123/volumes")
	assert.Equal(t, exp, got)

	got = kubelet.getPodVolumeDir("abc123", "plugin", "foobar")
	exp = filepath.Join(root, "pods/abc123/volumes/plugin/foobar")
	assert.Equal(t, exp, got)

	got = kubelet.getPodVolumeDevicesDir("abc123")
	exp = filepath.Join(root, "pods/abc123/volumeDevices")
	assert.Equal(t, exp, got)

	got = kubelet.getPodVolumeDeviceDir("abc123", "plugin")
	exp = filepath.Join(root, "pods/abc123/volumeDevices/plugin")
	assert.Equal(t, exp, got)

	got = kubelet.getPodPluginsDir("abc123")
	exp = filepath.Join(root, "pods/abc123/plugins")
	assert.Equal(t, exp, got)

	got = kubelet.getPodPluginDir("abc123", "foobar")
	exp = filepath.Join(root, "pods/abc123/plugins/foobar")
	assert.Equal(t, exp, got)

	got = kubelet.getVolumeDevicePluginsDir()
	exp = filepath.Join(root, "plugins")
	assert.Equal(t, exp, got)

	got = kubelet.getVolumeDevicePluginDir("foobar")
	exp = filepath.Join(root, "plugins", "foobar", "volumeDevices")
	assert.Equal(t, exp, got)

	got = kubelet.getPodContainerDir("abc123", "def456")
	exp = filepath.Join(root, "pods/abc123/containers/def456")
	assert.Equal(t, exp, got)

	got = kubelet.getPodResourcesDir()
	exp = filepath.Join(root, "pod-resources")
	assert.Equal(t, exp, got)

	got = kubelet.GetHostname()
	exp = "127.0.0.1"
	assert.Equal(t, exp, got)

	got = kubelet.getPodVolumeSubpathsDir("abc123")
	exp = filepath.Join(root, "pods/abc123/volume-subpaths")
	assert.Equal(t, exp, got)
}

func TestHandlerSupportsUserNamespaces(t *testing.T) {
	testKubelet := newTestKubelet(t, false /* controllerAttachDetachEnabled */)
	defer testKubelet.Cleanup()
	kubelet := testKubelet.kubelet

	kubelet.runtimeState.setRuntimeHandlers([]kubecontainer.RuntimeHandler{
		{
			Name:                   "has-support",
			SupportsUserNamespaces: true,
		},
		{
			Name:                   "has-no-support",
			SupportsUserNamespaces: false,
		},
	})

	got, err := kubelet.HandlerSupportsUserNamespaces("has-support")
	assert.True(t, got)
	assert.NoError(t, err)

	got, err = kubelet.HandlerSupportsUserNamespaces("has-no-support")
	assert.False(t, got)
	assert.NoError(t, err)

	got, err = kubelet.HandlerSupportsUserNamespaces("unknown")
	assert.False(t, got)
	assert.Error(t, err)
}
