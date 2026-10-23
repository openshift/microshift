package certificates

import (
	"encoding/json"
	"fmt"

	certificatesv1alpha1 "github.com/openshift/microshift/pkg/apis/certificates/v1alpha1"
)

// PendingRenewal is the private on-disk record, not a versioned public API.
// It contains hashes, never configuration or key data.
type PendingRenewal struct {
	Version      int                                            `json:"version"`
	ConfigHash   string                                         `json:"configHash"`
	OriginalHash string                                         `json:"originalHash"`
	RenewedHash  string                                         `json:"renewedHash"`
	Result       *certificatesv1alpha1.CertificateRenewalResult `json:"result"`
}

// ParsePendingRenewal validates the shared private metadata format without reading PKI.
func ParsePendingRenewal(contents []byte) (*PendingRenewal, error) {
	var state PendingRenewal
	if err := json.Unmarshal(contents, &state); err != nil {
		return nil, ErrInvalidPendingMetadata
	}
	if state.Version > 1 {
		return nil, fmt.Errorf("unsupported pending certificate renewal version %d", state.Version)
	}
	if state.Version != 1 || state.ConfigHash == "" || state.OriginalHash == "" || state.RenewedHash == "" || state.Result == nil ||
		state.Result.Status != certificatesv1alpha1.RenewalStatusPending ||
		(state.Result.Mode != certificatesv1alpha1.RenewalModeServing && state.Result.Mode != certificatesv1alpha1.RenewalModeCA) {
		return nil, ErrInvalidPendingMetadata
	}
	return &state, nil
}
