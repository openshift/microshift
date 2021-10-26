package printerstorage

import (
	"context"
	"fmt"

	"github.com/openshift/oauth-apiserver/pkg/printers"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// TableConvertor struct - converts objects to metav1beta1.Table using printers.TableGenerator
type TableConvertor struct {
	printers.TableGenerator
}

// ConvertToTable method - converts objects to metav1beta1.Table objects using TableGenerator
func (c TableConvertor) ConvertToTable(ctx context.Context, obj runtime.Object, tableOptions runtime.Object) (*metav1beta1.Table, error) {
	noHeaders := false
	if tableOptions != nil {
		switch t := tableOptions.(type) {
		case *metav1beta1.TableOptions:
			if t != nil {
				noHeaders = t.NoHeaders
			}
		default:
			return nil, fmt.Errorf("unrecognized type %T for table options, can't display tabular output", tableOptions)
		}
	}
	return c.TableGenerator.GenerateTable(obj, printers.GenerateOptions{Wide: true, NoHeaders: noHeaders})
}
