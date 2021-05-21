package internalversion

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	units "github.com/docker/go-units"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kprinters "k8s.io/kubernetes/pkg/printers"

	buildv1 "github.com/openshift/api/build/v1"
	buildapi "github.com/openshift/openshift-apiserver/pkg/build/apis/build"
	buildinternalhelpers "github.com/openshift/openshift-apiserver/pkg/build/apis/build/internal_helpers"
)

func AddBuildOpenShiftHandlers(h kprinters.PrintHandler) {
	buildColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Type", Type: "string", Description: "Describes a particular way of performing a build."},
		{Name: "From", Type: "string", Description: buildv1.CommonSpec{}.SwaggerDoc()["source"]},
		{Name: "Status", Type: "string", Description: buildv1.BuildStatus{}.SwaggerDoc()["phase"]},
		{Name: "Started", Type: "string", Description: buildv1.BuildStatus{}.SwaggerDoc()["startTimestamp"]},
		{Name: "Duration", Type: "string", Description: buildv1.BuildStatus{}.SwaggerDoc()["duration"]},
	}
	if err := h.TableHandler(buildColumnDefinitions, printBuildList); err != nil {
		panic(err)
	}
	if err := h.TableHandler(buildColumnDefinitions, printBuild); err != nil {
		panic(err)
	}

	buildConfigColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Type", Type: "string", Description: "Describes a particular way of performing a build."},
		{Name: "From", Type: "string", Description: buildv1.CommonSpec{}.SwaggerDoc()["source"]},
		{Name: "Latest", Type: "string", Description: buildv1.BuildConfigStatus{}.SwaggerDoc()["lastVersion"]},
	}
	if err := h.TableHandler(buildConfigColumnDefinitions, printBuildConfigList); err != nil {
		panic(err)
	}
	if err := h.TableHandler(buildConfigColumnDefinitions, printBuildConfig); err != nil {
		panic(err)
	}
}

func printBuild(build *buildapi.Build, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: build},
	}

	var created string
	if build.Status.StartTimestamp != nil {
		created = fmt.Sprintf("%s ago", formatRelativeTime(build.Status.StartTimestamp.Time))
	}
	var duration string
	if build.Status.Duration > 0 {
		duration = build.Status.Duration.String()
	}
	from := describeSourceShort(build.Spec.CommonSpec)
	status := string(build.Status.Phase)
	if len(build.Status.Reason) > 0 {
		status = fmt.Sprintf("%s (%s)", status, build.Status.Reason)
	}

	row.Cells = append(row.Cells, build.Name, buildinternalhelpers.StrategyType(build.Spec.Strategy),
		from, status, created, duration)

	return []metav1.TableRow{row}, nil
}

func printBuildList(list *buildapi.BuildList, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	sort.Sort(buildinternalhelpers.BuildSliceByCreationTimestamp(list.Items))
	rows := make([]metav1.TableRow, 0, len(list.Items))
	for i := range list.Items {
		r, err := printBuild(&list.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func printBuildConfig(bc *buildapi.BuildConfig, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: bc},
	}

	from := describeSourceShort(bc.Spec.CommonSpec)

	if bc.Spec.Strategy.CustomStrategy != nil {
		row.Cells = append(row.Cells, bc.Name, buildinternalhelpers.StrategyType(bc.Spec.Strategy),
			bc.Spec.Strategy.CustomStrategy.From.Name, bc.Status.LastVersion)
	} else {
		row.Cells = append(row.Cells, bc.Name, buildinternalhelpers.StrategyType(bc.Spec.Strategy), from,
			bc.Status.LastVersion)
	}

	return []metav1.TableRow{row}, nil

}

func printBuildConfigList(list *buildapi.BuildConfigList, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(list.Items))
	for i := range list.Items {
		r, err := printBuildConfig(&list.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func describeSourceShort(spec buildapi.CommonSpec) string {
	var from string
	switch source := spec.Source; {
	case source.Binary != nil:
		from = "Binary"
		if rev := describeSourceGitRevision(spec); len(rev) != 0 {
			from = fmt.Sprintf("%s@%s", from, rev)
		}
	case source.Dockerfile != nil && source.Git != nil:
		from = "Dockerfile,Git"
		if rev := describeSourceGitRevision(spec); len(rev) != 0 {
			from = fmt.Sprintf("%s@%s", from, rev)
		}
	case source.Dockerfile != nil:
		from = "Dockerfile"
	case source.Git != nil:
		from = "Git"
		if rev := describeSourceGitRevision(spec); len(rev) != 0 {
			from = fmt.Sprintf("%s@%s", from, rev)
		}
	default:
		from = buildSourceType(source)
	}
	return from
}

func buildSourceType(source buildapi.BuildSource) string {
	var sourceType string
	if source.Git != nil {
		sourceType = "Git"
	}
	if source.Dockerfile != nil {
		if len(sourceType) != 0 {
			sourceType = sourceType + ","
		}
		sourceType = sourceType + "Dockerfile"
	}
	if source.Binary != nil {
		if len(sourceType) != 0 {
			sourceType = sourceType + ","
		}
		sourceType = sourceType + "Binary"
	}
	return sourceType
}

var nonCommitRev = regexp.MustCompile("[^a-fA-F0-9]")

func describeSourceGitRevision(spec buildapi.CommonSpec) string {
	var rev string
	if spec.Revision != nil && spec.Revision.Git != nil {
		rev = spec.Revision.Git.Commit
	}
	if len(rev) == 0 && spec.Source.Git != nil {
		rev = spec.Source.Git.Ref
	}
	// if this appears to be a full Git commit hash, shorten it to 7 characters for brevity
	if !nonCommitRev.MatchString(rev) && len(rev) > 20 {
		rev = rev[:7]
	}
	return rev
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
