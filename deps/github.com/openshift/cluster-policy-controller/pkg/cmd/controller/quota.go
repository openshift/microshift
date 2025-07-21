package controller

import (
	"context"
	"math/rand"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kquota "k8s.io/apiserver/pkg/quota/v1"
	"k8s.io/apiserver/pkg/quota/v1/generic"
	"k8s.io/kubernetes/pkg/controller"
	kresourcequota "k8s.io/kubernetes/pkg/controller/resourcequota"
	quotainstall "k8s.io/kubernetes/pkg/quota/v1/install"

	"github.com/openshift/library-go/pkg/quota/clusterquotamapping"

	"github.com/openshift/cluster-policy-controller/pkg/quota/clusterquotareconciliation"
	image "github.com/openshift/cluster-policy-controller/pkg/quota/quotaimageexternal"
)

func RunResourceQuotaManager(ctx context.Context, controllerCtx *EnhancedControllerContext) (bool, error) {
	concurrentResourceQuotaSyncs := int(controllerCtx.OpenshiftControllerConfig.ResourceQuota.ConcurrentSyncs)
	resourceQuotaSyncPeriod := controllerCtx.OpenshiftControllerConfig.ResourceQuota.SyncPeriod.Duration
	replenishmentSyncPeriodFunc := calculateResyncPeriod(controllerCtx.OpenshiftControllerConfig.ResourceQuota.MinResyncPeriod.Duration)
	saName := "resourcequota-controller"
	listerFuncForResource := generic.ListerFuncForResourceFunc(controllerCtx.GenericResourceInformer.ForResource)
	quotaConfiguration := quotainstall.NewQuotaConfigurationForControllers(listerFuncForResource)
	resourceQuotaControllerClient := controllerCtx.ClientBuilder.ClientOrDie(saName)
	imageEvaluators := image.NewReplenishmentEvaluators(
		listerFuncForResource,
		controllerCtx.ImageInformers.Image().V1().ImageStreams(),
		controllerCtx.ClientBuilder.OpenshiftImageClientOrDie(saName).ImageV1())
	resourceQuotaRegistry := generic.NewRegistry(imageEvaluators)
	discoveryFunc := resourceQuotaDiscoveryWrapper(resourceQuotaRegistry, resourceQuotaControllerClient.Discovery().ServerPreferredNamespacedResources)

	resourceQuotaControllerOptions := &kresourcequota.ControllerOptions{
		QuotaClient:               resourceQuotaControllerClient.CoreV1(),
		ResourceQuotaInformer:     controllerCtx.KubernetesInformers.Core().V1().ResourceQuotas(),
		ResyncPeriod:              controller.StaticResyncPeriodFunc(resourceQuotaSyncPeriod),
		Registry:                  resourceQuotaRegistry,
		ReplenishmentResyncPeriod: replenishmentSyncPeriodFunc,
		IgnoredResourcesFunc:      quotaConfiguration.IgnoredResources,
		InformersStarted:          controllerCtx.InformersStarted,
		InformerFactory:           controllerCtx.GenericResourceInformer,
		DiscoveryFunc:             discoveryFunc,
	}
	ctrl, err := kresourcequota.NewController(ctx, resourceQuotaControllerOptions)
	if err != nil {
		return true, err
	}
	go ctrl.Run(ctx, concurrentResourceQuotaSyncs)
	go ctrl.Sync(ctx, discoveryFunc, 30*time.Second)

	return true, nil
}

func resourceQuotaDiscoveryWrapper(registry kquota.Registry, discoveryFunc kresourcequota.NamespacedResourcesFunc) kresourcequota.NamespacedResourcesFunc {
	return func() ([]*metav1.APIResourceList, error) {
		discoveryResources, discoveryErr := discoveryFunc()
		if discoveryErr != nil && len(discoveryResources) == 0 {
			return nil, discoveryErr
		}

		interestingResources := []*metav1.APIResourceList{}
		for _, resourceList := range discoveryResources {
			gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
			if err != nil {
				return nil, err
			}
			for i := range resourceList.APIResources {
				gr := schema.GroupResource{
					Group:    gv.Group,
					Resource: resourceList.APIResources[i].Name,
				}
				if evaluator := registry.Get(gr); evaluator != nil {
					interestingResources = append(interestingResources, resourceList)
				}
			}
		}
		return interestingResources, nil
	}
}

func RunClusterQuotaReconciliationController(ctx context.Context, controllerCtx *EnhancedControllerContext) (bool, error) {
	defaultResyncPeriod := 5 * time.Minute
	defaultReplenishmentSyncPeriod := 12 * time.Hour

	saName := infraClusterQuotaReconciliationControllerServiceAccountName

	clusterQuotaMappingController := clusterquotamapping.NewClusterQuotaMappingController(
		controllerCtx.KubernetesInformers.Core().V1().Namespaces(),
		controllerCtx.QuotaInformers.Quota().V1().ClusterResourceQuotas())
	resourceQuotaControllerClient := controllerCtx.ClientBuilder.ClientOrDie("resourcequota-controller")
	discoveryFunc := resourceQuotaControllerClient.Discovery().ServerPreferredNamespacedResources
	listerFuncForResource := generic.ListerFuncForResourceFunc(controllerCtx.GenericResourceInformer.ForResource)
	quotaConfiguration := quotainstall.NewQuotaConfigurationForControllers(listerFuncForResource)

	// TODO make a union registry
	resourceQuotaRegistry := generic.NewRegistry(quotaConfiguration.Evaluators())
	imageEvaluators := image.NewReplenishmentEvaluators(
		listerFuncForResource,
		controllerCtx.ImageInformers.Image().V1().ImageStreams(),
		controllerCtx.ClientBuilder.OpenshiftImageClientOrDie(saName).ImageV1())
	for i := range imageEvaluators {
		resourceQuotaRegistry.Add(imageEvaluators[i])
	}

	options := clusterquotareconciliation.ClusterQuotaReconcilationControllerOptions{
		ClusterQuotaInformer: controllerCtx.QuotaInformers.Quota().V1().ClusterResourceQuotas(),
		ClusterQuotaMapper:   clusterQuotaMappingController.GetClusterQuotaMapper(),
		ClusterQuotaClient:   controllerCtx.ClientBuilder.OpenshiftQuotaClientOrDie(saName).QuotaV1().ClusterResourceQuotas(),

		Registry:                  resourceQuotaRegistry,
		ResyncPeriod:              defaultResyncPeriod,
		ReplenishmentResyncPeriod: controller.StaticResyncPeriodFunc(defaultReplenishmentSyncPeriod),
		DiscoveryFunc:             discoveryFunc,
		IgnoredResourcesFunc:      quotaConfiguration.IgnoredResources,
		InformersStarted:          controllerCtx.InformersStarted,
		InformerFactory:           controllerCtx.GenericResourceInformer,
		UpdateFilter:              quotainstall.DefaultUpdateFilter(),
	}
	clusterQuotaReconciliationController, err := clusterquotareconciliation.NewClusterQuotaReconcilationController(ctx, options)
	if err != nil {
		return true, err
	}
	clusterQuotaMappingController.GetClusterQuotaMapper().AddListener(clusterQuotaReconciliationController)

	go clusterQuotaMappingController.Run(5, ctx.Done())
	go clusterQuotaReconciliationController.Run(5, ctx)
	go clusterQuotaReconciliationController.Sync(discoveryFunc, 30*time.Second, ctx)

	return true, nil
}

func calculateResyncPeriod(period time.Duration) func() time.Duration {
	return func() time.Duration {
		factor := rand.Float64() + 1
		return time.Duration(float64(period.Nanoseconds()) * factor)
	}
}
