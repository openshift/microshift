package ingress

import (
	"k8s.io/client-go/tools/record"
	"reflect"
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	"k8s.io/apimachinery/pkg/util/intstr"
	fake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	corelisters "k8s.io/client-go/listers/core/v1"
	networkingv1listers "k8s.io/client-go/listers/networking/v1"
	clientgotesting "k8s.io/client-go/testing"
	"k8s.io/client-go/util/workqueue"

	routev1 "github.com/openshift/api/route/v1"
	routev1fake "github.com/openshift/client-go/route/clientset/versioned/fake"
	routelisters "github.com/openshift/client-go/route/listers/route/v1"
)

type routeLister struct {
	Err   error
	Items []*routev1.Route
}

func (r *routeLister) List(selector labels.Selector) (ret []*routev1.Route, err error) {
	return r.Items, r.Err
}
func (r *routeLister) Routes(namespace string) routelisters.RouteNamespaceLister {
	return &nsRouteLister{r: r, ns: namespace}
}

type nsRouteLister struct {
	r  *routeLister
	ns string
}

func (r *nsRouteLister) List(selector labels.Selector) (ret []*routev1.Route, err error) {
	return r.r.Items, r.r.Err
}
func (r *nsRouteLister) Get(name string) (*routev1.Route, error) {
	for _, s := range r.r.Items {
		if s.Name == name && r.ns == s.Namespace {
			return s, nil
		}
	}
	return nil, kerrors.NewNotFound(schema.GroupResource{}, name)
}

type ingressLister struct {
	Err   error
	Items []*networkingv1.Ingress
}

func (r *ingressLister) List(selector labels.Selector) (ret []*networkingv1.Ingress, err error) {
	return r.Items, r.Err
}
func (r *ingressLister) Ingresses(namespace string) networkingv1listers.IngressNamespaceLister {
	return &nsIngressLister{r: r, ns: namespace}
}

type nsIngressLister struct {
	r  *ingressLister
	ns string
}

func (r *nsIngressLister) List(selector labels.Selector) (ret []*networkingv1.Ingress, err error) {
	return r.r.Items, r.r.Err
}
func (r *nsIngressLister) Get(name string) (*networkingv1.Ingress, error) {
	for _, s := range r.r.Items {
		if s.Name == name && r.ns == s.Namespace {
			return s, nil
		}
	}
	return nil, kerrors.NewNotFound(schema.GroupResource{}, name)
}

type ingressclassLister struct {
	Err   error
	Items []*networkingv1.IngressClass
}

func (r *ingressclassLister) List(selector labels.Selector) (ret []*networkingv1.IngressClass, err error) {
	return r.Items, r.Err
}
func (r *ingressclassLister) Get(name string) (*networkingv1.IngressClass, error) {
	for _, s := range r.Items {
		if s.Name == name {
			return s, nil
		}
	}
	return nil, kerrors.NewNotFound(schema.GroupResource{}, name)
}

type serviceLister struct {
	Err   error
	Items []*v1.Service
}

func (r *serviceLister) List(selector labels.Selector) (ret []*v1.Service, err error) {
	return r.Items, r.Err
}
func (r *serviceLister) Services(namespace string) corelisters.ServiceNamespaceLister {
	return &nsServiceLister{r: r, ns: namespace}
}

func (r *serviceLister) GetPodServices(pod *v1.Pod) ([]*v1.Service, error) {
	panic("unsupported")
}

type nsServiceLister struct {
	r  *serviceLister
	ns string
}

func (r *nsServiceLister) List(selector labels.Selector) (ret []*v1.Service, err error) {
	return r.r.Items, r.r.Err
}
func (r *nsServiceLister) Get(name string) (*v1.Service, error) {
	for _, s := range r.r.Items {
		if s.Name == name && r.ns == s.Namespace {
			return s, nil
		}
	}
	return nil, kerrors.NewNotFound(schema.GroupResource{}, name)
}

type secretLister struct {
	Err   error
	Items []*v1.Secret
}

func (r *secretLister) List(selector labels.Selector) (ret []*v1.Secret, err error) {
	return r.Items, r.Err
}
func (r *secretLister) Secrets(namespace string) corelisters.SecretNamespaceLister {
	return &nsSecretLister{r: r, ns: namespace}
}

type nsSecretLister struct {
	r  *secretLister
	ns string
}

func (r *nsSecretLister) List(selector labels.Selector) (ret []*v1.Secret, err error) {
	return r.r.Items, r.r.Err
}
func (r *nsSecretLister) Get(name string) (*v1.Secret, error) {
	for _, s := range r.r.Items {
		if s.Name == name && r.ns == s.Namespace {
			return s, nil
		}
	}
	return nil, kerrors.NewNotFound(schema.GroupResource{}, name)
}

const complexIngress = `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test-1
  namespace: test
spec:
  rules:
  - host: 1.ingress-test.com
    http:
      paths:
      - path: /test
        backend:
          service:
            name: ingress-endpoint-1
            port:
              number: 80
      - path: /other
        backend:
          service:
            name: ingress-endpoint-2
            port:
              number: 80
  - host: 2.ingress-test.com
    http:
      paths:
      - path: /
        backend:
          service:
            name: ingress-endpoint-1
            port:
              number: 80
  - host: 3.ingress-test.com
    http:
      paths:
      - path: /
        backend:
          service:
            name: ingress-endpoint-1
            port:
              number: 80
`

func TestController_stabilizeAfterCreate(t *testing.T) {
	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode([]byte(complexIngress), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	ingress := obj.(*networkingv1.Ingress)

	i := &ingressLister{
		Items: []*networkingv1.Ingress{
			ingress,
		},
	}
	ic := &ingressclassLister{Items: []*networkingv1.IngressClass{}}
	r := &routeLister{}
	s := &secretLister{}
	svc := &serviceLister{Items: []*v1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress-endpoint-1",
				Namespace: "test",
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Name:       "http",
						Port:       80,
						TargetPort: intstr.FromInt(8080),
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress-endpoint-2",
				Namespace: "test",
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Name:       "80-tcp",
						Port:       80,
						TargetPort: intstr.FromInt(8080),
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress-endpoint-3",
				Namespace: "test",
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Port:       80,
						TargetPort: intstr.FromString("tcp-8080"),
					},
				},
			},
		},
	}}

	var names []string
	routeClientset := &routev1fake.Clientset{}
	routeClientset.AddReactor("*", "routes", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
		switch a := action.(type) {
		case clientgotesting.CreateAction:
			obj := a.GetObject().DeepCopyObject()
			m := obj.(metav1.Object)
			if len(m.GetName()) == 0 {
				m.SetName(m.GetGenerateName())
			}
			names = append(names, m.GetName())
			return true, obj, nil
		}
		return true, nil, nil
	})
	kc := fake.NewSimpleClientset()
	kc.PrependReactor("*", "ingresses", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, nil
	})

	c := &Controller{
		queue:              workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ingress-to-route-test"),
		routeClient:        routeClientset.RouteV1(),
		ingressClient:      kc.NetworkingV1(),
		ingressLister:      i,
		ingressclassLister: ic,
		routeLister:        r,
		secretLister:       s,
		serviceLister:      svc,
		expectations:       newExpectations(),
	}
	defer c.queue.ShutDown()

	// load the ingresses for the namespace
	if err := c.sync(queueKey{namespace: "test"}); err != nil {
		t.Errorf("Controller.sync() error = %v", err)
	}
	if c.queue.Len() != 1 {
		t.Fatalf("Controller.sync() unexpected queue: %#v", c.queue.Len())
	}
	routeActions := routeClientset.Actions()
	if len(routeActions) != 0 {
		t.Fatalf("Controller.sync() unexpected actions: %#v", routeActions)
	}

	// process the ingress
	key, _ := c.queue.Get()
	expectKey := queueKey{namespace: ingress.Namespace, name: ingress.Name}
	if key.(queueKey) != expectKey {
		t.Fatalf("incorrect key: %v", key)
	}
	if err := c.sync(key.(queueKey)); err != nil {
		t.Fatalf("Controller.sync() error = %v", err)
	}
	c.queue.Done(key)
	if c.queue.Len() != 0 {
		t.Fatalf("Controller.sync() unexpected queue: %#v", c.queue.Len())
	}
	routeActions = routeClientset.Actions()
	if len(routeActions) == 0 {
		t.Fatalf("Controller.sync() unexpected actions: %#v", routeActions)
	}
	if !c.expectations.Expecting("test", "test-1") {
		t.Fatalf("Controller.sync() should be holding an expectation: %#v", c.expectations.expect)
	}

	for _, action := range routeActions {
		switch action.GetVerb() {
		case "create":
			switch o := action.(clientgotesting.CreateAction).GetObject().(type) {
			case *routev1.Route:
				r.Items = append(r.Items, o)
				c.processRoute(o)
			default:
				t.Fatalf("Unexpected create: %T", o)
			}
		default:
			t.Fatalf("Unexpected action: %#v", action)
		}
	}
	if c.queue.Len() != 1 {
		t.Fatalf("Controller.sync() unexpected queue: %#v", c.queue.Len())
	}
	if c.expectations.Expecting("test", "test-1") {
		t.Fatalf("Controller.sync() should have cleared all expectations: %#v", c.expectations.expect)
	}
	c.expectations.Expect("test", "test-1", names[0])

	// waiting for a single expected route, will do nothing
	key, _ = c.queue.Get()
	if err := c.sync(key.(queueKey)); err != nil {
		t.Errorf("Controller.sync() error = %v", err)
	}
	c.queue.Done(key)
	routeActions = routeClientset.Actions()
	if len(routeActions) == 0 {
		t.Fatalf("Controller.sync() unexpected actions: %#v", routeActions)
	}
	if c.queue.Len() != 1 {
		t.Fatalf("Controller.sync() unexpected queue: %#v", c.queue.Len())
	}
	c.expectations.Satisfied("test", "test-1", names[0])

	// steady state, nothing has changed
	key, _ = c.queue.Get()
	if err := c.sync(key.(queueKey)); err != nil {
		t.Errorf("Controller.sync() error = %v", err)
	}
	c.queue.Done(key)
	routeActions = routeClientset.Actions()
	if len(routeActions) == 0 {
		t.Fatalf("Controller.sync() unexpected actions: %#v", routeActions)
	}
	if c.queue.Len() != 0 {
		t.Fatalf("Controller.sync() unexpected queue: %#v", c.queue.Len())
	}
}

func newTestExpectations(fn func(*expectations)) *expectations {
	e := newExpectations()
	fn(e)
	return e
}

func TestController_sync(t *testing.T) {
	operatorv1GroupVersion := "operator.openshift.io/v1"
	ingressclasses := &ingressclassLister{Items: []*networkingv1.IngressClass{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "openshift-default",
			},
			Spec: networkingv1.IngressClassSpec{
				Controller: "openshift.io/ingress-to-route",
				Parameters: &networkingv1.IngressClassParametersReference{
					APIGroup: &operatorv1GroupVersion,
					Kind:     "IngressController",
					Name:     "default",
				},
			},
		},
	}}
	services := &serviceLister{Items: []*v1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service-1",
				Namespace: "test",
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Name:       "http",
						Port:       80,
						TargetPort: intstr.FromInt(8080),
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service-2",
				Namespace: "test",
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Name:       "80-tcp",
						Port:       80,
						TargetPort: intstr.FromInt(8080),
					},
				},
			},
		},
	}}
	secrets := &secretLister{Items: []*v1.Secret{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret-0",
				Namespace: "test",
			},
			Type: v1.SecretTypeOpaque,
			Data: map[string][]byte{
				v1.TLSCertKey:       []byte(`cert`),
				v1.TLSPrivateKeyKey: []byte(`key`),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret-1",
				Namespace: "test",
			},
			Type: v1.SecretTypeTLS,
			Data: map[string][]byte{
				v1.TLSCertKey:       []byte(`cert`),
				v1.TLSPrivateKeyKey: []byte(`key`),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret-1a",
				Namespace: "test",
			},
			Type: v1.SecretTypeTLS,
			Data: map[string][]byte{
				v1.TLSCertKey:       []byte(`cert`),
				v1.TLSPrivateKeyKey: []byte(`key2`),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret-2",
				Namespace: "test",
			},
			Type: v1.SecretTypeTLS,
			Data: map[string][]byte{
				v1.TLSPrivateKeyKey: []byte(`key`),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret-3",
				Namespace: "test",
			},
			Type: v1.SecretTypeTLS,
			Data: map[string][]byte{
				v1.TLSCertKey:       []byte(``),
				v1.TLSPrivateKeyKey: []byte(``),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret-ca-cert",
				Namespace: "test",
			},
			Type: v1.SecretTypeTLS,
			Data: map[string][]byte{
				v1.TLSCertKey: []byte(`CAcert`),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret-ca-cert-opaque",
				Namespace: "test",
			},
			Type: v1.SecretTypeOpaque,
			Data: map[string][]byte{
				v1.TLSCertKey: []byte(`CAcert-from-opaque`),
			},
		},
	}}
	boolTrue := true
	customIngressClassName := "custom"
	openshiftCustomIngressClassName := "openshift-custom"
	openshiftDefaultIngressClassName := "openshift-default"
	pathTypeExact := networkingv1.PathTypeExact
	pathTypePrefix := networkingv1.PathTypePrefix
	pathTypeImplementationSpecific := networkingv1.PathTypeImplementationSpecific
	type fields struct {
		i   networkingv1listers.IngressLister
		ic  networkingv1listers.IngressClassLister
		r   routelisters.RouteLister
		s   corelisters.SecretLister
		svc corelisters.ServiceLister
	}
	tests := []struct {
		name               string
		fields             fields
		args               queueKey
		expects            *expectations
		wantErr            bool
		wantRouteCreates   []*routev1.Route
		wantRoutePatches   []clientgotesting.PatchActionImpl
		wantRouteDeletes   []clientgotesting.DeleteActionImpl
		wantIngressUpdates []clientgotesting.UpdateActionImpl
		wantQueue          []queueKey
		wantExpectation    *expectations
		wantExpects        []queueKey
	}{
		{
			name:   "no changes",
			fields: fields{i: &ingressLister{}, r: &routeLister{}},
			args:   queueKey{namespace: "test", name: "1"},
		},
		{
			name:   "sync namespace - no ingress",
			fields: fields{i: &ingressLister{}, r: &routeLister{}},
			args:   queueKey{namespace: "test"},
		},
		{
			name: "sync namespace - two ingress",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "2",
							Namespace: "test",
						},
					},
				}},
				r: &routeLister{},
			},
			args: queueKey{namespace: "test"},
			wantQueue: []queueKey{
				{namespace: "test", name: "1"},
				{namespace: "test", name: "2"},
			},
		},
		{
			name: "ignores incomplete ingress - no host",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/deep",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{},
			},
			args: queueKey{namespace: "test", name: "1"},
		},
		{
			name: "ignores incomplete ingress - no service",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/deep",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{},
			},
			args: queueKey{namespace: "test", name: "1"},
		},
		{
			name: "ignores incomplete ingress - no paths",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{},
			},
			args: queueKey{namespace: "test", name: "1"},
		},
		{
			name: "ignores ingress with third-party ingressclass",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							IngressClassName: &customIngressClassName,
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/deep",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				ic: &ingressclassLister{Items: []*networkingv1.IngressClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "custom",
						},
						Spec: networkingv1.IngressClassSpec{
							Controller: "acme.io/ingress-controller",
						},
					},
				}},
				r: &routeLister{},
			},
			args: queueKey{namespace: "test", name: "1"},
		},
		{
			name: "ignores ingress with unsupported path type",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path: "/deep",
													// "Exact" is not implemented.
													PathType: &pathTypeExact,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{},
			},
			args: queueKey{namespace: "test", name: "1"},
		},
		{
			name: "ignores incomplete ingress - service does not exist",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-3",
															Port: networkingv1.ServiceBackendPort{
																Number: int32(80),
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{},
			},
			args: queueKey{namespace: "test", name: "1"},
		},
		{
			name: "create route",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path: "/deep",
													// Behavior for empty PathType is undefined;
													// treat it the same as "Prefix".
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
												{
													Path: "/",
													// Implementations may treat "ImplementationSpecific"
													// as "Exact" or "Prefix", so we treat it as "Prefix".
													PathType: &pathTypeImplementationSpecific,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{},
			},
			args:        queueKey{namespace: "test", name: "1"},
			wantExpects: []queueKey{{namespace: "test", name: "1"}},
			wantRouteCreates: []*routev1.Route{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "<generated>",
						Namespace:       "test",
						OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
					},
					Spec: routev1.RouteSpec{
						Host: "test.com",
						Path: "/deep",
						To: routev1.RouteTargetReference{
							Kind: "Service",
							Name: "service-1",
						},
						Port: &routev1.RoutePort{
							TargetPort: intstr.FromString("http"),
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "<generated>",
						Namespace:       "test",
						OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
					},
					Spec: routev1.RouteSpec{
						Host: "test.com",
						Path: "/",
						To: routev1.RouteTargetReference{
							Kind: "Service",
							Name: "service-1",
						},
						Port: &routev1.RoutePort{
							TargetPort: intstr.FromString("http"),
						},
					},
				},
			},
		},
		{
			name: "create route - with termination reencypt and destinationCaCert",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
							Annotations: map[string]string{
								"route.openshift.io/termination":                       "reencrypt",
								"route.openshift.io/destination-ca-certificate-secret": "secret-ca-cert",
							},
						},
						Spec: networkingv1.IngressSpec{
							TLS: []networkingv1.IngressTLS{
								{Hosts: []string{"test.com"}, SecretName: "secret-1"},
							},
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{},
			},
			args:        queueKey{namespace: "test", name: "1"},
			wantExpects: []queueKey{{namespace: "test", name: "1"}},
			wantRouteCreates: []*routev1.Route{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "<generated>",
						Namespace:       "test",
						OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						Annotations: map[string]string{
							"route.openshift.io/termination":                       "reencrypt",
							"route.openshift.io/destination-ca-certificate-secret": "secret-ca-cert",
						},
					},
					Spec: routev1.RouteSpec{
						Host: "test.com",
						Path: "/",
						To: routev1.RouteTargetReference{
							Kind: "Service",
							Name: "service-1",
						},
						Port: &routev1.RoutePort{
							TargetPort: intstr.FromString("http"),
						},
						TLS: &routev1.TLSConfig{
							Termination:                   routev1.TLSTerminationReencrypt,
							Key:                           "key",
							Certificate:                   "cert",
							DestinationCACertificate:      "CAcert",
							InsecureEdgeTerminationPolicy: "Redirect",
						},
					},
				},
			},
		},
		{
			name: "create route - targetPort string, service port with name",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-2",
															Port: networkingv1.ServiceBackendPort{
																Number: int32(80),
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{},
			},
			args:        queueKey{namespace: "test", name: "1"},
			wantExpects: []queueKey{{namespace: "test", name: "1"}},
			wantRouteCreates: []*routev1.Route{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "<generated>",
						Namespace:       "test",
						OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
					},
					Spec: routev1.RouteSpec{
						Host: "test.com",
						Path: "/",
						To: routev1.RouteTargetReference{
							Kind: "Service",
							Name: "service-2",
						},
						Port: &routev1.RoutePort{
							TargetPort: intstr.FromString("80-tcp"),
						},
					},
				},
			},
		},
		{
			name: "create route - default ingresscontroller",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							IngressClassName: &openshiftDefaultIngressClassName,
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				ic: &ingressclassLister{Items: []*networkingv1.IngressClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "openshift-default",
						},
						Spec: networkingv1.IngressClassSpec{
							Controller: "openshift.io/ingress-to-route",
							Parameters: &networkingv1.IngressClassParametersReference{
								APIGroup: &operatorv1GroupVersion,
								Kind:     "IngressController",
								Name:     "default",
							},
						},
					},
				}},
				r: &routeLister{},
			},
			args:        queueKey{namespace: "test", name: "1"},
			wantExpects: []queueKey{{namespace: "test", name: "1"}},
			wantRouteCreates: []*routev1.Route{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "<generated>",
						Namespace:       "test",
						OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
					},
					Spec: routev1.RouteSpec{
						Host: "test.com",
						To: routev1.RouteTargetReference{
							Kind: "Service",
							Name: "service-1",
						},
						Port: &routev1.RoutePort{
							TargetPort: intstr.FromString("http"),
						},
					},
				},
			},
		},
		{
			name: "create route - custom ingresscontroller",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							IngressClassName: &openshiftCustomIngressClassName,
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				ic: &ingressclassLister{Items: []*networkingv1.IngressClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "openshift-custom",
						},
						Spec: networkingv1.IngressClassSpec{
							Controller: "openshift.io/ingress-to-route",
							Parameters: &networkingv1.IngressClassParametersReference{
								APIGroup: &operatorv1GroupVersion,
								Kind:     "IngressController",
								Name:     "custom",
							},
						},
					},
				}},
				r: &routeLister{},
			},
			args:        queueKey{namespace: "test", name: "1"},
			wantExpects: []queueKey{{namespace: "test", name: "1"}},
			wantRouteCreates: []*routev1.Route{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "<generated>",
						Namespace:       "test",
						OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
					},
					Spec: routev1.RouteSpec{
						Host: "test.com",
						To: routev1.RouteTargetReference{
							Kind: "Service",
							Name: "service-1",
						},
						Port: &routev1.RoutePort{
							TargetPort: intstr.FromString("http"),
						},
					},
				},
			},
		},
		{
			name: "create route - blocked by expectation",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/deep",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{},
			},
			expects: newTestExpectations(func(e *expectations) {
				e.Expect("test", "1", "route-test-1")
			}),
			args:      queueKey{namespace: "test", name: "1"},
			wantQueue: []queueKey{{namespace: "test", name: "1"}},
			// preserves the expectations unchanged
			wantExpectation: newTestExpectations(func(e *expectations) {
				e.Expect("test", "1", "route-test-1")
			}),
		},
		{
			name: "update route",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromInt(80),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
			wantRoutePatches: []clientgotesting.PatchActionImpl{
				{
					Name:  "1-abcdef",
					Patch: []byte(`[{"op":"replace","path":"/spec","value":{"host":"test.com","path":"/","to":{"kind":"Service","name":"service-1","weight":null},"port":{"targetPort":"http"}}},{"op":"replace","path":"/metadata/annotations","value":null},{"op":"replace","path":"/metadata/ownerReferences","value":[{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","name":"1","uid":"","controller":true}]}]`),
				},
			},
		},
		{
			name: "no-op",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
		},
		{
			name: "no-op - ignore partially owned resource",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					// this route is identical to the ingress
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
					},
					// this route should be left as is because controller is not true
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-empty",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1"}},
						},
						Spec: routev1.RouteSpec{},
					},
					// this route should be ignored because it doesn't match the ingress name
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "2-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "2", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromInt(8080),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
		},
		{
			name: "no-op - ignore route created for an ingress with a third-party class",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							IngressClassName: &customIngressClassName,
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/foo",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/bar",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
		},
		{
			name: "no-op - destination CA certificate has been changed by the user",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
							Annotations: map[string]string{
								"route.openshift.io/termination": "reencrypt",
							},
						},
						Spec: networkingv1.IngressSpec{
							TLS: []networkingv1.IngressTLS{
								{Hosts: []string{"test.com"}, SecretName: "secret-1"},
							},
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						Status: networkingv1.IngressStatus{
							LoadBalancer: networkingv1.IngressLoadBalancerStatus{
								Ingress: []networkingv1.IngressLoadBalancerIngress{
									{Hostname: "apps.foo.com"},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
							Annotations: map[string]string{
								"route.openshift.io/termination": "reencrypt",
							},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
							TLS: &routev1.TLSConfig{
								Termination:              routev1.TLSTerminationReencrypt,
								Key:                      "key",
								Certificate:              "cert",
								DestinationCACertificate: "cert",
							},
						},
						Status: routev1.RouteStatus{
							Ingress: []routev1.RouteIngress{
								{
									RouterCanonicalHostname: "apps.foo.com",
									Conditions: []routev1.RouteIngressCondition{{
										Type:   routev1.RouteAdmitted,
										Status: v1.ConditionTrue,
									}},
								},
							},
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
		},
		{
			name: "update ingress with missing secret ref",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							TLS: []networkingv1.IngressTLS{
								{Hosts: []string{"test.com"}, SecretName: "secret-4"},
							},
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
			wantRouteDeletes: []clientgotesting.DeleteActionImpl{
				{
					Name: "1-abcdef",
				},
			},
		},
		{
			name: "update ingress with missing secret ref",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							TLS: []networkingv1.IngressTLS{
								{Hosts: []string{"test.com"}, SecretName: "secret-4"},
							},
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
			wantRouteDeletes: []clientgotesting.DeleteActionImpl{
				{
					Name: "1-abcdef",
				},
			},
		},
		{
			name: "update ingress to not reference secret",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							TLS: []networkingv1.IngressTLS{
								{Hosts: []string{"test.com1"}, SecretName: "secret-1"},
							},
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
							TLS: &routev1.TLSConfig{
								Termination:                   routev1.TLSTerminationEdge,
								InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
								Key:                           "key",
								Certificate:                   "cert",
							},
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
			wantRoutePatches: []clientgotesting.PatchActionImpl{
				{
					Name:  "1-abcdef",
					Patch: []byte(`[{"op":"replace","path":"/spec","value":{"host":"test.com","path":"/","to":{"kind":"Service","name":"service-1","weight":null},"port":{"targetPort":"http"}}},{"op":"replace","path":"/metadata/annotations","value":null},{"op":"replace","path":"/metadata/ownerReferences","value":[{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","name":"1","uid":"","controller":true}]}]`),
				},
			},
		},
		{
			name: "update route with old owner reference",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "extensions.k8s.io/v1beta1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
			wantRoutePatches: []clientgotesting.PatchActionImpl{
				{
					Name:  "1-abcdef",
					Patch: []byte(`[{"op":"replace","path":"/spec","value":{"host":"test.com","path":"/","to":{"kind":"Service","name":"service-1","weight":null},"port":{"targetPort":"http"}}},{"op":"replace","path":"/metadata/annotations","value":null},{"op":"replace","path":"/metadata/ownerReferences","value":[{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","name":"1","uid":"","controller":true}]}]`),
				},
			},
		},
		{
			name: "update route - tls config missing",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							TLS: []networkingv1.IngressTLS{
								{Hosts: []string{"test.com"}, SecretName: "secret-1"},
							},
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
			wantRoutePatches: []clientgotesting.PatchActionImpl{
				{
					Name:  "1-abcdef",
					Patch: []byte(`[{"op":"replace","path":"/spec","value":{"host":"test.com","path":"/","to":{"kind":"Service","name":"service-1","weight":null},"port":{"targetPort":"http"},"tls":{"termination":"edge","certificate":"cert","key":"key","insecureEdgeTerminationPolicy":"Redirect"}}},{"op":"replace","path":"/metadata/annotations","value":null},{"op":"replace","path":"/metadata/ownerReferences","value":[{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","name":"1","uid":"","controller":true}]}]`),
				},
			},
		},
		{
			name: "update route - termination policy changed to passthrough",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
							Annotations: map[string]string{
								"route.openshift.io/termination": "passthrough",
							},
						},
						Spec: networkingv1.IngressSpec{
							TLS: []networkingv1.IngressTLS{
								{Hosts: []string{"test.com"}, SecretName: "secret-1"},
							},
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							TLS: &routev1.TLSConfig{
								Termination:                   routev1.TLSTerminationEdge,
								Certificate:                   "cert",
								Key:                           "key",
								InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
							},
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
			wantRoutePatches: []clientgotesting.PatchActionImpl{
				{
					Name:  "1-abcdef",
					Patch: []byte(`[{"op":"replace","path":"/spec","value":{"host":"test.com","path":"/","to":{"kind":"Service","name":"service-1","weight":null},"port":{"targetPort":"http"},"tls":{"termination":"passthrough","insecureEdgeTerminationPolicy":"Redirect"}}},{"op":"replace","path":"/metadata/annotations","value":{"route.openshift.io/termination":"passthrough"}},{"op":"replace","path":"/metadata/ownerReferences","value":[{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","name":"1","uid":"","controller":true}]}]`),
				},
			},
		},
		{
			name: "update route - termination policy changed to reencrypt",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
							Annotations: map[string]string{
								"route.openshift.io/termination": "reencrypt",
							},
						},
						Spec: networkingv1.IngressSpec{
							TLS: []networkingv1.IngressTLS{
								{Hosts: []string{"test.com"}, SecretName: "secret-1"},
							},
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							TLS: &routev1.TLSConfig{
								Termination:                   routev1.TLSTerminationEdge,
								Certificate:                   "cert",
								Key:                           "key",
								InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
							},
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
			wantRoutePatches: []clientgotesting.PatchActionImpl{
				{
					Name:  "1-abcdef",
					Patch: []byte(`[{"op":"replace","path":"/spec","value":{"host":"test.com","path":"/","to":{"kind":"Service","name":"service-1","weight":null},"port":{"targetPort":"http"},"tls":{"termination":"reencrypt","certificate":"cert","key":"key","insecureEdgeTerminationPolicy":"Redirect"}}},{"op":"replace","path":"/metadata/annotations","value":{"route.openshift.io/termination":"reencrypt"}},{"op":"replace","path":"/metadata/ownerReferences","value":[{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","name":"1","uid":"","controller":true}]}]`),
				},
			},
		},
		{
			name: "update route - termination policy changed to reencrypt and no tls secret",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
							Annotations: map[string]string{
								"route.openshift.io/termination": "reencrypt",
							},
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							TLS: &routev1.TLSConfig{
								Termination:                   routev1.TLSTerminationEdge,
								Certificate:                   "cert",
								Key:                           "key",
								InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
							},
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
			wantRoutePatches: []clientgotesting.PatchActionImpl{
				{
					Name:  "1-abcdef",
					Patch: []byte(`[{"op":"replace","path":"/spec","value":{"host":"test.com","path":"/","to":{"kind":"Service","name":"service-1","weight":null},"port":{"targetPort":"http"},"tls":{"termination":"reencrypt","insecureEdgeTerminationPolicy":"Redirect"}}},` + `{"op":"replace","path":"/metadata/annotations","value":{"route.openshift.io/termination":"reencrypt"}},{"op":"replace","path":"/metadata/ownerReferences","value":[{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","name":"1","uid":"","controller":true}]}]`),
				},
			},
		},
		{
			name: "update route - termination policy changed to reencrypt with destCaCertCertificate",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
							Annotations: map[string]string{
								"route.openshift.io/termination":                       "reencrypt",
								"route.openshift.io/destination-ca-certificate-secret": "secret-ca-cert",
							},
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							TLS: &routev1.TLSConfig{
								Termination:                   routev1.TLSTerminationEdge,
								Certificate:                   "cert",
								Key:                           "key",
								InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
							},
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
			wantRoutePatches: []clientgotesting.PatchActionImpl{
				{
					Name: "1-abcdef",
					Patch: []byte(
						strings.Join(
							[]string{
								`[{"op":"replace","path":"/spec","value":{"host":"test.com","path":"/","to":{"kind":"Service","name":"service-1","weight":null},"port":{"targetPort":"http"},"tls":{"termination":"reencrypt","destinationCACertificate":"CAcert","insecureEdgeTerminationPolicy":"Redirect"}}}`,
								`{"op":"replace","path":"/metadata/annotations","value":{"route.openshift.io/destination-ca-certificate-secret":"secret-ca-cert","route.openshift.io/termination":"reencrypt"}}`,
								`{"op":"replace","path":"/metadata/ownerReferences","value":[{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","name":"1","uid":"","controller":true}]}]`,
							},
							",",
						),
					),
				},
			},
		},
		{
			name: "update route - termination policy changed from reencrypt to to edge - Must clear destinationCaCertificate",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
							Annotations: map[string]string{
								"route.openshift.io/termination": "edge",
							},
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1-abcdef",
							Namespace: "test",
							Annotations: map[string]string{
								"route.openshift.io/termination":                       "reencrypt",
								"route.openshift.io/destination-ca-certificate-secret": "secret-ca-cert",
							},
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							TLS: &routev1.TLSConfig{
								Termination:                   routev1.TLSTerminationReencrypt,
								Certificate:                   "cert",
								Key:                           "key",
								InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
								DestinationCACertificate:      "CACert",
							},
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
			wantRoutePatches: []clientgotesting.PatchActionImpl{
				{
					Name: "1-abcdef",
					Patch: []byte(
						strings.Join(
							[]string{
								`[{"op":"replace","path":"/spec","value":{"host":"test.com","path":"/","to":{"kind":"Service","name":"service-1","weight":null},"port":{"targetPort":"http"},"tls":{"termination":"edge","insecureEdgeTerminationPolicy":"Redirect"}}}`,
								`{"op":"replace","path":"/metadata/annotations","value":{"route.openshift.io/termination":"edge"}}`,
								`{"op":"replace","path":"/metadata/ownerReferences","value":[{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","name":"1","uid":"","controller":true}]}]`,
							},
							",",
						),
					),
				},
			},
		},
		{
			name: "update route - destination-ca-certificate-secret type changed",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
							Annotations: map[string]string{
								"route.openshift.io/termination":                       "reencrypt",
								"route.openshift.io/destination-ca-certificate-secret": "secret-ca-cert-opaque",
							},
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
							Annotations: map[string]string{
								"route.openshift.io/termination":                       "reencrypt",
								"route.openshift.io/destination-ca-certificate-secret": "secret-ca-cert",
							},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							TLS: &routev1.TLSConfig{
								Termination:                   routev1.TLSTerminationEdge,
								Certificate:                   "cert",
								DestinationCACertificate:      "CACert",
								Key:                           "key",
								InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
							},
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
			wantRoutePatches: []clientgotesting.PatchActionImpl{
				{
					Name: "1-abcdef",
					Patch: []byte(
						strings.Join(
							[]string{
								`[{"op":"replace","path":"/spec","value":{"host":"test.com","path":"/","to":{"kind":"Service","name":"service-1","weight":null},"port":{"targetPort":"http"},"tls":{"termination":"reencrypt","destinationCACertificate":"CAcert-from-opaque","insecureEdgeTerminationPolicy":"Redirect"}}}`,
								`{"op":"replace","path":"/metadata/annotations","value":{"route.openshift.io/destination-ca-certificate-secret":"secret-ca-cert-opaque","route.openshift.io/termination":"reencrypt"}}`,
								`{"op":"replace","path":"/metadata/ownerReferences","value":[{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","name":"1","uid":"","controller":true}]}]`,
							},
							",",
						),
					),
				},
			},
		},
		{
			name: "termination policy on ingress invalid, nothing happens",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
							Annotations: map[string]string{
								"route.openshift.io/termination": "Passthrough",
							},
						},
						Spec: networkingv1.IngressSpec{
							TLS: []networkingv1.IngressTLS{
								{Hosts: []string{"test.com"}, SecretName: "secret-1"},
							},
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
							Annotations:     map[string]string{"route.openshift.io/termination": "Passthrough"},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							TLS: &routev1.TLSConfig{
								Termination:                   routev1.TLSTerminationEdge,
								Certificate:                   "cert",
								Key:                           "key",
								InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
							},
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
		},
		{
			name: "termination policy on ingress invalid, disables tls",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
							Annotations: map[string]string{
								"route.openshift.io/termination": "Passthrough",
							},
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							TLS: &routev1.TLSConfig{
								Termination:                   routev1.TLSTerminationEdge,
								Certificate:                   "cert",
								Key:                           "key",
								InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
							},
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
			wantRoutePatches: []clientgotesting.PatchActionImpl{
				{
					Name:  "1-abcdef",
					Patch: []byte(`[{"op":"replace","path":"/spec","value":{"host":"test.com","path":"/","to":{"kind":"Service","name":"service-1","weight":null},"port":{"targetPort":"http"}}},` + `{"op":"replace","path":"/metadata/annotations","value":{"route.openshift.io/termination":"Passthrough"}},{"op":"replace","path":"/metadata/ownerReferences","value":[{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","name":"1","uid":"","controller":true}]}]`),
				},
			},
		},
		{
			name: "Empty tlsconfig enables edge termination without explicit cert",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							TLS: []networkingv1.IngressTLS{{Hosts: []string{"something-else"}}, {}},
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
			wantRoutePatches: []clientgotesting.PatchActionImpl{
				{
					Name:  "1-abcdef",
					Patch: []byte(`[{"op":"replace","path":"/spec","value":{"host":"test.com","path":"/","to":{"kind":"Service","name":"service-1","weight":null},"port":{"targetPort":"http"},"tls":{"termination":"edge","insecureEdgeTerminationPolicy":"Redirect"}}},{"op":"replace","path":"/metadata/annotations","value":null},{"op":"replace","path":"/metadata/ownerReferences","value":[{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","name":"1","uid":"","controller":true}]}]`),
				},
			},
		},
		{
			name: "update route - secret values changed",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							TLS: []networkingv1.IngressTLS{
								{Hosts: []string{"test.com"}, SecretName: "secret-1a"},
							},
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
							TLS: &routev1.TLSConfig{
								Termination: routev1.TLSTerminationEdge,
								Key:         "key",
								Certificate: "cert",
							},
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
			wantRoutePatches: []clientgotesting.PatchActionImpl{
				{
					Name:  "1-abcdef",
					Patch: []byte(`[{"op":"replace","path":"/spec","value":{"host":"test.com","path":"/","to":{"kind":"Service","name":"service-1","weight":null},"port":{"targetPort":"http"},"tls":{"termination":"edge","certificate":"cert","key":"key2"}}},{"op":"replace","path":"/metadata/annotations","value":null},{"op":"replace","path":"/metadata/ownerReferences","value":[{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","name":"1","uid":"","controller":true}]}]`),
				},
			},
		},
		{
			name: "no-op - has TLS",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							TLS: []networkingv1.IngressTLS{
								{Hosts: []string{"test.com"}, SecretName: "secret-1"},
							},
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
							TLS: &routev1.TLSConfig{
								Termination:                   routev1.TLSTerminationEdge,
								InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
								Key:                           "key",
								Certificate:                   "cert",
							},
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
		},
		{
			name: "no-op - has secret with empty keys",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							TLS: []networkingv1.IngressTLS{
								{Hosts: []string{"test.com"}, SecretName: "secret-3"},
							},
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
							TLS: &routev1.TLSConfig{
								Termination:                   routev1.TLSTerminationEdge,
								InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
								Key:                           "",
								Certificate:                   "",
							},
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
		},
		{
			name: "no-op - termination policy has been changed by the user",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							TLS: []networkingv1.IngressTLS{
								{Hosts: []string{"test.com"}, SecretName: "secret-1"},
							},
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						Status: networkingv1.IngressStatus{
							LoadBalancer: networkingv1.IngressLoadBalancerStatus{
								Ingress: []networkingv1.IngressLoadBalancerIngress{
									{Hostname: "apps.foo.com"},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
							TLS: &routev1.TLSConfig{
								Termination: routev1.TLSTerminationEdge,
								Key:         "key",
								Certificate: "cert",
							},
						},
						Status: routev1.RouteStatus{
							Ingress: []routev1.RouteIngress{
								{
									RouterCanonicalHostname: "apps.foo.com",
									Conditions: []routev1.RouteIngressCondition{{
										Type:   routev1.RouteAdmitted,
										Status: v1.ConditionTrue,
									}},
								},
							},
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
		},
		{
			name: "update route - router admitted route",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						Status: networkingv1.IngressStatus{},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
						Status: routev1.RouteStatus{
							Ingress: []routev1.RouteIngress{
								{
									RouterCanonicalHostname: "apps.foo.com",
									Conditions: []routev1.RouteIngressCondition{{
										Type:   routev1.RouteAdmitted,
										Status: v1.ConditionTrue,
									}},
								},
								{
									RouterCanonicalHostname: "apps.bar.com",
									Conditions: []routev1.RouteIngressCondition{{
										Type:   routev1.RouteAdmitted,
										Status: v1.ConditionFalse,
									}},
								},
							},
						},
					},
				}},
			},
			wantIngressUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: &networkingv1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Status: networkingv1.IngressStatus{
							LoadBalancer: networkingv1.IngressLoadBalancerStatus{
								Ingress: []networkingv1.IngressLoadBalancerIngress{{
									Hostname: "apps.foo.com",
								}},
							},
						},
					},
				},
			},
			args: queueKey{namespace: "test", name: "1"},
		},
		{
			name: "update route - second router admitted route",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						Status: networkingv1.IngressStatus{
							LoadBalancer: networkingv1.IngressLoadBalancerStatus{
								Ingress: []networkingv1.IngressLoadBalancerIngress{
									{Hostname: "apps.foo.com"},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
						Status: routev1.RouteStatus{
							Ingress: []routev1.RouteIngress{
								{
									RouterCanonicalHostname: "apps.foo.com",
									Conditions: []routev1.RouteIngressCondition{{
										Type:   routev1.RouteAdmitted,
										Status: v1.ConditionTrue,
									}},
								},
								{
									RouterCanonicalHostname: "apps.bar.com",
									Conditions: []routev1.RouteIngressCondition{{
										Type:   routev1.RouteAdmitted,
										Status: v1.ConditionTrue,
									}},
								},
							},
						},
					},
				}},
			},
			wantIngressUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: &networkingv1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Status: networkingv1.IngressStatus{
							LoadBalancer: networkingv1.IngressLoadBalancerStatus{
								Ingress: []networkingv1.IngressLoadBalancerIngress{
									{Hostname: "apps.bar.com"},
									{Hostname: "apps.foo.com"},
								},
							},
						},
					},
				},
			},
			args: queueKey{namespace: "test", name: "1"},
		},
		{
			name: "no-op - ingress status already updated",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						Status: networkingv1.IngressStatus{
							LoadBalancer: networkingv1.IngressLoadBalancerStatus{
								Ingress: []networkingv1.IngressLoadBalancerIngress{
									{Hostname: "apps.foo.com"},
									{Hostname: "apps.bar.com"},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
						Status: routev1.RouteStatus{
							Ingress: []routev1.RouteIngress{
								{
									RouterCanonicalHostname: "apps.foo.com",
									Conditions: []routev1.RouteIngressCondition{{
										Type:   routev1.RouteAdmitted,
										Status: v1.ConditionTrue,
									}},
								},
								{
									RouterCanonicalHostname: "apps.bar.com",
									Conditions: []routev1.RouteIngressCondition{{
										Type:   routev1.RouteAdmitted,
										Status: v1.ConditionTrue,
									}},
								},
							},
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
		},
		{
			name: "no-op - router rejected route",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
						Status: routev1.RouteStatus{
							Ingress: []routev1.RouteIngress{{
								RouterCanonicalHostname: "apps.testcluster.com",
								Conditions: []routev1.RouteIngressCondition{{
									Type:   routev1.RouteAdmitted,
									Status: v1.ConditionFalse,
								}},
							}},
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
		},
		{
			name: "delete route when referenced secret is not TLS",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							TLS: []networkingv1.IngressTLS{
								{Hosts: []string{"test.com"}, SecretName: "secret-0"},
							},
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						Status: networkingv1.IngressStatus{
							LoadBalancer: networkingv1.IngressLoadBalancerStatus{
								Ingress: []networkingv1.IngressLoadBalancerIngress{
									{Hostname: "apps.foo.com"},
									{Hostname: "apps.bar.com"},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
							TLS: &routev1.TLSConfig{
								Termination:                   routev1.TLSTerminationEdge,
								InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
								Key:                           "key",
								Certificate:                   "cert",
							},
						},
						Status: routev1.RouteStatus{
							Ingress: []routev1.RouteIngress{{
								RouterCanonicalHostname: "apps.foo.com",
								Conditions: []routev1.RouteIngressCondition{{
									Type:   routev1.RouteAdmitted,
									Status: v1.ConditionTrue,
								}},
							}},
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
			wantRouteDeletes: []clientgotesting.DeleteActionImpl{
				{
					Name: "1-abcdef",
				},
			},
			wantIngressUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: &networkingv1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Status: networkingv1.IngressStatus{
							LoadBalancer: networkingv1.IngressLoadBalancerStatus{
								Ingress: []networkingv1.IngressLoadBalancerIngress{
									{Hostname: "apps.bar.com"},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "delete route when referenced secret is not valid",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							TLS: []networkingv1.IngressTLS{
								{Hosts: []string{"test.com"}, SecretName: "secret-2"},
							},
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						Status: networkingv1.IngressStatus{
							LoadBalancer: networkingv1.IngressLoadBalancerStatus{
								Ingress: []networkingv1.IngressLoadBalancerIngress{
									{Hostname: "apps.foo.com"},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
							TLS: &routev1.TLSConfig{
								Termination:                   routev1.TLSTerminationEdge,
								InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
								Key:                           "key",
								Certificate:                   "",
							},
						},
						Status: routev1.RouteStatus{
							Ingress: []routev1.RouteIngress{{
								RouterCanonicalHostname: "apps.foo.com",
								Conditions: []routev1.RouteIngressCondition{{
									Type:   routev1.RouteAdmitted,
									Status: v1.ConditionTrue,
								}},
							}},
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
			wantRouteDeletes: []clientgotesting.DeleteActionImpl{
				{
					Name: "1-abcdef",
				},
			},
			wantIngressUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: &networkingv1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Status: networkingv1.IngressStatus{
							LoadBalancer: networkingv1.IngressLoadBalancerStatus{
								Ingress: []networkingv1.IngressLoadBalancerIngress{},
							},
						},
					},
				},
			},
		},
		{
			name: "ignore route when parent ingress no longer exists (gc will handle)",
			fields: fields{
				i: &ingressLister{},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
		},
		{
			name: "update route - termination policy changed to passthrough and timeout set",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
							Annotations: map[string]string{
								"route.openshift.io/termination":      "passthrough",
								"haproxy.router.openshift.io/timeout": "6m",
							},
						},
						Spec: networkingv1.IngressSpec{
							TLS: []networkingv1.IngressTLS{
								{Hosts: []string{"test.com"}, SecretName: "secret-1"},
							},
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
							Path: "/",
							TLS: &routev1.TLSConfig{
								Termination:                   routev1.TLSTerminationEdge,
								Certificate:                   "cert",
								Key:                           "key",
								InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
							},
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromString("http"),
							},
							WildcardPolicy: routev1.WildcardPolicyNone,
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
			wantRoutePatches: []clientgotesting.PatchActionImpl{
				{
					Name:  "1-abcdef",
					Patch: []byte(`[{"op":"replace","path":"/spec","value":{"host":"test.com","path":"/","to":{"kind":"Service","name":"service-1","weight":null},"port":{"targetPort":"http"},"tls":{"termination":"passthrough","insecureEdgeTerminationPolicy":"Redirect"}}},{"op":"replace","path":"/metadata/annotations","value":{"haproxy.router.openshift.io/timeout":"6m","route.openshift.io/termination":"passthrough"}},{"op":"replace","path":"/metadata/ownerReferences","value":[{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","name":"1","uid":"","controller":true}]}]`),
				},
			},
		},
		{
			name: "create wildcard route - targetPort string, service port with name",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "*.test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path: "/", Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-2",
															Port: networkingv1.ServiceBackendPort{
																Number: 80,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{},
			},
			args:        queueKey{namespace: "test", name: "1"},
			wantExpects: []queueKey{{namespace: "test", name: "1"}},
			wantRouteCreates: []*routev1.Route{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "<generated>",
						Namespace:       "test",
						OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
					},
					Spec: routev1.RouteSpec{
						Host: "wildcard.test.com",
						Path: "/",
						To: routev1.RouteTargetReference{
							Kind: "Service",
							Name: "service-2",
						},
						Port: &routev1.RoutePort{
							TargetPort: intstr.FromString("80-tcp")},
						WildcardPolicy: "Subdomain",
					},
				},
			},
		},
		{
			name: "create wildcard route with TLS config",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							TLS: []networkingv1.IngressTLS{
								{Hosts: []string{"*.test.com"}, SecretName: "secret-1"},
							},
							Rules: []networkingv1.IngressRule{
								{
									Host: "*.test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path: "/",
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{},
			},
			args:        queueKey{namespace: "test", name: "1"},
			wantExpects: []queueKey{{namespace: "test", name: "1"}},
			wantRouteCreates: []*routev1.Route{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "<generated>",
						Namespace:       "test",
						OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
					},
					Spec: routev1.RouteSpec{
						Host: "wildcard.test.com",
						Path: "/",
						TLS: &routev1.TLSConfig{
							Termination:                   routev1.TLSTerminationEdge,
							InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
							Key:                           "key",
							Certificate:                   "cert",
						},
						To: routev1.RouteTargetReference{
							Kind: "Service",
							Name: "service-1",
						},
						Port: &routev1.RoutePort{
							TargetPort: intstr.FromString("http"),
						},
						WildcardPolicy: "Subdomain",
					},
				},
			},
		},
		{
			name: "update wildcard route ",
			fields: fields{
				i: &ingressLister{Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "*.test.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path: "/", Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "service-1",
															Port: networkingv1.ServiceBackendPort{
																Name: "http",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
				r: &routeLister{Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "1-abcdef",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "1", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "wildcard.test.com",
							Path: "/",
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "service-1",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromInt(80),
							},
							WildcardPolicy: routev1.WildcardPolicySubdomain,
						},
					},
				}},
			},
			args: queueKey{namespace: "test", name: "1"},
			wantRoutePatches: []clientgotesting.PatchActionImpl{
				{
					Name:  "1-abcdef",
					Patch: []byte(`[{"op":"replace","path":"/spec","value":{"host":"wildcard.test.com","path":"/","to":{"kind":"Service","name":"service-1","weight":null},"port":{"targetPort":"http"},"wildcardPolicy":"Subdomain"}},{"op":"replace","path":"/metadata/annotations","value":null},{"op":"replace","path":"/metadata/ownerReferences","value":[{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","name":"1","uid":"","controller":true}]}]`),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var names []string
			routeClientset := &routev1fake.Clientset{}
			routeClientset.AddReactor("*", "routes", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				switch a := action.(type) {
				case clientgotesting.CreateAction:
					obj := a.GetObject().DeepCopyObject()
					m := obj.(metav1.Object)
					if len(m.GetName()) == 0 {
						m.SetName(m.GetGenerateName())
					}
					names = append(names, m.GetName())
					return true, obj, nil
				}
				return true, nil, nil
			})
			kc := fake.NewSimpleClientset()
			kc.PrependReactor("*", "ingresses", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, nil
			})

			c := &Controller{
				queue:              workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ingress-to-route-test"),
				routeClient:        routeClientset.RouteV1(),
				ingressClient:      kc.NetworkingV1(),
				ingressLister:      tt.fields.i,
				ingressclassLister: tt.fields.ic,
				routeLister:        tt.fields.r,
				secretLister:       tt.fields.s,
				serviceLister:      tt.fields.svc,
				expectations:       tt.expects,
				eventRecorder:      record.NewFakeRecorder(100),
			}
			// default these
			if c.expectations == nil {
				c.expectations = newExpectations()
			}
			if c.ingressclassLister == nil {
				c.ingressclassLister = ingressclasses
			}
			if c.secretLister == nil {
				c.secretLister = secrets
			}
			if c.serviceLister == nil {
				c.serviceLister = services
			}

			if err := c.sync(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("Controller.sync() error = %v, wantErr %v", err, tt.wantErr)
			}

			c.queue.ShutDown()
			var hasQueue []queueKey
			for {
				key, shutdown := c.queue.Get()
				if shutdown {
					break
				}
				hasQueue = append(hasQueue, key.(queueKey))
			}
			if !reflect.DeepEqual(tt.wantQueue, hasQueue) {
				t.Errorf("unexpected queue: %s", diff.ObjectReflectDiff(tt.wantQueue, hasQueue))
			}

			wants := tt.wantExpectation
			if wants == nil {
				wants = newTestExpectations(func(e *expectations) {
					for _, key := range tt.wantExpects {
						for _, routeName := range names {
							e.Expect(key.namespace, key.name, routeName)
						}
					}
				})
			}
			if !reflect.DeepEqual(wants, c.expectations) {
				t.Errorf("unexpected expectations: %s", diff.ObjectReflectDiff(wants.expect, c.expectations.expect))
			}

			routeActions := routeClientset.Actions()

			for i := range tt.wantRouteCreates {
				if i > len(routeActions)-1 {
					t.Fatalf("Controller.sync() unexpected action[%d]: %#v", i, tt.wantRouteCreates[i])
				}
				if routeActions[i].GetVerb() != "create" {
					t.Fatalf("Controller.sync() unexpected action[%d]: %#v", i, tt.wantRouteCreates[i])
				}
				action := routeActions[i].(clientgotesting.CreateAction)
				if action.GetNamespace() != tt.args.namespace {
					t.Errorf("unexpected action[%d]: %#v", i, action)
				}
				obj := action.GetObject()
				if tt.wantRouteCreates[i].Name == "<generated>" {
					tt.wantRouteCreates[i].Name = names[0]
					names = names[1:]
				}
				if !reflect.DeepEqual(tt.wantRouteCreates[i], obj) {
					t.Errorf("unexpected create: %s", diff.ObjectReflectDiff(tt.wantRouteCreates[i], obj))
				}
			}
			routeActions = routeActions[len(tt.wantRouteCreates):]

			for i := range tt.wantRoutePatches {
				if i > len(routeActions)-1 {
					t.Fatalf("Controller.sync() unexpected actions: %#v", routeClientset.Actions())
				}
				if routeActions[i].GetVerb() != "patch" {
					t.Fatalf("Controller.sync() unexpected actions: %#v", routeClientset.Actions())
				}
				action := routeActions[i].(clientgotesting.PatchAction)
				if action.GetNamespace() != tt.args.namespace || action.GetName() != tt.wantRoutePatches[i].Name {
					t.Errorf("unexpected action[%d]: %#v", i, action)
				}
				if !reflect.DeepEqual(string(action.GetPatch()), string(tt.wantRoutePatches[i].Patch)) {
					t.Errorf("unexpected action[%d]: %s", i, string(action.GetPatch()))
				}
			}
			routeActions = routeActions[len(tt.wantRoutePatches):]

			for i := range tt.wantRouteDeletes {
				if i > len(routeActions)-1 {
					t.Fatalf("Controller.sync() unexpected actions: %#v", routeClientset.Actions())
				}
				if routeActions[i].GetVerb() != "delete" {
					t.Fatalf("Controller.sync() unexpected actions: %#v", routeClientset.Actions())
				}
				action := routeActions[i].(clientgotesting.DeleteAction)
				if action.GetName() != tt.wantRouteDeletes[i].Name || action.GetNamespace() != tt.args.namespace {
					t.Errorf("unexpected action[%d]: %#v", i, action)
				}
			}
			routeActions = routeActions[len(tt.wantRouteDeletes):]

			if len(routeActions) != 0 {
				t.Fatalf("Controller.sync() unexpected actions: %#v", routeActions)
			}

			ingressActions := kc.Actions()
			for i := range tt.wantIngressUpdates {
				if i > len(ingressActions)-1 {
					t.Fatalf("Controller.sync() unexpected actions: %#v", kc.Actions())
				}
				if ingressActions[i].GetVerb() != "update" {
					t.Fatalf("Controller.sync() unexpected actions: %#v", kc.Actions())
				}
				action := ingressActions[i].(clientgotesting.UpdateAction)
				ingress, ok := action.GetObject().(*networkingv1.Ingress)
				if !ok {
					t.Fatalf("Controller.sync() unexpected actions: %#v", kc.Actions())
				}
				if ingress.Name != tt.wantIngressUpdates[i].Object.(*networkingv1.Ingress).Name || ingress.Namespace != tt.args.namespace || !reflect.DeepEqual(ingress.Status, tt.wantIngressUpdates[i].Object.(*networkingv1.Ingress).Status) {
					t.Errorf("unexpected ingress action[%d]: %#v", i, action)
				}
			}
			ingressActions = ingressActions[len(tt.wantIngressUpdates):]
			if len(ingressActions) != 0 {
				t.Fatalf("Controller.sync() unexpected actions: %#v", ingressActions)
			}
		})
	}
}
