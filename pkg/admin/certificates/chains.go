package certificates

import (
	"net"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
	"k8s.io/apiserver/pkg/authentication/serviceaccount"
	"k8s.io/apiserver/pkg/authentication/user"
	apiserveroptions "k8s.io/kubernetes/pkg/controlplane/apiserver/options"
)

// NewChains defines the managed PKI once for startup, administration, and observations.
// Callers choose whether to load public certificates or generate certificate material.
//
//nolint:ireturn // Callers choose read-only inventory loading or certificate creation.
func NewChains(cfg *config.Config, dataDir string) (certchains.CertificateChainsBuilder, error) {
	if err := cfg.Certificates.Validate(); err != nil {
		return nil, err
	}
	// Anchor certificate expiration to the next day. This forces
	// homogenous expiry dates for all certificates with the same validity.
	startTime := time.Now()
	nextMidnight := time.Date(
		startTime.Year(),
		startTime.Month(),
		startTime.Day()+1,
		0, 0, 0, 0,
		startTime.Location(),
	)
	alignValidity := func(baseValidity time.Duration) time.Duration {
		targetExpiration := nextMidnight.Add(baseValidity)
		return time.Until(targetExpiration)
	}

	_, svcNet, err := net.ParseCIDR(cfg.Network.ServiceNetwork[0])
	if err != nil {
		return nil, err
	}

	_, apiServerServiceIP, err := apiserveroptions.ServiceIPRange(*svcNet)
	if err != nil {
		return nil, err
	}

	externalCertNames := []string{
		cfg.Node.HostnameOverride,
		"api." + cfg.DNS.BaseDomain,
	}
	externalCertNames = append(externalCertNames, cfg.ApiServer.SubjectAltNames...)
	// When Kube apiserver advertise address matches the node IP we can not add
	// it to the certificates or else the internal pod access to apiserver is
	// broken. Because of client-go not using SNI and the way apiserver handles
	// which certificate to serve which destination IP, internal pods start
	// getting the external certificate, which is signed by a different CA and
	// does not match the hostname.
	if cfg.ApiServer.AdvertiseAddress != cfg.Node.NodeIP {
		externalCertNames = append(externalCertNames, cfg.Node.NodeIP)
	}

	certsDir := cryptomaterial.CertsDirectory(dataDir)

	builder := certchains.NewCertificateChains(
		// ------------------------------
		// CLIENT CERTIFICATE SIGNERS
		// ------------------------------

		// kube-control-plane-signer
		certchains.NewCertificateSigner(
			"kube-control-plane-signer",
			cryptomaterial.KubeControlPlaneSignerCertDir(certsDir),
			alignValidity(cfg.Certificates.CADuration()),
		).WithService("control-plane").WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "kube-controller-manager",
					Service:        "kube-controller-manager",
					Validity:       alignValidity(cryptomaterial.ShortLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyStandard,
				},
				UserInfo: &user.DefaultInfo{Name: "system:kube-controller-manager"},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "kube-scheduler",
					Service:        "kube-scheduler",
					Validity:       alignValidity(cryptomaterial.ShortLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyStandard,
				},
				UserInfo: &user.DefaultInfo{Name: "system:kube-scheduler"},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "cluster-policy-controller",
					Service:        "cluster-policy-controller",
					Validity:       alignValidity(cryptomaterial.ShortLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyStandard,
				},
				UserInfo: &user.DefaultInfo{Name: "system:kube-controller-manager"},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "route-controller-manager",
					Service:        "route-controller-manager",
					Validity:       alignValidity(cryptomaterial.ShortLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyStandard,
				},
				UserInfo: serviceaccount.UserInfo("openshift-route-controller-manager", "route-controller-manager-sa", ""),
			}),

		// kube-apiserver-to-kubelet-signer
		certchains.NewCertificateSigner(
			"kube-apiserver-to-kubelet-signer",
			cryptomaterial.KubeAPIServerToKubeletSignerCertDir(certsDir),
			alignValidity(cfg.Certificates.CADuration()),
		).WithService("kube-apiserver").WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "kube-apiserver-to-kubelet-client",
					Service:        "kube-apiserver",
					Validity:       alignValidity(cryptomaterial.ShortLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyStandard,
				},
				UserInfo: &user.DefaultInfo{Name: "system:kube-apiserver", Groups: []string{"kube-master"}},
			}).WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "metrics-server-kubelet-client",
					Service:        "metrics-server",
					Validity:       alignValidity(cryptomaterial.ShortLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyStandard,
				},
				UserInfo: &user.DefaultInfo{Name: "system:metrics-server"},
			}),

		// admin-kubeconfig-signer
		certchains.NewCertificateSigner(
			"admin-kubeconfig-signer",
			cryptomaterial.AdminKubeconfigSignerDir(certsDir),
			alignValidity(cfg.Certificates.CADuration()),
		).WithService("authentication").WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "admin-kubeconfig-client",
					Service:        "authentication",
					Validity:       alignValidity(cryptomaterial.LongLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyExtended,
				},
				UserInfo: &user.DefaultInfo{Name: "system:admin", Groups: []string{"system:masters"}},
			}).WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "openshift-observability-client",
					Service:        "observability",
					Validity:       alignValidity(cryptomaterial.ShortLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyStandard,
				},
				UserInfo: &user.DefaultInfo{Name: "openshift-observability-client"},
			},
		),

		// kubelet + CSR signing chain
		certchains.NewCertificateSigner(
			"kubelet-signer",
			cryptomaterial.KubeletCSRSignerSignerCertDir(certsDir),
			alignValidity(cfg.Certificates.CADuration()),
		).WithService("kubelet").WithSubCAs(
			certchains.NewCertificateSigner(
				"kube-csr-signer",
				cryptomaterial.CSRSignerCertDir(certsDir),
				alignValidity(cfg.Certificates.CADuration()),
			).WithService("kubelet").WithClientCertificates(
				&certchains.ClientCertificateSigningRequestInfo{
					CSRMeta: certchains.CSRMeta{
						Name:           "kubelet-client",
						Service:        "kubelet",
						Validity:       alignValidity(cryptomaterial.ShortLivedCertificateValidity),
						RotationPolicy: certchains.RotationPolicyStandard,
					},
					// userinfo per https://kubernetes.io/docs/reference/access-authn-authz/node/#overview
					UserInfo: &user.DefaultInfo{Name: "system:node:" + cfg.CanonicalNodeName(), Groups: []string{"system:nodes"}},
				},
			).WithServingCertificates(
				&certchains.ServingCertificateSigningRequestInfo{
					CSRMeta: certchains.CSRMeta{
						Name:     "kubelet-server",
						Service:  "kubelet",
						Validity: alignValidity(cfg.Certificates.ServingDuration()),
					},
					Hostnames: []string{cfg.Node.HostnameOverride, cfg.Node.NodeIP},
				},
			),
		),
		certchains.NewCertificateSigner(
			"aggregator-signer",
			cryptomaterial.AggregatorSignerDir(certsDir),
			alignValidity(cfg.Certificates.CADuration()),
		).WithService("kube-apiserver").WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "aggregator-client",
					Service:        "kube-apiserver",
					Validity:       alignValidity(cryptomaterial.ShortLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyStandard,
				},
				UserInfo: &user.DefaultInfo{Name: "system:openshift-aggregator"},
			},
		),

		//------------------------------
		// SERVING CERTIFICATE SIGNERS
		//------------------------------
		certchains.NewCertificateSigner(
			"service-ca",
			cryptomaterial.ServiceCADir(certsDir),
			alignValidity(cfg.Certificates.CADuration()),
		).WithService("service-ca").WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:     "route-controller-manager-serving",
					Service:  "route-controller-manager",
					Validity: alignValidity(cfg.Certificates.ServingDuration()),
				},
				Hostnames: []string{
					"route-controller-manager.openshift-route-controller-manager.svc",
					"route-controller-manager.openshift-route-controller-manager.svc.cluster.local",
				},
			},
		),

		certchains.NewCertificateSigner(
			"ingress-ca",
			cryptomaterial.IngressCADir(certsDir),
			alignValidity(cfg.Certificates.CADuration()),
		).WithService("ingress").WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:     "router-default-serving",
					Service:  "ingress",
					Validity: alignValidity(cfg.Certificates.ServingDuration()),
				},
				Hostnames: []string{
					"*.apps." + cfg.DNS.BaseDomain, // wildcard for any additional auto-generated domains
				},
			},
		),

		// this signer replaces the loadbalancer signers of OCP, we don't need those
		// in Microshift
		certchains.NewCertificateSigner(
			"kube-apiserver-external-signer",
			cryptomaterial.KubeAPIServerExternalSigner(certsDir),
			alignValidity(cfg.Certificates.CADuration()),
		).WithService("kube-apiserver").WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:     "kube-external-serving",
					Service:  "kube-apiserver",
					Validity: alignValidity(cfg.Certificates.ServingDuration()),
				},
				Hostnames: externalCertNames,
			},
		),

		certchains.NewCertificateSigner(
			"kube-apiserver-localhost-signer",
			cryptomaterial.KubeAPIServerLocalhostSigner(certsDir),
			alignValidity(cfg.Certificates.CADuration()),
		).WithService("kube-apiserver").WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:     "kube-apiserver-localhost-serving",
					Service:  "kube-apiserver",
					Validity: alignValidity(cfg.Certificates.ServingDuration()),
				},
				Hostnames: []string{
					"localhost",
				},
			},
		),

		certchains.NewCertificateSigner(
			"kube-apiserver-service-network-signer",
			cryptomaterial.KubeAPIServerServiceNetworkSigner(certsDir),
			alignValidity(cfg.Certificates.CADuration()),
		).WithService("kube-apiserver").WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:     "kube-apiserver-service-network-serving",
					Service:  "kube-apiserver",
					Validity: alignValidity(cfg.Certificates.ServingDuration()),
				},
				Hostnames: []string{
					"kubernetes",
					"kubernetes.default",
					"kubernetes.default.svc",
					"kubernetes.default.svc.cluster.local",
					"openshift",
					"openshift.default",
					"openshift.default.svc",
					"openshift.default.svc.cluster.local",
					"api." + cfg.DNS.BaseDomain,
					"api-int." + cfg.DNS.BaseDomain,
					cfg.ApiServer.AdvertiseAddress,
					apiServerServiceIP.String(),
				},
			},
		),

		//------------------------------
		// 	ETCD CERTIFICATE SIGNER
		//------------------------------
		certchains.NewCertificateSigner(
			"etcd-signer",
			cryptomaterial.EtcdSignerDir(certsDir),
			alignValidity(cfg.Certificates.CADuration()),
		).WithService("etcd").WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "apiserver-etcd-client",
					Service:        "kube-apiserver",
					Validity:       alignValidity(cryptomaterial.LongLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyExtended,
				},
				UserInfo: &user.DefaultInfo{Name: "etcd", Groups: []string{"etcd"}},
			},
		).WithPeerCertificiates(
			&certchains.PeerCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "etcd-peer",
					Service:        "etcd",
					Validity:       alignValidity(cryptomaterial.LongLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyExtended,
				},
				UserInfo:  &user.DefaultInfo{Name: "system:etcd-peer:etcd-client", Groups: []string{"system:etcd-peers"}},
				Hostnames: []string{"localhost", cfg.Node.HostnameOverride, cfg.Node.NodeIP},
			},
			&certchains.PeerCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "etcd-serving",
					Service:        "etcd",
					Validity:       alignValidity(cryptomaterial.LongLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyExtended,
				},
				UserInfo:  &user.DefaultInfo{Name: "system:etcd-server:etcd-client", Groups: []string{"system:etcd-servers"}},
				Hostnames: []string{"localhost", cfg.Node.HostnameOverride, cfg.Node.NodeIP},
			},
		),
	).WithCABundle(
		cryptomaterial.TotalClientCABundlePath(certsDir),
		[]string{"kube-control-plane-signer"},
		[]string{"kube-apiserver-to-kubelet-signer"},
		[]string{"admin-kubeconfig-signer"},
		[]string{"kubelet-signer"},
		[]string{"kubelet-signer", "kube-csr-signer"},
	).WithCABundle(
		cryptomaterial.KubeletClientCAPath(certsDir),
		[]string{"kube-control-plane-signer"},
		[]string{"kube-apiserver-to-kubelet-signer"},
		[]string{"admin-kubeconfig-signer"},
		[]string{"kubelet-signer"},
		[]string{"kubelet-signer", "kube-csr-signer"},
	).WithCABundle(
		cryptomaterial.KubeletServingCAPath(certsDir),
		[]string{"kubelet-signer"},
		[]string{"kubelet-signer", "kube-csr-signer"},
	).WithCABundle(
		cryptomaterial.ServiceAccountTokenCABundlePath(certsDir),
		[]string{"kube-apiserver-localhost-signer"},
		[]string{"kube-apiserver-service-network-signer"},
		[]string{"kube-apiserver-external-signer"},
	)

	return builder, nil
}
