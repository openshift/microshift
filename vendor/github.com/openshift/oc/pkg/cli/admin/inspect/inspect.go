package inspect

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	kcmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util"
	"k8s.io/kubectl/pkg/util/templates"

	configv1 "github.com/openshift/api/config/v1"
)

var (
	inspectLong = templates.LongDesc(`
		Gather debugging information for a resource.

		This command downloads the specified resource and any related
		resources for the purpose of gathering debugging information.

		Experimental: This command is under active development and may change without notice.
	`)

	inspectExample = templates.Examples(`
		# Collect debugging data for the "openshift-apiserver" clusteroperator
		oc adm inspect clusteroperator/openshift-apiserver

		# Collect debugging data for the "openshift-apiserver" and "kube-apiserver" clusteroperators
		oc adm inspect clusteroperator/openshift-apiserver clusteroperator/kube-apiserver

		# Collect debugging data for all clusteroperators
		oc adm inspect clusteroperator

		# Collect debugging data for all clusteroperators and clusterversions
		oc adm inspect clusteroperators,clusterversions
	`)
)

type InspectOptions struct {
	printFlags  *genericclioptions.PrintFlags
	configFlags *genericclioptions.ConfigFlags

	RESTConfig      *rest.Config
	kubeClient      kubernetes.Interface
	discoveryClient discovery.CachedDiscoveryInterface
	dynamicClient   dynamic.Interface

	fileWriter     *MultiSourceFileWriter
	builder        *resource.Builder
	since          time.Duration
	args           []string
	namespace      string
	sinceTime      string
	allNamespaces  bool
	rotatedPodLogs bool
	sinceInt       int64
	sinceTimestamp metav1.Time

	// directory where all gathered data will be stored
	DestDir string
	// whether or not to allow writes to an existing and populated base directory
	overwrite bool

	genericclioptions.IOStreams
	eventFile string
}

func NewInspectOptions(streams genericclioptions.IOStreams) *InspectOptions {
	printFlags := genericclioptions.NewPrintFlags("gathered").WithDefaultOutput("yaml").WithTypeSetter(scheme.Scheme)
	if printFlags.JSONYamlPrintFlags != nil {
		printFlags.JSONYamlPrintFlags.ShowManagedFields = true
	}
	return &InspectOptions{
		printFlags:  printFlags,
		configFlags: genericclioptions.NewConfigFlags(true),
		overwrite:   true,
		IOStreams:   streams,
	}
}

func NewCmdInspect(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewInspectOptions(streams)
	cmd := &cobra.Command{
		Use:     "inspect (TYPE[.VERSION][.GROUP] [NAME] | TYPE[.VERSION][.GROUP]/NAME ...) [flags]",
		Short:   "Collect debugging data for a given resource",
		Long:    inspectLong,
		Example: inspectExample,
		Run: func(c *cobra.Command, args []string) {
			kcmdutil.CheckErr(o.Complete(args))
			kcmdutil.CheckErr(o.Validate())
			kcmdutil.CheckErr(o.Run())
		},
	}

	cmd.Flags().StringVar(&o.DestDir, "dest-dir", o.DestDir, "Root directory used for storing all gathered cluster operator data. Defaults to $(PWD)/inspect.local.<rand>")
	cmd.Flags().StringVar(&o.eventFile, "events-file", o.eventFile, "A path to an events.json file to create a HTML page from")
	cmd.Flags().BoolVarP(&o.allNamespaces, "all-namespaces", "A", o.allNamespaces, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	cmd.Flags().StringVar(&o.sinceTime, "since-time", o.sinceTime, "Only return logs after a specific date (RFC3339). Defaults to all logs. Only one of since-time / since may be used.")
	cmd.Flags().DurationVar(&o.since, "since", o.since, "Only return logs newer than a relative duration like 5s, 2m, or 3h. Defaults to all logs. Only one of since-time / since may be used.")
	cmd.Flags().BoolVar(&o.rotatedPodLogs, "rotated-pod-logs", o.rotatedPodLogs, "Experimental: If present, retrieve rotated log files that are available for selected pods. This can significantly increase the collected logs size. since/since-time is ignored for rotated logs.")

	// The rotated-pod-logs option should be removed once support for retrieving rotated logs is added to kubelet
	// https://github.com/kubernetes/kubernetes/issues/59902
	cmd.Flags().MarkHidden("rotated-pod-logs")

	o.configFlags.AddFlags(cmd.Flags())
	return cmd
}

func (o *InspectOptions) Complete(args []string) error {
	o.args = args

	if len(o.eventFile) > 0 {
		return nil
	}

	var err error
	o.RESTConfig, err = o.configFlags.ToRESTConfig()
	if err != nil {
		return err
	}
	// we make lots and lots of client calls, don't slow down artificially.
	o.RESTConfig.QPS = 999999
	o.RESTConfig.Burst = 999999

	o.kubeClient, err = kubernetes.NewForConfig(o.RESTConfig)
	if err != nil {
		return err
	}

	o.dynamicClient, err = dynamic.NewForConfig(o.RESTConfig)
	if err != nil {
		return err
	}

	o.discoveryClient, err = o.configFlags.ToDiscoveryClient()
	if err != nil {
		return err
	}

	o.namespace, _, err = o.configFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	if o.since != 0 {
		o.sinceInt = (int64(o.since.Round(time.Second).Seconds()))
	}
	if len(o.sinceTime) > 0 {
		o.sinceTimestamp, err = util.ParseRFC3339(o.sinceTime, metav1.Now)
		if err != nil {
			return err
		}
	}

	printer, err := o.printFlags.ToPrinter()
	if err != nil {
		return err
	}
	o.fileWriter = NewMultiSourceWriter(printer)

	o.builder = resource.NewBuilder(o.configFlags)

	if len(o.DestDir) == 0 {
		o.DestDir = fmt.Sprintf("inspect.local.%06d", rand.Int63())
	}
	return nil
}

func (o *InspectOptions) Validate() error {
	if len(o.DestDir) == 0 {
		return fmt.Errorf("--dest-dir must not be empty")
	}
	if len(o.sinceTime) > 0 && o.since != 0 {
		return fmt.Errorf("at most one of `sinceTime` or `since` may be specified")
	}
	return nil
}

func (o *InspectOptions) Run() error {
	if len(o.eventFile) > 0 {
		return createEventFilterPageFromFile(o.eventFile, o.DestDir)
	}
	r := o.builder.
		Unstructured().
		NamespaceParam(o.namespace).DefaultNamespace().AllNamespaces(o.allNamespaces).
		ResourceTypeOrNameArgs(true, o.args...).
		Flatten().
		ContinueOnError().
		Latest().Do()

	allErrs := []error{}
	infos, err := r.Infos()
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		allErrs = append(allErrs, err)
	}

	// ensure we're able to proceed writing data to specified destination
	if err := o.ensureDirectoryViable(); err != nil {
		return err
	}

	// ensure destination path exists
	if err := os.MkdirAll(o.DestDir, os.ModePerm); err != nil {
		return err
	}

	if err := o.logTimestamp(); err != nil {
		return err
	}
	defer o.logTimestamp()

	// Collect all resources served by the server
	discoveryClient, err := o.configFlags.ToDiscoveryClient()
	if err != nil {
		return err
	}

	// Check if the resource is served by the server
	_, rList, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return fmt.Errorf("unable to retrieve served resources: %v", err)
	}
	serverResources := sets.NewString()
	for _, rItem := range rList {
		for _, item := range rItem.APIResources {
			// The inspection checks whether a resource of a given name exist
			// independent of the version/group.
			serverResources.Insert(item.Name)
		}
	}

	// finally, gather polymorphic resources specified by the user
	ctx := NewResourceContext(serverResources)
	for _, info := range infos {
		err := InspectResource(info, ctx, o)
		if err != nil {
			allErrs = append(allErrs, err)
		}
	}

	// now gather all the events into a single file and produce a unified file
	if err := CreateEventFilterPage(o.DestDir); err != nil {
		allErrs = append(allErrs, err)
	}

	fmt.Fprintf(o.Out, "Wrote inspect data to %s.\n", o.DestDir)
	if len(allErrs) > 0 {
		return fmt.Errorf("errors occurred while gathering data:\n    %v", errors.NewAggregate(allErrs))
	}

	return nil
}

// gatherConfigResourceData gathers all config.openshift.io resources
func (o *InspectOptions) gatherConfigResourceData(destDir string, ctx *resourceContext) error {
	// determine if we've already collected configResourceData
	if ctx.visited.Has(configResourceDataKey) {
		klog.V(1).Infof("Skipping previously-collected config.openshift.io resource data")
		return nil
	}
	ctx.visited.Insert(configResourceDataKey)

	klog.V(1).Infof("Gathering config.openshift.io resource data...\n")

	// ensure destination path exists
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return err
	}

	resources, err := retrieveAPIGroupVersionResourceNames(o.discoveryClient, configv1.GroupName)
	if err != nil {
		return err
	}

	errs := []error{}
	for _, resource := range resources {
		resourceList, err := o.dynamicClient.Resource(resource).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			errs = append(errs, err)
			continue
		}

		objToPrint := runtime.Object(resourceList)
		filename := fmt.Sprintf("%s.yaml", resource.Resource)
		if err := o.fileWriter.WriteFromResource(path.Join(destDir, "/"+filename), objToPrint); err != nil {
			errs = append(errs, err)
			continue
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("one or more errors occurred while gathering config.openshift.io resource data:\n\n    %v", errors.NewAggregate(errs))
	}
	return nil
}

// gatherOperatorResourceData gathers all kubeapiserver.operator.openshift.io resources
func (o *InspectOptions) gatherOperatorResourceData(destDir string, ctx *resourceContext) error {
	// determine if we've already collected operatorResourceData
	if ctx.visited.Has(operatorResourceDataKey) {
		klog.V(1).Infof("Skipping previously-collected operator.openshift.io resource data")
		return nil
	}
	ctx.visited.Insert(operatorResourceDataKey)

	// ensure destination path exists
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return err
	}

	resources, err := retrieveAPIGroupVersionResourceNames(o.discoveryClient, "kubeapiserver.operator.openshift.io")
	if err != nil {
		return err
	}

	errs := []error{}
	for _, resource := range resources {
		resourceList, err := o.dynamicClient.Resource(resource).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			errs = append(errs, err)
			continue
		}

		objToPrint := runtime.Object(resourceList)
		filename := fmt.Sprintf("%s.yaml", resource.Resource)
		if err := o.fileWriter.WriteFromResource(path.Join(destDir, "/"+filename), objToPrint); err != nil {
			errs = append(errs, err)
			continue
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("one or more errors occurred while gathering operator.openshift.io resource data:\n\n    %v", errors.NewAggregate(errs))
	}
	return nil
}

// ensureDirectoryViable returns an error if DestDir:
// 1. already exists AND is a file (not a directory)
// 2. already exists AND is NOT empty, unless overwrite was passed
// 3. an IO error occurs
func (o *InspectOptions) ensureDirectoryViable() error {
	baseDirInfo, err := os.Stat(o.DestDir)
	if err != nil && os.IsNotExist(err) {
		// no error, directory simply does not exist yet
		return nil
	}
	if err != nil {
		return err
	}

	if !baseDirInfo.IsDir() {
		return fmt.Errorf("%q exists and is a file", o.DestDir)
	}
	files, err := ioutil.ReadDir(o.DestDir)
	if err != nil {
		return err
	}
	if len(files) > 0 && !o.overwrite {
		return fmt.Errorf("%q exists and is not empty. Pass --overwrite to allow data overwrites", o.DestDir)
	}
	return nil
}

func (o *InspectOptions) logTimestamp() error {
	f, err := os.OpenFile(path.Join(o.DestDir, "timestamp"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	_, err = f.WriteString(fmt.Sprintf("%v\n", time.Now()))
	return err
}

// supportedResourceFinder provides a way to discover supported resources by the server.
// it exists to allow for easier testability.
type supportedResourceFinder interface {
	ServerPreferredResources() ([]*metav1.APIResourceList, error)
}

func retrieveAPIGroupVersionResourceNames(discoveryClient supportedResourceFinder, apiGroup string) ([]schema.GroupVersionResource, error) {
	lists, discoveryErr := discoveryClient.ServerPreferredResources()

	foundResources := sets.String{}
	resources := []schema.GroupVersionResource{}
	for _, list := range lists {
		if len(list.APIResources) == 0 {
			continue
		}
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			/// something went seriously wrong
			return nil, err
		}
		for _, resource := range list.APIResources {
			// filter groups outside of the provided apiGroup
			if !strings.HasSuffix(gv.Group, apiGroup) {
				continue
			}
			verbs := sets.NewString(([]string(resource.Verbs))...)
			if !verbs.Has("list") {
				continue
			}
			// if we've already seen this resource in another version, don't add it again
			if foundResources.Has(resource.Name) {
				continue
			}

			foundResources.Insert(resource.Name)
			resources = append(resources, schema.GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: resource.Name})
		}
	}
	// we only care about discovery errors if we don't find what we want
	if len(resources) == 0 {
		return nil, discoveryErr
	}

	return resources, nil
}
