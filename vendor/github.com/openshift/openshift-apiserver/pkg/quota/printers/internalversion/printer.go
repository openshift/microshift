package internalversion

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kprinters "k8s.io/kubernetes/pkg/printers"

	quotav1 "github.com/openshift/api/quota/v1"

	quotaapi "github.com/openshift/openshift-apiserver/pkg/quota/apis/quota"
)

func AddQuotaOpenShiftHandler(h kprinters.PrintHandler) {
	addClusterResourceQuota(h)
}

func addClusterResourceQuota(h kprinters.PrintHandler) {
	clusterResourceQuotaColumnsDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Labels Selector", Type: "string", Description: quotav1.ClusterResourceQuotaSelector{}.SwaggerDoc()["labels"]},
		{Name: "Annotations Selector", Type: "string", Description: quotav1.ClusterResourceQuotaSelector{}.SwaggerDoc()["annotations"]},
	}
	if err := h.TableHandler(clusterResourceQuotaColumnsDefinitions, printClusterResourceQuota); err != nil {
		panic(err)
	}
	if err := h.TableHandler(clusterResourceQuotaColumnsDefinitions, printClusterResourceQuotaList); err != nil {
		panic(err)
	}
	if err := h.TableHandler(clusterResourceQuotaColumnsDefinitions, printAppliedClusterResourceQuota); err != nil {
		panic(err)
	}
	if err := h.TableHandler(clusterResourceQuotaColumnsDefinitions, printAppliedClusterResourceQuotaList); err != nil {
		panic(err)
	}
}

func printClusterResourceQuota(clusterResourceQuota *quotaapi.ClusterResourceQuota, _ kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: clusterResourceQuota},
	}

	row.Cells = append(row.Cells,
		clusterResourceQuota.Name,
		metav1.FormatLabelSelector(clusterResourceQuota.Spec.Selector.LabelSelector),
		clusterResourceQuota.Spec.Selector.AnnotationSelector,
	)

	return []metav1.TableRow{row}, nil
}

func printClusterResourceQuotaList(clusterResourceQuotaList *quotaapi.ClusterResourceQuotaList, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(clusterResourceQuotaList.Items))
	for i := range clusterResourceQuotaList.Items {
		r, err := printClusterResourceQuota(&clusterResourceQuotaList.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func printAppliedClusterResourceQuota(appliedClusterResourceQuota *quotaapi.AppliedClusterResourceQuota, _ kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: appliedClusterResourceQuota},
	}

	row.Cells = append(row.Cells,
		appliedClusterResourceQuota.Name,
		metav1.FormatLabelSelector(appliedClusterResourceQuota.Spec.Selector.LabelSelector),
		appliedClusterResourceQuota.Spec.Selector.AnnotationSelector,
	)

	return []metav1.TableRow{row}, nil
}

func printAppliedClusterResourceQuotaList(appliedClusterResourceQuotaList *quotaapi.AppliedClusterResourceQuotaList, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(appliedClusterResourceQuotaList.Items))
	for i := range appliedClusterResourceQuotaList.Items {
		r, err := printAppliedClusterResourceQuota(&appliedClusterResourceQuotaList.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}
