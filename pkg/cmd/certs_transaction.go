package cmd

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/openshift/library-go/pkg/crypto"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openshift/microshift/pkg/admin/certificates"
	certificatesv1alpha1 "github.com/openshift/microshift/pkg/apis/certificates/v1alpha1"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
)

// Keep the lock outside DataDir, which startup may replace during restore.
const certificateLockPath = config.BackupsDir + "/certs.lock"

func checkCertificateStartup(transaction *certificates.Transaction) error {
	if pending, err := transaction.Pending(); err != nil {
		return err
	} else if pending {
		return fmt.Errorf("interrupted certificate renewal: stop MicroShift and run microshift certs status to recover before starting")
	}
	return nil
}

func prepareCertificateAccess(writing bool) (func(), error) {
	transaction := &certificates.Transaction{DataDir: config.DataDir}
	return prepareCertificateTransaction(transaction, certificateLockPath, writing, certificateServicesStopped)
}

func prepareCertificateTransaction(transaction *certificates.Transaction, lockPath string, writing bool, stopped func() error) (func(), error) {
	if writing {
		if err := stopped(); err != nil {
			return nil, err
		}
	}
	lock, err := certificates.Lock(lockPath, writing)
	if err != nil {
		return nil, &certificateCommandError{certificatesv1alpha1.ErrorCodeRenewalFailed, err}
	}
	release := func() { _ = lock.Close() }
	pending, err := transaction.Pending()
	if err != nil {
		release()
		return nil, &certificateCommandError{certificatesv1alpha1.ErrorCodeRecoveryFailed, err}
	}
	if !pending {
		return release, nil
	}
	if !writing {
		release()
		lock, err = certificates.Lock(lockPath, true)
		if err != nil {
			return nil, &certificateCommandError{certificatesv1alpha1.ErrorCodeRecoveryFailed, err}
		}
		if err := stopped(); err != nil {
			release()
			return nil, &certificateCommandError{certificatesv1alpha1.ErrorCodeRecoveryFailed, err}
		}
	}
	if err := transaction.Recover(); err != nil {
		release()
		return nil, &certificateCommandError{certificatesv1alpha1.ErrorCodeRecoveryFailed, err}
	}
	return release, nil
}

func certificateServicesStopped() error {
	for _, service := range []string{"microshift.service", "microshift-etcd.scope"} {
		out, err := exec.Command("systemctl", "show", "-p", "ActiveState", "--value", service).Output()
		if err != nil {
			return &certificateCommandError{certificatesv1alpha1.ErrorCodeInternalError,
				fmt.Errorf("cannot determine whether %s is stopped", service)}
		}
		state := strings.TrimSpace(string(out))
		if state != "inactive" && state != systemdStateFailed {
			return &certificateCommandError{certificatesv1alpha1.ErrorCodeMicroShiftRunning,
				fmt.Errorf("MicroShift must be stopped before certificates can be renewed (%s is %s)", service, state)}
		}
	}
	return nil
}

func applyCertificateRenewal(cfg *config.Config, dataDir string, renewCAs bool) (inventory certchains.CertificateInventory, retErr error) {
	transaction := &certificates.Transaction{DataDir: dataDir}
	defer func() { retErr = errors.Join(retErr, transaction.Discard()) }()
	if err := transaction.Prepare(); err != nil {
		return nil, err
	}
	stage := transaction.Stage()
	builder, err := certificateChainsSetup(cfg, stage)
	if err != nil {
		return nil, err
	}
	before, err := builder.LoadInventory()
	if err != nil {
		return nil, err
	}
	// Existing PKI writers warn on stderr about long-lived certificates. The
	// command owns its output contract, so keep those internal diagnostics out
	// of versioned results (the same convention used by Complete).
	if err := quietCertificateGeneration(func() error {
		chains, err := builder.Complete()
		if err != nil {
			return err
		}
		for _, entry := range before {
			selected := renewCAs && len(entry.Path) == 1 || !renewCAs && entry.Role != certchains.CertificateRoleCA
			if selected {
				if err := chains.RegenerateForRenewal(entry.Path...); err != nil {
					return err
				}
			}
		}
		return initKubeconfigs(cfg, chains, stage)
	}); err != nil {
		return nil, err
	}
	staged, err := validateCertificateRenewal(cfg, stage, before, renewCAs)
	if err != nil {
		return nil, err
	}
	if err := transaction.Commit(func() error {
		committed, err := validateCertificateRenewal(cfg, dataDir, before, renewCAs)
		if err != nil {
			return err
		}
		if len(committed) != len(staged) {
			return fmt.Errorf("committed inventory differs from validated staging")
		}
		for i := range staged {
			if !bytes.Equal(staged[i].Certificate.Raw, committed[i].Certificate.Raw) {
				return fmt.Errorf("committed certificate %q differs from validated staging", staged[i].Name)
			}
		}
		inventory = committed
		return nil
	}); err != nil {
		return nil, err
	}
	return inventory, nil
}

func quietCertificateGeneration(generate func() error) error {
	null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	original := os.Stderr
	os.Stderr = null
	defer func() { os.Stderr = original; _ = null.Close() }()
	return generate()
}

func validateCertificateRenewal(cfg *config.Config, dataDir string, before certchains.CertificateInventory, renewCAs bool) (certchains.CertificateInventory, error) {
	builder, err := certificateChainsSetup(cfg, dataDir)
	if err != nil {
		return nil, err
	}
	after, err := builder.ValidateRenewal(renewCAs, time.Now())
	if err != nil {
		return nil, err
	}
	if len(before) != len(after) {
		return nil, fmt.Errorf("certificate inventory changed during renewal")
	}
	selected := make(certchains.CertificateInventory, 0, len(after))
	for i, entry := range after {
		if entry.Name != before[i].Name || entry.ParentCA != before[i].ParentCA {
			return nil, fmt.Errorf("certificate identity changed during renewal")
		}
		changed := !bytes.Equal(entry.Certificate.Raw, before[i].Certificate.Raw)
		if renewCAs || entry.Role != certchains.CertificateRoleCA {
			if !changed {
				return nil, fmt.Errorf("selected certificate %q was not renewed", entry.Name)
			}
			selected = append(selected, entry)
		} else if changed {
			return nil, fmt.Errorf("leaf renewal unexpectedly changed CA %q", entry.Name)
		}
	}
	if err := validateRenewedKubeconfigs(dataDir, after); err != nil {
		return nil, err
	}
	return selected, nil
}

func validateRenewedKubeconfigs(dataDir string, inventory certchains.CertificateInventory) error {
	resources := filepath.Join(dataDir, "resources")
	for _, id := range []config.KubeConfigID{config.KubeAdmin, config.KubeControllerManager, config.KubeScheduler,
		config.Kubelet, config.ClusterPolicyController, config.RouteControllerManager, config.ObservabilityClient} {
		if err := validateRenewedKubeconfig(filepath.Join(resources, string(id), "kubeconfig"), inventory); err != nil {
			return err
		}
	}
	// Administrator kubeconfigs also have one copy per serving name. Do not
	// interpret unrelated component resources as kubeconfigs owned by this CLI.
	return filepath.WalkDir(filepath.Join(resources, string(config.KubeAdmin)), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() != "kubeconfig" {
			return nil
		}
		return validateRenewedKubeconfig(path, inventory)
	})
}

func validateRenewedKubeconfig(path string, inventory certchains.CertificateInventory) error {
	cfg, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return fmt.Errorf("cannot read generated kubeconfig %q", path)
	}
	if err := clientcmd.Validate(*cfg); err != nil {
		return fmt.Errorf("invalid generated kubeconfig %q", path)
	}
	for _, auth := range cfg.AuthInfos {
		pair, err := tls.X509KeyPair(auth.ClientCertificateData, auth.ClientKeyData)
		if err != nil || len(pair.Certificate) == 0 || !inventoryContainsCertificate(inventory, pair.Certificate[0]) {
			return fmt.Errorf("generated kubeconfig %q does not contain a current client certificate/key pair", path)
		}
	}
	for _, cluster := range cfg.Clusters {
		if len(cluster.CertificateAuthorityData) == 0 {
			continue // Named/custom serving certificates use the system trust store.
		}
		certs, err := crypto.CertsFromPEM(cluster.CertificateAuthorityData)
		if err != nil {
			return fmt.Errorf("invalid trust bundle in generated kubeconfig %q", path)
		}
		for _, cert := range certs {
			if !inventoryContainsCertificate(inventory, cert.Raw) || !cert.IsCA {
				return fmt.Errorf("stale trust bundle in generated kubeconfig %q", path)
			}
		}
	}
	return nil
}

func inventoryContainsCertificate(inventory certchains.CertificateInventory, raw []byte) bool {
	for _, entry := range inventory {
		if bytes.Equal(entry.Certificate.Raw, raw) {
			return true
		}
	}
	return false
}
