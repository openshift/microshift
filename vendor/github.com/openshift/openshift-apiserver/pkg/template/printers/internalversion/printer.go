package internalversion

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kprinters "k8s.io/kubernetes/pkg/printers"

	templatev1 "github.com/openshift/api/template/v1"

	templateapi "github.com/openshift/openshift-apiserver/pkg/template/apis/template"
)

const templateDescriptionLen = 80

func AddTemplateOpenShiftHandlers(h kprinters.PrintHandler) {
	templateColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Description", Type: "string", Description: "Template description."},
		{Name: "Parameters", Type: "string", Description: templatev1.Template{}.SwaggerDoc()["parameters"]},
		{Name: "Objects", Type: "string", Description: "Number of resources in this template."},
	}
	if err := h.TableHandler(templateColumnDefinitions, printTemplateList); err != nil {
		panic(err)
	}
	if err := h.TableHandler(templateColumnDefinitions, printTemplate); err != nil {
		panic(err)
	}

	templateInstanceColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Template", Type: "string", Description: "Template name to be instantiated."},
	}
	if err := h.TableHandler(templateInstanceColumnDefinitions, printTemplateInstanceList); err != nil {
		panic(err)
	}
	if err := h.TableHandler(templateInstanceColumnDefinitions, printTemplateInstance); err != nil {
		panic(err)
	}

	brokerTemplateInstanceColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Template Instance", Type: "string", Description: "Template instance name."},
	}
	if err := h.TableHandler(brokerTemplateInstanceColumnDefinitions, printBrokerTemplateInstanceList); err != nil {
		panic(err)
	}
	if err := h.TableHandler(brokerTemplateInstanceColumnDefinitions, printBrokerTemplateInstance); err != nil {
		panic(err)
	}
}

func printTemplate(t *templateapi.Template, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: t},
	}

	description := ""
	if t.Annotations != nil {
		description = t.Annotations["description"]
	}
	// Only print the first line of description
	if lines := strings.SplitN(description, "\n", 2); len(lines) > 1 {
		description = lines[0] + "..."
	}
	if len(description) > templateDescriptionLen {
		description = strings.TrimSpace(description[:templateDescriptionLen-3]) + "..."
	}
	empty, generated, total := 0, 0, len(t.Parameters)
	for _, p := range t.Parameters {
		if len(p.Value) > 0 {
			continue
		}
		if len(p.Generate) > 0 {
			generated++
			continue
		}
		empty++
	}
	params := ""
	switch {
	case empty > 0:
		params = fmt.Sprintf("%d (%d blank)", total, empty)
	case generated > 0:
		params = fmt.Sprintf("%d (%d generated)", total, generated)
	default:
		params = fmt.Sprintf("%d (all set)", total)
	}

	row.Cells = append(row.Cells, t.Name, description, params, int64(len(t.Objects)))

	return []metav1.TableRow{row}, nil
}

func printTemplateList(list *templateapi.TemplateList, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(list.Items))
	for i := range list.Items {
		r, err := printTemplate(&list.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func printTemplateInstance(templateInstance *templateapi.TemplateInstance, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: templateInstance},
	}

	row.Cells = append(row.Cells, templateInstance.Name, templateInstance.Spec.Template.Name)

	return []metav1.TableRow{row}, nil
}

func printTemplateInstanceList(list *templateapi.TemplateInstanceList, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(list.Items))
	for i := range list.Items {
		r, err := printTemplateInstance(&list.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func printBrokerTemplateInstance(brokerTemplateInstance *templateapi.BrokerTemplateInstance, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: brokerTemplateInstance},
	}

	row.Cells = append(row.Cells, brokerTemplateInstance.Name, brokerTemplateInstance.Spec.TemplateInstance.Namespace, brokerTemplateInstance.Spec.TemplateInstance.Name)

	return []metav1.TableRow{row}, nil
}

func printBrokerTemplateInstanceList(list *templateapi.BrokerTemplateInstanceList, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(list.Items))
	for i := range list.Items {
		r, err := printBrokerTemplateInstance(&list.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}
