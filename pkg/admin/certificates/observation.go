package certificates

import (
	"fmt"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
)

// Observation describes active on-disk certificates, not TLS identities in memory.
// Inventory and pending metadata can fail independently. A nil PendingError makes
// RenewalPending meaningful; it says nothing about whether activation will succeed.
type Observation struct {
	ObservedAt     time.Time
	Inventory      certchains.CertificateInventory
	InventoryError error
	RenewalPending bool
	PendingError   error
}

// Observe reads public certificates and pending metadata under the existing lock.
// It never generates certificates, reads private keys, or repairs transactions.
// cfg is the startup configuration; encoded certificate dates determine status.
func Observe(cfg *config.Config, dataDir, lockPath string, now time.Time) Observation {
	result := Observation{ObservedAt: now.UTC().Truncate(time.Second)}
	lock, err := ReadLock(lockPath)
	if err != nil {
		result.InventoryError, result.PendingError = err, err
		return result
	}
	defer func() { _ = lock.Close() }()

	transaction := &Transaction{DataDir: dataDir}
	interrupted, err := transaction.Pending()
	if err == nil && interrupted {
		err = fmt.Errorf("certificate activation requires recovery")
	}
	if err != nil {
		result.InventoryError, result.PendingError = err, err
		return result
	}

	builder, err := NewChains(cfg, dataDir)
	if err == nil {
		result.Inventory, err = builder.LoadInventory()
	}
	result.InventoryError = err
	metadata, err := transaction.LoadPendingMetadata()
	if err == nil && metadata != nil {
		_, err = ParsePendingRenewal(metadata)
		result.RenewalPending = err == nil
	}
	result.PendingError = err
	return result
}
