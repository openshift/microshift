package main

import (
	"testing"

	"github.com/openshift/microshift/pkg/config"
	etcd "go.etcd.io/etcd/server/v3/embed"
)

func TestConfigureWarningUnaryRequestDuration(t *testing.T) {
	s := NewEtcd(config.NewDefault())

	if got, want := s.etcdCfg.WarningUnaryRequestDuration, etcd.DefaultWarningUnaryRequestDuration; got != want {
		t.Errorf("WarningUnaryRequestDuration = %v, want %v", got, want)
	}
}
