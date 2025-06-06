package ingress

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/component-base/metrics/legacyregistry"
)

type fakeResponseWriter struct {
	bytes.Buffer
	statusCode int
	header     http.Header
}

func (f *fakeResponseWriter) Header() http.Header {
	return f.header
}

func (f *fakeResponseWriter) WriteHeader(statusCode int) {
	f.statusCode = statusCode
}

func TestMetrics(t *testing.T) {
	boolTrue := true
	emptyString := ""
	customIngressClassName := "custom"
	openshiftDefaultIngressClassName := "openshift-default"

	testCases := []struct {
		name               string
		ingressLister      *ingressLister
		ingressclassLister *ingressclassLister
		routeLister        *routeLister
		expectedResponse   string
	}{
		{
			name: "Ingress with nil IngressClassName should return 1",
			ingressLister: &ingressLister{
				Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "nil-ingressclassname",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							IngressClassName: nil,
						},
					},
				},
			},
			ingressclassLister: &ingressclassLister{},
			routeLister:        &routeLister{},
			expectedResponse:   "openshift_ingress_to_route_controller_ingress_without_class_name{name=\"nil-ingressclassname\",namespace=\"test\"} 1",
		},
		{
			name: "Ingress with empty IngressClassName should return 1",
			ingressLister: &ingressLister{
				Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "empty-ingressclassname",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{},
					},
				},
			},
			ingressclassLister: &ingressclassLister{},
			routeLister:        &routeLister{},
			expectedResponse:   "openshift_ingress_to_route_controller_ingress_without_class_name{name=\"empty-ingressclassname\",namespace=\"test\"} 1",
		},
		{
			name: "Ingress with empty string IngressClassName should return 1",
			ingressLister: &ingressLister{
				Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "emptystring-ingressclassname",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							IngressClassName: &emptyString,
						},
					},
				},
			},
			ingressclassLister: &ingressclassLister{},
			routeLister:        &routeLister{},
			expectedResponse:   "openshift_ingress_to_route_controller_ingress_without_class_name{name=\"emptystring-ingressclassname\",namespace=\"test\"} 1",
		},
		{
			name: "Ingress with set IngressClassName should return 0",
			ingressLister: &ingressLister{
				Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "set-ingressclassname",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							IngressClassName: &openshiftDefaultIngressClassName,
						},
					},
				},
			},
			ingressclassLister: &ingressclassLister{},
			routeLister:        &routeLister{},
			expectedResponse:   "openshift_ingress_to_route_controller_ingress_without_class_name{name=\"set-ingressclassname\",namespace=\"test\"} 0",
		},
		{
			name: "Route with an unmanaged Ingress owner should return 1",
			ingressLister: &ingressLister{
				Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "not-managed",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							IngressClassName: &customIngressClassName,
						},
					},
				},
			},
			ingressclassLister: &ingressclassLister{
				Items: []*networkingv1.IngressClass{
					{ // IngressClass specifying "acme.io/ingress-controller" controller
						ObjectMeta: metav1.ObjectMeta{
							Name: customIngressClassName,
						},
						Spec: networkingv1.IngressClassSpec{
							Controller: "acme.io/ingress-controller",
						},
					},
				},
			},
			routeLister: &routeLister{
				Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "owned-by-unmanaged",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "not-managed", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
						},
					},
				},
			},
			expectedResponse: "openshift_ingress_to_route_controller_route_with_unmanaged_owner{host=\"test.com\",name=\"owned-by-unmanaged\",namespace=\"test\"} 1",
		},
		{
			name: "Route with a managed Ingress owner should return 0",
			ingressLister: &ingressLister{
				Items: []*networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "managed",
							Namespace: "test",
						},
						Spec: networkingv1.IngressSpec{
							IngressClassName: &openshiftDefaultIngressClassName,
						},
					},
				},
			},
			ingressclassLister: &ingressclassLister{
				Items: []*networkingv1.IngressClass{
					{ // IngressClass specifying "openshift.io/ingress-to-route" controller
						ObjectMeta: metav1.ObjectMeta{
							Name: openshiftDefaultIngressClassName,
						},
						Spec: networkingv1.IngressClassSpec{
							Controller: "openshift.io/ingress-to-route",
						},
					},
				},
			},
			routeLister: &routeLister{
				Items: []*routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "owned-by-managed",
							Namespace:       "test",
							OwnerReferences: []metav1.OwnerReference{{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Name: "managed", Controller: &boolTrue}},
						},
						Spec: routev1.RouteSpec{
							Host: "test.com",
						},
					},
				},
			},
			expectedResponse: "openshift_ingress_to_route_controller_route_with_unmanaged_owner{host=\"test.com\",name=\"owned-by-managed\",namespace=\"test\"} 0",
		},
	}

	i := ingressLister{}
	ic := ingressclassLister{}
	r := routeLister{}

	c := &Controller{
		ingressLister:      &i,
		ingressclassLister: &ic,
		routeLister:        &r,
	}
	legacyregistry.MustRegister(c)
	t.Cleanup(func() {
		// Unregister the collector when this test is done so that
		// subsequent tests can register it.
		legacyregistry.Registerer().Unregister(c)
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			i.Items = tc.ingressLister.Items
			ic.Items = tc.ingressclassLister.Items
			r.Items = tc.routeLister.Items
			h := promhttp.HandlerFor(legacyregistry.DefaultGatherer, promhttp.HandlerOpts{ErrorHandling: promhttp.HTTPErrorOnError})

			rw := &fakeResponseWriter{header: http.Header{}}
			h.ServeHTTP(rw, &http.Request{})

			respStr := rw.String()
			if !strings.Contains(respStr, tc.expectedResponse) {
				t.Errorf("expected string %s did not appear in %s", tc.expectedResponse, respStr)
			}
		})
	}
}

func Test_ResetIngressMetrics(t *testing.T) {
	ingress := func(namespace, name string, ingressClassName *string) *networkingv1.Ingress {
		return &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: networkingv1.IngressSpec{
				IngressClassName: ingressClassName,
			},
		}
	}
	defaultIngressClassName := "openshift-default"
	i1 := ingress("test", "ingress1", &defaultIngressClassName)
	i2 := ingress("test", "ingress2", nil)
	i3 := ingress("test", "ingress3", nil)
	c := &Controller{
		ingressLister: &ingressLister{
			Items: []*networkingv1.Ingress{i1, i2, i3},
		},
		ingressclassLister: &ingressclassLister{
			Items: []*networkingv1.IngressClass{},
		},
		routeLister: &routeLister{},
	}
	legacyregistry.MustRegister(c)
	t.Cleanup(func() {
		// Unregister the collector when this test is done so that
		// subsequent tests can register it.
		legacyregistry.Registerer().Unregister(c)
	})

	assertMetrics := func(t *testing.T, expectedResponse string) {
		h := promhttp.HandlerFor(legacyregistry.DefaultGatherer, promhttp.HandlerOpts{ErrorHandling: promhttp.HTTPErrorOnError})

		rw := &fakeResponseWriter{header: http.Header{}}
		h.ServeHTTP(rw, &http.Request{})

		respStr := rw.String()
		if !strings.Contains(respStr, expectedResponse) {
			t.Fatalf("expected string did not appear in response\nexpected:\n%s\n\nresponse:\n %s", expectedResponse, respStr)
		}
	}

	t.Log("Initially, ingress1 has openshift-default, ingress2 has null, ingress3 has null.")

	assertMetrics(t,
		`openshift_ingress_to_route_controller_ingress_without_class_name{name="ingress1",namespace="test"} 0`+"\n"+
			`openshift_ingress_to_route_controller_ingress_without_class_name{name="ingress2",namespace="test"} 1`+"\n"+
			`openshift_ingress_to_route_controller_ingress_without_class_name{name="ingress3",namespace="test"} 1`+"\n")

	t.Log("Simulate deletion of ingress2 and update of ingress3.")

	c.ingressLister = &ingressLister{Items: []*networkingv1.Ingress{i1, i3}}
	c.ResetIngressMetrics(i2.Namespace, i2.Name)
	i3.Spec.IngressClassName = &defaultIngressClassName

	assertMetrics(t,
		`openshift_ingress_to_route_controller_ingress_without_class_name{name="ingress1",namespace="test"} 0`+"\n"+
			`openshift_ingress_to_route_controller_ingress_without_class_name{name="ingress2",namespace="test"} 0`+"\n"+
			`openshift_ingress_to_route_controller_ingress_without_class_name{name="ingress3",namespace="test"} 0`+"\n")
}
