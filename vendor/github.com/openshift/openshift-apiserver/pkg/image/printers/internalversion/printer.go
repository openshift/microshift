package internalversion

import (
	"fmt"
	"strings"
	"time"

	units "github.com/docker/go-units"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubernetes/pkg/apis/core"
	kprinters "k8s.io/kubernetes/pkg/printers"

	imagev1 "github.com/openshift/api/image/v1"
	"github.com/openshift/library-go/pkg/image/imageutil"
	imageapi "github.com/openshift/openshift-apiserver/pkg/image/apis/image"
)

func AddImageOpenShiftHandlers(h kprinters.PrintHandler) {
	imageColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Image Reference", Type: "string", Description: imagev1.Image{}.SwaggerDoc()["dockerImageReference"]},
	}
	if err := h.TableHandler(imageColumnDefinitions, printImageList); err != nil {
		panic(err)
	}
	if err := h.TableHandler(imageColumnDefinitions, printImage); err != nil {
		panic(err)
	}

	imageStreamColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Image Repository", Type: "string", Description: imagev1.ImageStreamStatus{}.SwaggerDoc()["dockerImageRepository"]},
		{Name: "Tags", Type: "string", Description: "Human readable list of tags."},
		{Name: "Updated", Type: "string", Description: "Last update time."},
	}
	if err := h.TableHandler(imageStreamColumnDefinitions, printImageStreamList); err != nil {
		panic(err)
	}
	if err := h.TableHandler(imageStreamColumnDefinitions, printImageStream); err != nil {
		panic(err)
	}

	imageStreamTagColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Image Reference", Type: "string", Description: imagev1.Image{}.SwaggerDoc()["dockerImageReference"]},
		{Name: "Updated", Type: "string", Description: "Last update time."},
		{Name: "Image Name", Type: "string", Priority: 1, Description: imagev1.ImageStreamTag{}.SwaggerDoc()["image"]},
	}
	if err := h.TableHandler(imageStreamTagColumnDefinitions, printImageStreamTagList); err != nil {
		panic(err)
	}
	if err := h.TableHandler(imageStreamTagColumnDefinitions, printImageStreamTag); err != nil {
		panic(err)
	}

	imageTagColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Spec", Type: "string", Description: imagev1.ImageTag{}.SwaggerDoc()["spec"]},
		{Name: "Status", Type: "string", Description: imagev1.ImageTag{}.SwaggerDoc()["status"]},
		{Name: "History", Type: "number", Description: "Number of history entries"},
		{Name: "Updated", Type: "string", Description: "Last updated at"},
	}
	if err := h.TableHandler(imageTagColumnDefinitions, printImageTagList); err != nil {
		panic(err)
	}
	if err := h.TableHandler(imageTagColumnDefinitions, printImageTag); err != nil {
		panic(err)
	}

	imageStreamImageColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Updated", Type: "string", Description: "Last update time."},
		{Name: "Image Reference", Type: "string", Priority: 1, Description: imagev1.Image{}.SwaggerDoc()["dockerImageReference"]},
		{Name: "Image Name", Type: "string", Priority: 1, Description: imagev1.ImageStreamImage{}.SwaggerDoc()["image"]},
	}
	if err := h.TableHandler(imageStreamImageColumnDefinitions, printImageStreamImage); err != nil {
		panic(err)
	}
}

func printImage(image *imageapi.Image, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: image},
	}

	row.Cells = append(row.Cells, image.Name, image.DockerImageReference)

	return []metav1.TableRow{row}, nil
}

func printImageList(list *imageapi.ImageList, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(list.Items))
	for i := range list.Items {
		r, err := printImage(&list.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func printImageStream(stream *imageapi.ImageStream, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: stream},
	}

	var latest metav1.Time
	for _, list := range stream.Status.Tags {
		if len(list.Items) > 0 {
			if list.Items[0].Created.After(latest.Time) {
				latest = list.Items[0].Created
			}
		}
	}
	latestTime := ""
	if !latest.IsZero() {
		latestTime = fmt.Sprintf("%s ago", formatRelativeTime(latest.Time))
	}

	tags := printTagsUpToWidth(stream.Status.Tags, 40)

	repo := stream.Spec.DockerImageRepository
	if len(repo) == 0 {
		repo = stream.Status.DockerImageRepository
	}
	if len(stream.Status.PublicDockerImageRepository) > 0 {
		repo = stream.Status.PublicDockerImageRepository
	}

	row.Cells = append(row.Cells, stream.Name, repo, tags, latestTime)

	return []metav1.TableRow{row}, nil
}

func printImageStreamList(list *imageapi.ImageStreamList, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(list.Items))
	for i := range list.Items {
		r, err := printImageStream(&list.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func printImageStreamTag(ist *imageapi.ImageStreamTag, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: ist},
	}

	created := fmt.Sprintf("%s ago", formatRelativeTime(ist.CreationTimestamp.Time))

	row.Cells = append(row.Cells, ist.Name, ist.Image.DockerImageReference, created)

	if options.Wide {
		row.Cells = append(row.Cells, ist.Image.Name)
	}

	return []metav1.TableRow{row}, nil
}

func printImageStreamTagList(list *imageapi.ImageStreamTagList, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(list.Items))
	for i := range list.Items {
		r, err := printImageStreamTag(&list.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func printImageTag(ist *imageapi.ImageTag, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: ist},
	}

	var generation int64
	var source string
	var status string
	var created string
	var count int
	var spec string
	if ist.Spec != nil {
		spec = "Tag"
		if ist.Spec.Reference {
			spec = "Ref"
		}
		if ist.Spec.ImportPolicy.Scheduled {
			spec = "Scheduled"
		}
		if ist.Spec.Generation != nil {
			generation = *ist.Spec.Generation
		}
		if ist.Spec.From != nil {
			switch ist.Spec.From.Kind {
			case "DockerImage":
				source = ist.Spec.From.Name
			case "ImageStreamImage":
				if _, id, ok := imageutil.SplitImageStreamImage(ist.Spec.From.Name); ok {
					source = fmt.Sprintf("image/%s", id)
				} else {
					status = "InvalidRefName"
				}
			case "ImageStreamTag":
				if len(ist.Spec.From.Namespace) > 0 {
					source = fmt.Sprintf("istag %s/%s", ist.Spec.From.Namespace, ist.Spec.From.Name)
				} else {
					source = fmt.Sprintf("istag/%s", ist.Spec.From.Name)
				}
				if _, _, ok := imageutil.SplitImageStreamTag(ist.Spec.From.Name); !ok {
					// tags that reference other internal tags are "tracking tags" and should
					// be denoted as such
					spec = "Track"
				}
			default:
				status = "InvalidRefKind"
			}
		} else {
			spec = ""
		}
	}
	if ist.Status != nil {
		count = len(ist.Status.Items)
		var last time.Time
		if count > 0 {
			event := ist.Status.Items[0]
			last = event.Created.Time
			if generation > 0 && event.Generation > 0 && event.Generation < generation {
				status = fmt.Sprintf("Importing")
			} else if len(event.Image) > 0 {
				source = fmt.Sprintf("image/%s", event.Image)
			} else if len(event.DockerImageReference) > 0 {
				source = event.DockerImageReference
			}
			switch {
			case len(source) > 0 && len(spec) == 0:
				// default to push if no spec
				spec = "Push"
			case spec == "Track":
				// TODO: tracking tags do not have a generation that matches their tag, we
				//   can't be sure that the tag matches
			case generation > 0 && event.Generation > 0 && event.Generation > generation:
				// if the generation is newer than the spec tag, it was either reset or
				// another user has pushed to this tag
				spec = "Push"
			}
		}
		for _, condition := range ist.Status.Conditions {
			if condition.Type == imageapi.ImportSuccess && condition.Status == core.ConditionFalse {
				status = fmt.Sprintf("ImportFailed (%s)", condition.Reason)
			}
			if condition.LastTransitionTime.After(last) {
				last = condition.LastTransitionTime.Time
			}
		}
		if !last.IsZero() {
			created = fmt.Sprintf("%s ago", formatRelativeTime(last))
		}
	}
	if len(status) == 0 {
		status = source
	}

	if count > 0 {
		row.Cells = append(row.Cells, ist.Name, spec, status, count, created)
	} else {
		row.Cells = append(row.Cells, ist.Name, spec, status, nil, created)
	}

	return []metav1.TableRow{row}, nil
}

func printImageTagList(list *imageapi.ImageTagList, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(list.Items))
	for i := range list.Items {
		r, err := printImageTag(&list.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func printImageStreamImage(isi *imageapi.ImageStreamImage, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: isi},
	}

	created := fmt.Sprintf("%s ago", formatRelativeTime(isi.CreationTimestamp.Time))

	row.Cells = append(row.Cells, isi.Name, created)

	if options.Wide {
		row.Cells = append(row.Cells, isi.Image.DockerImageReference, isi.Image.Name)
	}

	return []metav1.TableRow{row}, nil
}

// printTagsUpToWidth displays a human readable list of tags with as many tags as will fit in the
// width we budget. It will always display at least one tag, and will allow a slightly wider width
// if it's less than 25% of the total width to feel more even.
func printTagsUpToWidth(statusTags map[string]imageapi.TagEventList, preferredWidth int) string {
	tags := imageapi.SortStatusTags(statusTags)
	remaining := preferredWidth
	for i, tag := range tags {
		remaining -= len(tag) + 1
		if remaining >= 0 {
			continue
		}
		if i == 0 {
			tags = tags[:1]
			break
		}
		// if we've left more than 25% of the width unfilled, and adding the current tag would be
		// less than 125% of the preferred width, keep going in order to make the edges less ragged.
		margin := preferredWidth / 4
		if margin < (remaining+len(tag)) && margin >= (-remaining) {
			continue
		}
		tags = tags[:i]
		break
	}
	if hiddenTags := len(statusTags) - len(tags); hiddenTags > 0 {
		return fmt.Sprintf("%s + %d more...", strings.Join(tags, ","), hiddenTags)
	}
	return strings.Join(tags, ",")
}

// formatResourceName receives a resource kind, name, and boolean specifying
// whether or not to update the current name to "kind/name"
func formatResourceName(kind schema.GroupKind, name string, withKind bool) string {
	if !withKind || kind.Empty() {
		return name
	}

	return strings.ToLower(kind.String()) + "/" + name
}

func formatRelativeTime(t time.Time) string {
	return units.HumanDuration(timeNowFn().Sub(t))
}

var timeNowFn = func() time.Time {
	return time.Now()
}
