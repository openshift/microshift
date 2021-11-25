package controller

import (
	"context"
	"fmt"
	"time"

	kappsv1 "k8s.io/api/apps/v1"
	kappsv1beta1 "k8s.io/api/apps/v1beta1"
	kappsv1beta2 "k8s.io/api/apps/v1beta2"
	kbatchv1 "k8s.io/api/batch/v1"
	kbatchv1beta1 "k8s.io/api/batch/v1beta1"
	kapiv1 "k8s.io/api/core/v1"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kclientsetexternal "k8s.io/client-go/kubernetes"

	triggerutil "github.com/openshift/library-go/pkg/image/trigger"
	imagecontroller "github.com/openshift/openshift-controller-manager/pkg/image/controller"
	imagesignaturecontroller "github.com/openshift/openshift-controller-manager/pkg/image/controller/signature"
	imagetriggercontroller "github.com/openshift/openshift-controller-manager/pkg/image/controller/trigger"
	triggerannotations "github.com/openshift/openshift-controller-manager/pkg/image/trigger/annotations"
	triggerbuildconfigs "github.com/openshift/openshift-controller-manager/pkg/image/trigger/buildconfigs"
	triggerdeploymentconfigs "github.com/openshift/openshift-controller-manager/pkg/image/trigger/deploymentconfigs"
)

func RunImageTriggerController(ctx *ControllerContext) (bool, error) {
	informer := ctx.ImageInformers.Image().V1().ImageStreams()

	buildClient, err := ctx.ClientBuilder.OpenshiftBuildClient(infraImageTriggerControllerServiceAccountName)
	if err != nil {
		return true, err
	}

	appsClient, err := ctx.ClientBuilder.OpenshiftAppsClient(infraImageTriggerControllerServiceAccountName)
	if err != nil {
		return true, err
	}
	kclient := ctx.ClientBuilder.ClientOrDie(infraImageTriggerControllerServiceAccountName)

	updater := podSpecUpdater{kclient}
	broadcaster := imagetriggercontroller.NewTriggerEventBroadcaster(kclient.CoreV1())

	sources := []imagetriggercontroller.TriggerSource{
		{
			Resource:  schema.GroupResource{Group: "apps.openshift.io", Resource: "deploymentconfigs"},
			Informer:  ctx.AppsInformers.Apps().V1().DeploymentConfigs().Informer(),
			Store:     ctx.AppsInformers.Apps().V1().DeploymentConfigs().Informer().GetIndexer(),
			TriggerFn: triggerdeploymentconfigs.NewDeploymentConfigTriggerIndexer,
			Reactor:   &triggerdeploymentconfigs.DeploymentConfigReactor{Client: appsClient.AppsV1()},
		},
	}
	sources = append(sources, imagetriggercontroller.TriggerSource{
		Resource:  schema.GroupResource{Group: "build.openshift.io", Resource: "buildconfigs"},
		Informer:  ctx.BuildInformers.Build().V1().BuildConfigs().Informer(),
		Store:     ctx.BuildInformers.Build().V1().BuildConfigs().Informer().GetIndexer(),
		TriggerFn: triggerbuildconfigs.NewBuildConfigTriggerIndexer,
		Reactor:   triggerbuildconfigs.NewBuildConfigReactor(buildClient.BuildV1(), kclient.CoreV1().RESTClient()),
	})
	sources = append(sources, imagetriggercontroller.TriggerSource{
		Resource:  schema.GroupResource{Group: "apps", Resource: "deployments"},
		Informer:  ctx.KubernetesInformers.Apps().V1().Deployments().Informer(),
		Store:     ctx.KubernetesInformers.Apps().V1().Deployments().Informer().GetIndexer(),
		TriggerFn: triggerannotations.NewAnnotationTriggerIndexer,
		Reactor:   &triggerutil.AnnotationReactor{Updater: updater},
	})
	sources = append(sources, imagetriggercontroller.TriggerSource{
		Resource:  schema.GroupResource{Group: "apps", Resource: "daemonsets"},
		Informer:  ctx.KubernetesInformers.Apps().V1().DaemonSets().Informer(),
		Store:     ctx.KubernetesInformers.Apps().V1().DaemonSets().Informer().GetIndexer(),
		TriggerFn: triggerannotations.NewAnnotationTriggerIndexer,
		Reactor:   &triggerutil.AnnotationReactor{Updater: updater},
	})
	sources = append(sources, imagetriggercontroller.TriggerSource{
		Resource:  schema.GroupResource{Group: "apps", Resource: "statefulsets"},
		Informer:  ctx.KubernetesInformers.Apps().V1().StatefulSets().Informer(),
		Store:     ctx.KubernetesInformers.Apps().V1().StatefulSets().Informer().GetIndexer(),
		TriggerFn: triggerannotations.NewAnnotationTriggerIndexer,
		Reactor:   &triggerutil.AnnotationReactor{Updater: updater},
	})
	sources = append(sources, imagetriggercontroller.TriggerSource{
		Resource:  schema.GroupResource{Group: "batch", Resource: "cronjobs"},
		Informer:  ctx.KubernetesInformers.Batch().V1().CronJobs().Informer(),
		Store:     ctx.KubernetesInformers.Batch().V1().CronJobs().Informer().GetIndexer(),
		TriggerFn: triggerannotations.NewAnnotationTriggerIndexer,
		Reactor:   &triggerutil.AnnotationReactor{Updater: updater},
	})

	go imagetriggercontroller.NewTriggerController(
		ctx.OpenshiftControllerConfig.DockerPullSecret.InternalRegistryHostname,
		broadcaster,
		informer,
		sources...,
	).Run(5, ctx.Stop)

	return true, nil
}

type podSpecUpdater struct {
	kclient kclientsetexternal.Interface
}

func (u podSpecUpdater) Update(obj runtime.Object) error {
	switch t := obj.(type) {
	case *kappsv1.DaemonSet:
		_, err := u.kclient.AppsV1().DaemonSets(t.Namespace).Update(context.TODO(), t, kmetav1.UpdateOptions{})
		return err
	case *kappsv1.Deployment:
		_, err := u.kclient.AppsV1().Deployments(t.Namespace).Update(context.TODO(), t, kmetav1.UpdateOptions{})
		return err
	case *kappsv1beta1.Deployment:
		_, err := u.kclient.AppsV1beta1().Deployments(t.Namespace).Update(context.TODO(), t, kmetav1.UpdateOptions{})
		return err
	case *kappsv1beta2.Deployment:
		_, err := u.kclient.AppsV1beta2().Deployments(t.Namespace).Update(context.TODO(), t, kmetav1.UpdateOptions{})
		return err
	case *kappsv1.StatefulSet:
		_, err := u.kclient.AppsV1().StatefulSets(t.Namespace).Update(context.TODO(), t, kmetav1.UpdateOptions{})
		return err
	case *kappsv1beta1.StatefulSet:
		_, err := u.kclient.AppsV1beta1().StatefulSets(t.Namespace).Update(context.TODO(), t, kmetav1.UpdateOptions{})
		return err
	case *kappsv1beta2.StatefulSet:
		_, err := u.kclient.AppsV1beta2().StatefulSets(t.Namespace).Update(context.TODO(), t, kmetav1.UpdateOptions{})
		return err
	case *kbatchv1.Job:
		_, err := u.kclient.BatchV1().Jobs(t.Namespace).Update(context.TODO(), t, kmetav1.UpdateOptions{})
		return err
	case *kbatchv1.CronJob:
		_, err := u.kclient.BatchV1().CronJobs(t.Namespace).Update(context.TODO(), t, kmetav1.UpdateOptions{})
		return err
	case *kbatchv1beta1.CronJob:
		_, err := u.kclient.BatchV1beta1().CronJobs(t.Namespace).Update(context.TODO(), t, kmetav1.UpdateOptions{})
		return err
	case *kapiv1.Pod:
		_, err := u.kclient.CoreV1().Pods(t.Namespace).Update(context.TODO(), t, kmetav1.UpdateOptions{})
		return err
	default:
		return fmt.Errorf("unrecognized object - no trigger update possible for %T", obj)
	}
}

func RunImageSignatureImportController(ctx *ControllerContext) (bool, error) {
	// TODO these should really be configurable
	resyncPeriod := 1 * time.Hour
	signatureFetchTimeout := 1 * time.Minute
	signatureImportLimit := 10

	controller := imagesignaturecontroller.NewSignatureImportController(
		context.Background(),
		ctx.ClientBuilder.OpenshiftImageClientOrDie(infraImageImportControllerServiceAccountName),
		ctx.ImageInformers.Image().V1().Images(),
		resyncPeriod,
		signatureFetchTimeout,
		signatureImportLimit,
	)
	go controller.Run(5, ctx.Stop)
	return true, nil
}

func RunImageImportController(ctx *ControllerContext) (bool, error) {
	informer := ctx.ImageInformers.Image().V1().ImageStreams()
	controller := imagecontroller.NewImageStreamController(
		ctx.ClientBuilder.OpenshiftImageClientOrDie(infraImageImportControllerServiceAccountName),
		informer,
	)
	go controller.Run(50, ctx.Stop)

	// TODO control this using enabled and disabled controllers
	if ctx.OpenshiftControllerConfig.ImageImport.DisableScheduledImport {
		return true, nil
	}

	scheduledController := imagecontroller.NewScheduledImageStreamController(
		ctx.ClientBuilder.OpenshiftImageClientOrDie(infraImageImportControllerServiceAccountName),
		informer,
		imagecontroller.ScheduledImageStreamControllerOptions{
			Resync: time.Duration(ctx.OpenshiftControllerConfig.ImageImport.ScheduledImageImportMinimumIntervalSeconds) * time.Second,

			Enabled:                  !ctx.OpenshiftControllerConfig.ImageImport.DisableScheduledImport,
			DefaultBucketSize:        4,
			MaxImageImportsPerMinute: ctx.OpenshiftControllerConfig.ImageImport.MaxScheduledImageImportsPerMinute,
		},
	)

	controller.SetNotifier(scheduledController)
	go scheduledController.Run(ctx.Stop)

	return true, nil
}
