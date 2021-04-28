// Package image implements evaluators of usage for imagestreams and images. They are supposed
// to be passed to resource quota controller and origin resource quota admission plugin.
package image

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apiserverquotav1 "k8s.io/apiserver/pkg/quota/v1"
	"k8s.io/apiserver/pkg/quota/v1/generic"

	imagev1 "github.com/openshift/api/image/v1"
	imagev1typedclient "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	imagev1informer "github.com/openshift/client-go/image/informers/externalversions/image/v1"
)

var legacyObjectCountAliases = map[schema.GroupVersionResource]corev1.ResourceName{
	imagev1.GroupVersion.WithResource("imagestreams"): imagev1.ResourceImageStreams,
}

// NewEvaluators returns the list of static evaluators that manage more than counts
func NewReplenishmentEvaluators(f apiserverquotav1.ListerForResourceFunc, isInformer imagev1informer.ImageStreamInformer, imageClient imagev1typedclient.ImageStreamTagsGetter) []apiserverquotav1.Evaluator {
	// these evaluators have special logic
	result := []apiserverquotav1.Evaluator{
		NewImageStreamTagEvaluator(isInformer.Lister(), imageClient),
		NewImageStreamImportEvaluator(isInformer.Lister()),
	}
	// these evaluators require an alias for backwards compatibility
	for gvr, alias := range legacyObjectCountAliases {
		result = append(result,
			generic.NewObjectCountEvaluator(gvr.GroupResource(), generic.ListResourceUsingListerFunc(f, gvr), alias))
	}
	return result
}
