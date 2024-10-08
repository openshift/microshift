// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/openshift/api/template/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/listers"
	"k8s.io/client-go/tools/cache"
)

// BrokerTemplateInstanceLister helps list BrokerTemplateInstances.
// All objects returned here must be treated as read-only.
type BrokerTemplateInstanceLister interface {
	// List lists all BrokerTemplateInstances in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.BrokerTemplateInstance, err error)
	// Get retrieves the BrokerTemplateInstance from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.BrokerTemplateInstance, error)
	BrokerTemplateInstanceListerExpansion
}

// brokerTemplateInstanceLister implements the BrokerTemplateInstanceLister interface.
type brokerTemplateInstanceLister struct {
	listers.ResourceIndexer[*v1.BrokerTemplateInstance]
}

// NewBrokerTemplateInstanceLister returns a new BrokerTemplateInstanceLister.
func NewBrokerTemplateInstanceLister(indexer cache.Indexer) BrokerTemplateInstanceLister {
	return &brokerTemplateInstanceLister{listers.New[*v1.BrokerTemplateInstance](indexer, v1.Resource("brokertemplateinstance"))}
}
