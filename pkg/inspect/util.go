package inspect

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
)

// resourceContext is used to keep track of previously seen objects
type resourceContext struct {
	visited sets.String
}

func NewResourceContext() *resourceContext {
	return &resourceContext{
		visited: sets.NewString(),
	}
}

func objectReferenceToString(ref *configv1.ObjectReference) string {
	resource := ref.Resource
	group := ref.Group
	name := ref.Name
	if len(name) > 0 {
		name = "/" + name
	}
	if len(group) > 0 {
		group = "." + group
	}
	return resource + group + name
}

func unstructuredToString(obj *unstructured.Unstructured) string {
	resource := obj.GetKind()
	var group string
	if gv, err := schema.ParseGroupVersion(obj.GetAPIVersion()); err != nil {
		group = gv.Group
	}
	name := obj.GetName()
	if len(name) > 0 {
		name = "/" + name
	}
	if len(group) > 0 {
		group = "." + group
	}
	return resource + group + name

}

func objectReferenceToResourceInfos(clientGetter genericclioptions.RESTClientGetter, ref *configv1.ObjectReference) ([]*resource.Info, error) {
	b := resource.NewBuilder(clientGetter).
		Unstructured().
		ResourceTypeOrNameArgs(true, objectReferenceToString(ref)).
		NamespaceParam(ref.Namespace).DefaultNamespace().AllNamespaces(len(ref.Namespace) == 0).
		Flatten().
		Latest()

	infos, err := b.Do().Infos()
	if err != nil {
		return nil, err
	}

	return infos, nil
}

func groupResourceToInfos(clientGetter genericclioptions.RESTClientGetter, ref schema.GroupResource, namespace string) ([]*resource.Info, error) {
	resourceString := ref.Resource
	if len(ref.Group) > 0 {
		resourceString = fmt.Sprintf("%s.%s", resourceString, ref.Group)
	}
	b := resource.NewBuilder(clientGetter).
		Unstructured().
		ResourceTypeOrNameArgs(false, resourceString).
		SelectAllParam(true).
		NamespaceParam(namespace).
		Latest()

	return b.Do().Infos()
}

// infoToContextKey receives a resource.Info and returns a unique string for use in keeping track of objects previously seen
func infoToContextKey(info *resource.Info) string {
	name := info.Name
	if meta.IsListType(info.Object) {
		name = "*"
	}
	return fmt.Sprintf("%s/%s/%s/%s", info.Namespace, info.ResourceMapping().GroupVersionKind.Group, info.ResourceMapping().Resource.Resource, name)
}

// objectRefToContextKey is a variant of infoToContextKey that receives a configv1.ObjectReference and returns a unique string for use in keeping track of object references previously seen
func objectRefToContextKey(objRef *configv1.ObjectReference) string {
	return fmt.Sprintf("%s/%s/%s/%s", objRef.Namespace, objRef.Group, objRef.Resource, objRef.Name)
}

func resourceToContextKey(resource schema.GroupResource, namespace string) string {
	return fmt.Sprintf("%s/%s/%s/%s", namespace, resource.Group, resource.Resource, "*")
}

// dirPathForInfo receives a *resource.Info and returns a relative path
// corresponding to the directory location of that object on disk
func dirPathForInfo(baseDir string, info *resource.Info) string {
	groupName := "core"
	if len(info.Mapping.GroupVersionKind.Group) > 0 {
		groupName = info.Mapping.GroupVersionKind.Group
	}

	groupPath := path.Join(baseDir, namespaceResourcesDirname, info.Namespace, groupName)
	if len(info.Namespace) == 0 {
		groupPath = path.Join(baseDir, clusterScopedResourcesDirname, "/"+groupName)
	}
	if meta.IsListType(info.Object) {
		return groupPath
	}

	objPath := path.Join(groupPath, info.ResourceMapping().Resource.Resource)
	if len(info.Namespace) == 0 {
		objPath = path.Join(groupPath, info.ResourceMapping().Resource.Resource)
	}
	return objPath
}

// filenameForInfo receives a *resource.Info and returns the basename
func filenameForInfo(info *resource.Info) string {
	if meta.IsListType(info.Object) {
		return info.ResourceMapping().Resource.Resource + ".yaml"
	}

	return info.Name + ".yaml"
}

// getAllEventsRecursive returns a union (not deconflicted) or all events under a directory
func getAllEventsRecursive(rootDir string) (*corev1.EventList, error) {
	// now gather all the events into a single file and produce a unified file
	eventLists := &corev1.EventList{}
	err := filepath.Walk(rootDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Name() != "events.yaml" {
				return nil
			}
			eventBytes, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			events, err := readEvents(eventBytes)
			if err != nil {
				return err
			}
			eventLists.Items = append(eventLists.Items, events.Items...)
			return nil
		})
	if err != nil {
		return nil, err
	}

	return eventLists, nil
}

func createEventFilterPageFromFile(eventFile string, rootDir string) error {
	var jsonStream io.Reader
	var err error

	if strings.HasPrefix(eventFile, "https://") || strings.HasPrefix(eventFile, "http://") {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		resp, err := client.Get(eventFile)
		if err != nil {
			return err
		}
		jsonStream = resp.Body
		defer resp.Body.Close()
	} else {
		jsonStream, err = os.Open(eventFile)
		if err != nil {
			return err
		}
	}

	decoder := json.NewDecoder(jsonStream)
	var events corev1.EventList
	if err := decoder.Decode(&events); err != nil {
		return err
	}
	return createEventFilterPage(&events, rootDir)
}

func createEventFilterPage(events *corev1.EventList, rootDir string) error {
	sort.Slice(events.Items, func(i, j int) bool {
		return events.Items[i].LastTimestamp.Time.Before(events.Items[j].LastTimestamp.Time)
	})

	t := template.Must(template.New("events").Funcs(template.FuncMap{
		"formatTime": func(created, firstSeen, lastSeen metav1.Time, count int32) template.HTML {
			countMsg := ""
			if count > 1 {
				countMsg = fmt.Sprintf(" <small>(x%d)</small>", count)
			}
			if lastSeen.IsZero() {
				lastSeen = created
			}
			return template.HTML(fmt.Sprintf(`<time datetime="%s" title="First Seen: %s">%s</time>%s`, lastSeen.String(), firstSeen.Format("15:04:05"), lastSeen.Format("15:04:05"), countMsg))
		},
		"formatReason": func(r string) template.HTML {
			switch {
			case strings.Contains(strings.ToLower(r), "fail"),
				strings.Contains(strings.ToLower(r), "error"),
				strings.Contains(strings.ToLower(r), "kill"),
				strings.Contains(strings.ToLower(r), "backoff"):
				return template.HTML(`<p class="text-danger">` + r + `</p>`)
			case strings.Contains(strings.ToLower(r), "notready"),
				strings.Contains(strings.ToLower(r), "unhealthy"),
				strings.Contains(strings.ToLower(r), "missing"):
				return template.HTML(`<p class="text-warning">` + r + `</p>`)
			}
			return template.HTML(`<p class="text-muted">` + r + `</p>`)
		},
	}).Parse(eventHTMLPage))

	out := bytes.NewBuffer([]byte{})
	if err := t.Execute(out, events); err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(rootDir, "event-filter.html"), out.Bytes(), 0644)
}

// CreateEventFilterPage reads all events in rootDir recursively, produces a single file, and produces a webpage
// that can be viewed locally to filter the events.
func CreateEventFilterPage(rootDir string) error {
	events, err := getAllEventsRecursive(rootDir)
	if err != nil {
		return err
	}
	return createEventFilterPage(events, rootDir)
}

var (
	coreScheme = runtime.NewScheme()
	coreCodecs = serializer.NewCodecFactory(coreScheme)
)

func init() {
	if err := corev1.AddToScheme(coreScheme); err != nil {
		panic(err)
	}
}

func readEvents(objBytes []byte) (*corev1.EventList, error) {
	requiredObj, err := runtime.Decode(coreCodecs.UniversalDecoder(corev1.SchemeGroupVersion), objBytes)
	if err != nil {
		return nil, err
	}
	return requiredObj.(*corev1.EventList), nil
}
