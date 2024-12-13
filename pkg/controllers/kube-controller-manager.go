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
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	embedded "github.com/openshift/microshift/assets"
	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"

	kubecontrolplanev1 "github.com/openshift/api/kubecontrolplane/v1"
	"github.com/openshift/library-go/pkg/operator/resource/resourcemerge"
	kubecm "k8s.io/kubernetes/cmd/kube-controller-manager/app"

	klog "k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	kcmDefaultConfigAsset = "controllers/kube-controller-manager/defaultconfig.yaml"
)

type KubeControllerManager struct {
	args    []string
	applyFn func() error

	// TODO: report configuration errors immediately
	configureErr error
}

func NewKubeControllerManager(ctx context.Context, cfg *config.Config) *KubeControllerManager {
	s := &KubeControllerManager{}
	// TODO: manage and invoke the configure bits independently outside of this.
	s.args, s.applyFn, s.configureErr = configure(ctx, cfg)
	return s
}

func (s *KubeControllerManager) Name() string           { return "kube-controller-manager" }
func (s *KubeControllerManager) Dependencies() []string { return []string{"kube-apiserver"} }

func kcmRootCAFile() string {
	certsDir := cryptomaterial.CertsDirectory(config.DataDir)
	return cryptomaterial.ServiceAccountTokenCABundlePath(certsDir)
}

func kcmClusterSigningCertKeyAndFile() (string, string) {
	certsDir := cryptomaterial.CertsDirectory(config.DataDir)
	csrSignerDir := cryptomaterial.CSRSignerCertDir(certsDir)
	return cryptomaterial.CAKeyPath(csrSignerDir), cryptomaterial.CACertPath(csrSignerDir)
}

func kcmServiceAccountPrivateKeyFile() string {
	return filepath.Join(config.DataDir, "/resources/kube-apiserver/secrets/service-account-key/service-account.key")
}

func configure(ctx context.Context, cfg *config.Config) (args []string, applyFn func() error, err error) {
	kubeConfig := cfg.KubeConfigPath(config.KubeControllerManager)
	clusterSigningKey, clusterSigningCert := kcmClusterSigningCertKeyAndFile()

	overrides := &kubecontrolplanev1.KubeControllerManagerConfig{
		ExtendedArguments: map[string]kubecontrolplanev1.Arguments{
			"kubeconfig":                       {kubeConfig},
			"authentication-kubeconfig":        {kubeConfig},
			"authorization-kubeconfig":         {kubeConfig},
			"service-account-private-key-file": {kcmServiceAccountPrivateKeyFile()},
			"allocate-node-cidrs":              {"true"},
			"cluster-cidr":                     {strings.Join(cfg.Network.ClusterNetwork, ",")},
			"service-cluster-ip-range":         {strings.Join(cfg.Network.ServiceNetwork, ",")},
			"root-ca-file":                     {kcmRootCAFile()},
			"secure-port":                      {"10257"},
			"leader-elect":                     {"false"},
			"use-service-account-credentials":  {"true"},
			"cluster-signing-cert-file":        {clusterSigningCert},
			"cluster-signing-key-file":         {clusterSigningKey},
			"v":                                {strconv.Itoa(cfg.GetVerbosity())},
			"tls-cipher-suites":                {strings.Join(cfg.ApiServer.TLS.CipherSuites, ",")},
			"tls-min-version":                  {cfg.ApiServer.TLS.MinVersion},
		},
	}

	args, err = mergeAndConvertToArgs(overrides)
	applyFn = func() error {
		return assets.ApplyNamespaces(ctx, []string{
			"controllers/kube-controller-manager/namespace-openshift-kube-controller-manager.yaml",
			"core/namespace-openshift-infra.yaml",
		}, cfg.KubeConfigPath(config.KubeAdmin))
	}
	return args, applyFn, err
}

func (s *KubeControllerManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	if s.configureErr != nil {
		return fmt.Errorf("configuration failed: %w", s.configureErr)
	}

	defer close(stopped)
	errorChannel := make(chan error, 1)

	// run readiness check
	go func() {
		// This endpoint uses a self-signed certificate on purpose, we need to skip verification.
		healthcheckStatus := util.RetryInsecureGet(ctx, "https://localhost:10257/healthz")
		if healthcheckStatus != 200 {
			klog.Errorf("kube-controller-manager failed to start")
			errorChannel <- errors.New("kube-controller-manager failed to start")
		}

		klog.Infof("%s is ready", s.Name())
		close(ready)
	}()

	// Carrying a patch for NewControllerManagerCommand to use cmd.Context().Done()
	// as the stop channel instead of the channel returned by SetupSignalHandler,
	// which expects to be called at most once in a process.
	cmd := kubecm.NewControllerManagerCommand()
	cmd.SetArgs(s.args)

	panicChannel := make(chan any, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicChannel <- r
			}
		}()
		errorChannel <- cmd.ExecuteContext(ctx)
	}()

	if err := s.applyFn(); err != nil {
		return fmt.Errorf("failed to apply openshift namespaces: %w", err)
	}

	select {
	case err := <-errorChannel:
		return err
	case perr := <-panicChannel:
		panic(perr)
	}
}

func mergeAndConvertToArgs(overrides *kubecontrolplanev1.KubeControllerManagerConfig) ([]string, error) {
	defaultConfigBytes, err := embedded.Asset(kcmDefaultConfigAsset)
	if err != nil {
		return nil, fmt.Errorf("invalid asset: %q, error: %w", kcmDefaultConfigAsset, err)
	}
	overridesBytes, err := json.Marshal(overrides)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal KubeControllerManagerConfig, error: %w", err)
	}
	mergedBytes, err := resourcemerge.MergePrunedProcessConfig(
		&kubecontrolplanev1.KubeControllerManagerConfig{}, nil, defaultConfigBytes, overridesBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to merge kube-controller-manager configuration: error: %w", err)
	}

	var kubeControllerManagerConfig map[string]interface{}
	if err := yaml.Unmarshal(mergedBytes, &kubeControllerManagerConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal the kube-controller-manager config: %v", err)
	}
	return GetKubeControllerManagerArgs(kubeControllerManagerConfig), nil
}

// This is a straight copy from KCM operator repo
func GetKubeControllerManagerArgs(config map[string]interface{}) []string {
	extendedArguments, ok := config["extendedArguments"]
	if !ok || extendedArguments == nil {
		return nil
	}
	args := []string{}
	for key, value := range extendedArguments.(map[string]interface{}) {
		for _, arrayValue := range value.([]interface{}) {
			if len(key) == 1 {
				args = append(args, fmt.Sprintf("-%s=%s", key, arrayValue.(string)))
			} else {
				args = append(args, fmt.Sprintf("--%s=%s", key, arrayValue.(string)))
			}
		}
	}
	// make sure to sort the arguments, otherwise we might get mismatch
	// when comparing revisions leading to new ones being created, unnecessarily
	sort.Strings(args)
	return args
}
