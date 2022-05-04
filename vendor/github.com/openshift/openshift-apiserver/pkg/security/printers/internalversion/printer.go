package internalversion

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kprinters "k8s.io/kubernetes/pkg/printers"

	securityv1 "github.com/openshift/api/security/v1"
	securityapi "github.com/openshift/openshift-apiserver/pkg/security/apis/security"
)

func AddSecurityOpenShiftHandler(h kprinters.PrintHandler) {
	addSecurityContextConstraint(h)
	addRangeAllocation(h)
}

func addRangeAllocation(h kprinters.PrintHandler) {
	rangeAllocationColumnsDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Range", Type: "string", Format: "name", Description: securityv1.RangeAllocation{}.SwaggerDoc()["range"]},
		{Name: "Data", Type: "string", Description: securityv1.RangeAllocation{}.SwaggerDoc()["data"]},
	}
	if err := h.TableHandler(rangeAllocationColumnsDefinitions, printRangeAllocation); err != nil {
		panic(err)
	}
	if err := h.TableHandler(rangeAllocationColumnsDefinitions, printRangeAllocationList); err != nil {
		panic(err)
	}
}

func printRangeAllocation(allocation *securityapi.RangeAllocation, _ kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: allocation},
	}

	row.Cells = append(row.Cells,
		allocation.Name,
		allocation.Range,
		fmt.Sprintf("0x%x", allocation.Data),
	)

	return []metav1.TableRow{row}, nil
}

func printRangeAllocationList(allocationList *securityapi.RangeAllocationList, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(allocationList.Items))
	for i := range allocationList.Items {
		r, err := printRangeAllocation(&allocationList.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func addSecurityContextConstraint(h kprinters.PrintHandler) {
	securityContextConstraintsColumnsDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Privileged", Type: "bool", Description: securityv1.SecurityContextConstraints{}.SwaggerDoc()["allowPrivilegedContainer"]},
		{Name: "Capabilities", Type: "string", Description: securityv1.SecurityContextConstraints{}.SwaggerDoc()["allowedCapabilities"]},
		{Name: "SELinux", Type: "string", Description: securityv1.SecurityContextConstraints{}.SwaggerDoc()["seLinuxContext"]},
		{Name: "RunAsUser", Type: "string", Description: securityv1.SecurityContextConstraints{}.SwaggerDoc()["runAsUser"]},
		{Name: "FSGroup", Type: "string", Description: securityv1.SecurityContextConstraints{}.SwaggerDoc()["fsGroup"]},
		{Name: "SupplementalGroups", Type: "string", Description: securityv1.SecurityContextConstraints{}.SwaggerDoc()["supplementalGroups"]},
		{Name: "Priority", Type: "integer", Description: securityv1.SecurityContextConstraints{}.SwaggerDoc()["priority"]},
		{Name: "ReadOnlyFS", Type: "bool", Description: securityv1.SecurityContextConstraints{}.SwaggerDoc()["readOnlyRootFilesystem"]},
		{Name: "Volumes", Type: "string", Description: securityv1.SecurityContextConstraints{}.SwaggerDoc()["volumes"]},
	}

	if err := h.TableHandler(securityContextConstraintsColumnsDefinitions, printSecurityContextConstraint); err != nil {
		panic(err)
	}
	if err := h.TableHandler(securityContextConstraintsColumnsDefinitions, printSecurityContextConstraintList); err != nil {
		panic(err)
	}
}

func printSecurityContextConstraint(scc *securityapi.SecurityContextConstraints, _ kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: scc},
	}

	capabilities := []string{}
	for _, c := range scc.AllowedCapabilities {
		capabilities = append(capabilities, string(c))
	}

	priority := "<none>"
	if scc.Priority != nil {
		priority = fmt.Sprintf("%d", *scc.Priority)
	}

	volumes := []string{}
	for _, v := range scc.Volumes {
		volumes = append(volumes, string(v))
	}

	row.Cells = append(row.Cells,
		scc.Name,
		scc.AllowPrivilegedContainer,
		strings.Join(capabilities, ","),
		string(scc.SELinuxContext.Type),
		string(scc.RunAsUser.Type),
		string(scc.FSGroup.Type),
		string(scc.SupplementalGroups.Type),
		priority,
		scc.ReadOnlyRootFilesystem,
		strings.Join(volumes, ","),
	)

	return []metav1.TableRow{row}, nil
}

func printSecurityContextConstraintList(sccList *securityapi.SecurityContextConstraintsList, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(sccList.Items))
	for i := range sccList.Items {
		r, err := printSecurityContextConstraint(&sccList.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}
