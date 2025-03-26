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
	"errors"
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

	// Obtain the executable path
	etcdPath, err := getEtcdServicePath()
	if err != nil {
		return err
	}

	// If MicroShift is launched as a service, etcd should be launched in the
	// same manner. This is done by wrapping etcd in a transient systemd-unit
	// that is tied to the MicroShift service lifetime.
	var exe string
	var args []string
	if runningAsSvc {
		if err := stopMicroshiftEtcdScopeIfExists(); err != nil {
			return err
		}

		// Append systemd arguments
		args = append(args,
			"--uid=root",
			"--scope",
			"--collect",
			"--unit", "microshift-etcd",
			"--property", "Before=microshift.service",
			"--property", "BindsTo=microshift.service",
		)
		if s.memoryLimit > 0 {
			args = append(args, "--property", fmt.Sprintf("MemoryHigh=%vM", s.memoryLimit))
		}
		// Mark the end of the systemd-run options
		args = append(args, "--")
		// Put the etcd executable as an argument to systemd
		args = append(args, etcdPath)

		// Select the systemd runner executable
		exe = "systemd-run"
	} else {
		// Select the etcd executable
		exe = etcdPath
	}

	// Append the etcd service arguments
	etcdArgs := getEtcdServiceArgs()
	args = append(args, etcdArgs...)

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

		if !errors.Is(ctx.Err(), context.Canceled) {
			// Exit microshift to trigger microshift-etcd restart
			klog.Warning("microshift-etcd process terminated prematurely, restarting MicroShift")
			os.Exit(0)
		} else {
			klog.Info("MicroShift is mid shutdown - ignoring etcd termination")
		}
	}()

	// Ensures microshift-etcd unit stopped after microshift
	defer func() {
		klog.Info("stopping microshift-etcd")
		cmd := exec.Command("systemctl", "stop", "microshift-etcd.scope", "--no-block")

		if out, err := cmd.CombinedOutput(); err != nil {
			klog.ErrorS(err, "failed to stop microshift-etcd", "output", string(out))
			return
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

func getEtcdServicePath() (string, error) {
	etcdPath := os.Getenv("MICROSHIFT_ETCD_SERVICE_PATH")
	if etcdPath == "" {
		// Get the path to the etcd binary based on the MicroShift binary location
		microshiftExecPath, err := os.Executable()
		if err != nil {
			return "", fmt.Errorf("failed to get exec path: %v", err)
		}
		return filepath.Join(filepath.Dir(microshiftExecPath), "microshift-etcd"), nil
	}
	return etcdPath, nil
}

func getEtcdServiceArgs() []string {
	args := []string{}

	etcdArgs := os.Getenv("MICROSHIFT_ETCD_SERVICE_ARGS")
	if etcdArgs == "" {
		// The proper etcd arguments are handled in etcd/cmd/microshift-etcd/run.go.
		args = append(args, "run")
	} else {
		// Non-default backend arguments should all be handled by it
	}
	return args
}

func stopMicroshiftEtcdScopeIfExists() error {
	// There are several codes that systemctl can return like
	// 0 - unit is active, 3 - unit is not active, 4 - no such unit.
	// Because microshift-etcd.scope is transient unit it's either active or doesn't exist,
	// just check for active (existing) to simplify procedure.
	statusCmd := exec.Command("systemctl", "status", "microshift-etcd.scope")
	if err := statusCmd.Run(); err != nil {
		// nolint:nilerr
		return nil
	}

	klog.InfoS("microshift-etcd.scope is already active - stopping")
	stopCmd := exec.Command("systemctl", "stop", "microshift-etcd.scope")
	if out, err := stopCmd.CombinedOutput(); err != nil {
		klog.ErrorS(err, "failed to stop microshift-etcd", "output", string(out))
		return err
	}
	return nil
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
			if err == context.Canceled {
				return err
			}
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
