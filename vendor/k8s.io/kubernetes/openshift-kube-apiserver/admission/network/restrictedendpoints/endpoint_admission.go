package restrictedendpoints

import (
	"context"
	"fmt"
	"io"
	"net"
	"reflect"

	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/admission/initializer"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/klog/v2"
	kapi "k8s.io/kubernetes/pkg/apis/core"

	"github.com/openshift/library-go/pkg/config/helpers"
	"k8s.io/kubernetes/openshift-kube-apiserver/admission/network/apis/restrictedendpoints"
	v1 "k8s.io/kubernetes/openshift-kube-apiserver/admission/network/apis/restrictedendpoints/v1"
)

const RestrictedEndpointsPluginName = "network.openshift.io/RestrictedEndpointsAdmission"

func RegisterRestrictedEndpoints(plugins *admission.Plugins) {
	plugins.Register(RestrictedEndpointsPluginName,
		func(config io.Reader) (admission.Interface, error) {
			pluginConfig, err := readConfig(config)
			if err != nil {
				return nil, err
			}
			if pluginConfig == nil {
				klog.Infof("Admission plugin %q is not configured so it will be disabled.", RestrictedEndpointsPluginName)
				return nil, nil
			}
			restrictedNetworks, err := ParseSimpleCIDRRules(pluginConfig.RestrictedCIDRs)
			if err != nil {
				// should have been caught with validation
				return nil, err
			}

			return NewRestrictedEndpointsAdmission(restrictedNetworks), nil
		})
}

func readConfig(reader io.Reader) (*restrictedendpoints.RestrictedEndpointsAdmissionConfig, error) {
	obj, err := helpers.ReadYAMLToInternal(reader, restrictedendpoints.Install, v1.Install)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, nil
	}
	config, ok := obj.(*restrictedendpoints.RestrictedEndpointsAdmissionConfig)
	if !ok {
		return nil, fmt.Errorf("unexpected config object: %#v", obj)
	}
	// No validation needed since config is just list of strings
	return config, nil
}

type restrictedEndpointsAdmission struct {
	*admission.Handler

	authorizer         authorizer.Authorizer
	restrictedNetworks []*net.IPNet
}

var _ = initializer.WantsAuthorizer(&restrictedEndpointsAdmission{})
var _ = admission.ValidationInterface(&restrictedEndpointsAdmission{})

// ParseSimpleCIDRRules parses a list of CIDR strings
func ParseSimpleCIDRRules(rules []string) (networks []*net.IPNet, err error) {
	for _, s := range rules {
		_, cidr, err := net.ParseCIDR(s)
		if err != nil {
			return nil, err
		}
		networks = append(networks, cidr)
	}
	return networks, nil
}

// NewRestrictedEndpointsAdmission creates a new endpoints admission plugin.
func NewRestrictedEndpointsAdmission(restrictedNetworks []*net.IPNet) *restrictedEndpointsAdmission {
	return &restrictedEndpointsAdmission{
		Handler:            admission.NewHandler(admission.Create, admission.Update),
		restrictedNetworks: restrictedNetworks,
	}
}

func (r *restrictedEndpointsAdmission) SetAuthorizer(a authorizer.Authorizer) {
	r.authorizer = a
}

func (r *restrictedEndpointsAdmission) ValidateInitialization() error {
	if r.authorizer == nil {
		return fmt.Errorf("missing authorizer")
	}
	return nil
}

var (
	defaultRestrictedPorts = []kapi.EndpointPort{
		// MCS ports
		{Protocol: kapi.ProtocolTCP, Port: 22623},
		{Protocol: kapi.ProtocolTCP, Port: 22624},
	}
	defaultRestrictedNetworks = []*net.IPNet{
		// IPv4 link-local range 169.254.0.0/16 (including cloud metadata IP)
		{IP: net.ParseIP("169.254.0.0"), Mask: net.CIDRMask(16, 32)},
	}
)

func (r *restrictedEndpointsAdmission) findRestrictedIP(ep *kapi.Endpoints, restricted []*net.IPNet) error {
	for _, subset := range ep.Subsets {
		for _, addr := range subset.Addresses {
			ip := net.ParseIP(addr.IP)
			if ip == nil {
				continue
			}
			for _, net := range restricted {
				if net.Contains(ip) {
					return fmt.Errorf("endpoint address %s is not allowed", addr.IP)
				}
			}
		}
	}
	return nil
}

func (r *restrictedEndpointsAdmission) findRestrictedPort(ep *kapi.Endpoints, restricted []kapi.EndpointPort) error {
	for _, subset := range ep.Subsets {
		for _, port := range subset.Ports {
			for _, restricted := range restricted {
				if port.Protocol == restricted.Protocol && port.Port == restricted.Port {
					return fmt.Errorf("endpoint port %s:%d is not allowed", string(port.Protocol), port.Port)
				}
			}
		}
	}
	return nil
}

func (r *restrictedEndpointsAdmission) checkAccess(ctx context.Context, attr admission.Attributes) (bool, error) {
	authzAttr := authorizer.AttributesRecord{
		User:            attr.GetUserInfo(),
		Verb:            "create",
		Namespace:       attr.GetNamespace(),
		Resource:        "endpoints",
		Subresource:     "restricted",
		APIGroup:        kapi.GroupName,
		Name:            attr.GetName(),
		ResourceRequest: true,
	}
	authorized, _, err := r.authorizer.Authorize(ctx, authzAttr)
	return authorized == authorizer.DecisionAllow, err
}

// Admit determines if the endpoints object should be admitted
func (r *restrictedEndpointsAdmission) Validate(ctx context.Context, a admission.Attributes, _ admission.ObjectInterfaces) error {
	if a.GetResource().GroupResource() != kapi.Resource("endpoints") {
		return nil
	}
	ep, ok := a.GetObject().(*kapi.Endpoints)
	if !ok {
		return nil
	}
	old, ok := a.GetOldObject().(*kapi.Endpoints)
	if ok && reflect.DeepEqual(ep.Subsets, old.Subsets) {
		return nil
	}

	restrictedErr := r.findRestrictedIP(ep, r.restrictedNetworks)
	if restrictedErr == nil {
		restrictedErr = r.findRestrictedIP(ep, defaultRestrictedNetworks)
	}
	if restrictedErr == nil {
		restrictedErr = r.findRestrictedPort(ep, defaultRestrictedPorts)
	}
	if restrictedErr == nil {
		return nil
	}

	allow, err := r.checkAccess(ctx, a)
	if err != nil {
		return err
	}
	if !allow {
		return admission.NewForbidden(a, restrictedErr)
	}
	return nil
}
