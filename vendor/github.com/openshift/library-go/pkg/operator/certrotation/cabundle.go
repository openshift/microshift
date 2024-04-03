package certrotation

import (
	"bytes"
	"context"
	"crypto/x509"
	"fmt"
	"reflect"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1informers "k8s.io/client-go/informers/core/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/util/cert"
	"k8s.io/klog/v2"

	"github.com/openshift/library-go/pkg/certs"
	"github.com/openshift/library-go/pkg/crypto"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
)

// CABundleConfigMap maintains a CA bundle config map, by adding new CA certs coming from RotatedSigningCASecret, and by removing expired old ones.
type CABundleConfigMap struct {
	// Namespace is the namespace of the ConfigMap to maintain.
	Namespace string
	// Name is the name of the ConfigMap to maintain.
	Name string
	// Owner is an optional reference to add to the secret that this rotator creates.
	Owner *metav1.OwnerReference
	// AdditionalAnnotations is a collection of annotations set for the secret
	AdditionalAnnotations AdditionalAnnotations
	// Plumbing:
	Informer      corev1informers.ConfigMapInformer
	Lister        corev1listers.ConfigMapLister
	Client        corev1client.ConfigMapsGetter
	EventRecorder events.Recorder
}

func (c CABundleConfigMap) EnsureConfigMapCABundle(ctx context.Context, signingCertKeyPair *crypto.CA) ([]*x509.Certificate, error) {
	// by this point we have current signing cert/key pair.  We now need to make sure that the ca-bundle configmap has this cert and
	// doesn't have any expired certs
	originalCABundleConfigMap, err := c.Lister.ConfigMaps(c.Namespace).Get(c.Name)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	caBundleConfigMap := originalCABundleConfigMap.DeepCopy()
	if apierrors.IsNotFound(err) {
		// create an empty one
		caBundleConfigMap = &corev1.ConfigMap{ObjectMeta: NewTLSArtifactObjectMeta(
			c.Name,
			c.Namespace,
			c.AdditionalAnnotations,
		)}
	}

	needsMetadataUpdate := false
	if c.Owner != nil {
		needsMetadataUpdate = ensureOwnerReference(&caBundleConfigMap.ObjectMeta, c.Owner)
	}
	needsMetadataUpdate = c.AdditionalAnnotations.EnsureTLSMetadataUpdate(&caBundleConfigMap.ObjectMeta) || needsMetadataUpdate
	if needsMetadataUpdate && len(caBundleConfigMap.ResourceVersion) > 0 {
		_, _, err := resourceapply.ApplyConfigMap(ctx, c.Client, c.EventRecorder, caBundleConfigMap)
		if err != nil {
			return nil, err
		}
	}

	updatedCerts, err := manageCABundleConfigMap(caBundleConfigMap, signingCertKeyPair.Config.Certs[0])
	if err != nil {
		return nil, err
	}
	if originalCABundleConfigMap == nil || originalCABundleConfigMap.Data == nil || !equality.Semantic.DeepEqual(originalCABundleConfigMap.Data, caBundleConfigMap.Data) {
		c.EventRecorder.Eventf("CABundleUpdateRequired", "%q in %q requires a new cert", c.Name, c.Namespace)
		LabelAsManagedConfigMap(caBundleConfigMap, CertificateTypeCABundle)

		actualCABundleConfigMap, modified, err := resourceapply.ApplyConfigMap(ctx, c.Client, c.EventRecorder, caBundleConfigMap)
		if err != nil {
			return nil, err
		}
		if modified {
			klog.V(2).Infof("Updated ca-bundle.crt configmap %s/%s with:\n%s", certs.CertificateBundleToString(updatedCerts), caBundleConfigMap.Namespace, caBundleConfigMap.Name)
		}

		caBundleConfigMap = actualCABundleConfigMap
	}

	caBundle := caBundleConfigMap.Data["ca-bundle.crt"]
	if len(caBundle) == 0 {
		return nil, fmt.Errorf("configmap/%s -n%s missing ca-bundle.crt", caBundleConfigMap.Name, caBundleConfigMap.Namespace)
	}
	certificates, err := cert.ParseCertsPEM([]byte(caBundle))
	if err != nil {
		return nil, err
	}

	return certificates, nil
}

// manageCABundleConfigMap adds the new certificate to the list of cabundles, eliminates duplicates, and prunes the list of expired
// certs to trust as signers
func manageCABundleConfigMap(caBundleConfigMap *corev1.ConfigMap, currentSigner *x509.Certificate) ([]*x509.Certificate, error) {
	if caBundleConfigMap.Data == nil {
		caBundleConfigMap.Data = map[string]string{}
	}

	certificates := []*x509.Certificate{}
	caBundle := caBundleConfigMap.Data["ca-bundle.crt"]
	if len(caBundle) > 0 {
		var err error
		certificates, err = cert.ParseCertsPEM([]byte(caBundle))
		if err != nil {
			return nil, err
		}
	}
	certificates = append([]*x509.Certificate{currentSigner}, certificates...)
	certificates = crypto.FilterExpiredCerts(certificates...)

	finalCertificates := []*x509.Certificate{}
	// now check for duplicates. n^2, but super simple
	for i := range certificates {
		found := false
		for j := range finalCertificates {
			if reflect.DeepEqual(certificates[i].Raw, finalCertificates[j].Raw) {
				found = true
				break
			}
		}
		if !found {
			finalCertificates = append(finalCertificates, certificates[i])
		}
	}

	// sorting ensures we don't continuously swap the certificates in the bundle, which might cause revision rollouts
	sort.SliceStable(finalCertificates, func(i, j int) bool {
		return bytes.Compare(finalCertificates[i].Raw, finalCertificates[j].Raw) < 0
	})
	caBytes, err := crypto.EncodeCertificates(finalCertificates...)
	if err != nil {
		return nil, err
	}

	caBundleConfigMap.Data["ca-bundle.crt"] = string(caBytes)

	return finalCertificates, nil
}
