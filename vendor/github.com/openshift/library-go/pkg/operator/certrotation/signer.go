package certrotation

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/openshift/library-go/pkg/crypto"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1informers "k8s.io/client-go/informers/core/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
)

// RotatedSigningCASecret rotates a self-signed signing CA stored in a secret. It creates a new one when
// - refresh duration is over
// - or 80% of validity is over (if RefreshOnlyWhenExpired is false)
// - or the CA is expired.
type RotatedSigningCASecret struct {
	// Namespace is the namespace of the Secret.
	Namespace string
	// Name is the name of the Secret.
	Name string
	// Validity is the duration from time.Now() until the signing CA expires. If RefreshOnlyWhenExpired
	// is false, the signing cert is rotated when 80% of validity is reached.
	Validity time.Duration
	// Refresh is the duration after signing CA creation when it is rotated at the latest. It is ignored
	// if RefreshOnlyWhenExpired is true, or if Refresh > Validity.
	Refresh time.Duration
	// RefreshOnlyWhenExpired set to true means to ignore 80% of validity and the Refresh duration for rotation,
	// but only rotate when the signing CA expires. This is useful for auto-recovery when we want to enforce
	// rotation on expiration only, but not interfere with the ordinary rotation controller.
	RefreshOnlyWhenExpired bool

	// Owner is an optional reference to add to the secret that this rotator creates. Use this when downstream
	// consumers of the signer CA need to be aware of changes to the object.
	// WARNING: be careful when using this option, as deletion of the owning object will cascade into deletion
	// of the signer. If the lifetime of the owning object is not a superset of the lifetime in which the signer
	// is used, early deletion will be catastrophic.
	Owner *metav1.OwnerReference

	// AdditionalAnnotations is a collection of annotations set for the secret
	AdditionalAnnotations AdditionalAnnotations

	// Plumbing:
	Informer      corev1informers.SecretInformer
	Lister        corev1listers.SecretLister
	Client        corev1client.SecretsGetter
	EventRecorder events.Recorder

	// Deprecated: DO NOT enable, it is intended as a short term hack for a very specific use case,
	// and it works in tandem with a particular carry patch applied to the openshift kube-apiserver.
	// we will remove this when we migrate all of the affected secret
	// objects to their intended type: https://issues.redhat.com/browse/API-1800
	UseSecretUpdateOnly bool
}

// EnsureSigningCertKeyPair manages the entire lifecycle of a signer cert as a secret, from creation to continued rotation.
// It always returns the currently used CA pair, a bool indicating whether it was created/updated within this function call and an error.
func (c RotatedSigningCASecret) EnsureSigningCertKeyPair(ctx context.Context) (*crypto.CA, bool, error) {
	originalSigningCertKeyPairSecret, err := c.Lister.Secrets(c.Namespace).Get(c.Name)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, false, err
	}
	signingCertKeyPairSecret := originalSigningCertKeyPairSecret.DeepCopy()
	if apierrors.IsNotFound(err) {
		// create an empty one
		signingCertKeyPairSecret = &corev1.Secret{
			ObjectMeta: NewTLSArtifactObjectMeta(
				c.Name,
				c.Namespace,
				c.AdditionalAnnotations,
			),
			Type: corev1.SecretTypeTLS,
		}
	}

	applyFn := resourceapply.ApplySecret
	if c.UseSecretUpdateOnly {
		applyFn = resourceapply.ApplySecretDoNotUse
	}

	// apply necessary metadata (possibly via delete+recreate) if secret exists
	// this is done before content update to prevent unexpected rollouts
	if ensureMetadataUpdate(signingCertKeyPairSecret, c.Owner, c.AdditionalAnnotations) && ensureSecretTLSTypeSet(signingCertKeyPairSecret) {
		actualSigningCertKeyPairSecret, _, err := applyFn(ctx, c.Client, c.EventRecorder, signingCertKeyPairSecret)
		if err != nil {
			return nil, false, err
		}
		signingCertKeyPairSecret = actualSigningCertKeyPairSecret
	}

	signerUpdated := false
	if needed, reason := needNewSigningCertKeyPair(signingCertKeyPairSecret.Annotations, c.Refresh, c.RefreshOnlyWhenExpired); needed {
		c.EventRecorder.Eventf("SignerUpdateRequired", "%q in %q requires a new signing cert/key pair: %v", c.Name, c.Namespace, reason)
		if err := setSigningCertKeyPairSecret(signingCertKeyPairSecret, c.Validity); err != nil {
			return nil, false, err
		}

		LabelAsManagedSecret(signingCertKeyPairSecret, CertificateTypeSigner)

		actualSigningCertKeyPairSecret, _, err := applyFn(ctx, c.Client, c.EventRecorder, signingCertKeyPairSecret)
		if err != nil {
			return nil, false, err
		}
		signingCertKeyPairSecret = actualSigningCertKeyPairSecret
		signerUpdated = true
	}
	// at this point, the secret has the correct signer, so we should read that signer to be able to sign
	signingCertKeyPair, err := crypto.GetCAFromBytes(signingCertKeyPairSecret.Data["tls.crt"], signingCertKeyPairSecret.Data["tls.key"])
	if err != nil {
		return nil, signerUpdated, err
	}

	return signingCertKeyPair, signerUpdated, nil
}

// ensureOwnerReference adds the owner to the list of owner references in meta, if necessary
func ensureOwnerReference(meta *metav1.ObjectMeta, owner *metav1.OwnerReference) bool {
	var found bool
	for _, ref := range meta.OwnerReferences {
		if ref == *owner {
			found = true
			break
		}
	}
	if !found {
		meta.OwnerReferences = append(meta.OwnerReferences, *owner)
		return true
	}
	return false
}

func needNewSigningCertKeyPair(annotations map[string]string, refresh time.Duration, refreshOnlyWhenExpired bool) (bool, string) {
	notBefore, notAfter, reason := getValidityFromAnnotations(annotations)
	if len(reason) > 0 {
		return true, reason
	}

	if time.Now().After(notAfter) {
		return true, "already expired"
	}

	if refreshOnlyWhenExpired {
		return false, ""
	}

	validity := notAfter.Sub(notBefore)
	at80Percent := notAfter.Add(-validity / 5)
	if time.Now().After(at80Percent) {
		return true, fmt.Sprintf("past its latest possible time %v", at80Percent)
	}

	developerSpecifiedRefresh := notBefore.Add(refresh)
	if time.Now().After(developerSpecifiedRefresh) {
		return true, fmt.Sprintf("past its refresh time %v", developerSpecifiedRefresh)
	}

	return false, ""
}

func getValidityFromAnnotations(annotations map[string]string) (notBefore time.Time, notAfter time.Time, reason string) {
	notAfterString := annotations[CertificateNotAfterAnnotation]
	if len(notAfterString) == 0 {
		return notBefore, notAfter, "missing notAfter"
	}
	notAfter, err := time.Parse(time.RFC3339, notAfterString)
	if err != nil {
		return notBefore, notAfter, fmt.Sprintf("bad expiry: %q", notAfterString)
	}
	notBeforeString := annotations[CertificateNotBeforeAnnotation]
	if len(notAfterString) == 0 {
		return notBefore, notAfter, "missing notBefore"
	}
	notBefore, err = time.Parse(time.RFC3339, notBeforeString)
	if err != nil {
		return notBefore, notAfter, fmt.Sprintf("bad expiry: %q", notBeforeString)
	}

	return notBefore, notAfter, ""
}

// setSigningCertKeyPairSecret creates a new signing cert/key pair and sets them in the secret
func setSigningCertKeyPairSecret(signingCertKeyPairSecret *corev1.Secret, validity time.Duration) error {
	signerName := fmt.Sprintf("%s_%s@%d", signingCertKeyPairSecret.Namespace, signingCertKeyPairSecret.Name, time.Now().Unix())
	ca, err := crypto.MakeSelfSignedCAConfigForDuration(signerName, validity)
	if err != nil {
		return err
	}

	certBytes := &bytes.Buffer{}
	keyBytes := &bytes.Buffer{}
	if err := ca.WriteCertConfig(certBytes, keyBytes); err != nil {
		return err
	}

	if signingCertKeyPairSecret.Annotations == nil {
		signingCertKeyPairSecret.Annotations = map[string]string{}
	}
	if signingCertKeyPairSecret.Data == nil {
		signingCertKeyPairSecret.Data = map[string][]byte{}
	}
	signingCertKeyPairSecret.Data["tls.crt"] = certBytes.Bytes()
	signingCertKeyPairSecret.Data["tls.key"] = keyBytes.Bytes()
	signingCertKeyPairSecret.Annotations[CertificateNotAfterAnnotation] = ca.Certs[0].NotAfter.Format(time.RFC3339)
	signingCertKeyPairSecret.Annotations[CertificateNotBeforeAnnotation] = ca.Certs[0].NotBefore.Format(time.RFC3339)
	signingCertKeyPairSecret.Annotations[CertificateIssuer] = ca.Certs[0].Issuer.CommonName

	return nil
}
