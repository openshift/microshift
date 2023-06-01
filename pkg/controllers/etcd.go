/*
Copyright © 2021 MicroShift Contributors

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
package controllers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	klog "k8s.io/klog/v2"

	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	HealthCheckRetries = 10
	HealthCheckWait    = 3 * time.Second
)

type EtcdService struct {
	memoryLimit uint64
}

func NewEtcd(cfg *config.Config) *EtcdService {
	return &EtcdService{
		memoryLimit: cfg.Etcd.MemoryLimitMB,
	}
}

func (s *EtcdService) Name() string           { return "etcd" }
func (s *EtcdService) Dependencies() []string { return []string{} }

func (s *EtcdService) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	// Check to see if we should run as a systemd run or directly as a binary.
	runningAsSvc := os.Getenv("INVOCATION_ID") != ""

	// Get the path to the etcd binary based on the MicroShift binary location.
	microshiftExecPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("%v failed to get exec path: %v", s.Name(), err)
	}
	etcdPath := filepath.Join(filepath.Dir(microshiftExecPath), "microshift-etcd")
	// Not running the etcd binary directly, the proper etcd arguments
	// are handled in etcd/cmd/microshift-etcd/run.go.
	args := []string{}

	// If we're launching MicroShift as a service, we need to do the
	// same with etcd, so wrap it in a transient systemd-unit that's
	// tied to the MicroShift service lifetime.
	var exe string
	if runningAsSvc {
		args = append(args,
			"--uid=root",
			"--scope",
			"--unit", "microshift-etcd",
			"--property", "BindsTo=microshift.service",
		)

		if s.memoryLimit > 0 {
			args = append(args, "--property", fmt.Sprintf("MemoryHigh=%vM", s.memoryLimit))
		}

		args = append(args, etcdPath)

		exe = "systemd-run"
	} else {
		exe = etcdPath
	}
	args = append(args, "run")
	// Not using context as canceling ctx sends SIGKILL to process
	klog.Infof("starting etcd via %s with args %v", exe, args)
	cmd := exec.Command(exe, args...)

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("%s failed to get workdir: %v", s.Name(), err)
	}
	cmd.Dir = wd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("%s failed to start: %v", s.Name(), err)
	}

	// Handle microshift-etcd termination before microshift process exits
	go func() {
		if err := cmd.Wait(); err != nil {
			klog.Warningf("%v failed waiting on process to finish: %+v", s.Name(), err)
		}
		klog.Infof("%v process quit: %v", s.Name(), cmd.ProcessState.String())

		// Exit microshift to trigger microshift-etcd restart
		klog.Warning("microshift-etcd process terminated prematurely, restarting MicroShift")
		os.Exit(0)
	}()

	// Handle microshift-etcd termination after microshift process exits
	defer func() {
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			klog.Warningf("%v failed interrupting the process: %+v", s.Name(), err)
			if err := cmd.Process.Kill(); err != nil {
				klog.Warningf("%v failed killing the process: %+v", s.Name(), err)
			}
		}
	}()

	if err := checkIfEtcdIsReady(ctx); err != nil {
		return err
	}
	klog.Info("etcd is ready!")
	close(ready)

	// Wait for MicroShift to be done
	<-ctx.Done()
	return ctx.Err()
}

func checkIfEtcdIsReady(ctx context.Context) error {
	client, err := getEtcdClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to obtain etcd client: %v", err)
	}
	defer client.Close()

	for i := 0; i < HealthCheckRetries; i++ {
		time.Sleep(HealthCheckWait)
		if _, err = client.Get(ctx, "health"); err == nil {
			return nil
		} else {
			klog.Infof("etcd not ready yet: %v", err)
		}
	}
	return fmt.Errorf("etcd still not healthy after checking %d times", HealthCheckRetries)
}

func getEtcdClient(ctx context.Context) (*clientv3.Client, error) {
	certsDir := cryptomaterial.CertsDirectory(config.DataDir)
	etcdAPIServerClientCertDir := cryptomaterial.EtcdAPIServerClientCertDir(certsDir)

	tlsInfo := transport.TLSInfo{
		CertFile:      cryptomaterial.ClientCertPath(etcdAPIServerClientCertDir),
		KeyFile:       cryptomaterial.ClientKeyPath(etcdAPIServerClientCertDir),
		TrustedCAFile: cryptomaterial.CACertPath(cryptomaterial.EtcdSignerDir(certsDir)),
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		return nil, err
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"https://localhost:2379"},
		DialTimeout: 5 * time.Second,
		TLS:         tlsConfig,
		Context:     ctx,
	})
	if err != nil {
		return nil, err
	}
	return cli, nil
}
