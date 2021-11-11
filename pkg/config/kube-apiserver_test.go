// kube-apiserver_test.go
package config

import (
	"errors"
	"os"
	"testing"
)

// test that kubeAPIAuditPolicyFile is successfully created at the directory /tmp/audit-policy.yaml
func TestKubeAPIAuditPolicyFile(t *testing.T) {
	var kubeAPIAuditPolicyFileTests = []struct {
		path string
		err  error
	}{
		{"/tmp/audit-policy.yaml", nil},
		{"/dev/null", nil},
		{"/this/directory/doesnt/exist/unless/it/does", errors.New("success")},
	}
	os.MkdirAll("/tmp", os.FileMode(0755))
	for _, tt := range kubeAPIAuditPolicyFileTests {
		err := kubeAPIAuditPolicyFile(tt.path)
		if (err != nil && tt.err == nil) || (err == nil && tt.err != nil) {
			t.Errorf("%+v: does not match %+v\n", tt, err)
		}
	}
}

// test that kubeAPIOAuthMetadataFile is successfully created at the directory /tmp/oauth-metadata.json
func TestKubeAPIOAuthMetadataFile(t *testing.T) {
	var MetadataFileTests = []struct {
		path string
		err  error
	}{
		{"/tmp/oauth-metadata.json", nil},
		{"/dev/null", nil},
		{"/this/directory/doesnt/exist/unless/it/does/oauth.json", errors.New("success")},
	}
	os.MkdirAll("/tmp", os.FileMode(0755))
	for _, tt := range MetadataFileTests {
		err := kubeAPIOAuthMetadataFile(tt.path)
		if (err != nil && tt.err == nil) || (err == nil && tt.err != nil) {
			t.Errorf("%+v: does not match %+v\n", tt, err)
		}
	}
}

// test that KubeAPIServerConfig can successfully create the necessary files when a valid MicroshiftConfig is provided
func TestKubeAPIServerConfig(t *testing.T) {
	os.MkdirAll("/tmp", os.FileMode(0755))
	var KubeAPIServerConfigTests = []struct {
		config MicroshiftConfig
		err    error
	}{
		{MicroshiftConfig{ConfigFile: "/tmp/config.yaml", DataDir: "/tmp/data/"}, nil},
		{*NewMicroshiftConfig(), nil},
		{MicroshiftConfig{}, errors.New("config is empty")},
	}
	for _, tt := range KubeAPIServerConfigTests {
		err := KubeAPIServerConfig(&tt.config)
		if (err != nil && tt.err == nil) || (err == nil && tt.err != nil) {
			t.Errorf("%+v: does not match %+v\n", tt, err)
		}
	}
}
