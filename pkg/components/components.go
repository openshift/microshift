package components

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/sirupsen/logrus"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	batchclientv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openshift/microshift/pkg/config"
)

func startDefaultComponents(cfg *config.MicroshiftConfig) error {
	if err := startFlannel(cfg.DataDir + "/resources/kubeadmin/kubeconfig"); err != nil {
		logrus.Warningf("failed to start Flannel: %v", err)
		return err
	}
	return nil
}

func StartComponents(cfg *config.MicroshiftConfig) error {
	if err := startDefaultComponents(cfg); err != nil {
		return err
	}

	if len(cfg.Components) == 0 {
		return nil
	}

	componentLoadNamespace := "component-loader-ns"
	restConfig, err := clientcmd.BuildConfigFromFlags("", cfg.DataDir+"/resources/kubeadmin/kubeconfig")
	if err != nil {
		return err
	}
	// create the namespace for component loader
	coreclient := coreclientv1.NewForConfigOrDie(rest.AddUserAgent(restConfig, "core-agent"))
	ns := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: componentLoadNamespace,
		},
	}
	_, err = coreclient.Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}
	// embed kubeadm in a secret volume
	kubeconfigBuf, err := ioutil.ReadFile(cfg.DataDir + "/resources/kubeadmin/kubeconfig")
	if err != nil {
		return err
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "microshift-component-loader-kubeconfig",
			Namespace: componentLoadNamespace,
		},
		Data: map[string][]byte{
			"kubeconfig": kubeconfigBuf,
		},
	}
	_, err = coreclient.Secrets(componentLoadNamespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	// create a KUBECONFIG env var
	kubeconfigVar := corev1.EnvVar{Name: "KUBECONFIG", Value: "/var/lib/microshift/kubeconfig"}

	batchclient := batchclientv1.NewForConfigOrDie(rest.AddUserAgent(restConfig, "batch-agent"))

	for _, component := range cfg.Components {
		image := component.Image
		args := []string{}
		var mode int32 = 0666
		volumes := []corev1.Volume{
			{
				Name: "kubeconfig",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: "microshift-component-loader-kubeconfig",
						Items: []corev1.KeyToPath{
							{
								Key:  "kubeconfig",
								Path: "kubeconfig",
							},
						},
						DefaultMode: &mode,
					},
				},
			},
		}

		volMounts := []corev1.VolumeMount{
			{
				Name:      "kubeconfig",
				MountPath: "/var/lib/microshift",
			},
		}

		logrus.Infof("creating loader job %s", component.Name)
		if len(component.Parameters) > 0 {
			// if the component loader has parameters,
			// create a ConfigMap and volume-mount it
			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "microshift-component-loader-" + component.Name,
					Namespace: componentLoadNamespace,
				},
				Data: component.Parameters,
			}
			logrus.Infof("applying cm for loader job %s", component.Name)

			_, err = coreclient.ConfigMaps(componentLoadNamespace).Create(context.TODO(), cm, metav1.CreateOptions{})
			if err != nil && !apierrors.IsAlreadyExists(err) {
				return err
			}
			mountPath := "/var/lib/loader.config"
			args = []string{"-c", mountPath}
			vol := corev1.VolumeMount{Name: "config", MountPath: mountPath}
			volMounts = append(volMounts, vol)
			volume := corev1.Volume{
				Name: "config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "microshift-component-loader-" + component.Name,
						},
					},
				},
			}
			volumes = append(volumes, volume)
		}
		container := corev1.Container{
			Name:  "microshift-component-loader",
			Image: image,
			// use container's entrypoint for `Command`
			Args: args,
			Env: []corev1.EnvVar{
				kubeconfigVar,
			},
			VolumeMounts: volMounts,
		}
		// create a job to start component loader
		job := &batchv1.Job{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Job",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "microshift-component-loader-" + component.Name,
				Namespace: componentLoadNamespace,
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							container,
						},
						// use hostnetwork so the kubeconfig points to the same api endpoint as created earlier
						HostNetwork: true,
						Volumes:     volumes,
						// don't restart the job upon failure
						// watcher will count the failures and exit after 3 retries
						RestartPolicy: corev1.RestartPolicyNever,
					},
				},
			},
		}
		logrus.Infof("applying loader job %s", component.Name)
		_, err = batchclient.Jobs(componentLoadNamespace).Create(context.TODO(), job, metav1.CreateOptions{})
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
		// watch job status
		watcher, err := batchclient.Jobs(componentLoadNamespace).Watch(
			context.TODO(),
			metav1.SingleObject(metav1.ObjectMeta{
				Name: "microshift-component-loader-" + component.Name, Namespace: componentLoadNamespace}))
		if err != nil {
			return err
		}
		ch := watcher.ResultChan()
	loop:
		for {
			event, active := <-ch
			if active {
				switch event.Type {
				case watch.Modified:
					newJob, ok := event.Object.(*batchv1.Job)
					if !ok {
						logrus.Warning("failed to get loader job after watch")
						continue
					}
					logrus.Infof("component loader job status %v", newJob.Status)
					if newJob.Status.Succeeded >= 1 {
						logrus.Infof("component loader job succeeded")
						break loop
					}
					if newJob.Status.Failed > 3 {
						return fmt.Errorf("component loader job failed")
					}

				case watch.Deleted:
					logrus.Infof("component loader job deleted")
					break loop
				default:
					// skip
				}
			} else {
				logrus.Infof("component loader job watcher closed")
				break
			}
		}
	}
	return nil
}
