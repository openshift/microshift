package csr

import (
	"context"
	"crypto/tls"
	"crypto/x509/pkix"
	"fmt"
	"math/rand"
	"time"

	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"

	certificates "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	certificatesinformers "k8s.io/client-go/informers/certificates/v1"
	corev1informers "k8s.io/client-go/informers/core/v1"
	csrclient "k8s.io/client-go/kubernetes/typed/certificates/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	certificateslisters "k8s.io/client-go/listers/certificates/v1"
	cache "k8s.io/client-go/tools/cache"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"k8s.io/klog/v2"
)

const (
	// TLSKeyFile is the name of tls key file in kubeconfigSecret
	TLSKeyFile = "tls.key"
	// TLSCertFile is the name of the tls cert file in kubeconfigSecret
	TLSCertFile = "tls.crt"
)

// ControllerResyncInterval is exposed so that integration tests can crank up the constroller sync speed.
var ControllerResyncInterval = 5 * time.Minute

// CSROption includes options that is used to create and monitor csrs
type CSROption struct {
	// ObjectMeta is the ObjectMeta shared by all created csrs. It should use GenerateName instead of Name
	// to generate random csr names
	ObjectMeta metav1.ObjectMeta
	// Subject represents the subject of the client certificate used to create csrs
	Subject *pkix.Name
	// DNSNames represents DNS names used to create the client certificate
	DNSNames []string
	// SignerName is the name of the signer specified in the created csrs
	SignerName string

	// EventFilterFunc matches csrs created with above options
	EventFilterFunc factory.EventFilterFunc
}

// ClientCertOption includes options that is used to create client certificate
type ClientCertOption struct {
	// SecretNamespace is the namespace of the secret containing client certificate.
	SecretNamespace string
	// SecretName is the name of the secret containing client certificate. The secret will be created if
	// it does not exist.
	SecretName string
	// AdditonalSecretData contains data that will be added into client certificate secret besides tls.key/tls.crt
	AdditonalSecretData map[string][]byte
}

// clientCertificateController implements the common logic of hub client certification creation/rotation. It
// creates a client certificate and rotates it before it becomes expired by using csrs. The client
// certificate generated is stored in a specific secret with the keys below:
// 1). tls.key: tls key file
// 2). tls.crt: tls cert file
type clientCertificateController struct {
	ClientCertOption
	CSROption

	hubCSRLister    certificateslisters.CertificateSigningRequestLister
	hubCSRClient    csrclient.CertificateSigningRequestInterface
	spokeCoreClient corev1client.CoreV1Interface
	controllerName  string

	// csrName is the name of csr created by controller and waiting for approval.
	csrName string

	// keyData is the private key data used to created a csr
	// csrName and keyData store the internal state of the controller. They are set after controller creates a new csr
	// and cleared once the csr is approved and processed by controller. There are 4 combination of their values:
	//   1. csrName empty, keyData empty: means we aren't trying to create a new client cert, our current one is valid
	//   2. csrName set, keyData empty: there was bug
	//   3. csrName set, keyData set: we are waiting for a new cert to be signed.
	//   4. csrName empty, keydata set: the CSR failed to create, this shouldn't happen, it's a bug.
	keyData []byte
}

// NewClientCertificateController return an instance of clientCertificateController
func NewClientCertificateController(
	clientCertOption ClientCertOption,
	csrOption CSROption,
	hubCSRInformer certificatesinformers.CertificateSigningRequestInformer,
	hubCSRClient csrclient.CertificateSigningRequestInterface,
	spokeSecretInformer corev1informers.SecretInformer,
	spokeCoreClient corev1client.CoreV1Interface,
	recorder events.Recorder,
	controllerName string,
) factory.Controller {
	c := clientCertificateController{
		ClientCertOption: clientCertOption,
		CSROption:        csrOption,
		hubCSRLister:     hubCSRInformer.Lister(),
		hubCSRClient:     hubCSRClient,
		spokeCoreClient:  spokeCoreClient,
		controllerName:   controllerName,
	}

	return factory.New().
		WithFilteredEventsInformersQueueKeyFunc(func(obj runtime.Object) string {
			key, _ := cache.MetaNamespaceKeyFunc(obj)
			return key
		}, func(obj interface{}) bool {
			accessor, err := meta.Accessor(obj)
			if err != nil {
				return false
			}
			// only enqueue a specific secret
			if accessor.GetNamespace() == c.SecretNamespace && accessor.GetName() == c.SecretName {
				return true
			}
			return false
		}, spokeSecretInformer.Informer()).
		WithFilteredEventsInformersQueueKeyFunc(func(obj runtime.Object) string {
			accessor, _ := meta.Accessor(obj)
			return accessor.GetName()
		}, c.EventFilterFunc, hubCSRInformer.Informer()).
		WithSync(c.sync).
		ResyncEvery(ControllerResyncInterval).
		ToController(controllerName, recorder)
}

func (c *clientCertificateController) sync(ctx context.Context, syncCtx factory.SyncContext) error {
	// get secret containing client certificate
	secret, err := c.spokeCoreClient.Secrets(c.SecretNamespace).Get(ctx, c.SecretName, metav1.GetOptions{})
	switch {
	case errors.IsNotFound(err):
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: c.SecretNamespace,
				Name:      c.SecretName,
			},
		}
	case err != nil:
		return fmt.Errorf("unable to get secret %q: %w", c.SecretNamespace+"/"+c.SecretName, err)
	}

	// reconcile pending csr if exists
	if len(c.csrName) > 0 {
		newSecretConfig, err := c.syncCSR(secret)
		if err != nil {
			c.reset()
			return err
		}
		if len(newSecretConfig) == 0 {
			return nil
		}
		// append additional data into client certificate secret
		for k, v := range c.AdditonalSecretData {
			newSecretConfig[k] = v
		}
		secret.Data = newSecretConfig
		// save the changes into secret
		if err := c.saveSecret(secret); err != nil {
			return err
		}
		syncCtx.Recorder().Eventf("ClientCertificateCreated", "A new client certificate for %s is available", c.controllerName)
		c.reset()
		return nil
	}

	// create a csr to request new client certificate if
	// a. there is no valid client certificate issued for the current cluster/agent
	// b. client certificate exists and has less than a random percentage range from 20% to 25% of its life remaining
	if c.hasValidClientCertificate(secret) {
		notBefore, notAfter, err := getCertValidityPeriod(secret)
		if err != nil {
			return err
		}

		total := notAfter.Sub(*notBefore)
		remaining := notAfter.Sub(time.Now())
		klog.V(4).Infof("Client certificate for %s: time total=%v, remaining=%v, remaining/total=%v", c.controllerName, total, remaining, remaining.Seconds()/total.Seconds())
		threshold := jitter(0.2, 0.25)
		if remaining.Seconds()/total.Seconds() > threshold {
			// Do nothing if the client certificate is valid and has more than a random percentage range from 20% to 25% of its life remaining
			klog.V(4).Infof("Client certificate for %s is valid and has more than %.2f%% of its life remaining", c.controllerName, threshold*100)
			return nil
		}
		syncCtx.Recorder().Eventf("CertificateRotationStarted", "The current client certificate for %s expires in %v. Start certificate rotation", c.controllerName, remaining.Round(time.Second))
	} else {
		syncCtx.Recorder().Eventf("NoValidCertificateFound", "No valid client certificate for %s is found. Bootstrap is required", c.controllerName)
	}

	// create a new private key
	c.keyData, err = keyutil.MakeEllipticPrivateKeyPEM()
	if err != nil {
		return err
	}

	// create a csr
	c.csrName, err = c.createCSR(ctx)
	if err != nil {
		c.reset()
		return err
	}
	syncCtx.Recorder().Eventf("CSRCreated", "A csr %q is created for %s", c.csrName, c.controllerName)
	return nil
}

func (c *clientCertificateController) syncCSR(secret *corev1.Secret) (map[string][]byte, error) {
	// skip if there is no ongoing csr
	if len(c.csrName) == 0 {
		return nil, fmt.Errorf("no ongoing csr")
	}

	// skip if csr no longer exists
	csr, err := c.hubCSRLister.Get(c.csrName)
	switch {
	case errors.IsNotFound(err):
		// fallback to fetching csr from hub apiserver in case it is not cached by informer yet
		csr, err = c.hubCSRClient.Get(context.Background(), c.csrName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("unable to get csr %q. It might have already been deleted.", c.csrName)
		}
	case err != nil:
		return nil, err
	}

	// skip if csr is not approved yet
	if !isCSRApproved(csr) {
		return nil, nil
	}

	// skip if csr has no certificate in its status yet
	if len(csr.Status.Certificate) == 0 {
		return nil, nil
	}

	klog.V(4).Infof("Sync csr %v", c.csrName)
	// check if cert in csr status matches with the corresponding private key
	if c.keyData == nil {
		return nil, fmt.Errorf("No private key found for certificate in csr: %s", c.csrName)
	}
	_, err = tls.X509KeyPair(csr.Status.Certificate, c.keyData)
	if err != nil {
		return nil, fmt.Errorf("Private key does not match with the certificate in csr: %s", c.csrName)
	}

	data := map[string][]byte{
		TLSCertFile: csr.Status.Certificate,
		TLSKeyFile:  c.keyData,
	}

	return data, nil
}

func (c *clientCertificateController) createCSR(ctx context.Context) (string, error) {
	privateKey, err := keyutil.ParsePrivateKeyPEM(c.keyData)
	if err != nil {
		return "", fmt.Errorf("invalid private key for certificate request: %w", err)
	}
	csrData, err := certutil.MakeCSR(privateKey, c.Subject, c.DNSNames, nil)
	if err != nil {
		return "", fmt.Errorf("unable to generate certificate request: %w", err)
	}

	csr := &certificates.CertificateSigningRequest{
		ObjectMeta: c.ObjectMeta,
		Spec: certificates.CertificateSigningRequestSpec{
			Request: csrData,
			Usages: []certificates.KeyUsage{
				certificates.UsageDigitalSignature,
				certificates.UsageKeyEncipherment,
				certificates.UsageClientAuth,
			},
			SignerName: c.SignerName,
		},
	}

	req, err := c.hubCSRClient.Create(ctx, csr, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}
	return req.Name, nil
}

func (c *clientCertificateController) saveSecret(secret *corev1.Secret) error {
	var err error
	if secret.ResourceVersion == "" {
		_, err = c.spokeCoreClient.Secrets(c.SecretNamespace).Create(context.Background(), secret, metav1.CreateOptions{})
		return err
	}
	_, err = c.spokeCoreClient.Secrets(c.SecretNamespace).Update(context.Background(), secret, metav1.UpdateOptions{})
	return err
}

func (c *clientCertificateController) reset() {
	c.csrName = ""
	c.keyData = nil
}

func (c *clientCertificateController) hasValidClientCertificate(secret *corev1.Secret) bool {
	if valid, err := IsCertificateValid(secret.Data[TLSCertFile], c.Subject); err == nil {
		return valid
	}
	return false
}

func jitter(percentage float64, maxFactor float64) float64 {
	if maxFactor <= 0.0 {
		maxFactor = 1.0
	}
	newPercentage := percentage + percentage*rand.Float64()*maxFactor
	return newPercentage
}
