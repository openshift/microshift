package cmd

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"k8s.io/klog/v2"

	"github.com/openshift/microshift/pkg/admin/certificates"
	certificatesv1alpha1 "github.com/openshift/microshift/pkg/apis/certificates/v1alpha1"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
)

const invalidPendingRenewalWarning = "Pending certificate renewal metadata is invalid; the pending certificates will not be activated. Restart MicroShift to quarantine this renewal, then run microshift certs renew again."
const rejectedPendingRenewalWarning = "A pending certificate renewal was rejected because its metadata was invalid. The rejected certificates were not activated; normal certificate startup policy applies. Check active certificate status, then prepare and activate a replacement renewal to clear this warning."

// This private on-disk record contains hashes, never configuration or key data.
type pendingCertificateRenewal struct {
	Version      int                                            `json:"version"`
	ConfigHash   string                                         `json:"configHash"`
	OriginalHash string                                         `json:"originalHash"`
	RenewedHash  string                                         `json:"renewedHash"`
	Result       *certificatesv1alpha1.CertificateRenewalResult `json:"result"`
}

func certificateConfigHash(cfg *config.Config) (string, error) {
	// Only inputs used by certificateChainsSetup affect the prepared PKI.
	// Kubeconfigs and unrelated resources are rebuilt at activation.
	contents, err := json.Marshal(struct {
		ServiceNetwork   []string
		Hostname         string
		NodeIP           string
		BaseDomain       string
		AdvertiseAddress string
		SubjectAltNames  []string
		ServingValidity  time.Duration
		CAValidity       time.Duration
	}{
		ServiceNetwork:   cfg.Network.ServiceNetwork,
		Hostname:         cfg.Node.HostnameOverride,
		NodeIP:           cfg.Node.NodeIP,
		BaseDomain:       cfg.DNS.BaseDomain,
		AdvertiseAddress: cfg.ApiServer.AdvertiseAddress,
		SubjectAltNames:  cfg.ApiServer.SubjectAltNames,
		ServingValidity:  cfg.Certificates.ServingDuration(),
		CAValidity:       cfg.Certificates.CADuration(),
	})
	if err != nil {
		return "", invalidCertificateConfiguration()
	}
	return fmt.Sprintf("%x", sha256.Sum256(contents)), nil
}

func newPendingCertificateRenewal(cfg *config.Config, dataDir string) (*pendingCertificateRenewal, error) {
	configHash, err := certificateConfigHash(cfg)
	if err != nil {
		return nil, err
	}
	originalHash, err := certificates.DigestTree(filepath.Join(dataDir, "certs"))
	if err != nil {
		return nil, err
	}
	return &pendingCertificateRenewal{Version: 1, ConfigHash: configHash, OriginalHash: originalHash}, nil
}

func publishCertificateRenewal(transaction *certificates.Transaction, state *pendingCertificateRenewal) error {
	var err error
	state.RenewedHash, err = certificates.DigestTree(filepath.Join(transaction.Stage(), "certs"))
	if err != nil {
		return err
	}
	// Detect an external PKI change while generation was in progress too.
	currentHash, err := certificates.DigestTree(filepath.Join(transaction.DataDir, "certs"))
	if err != nil {
		return err
	}
	if currentHash != state.OriginalHash {
		return fmt.Errorf("active PKI changed while preparing renewal; retry renewal")
	}
	contents, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return transaction.Publish(contents)
}

func readPendingCertificateRenewal(transaction *certificates.Transaction) (*pendingCertificateRenewal, error) {
	contents, err := transaction.LoadPending()
	if err != nil || contents == nil {
		return nil, err
	}
	var state pendingCertificateRenewal
	if err := json.Unmarshal(contents, &state); err != nil {
		return nil, certificates.ErrInvalidPendingMetadata
	}
	if state.Version > 1 {
		return nil, fmt.Errorf("unsupported pending certificate renewal version %d", state.Version)
	}
	if state.Version != 1 || state.ConfigHash == "" || state.OriginalHash == "" || state.RenewedHash == "" || state.Result == nil ||
		state.Result.Status != certificatesv1alpha1.RenewalStatusPending ||
		(state.Result.Mode != certificatesv1alpha1.RenewalModeServing && state.Result.Mode != certificatesv1alpha1.RenewalModeCA) {
		return nil, certificates.ErrInvalidPendingMetadata
	}
	return &state, nil
}

func loadPendingCertificateRenewal(dataDir string) (*certificatesv1alpha1.CertificateRenewalResult, error) {
	state, err := readPendingCertificateRenewal(&certificates.Transaction{DataDir: dataDir})
	if err != nil || state == nil {
		return nil, err
	}
	return state.Result, nil
}

// loadCertificateRenewalStatus keeps active inventory available when only the
// pending JSON is corrupt. It is read-only, including before startup quarantine.
func loadCertificateRenewalStatus(dataDir string) (*certificatesv1alpha1.CertificateRenewalResult, []string, error) {
	pending, err := loadPendingCertificateRenewal(dataDir)
	if errors.Is(err, certificates.ErrInvalidPendingMetadata) {
		return nil, []string{invalidPendingRenewalWarning}, nil
	}
	if err != nil {
		return nil, nil, err
	}
	rejected, err := (&certificates.Transaction{DataDir: dataDir}).HasRejected()
	if err != nil {
		return nil, nil, err
	}
	if rejected {
		return pending, []string{rejectedPendingRenewalWarning}, nil
	}
	return pending, nil, nil
}

// activateCertificateRenewal runs only at startup, under both the runtime and
// operation locks, after restore and before any component reads active PKI.
// It reports whether a pending generation was activated, not merely present.
func activateCertificateRenewal(cfg *config.Config, transaction *certificates.Transaction, etcdStopped func() error, now time.Time) (bool, error) {
	state, err := readPendingCertificateRenewal(transaction)
	if errors.Is(err, certificates.ErrInvalidPendingMetadata) {
		if err := transaction.QuarantinePending(); err != nil {
			return false, fmt.Errorf("failed to quarantine invalid pending certificate renewal: %w", err)
		}
		klog.Warningf("%s Rejected files were quarantined under %s/.cert-renewal-rejected; continuing normal certificate initialization", rejectedPendingRenewalWarning, transaction.DataDir)
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if state == nil {
		return false, transaction.Discard() // Remove abandoned, unpublished staging.
	}
	return activatePendingCertificateRenewal(cfg, transaction, state, etcdStopped, now)
}

func activatePendingCertificateRenewal(cfg *config.Config, transaction *certificates.Transaction, state *pendingCertificateRenewal, etcdStopped func() error, now time.Time) (bool, error) {
	if err := etcdStopped(); err != nil {
		return false, err
	}
	current, err := newPendingCertificateRenewal(cfg, transaction.DataDir)
	if err != nil {
		return false, err
	}
	if current.ConfigHash != state.ConfigHash || current.OriginalHash != state.OriginalHash {
		return false, discardPendingCertificateRenewal(transaction, "certificate configuration or active PKI changed since renewal")
	}
	if err := validatePendingCertificateHash(transaction.Stage(), state.RenewedHash); err != nil {
		return false, err
	}
	renewCAs := state.Result.Mode == certificatesv1alpha1.RenewalModeCA
	builder, err := certificateChainsSetup(cfg, transaction.Stage())
	if err != nil {
		return false, err
	}
	// Recheck validity at activation, which may be much later than generation.
	if _, err := builder.ValidateRenewal(renewCAs, now); err != nil {
		if errors.Is(err, certchains.ErrCertificateExpired) {
			return false, discardPendingCertificateRenewal(transaction, "pending certificates expired before activation")
		}
		return false, fmt.Errorf("pending certificates cannot be activated; run microshift certs renew again: %w", err)
	}
	activeBuilder, err := certificateChainsSetup(cfg, transaction.DataDir)
	if err != nil {
		return false, err
	}
	before, err := activeBuilder.LoadInventory()
	if err != nil {
		return false, err
	}
	if err := transaction.RefreshResources(); err != nil {
		return false, err
	}
	chains, err := builder.Load()
	if err != nil {
		return false, err
	}
	if err := initKubeconfigs(cfg, chains, transaction.Stage()); err != nil {
		return false, err
	}
	if err := validatePendingCertificateHash(transaction.Stage(), state.RenewedHash); err != nil {
		return false, err
	}
	if _, err := validateCertificateRenewal(cfg, transaction.Stage(), before, renewCAs); err != nil {
		return false, err
	}
	err = transaction.Commit(func() error {
		if _, err := validateCertificateRenewal(cfg, transaction.DataDir, before, renewCAs); err != nil {
			return err
		}
		return validatePendingCertificateHash(transaction.DataDir, state.RenewedHash)
	})
	if err == nil {
		if cleanupErr := transaction.ClearRejected(); cleanupErr != nil {
			// Activation succeeded. A stale warning must not fail startup.
			klog.Warningf("Certificate renewal activated, but rejected renewal cleanup failed; the status warning may remain: %v", cleanupErr)
		}
	}
	return err == nil, err
}

func discardPendingCertificateRenewal(transaction *certificates.Transaction, reason string) error {
	if err := transaction.DiscardPending(); err != nil {
		return fmt.Errorf("failed to discard pending certificate renewal: %w", err)
	}
	klog.Warningf("Discarded pending certificate renewal: %s; continuing normal certificate initialization. Run microshift certs renew again if needed", reason)
	return nil
}

func validatePendingCertificateHash(dataDir, expected string) error {
	digest, err := certificates.DigestTree(filepath.Join(dataDir, "certs"))
	if err != nil {
		return err
	}
	if digest != expected {
		return fmt.Errorf("pending certificate material changed; run microshift certs renew again")
	}
	return nil
}

func loadActivatedCertificates(cfg *config.Config, dataDir string) (*certchains.CertificateChains, error) {
	builder, err := certificateChainsSetup(cfg, dataDir)
	if err != nil {
		return nil, err
	}
	chains, err := builder.Load()
	if err != nil {
		return nil, err
	}
	cfg.Ingress.ServingCertificate, cfg.Ingress.ServingKey, err = chains.GetCertKey("ingress-ca", "router-default-serving")
	return chains, err
}
