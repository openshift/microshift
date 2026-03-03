package manifestcomparators

import (
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

type crdComparatorRegistry struct {
	comparators map[string]CRDComparator
}

func NewRegistry() CRDComparatorRegistry {
	return &crdComparatorRegistry{
		comparators: map[string]CRDComparator{},
	}
}

func (r *crdComparatorRegistry) AddComparator(comparator CRDComparator) error {
	if _, ok := r.comparators[comparator.Name()]; ok {
		return fmt.Errorf("comparator/%v is already registered", comparator.Name())
	}

	r.comparators[comparator.Name()] = comparator
	return nil
}

func (r *crdComparatorRegistry) GetComparator(name string) (CRDComparator, error) {
	ret, ok := r.comparators[name]
	if !ok {
		return nil, fmt.Errorf("comparator/%v is not registered", name)

	}
	return ret, nil
}

func (r *crdComparatorRegistry) KnownComparators() []string {
	keys := sets.StringKeySet(r.comparators)
	return keys.List()
}

func (r *crdComparatorRegistry) AllComparators() []CRDComparator {
	ret := []CRDComparator{}

	keys := sets.StringKeySet(r.comparators)
	for _, name := range keys.List() {
		ret = append(ret, r.comparators[name])
	}

	return ret
}

func (r *crdComparatorRegistry) Compare(existingCRD, newCRD *apiextensionsv1.CustomResourceDefinition, names ...string) ([]ComparisonResults, []error) {
	comparators := []CRDComparator{}
	if len(names) == 0 {
		comparators = r.AllComparators()
	} else {
		for _, name := range names {
			comparator, err := r.GetComparator(name)
			if err != nil {
				return nil, []error{err}
			}
			comparators = append(comparators, comparator)
		}
	}

	ret := []ComparisonResults{}
	errs := []error{}
	for _, comparator := range comparators {
		currResults, err := comparator.Compare(existingCRD, newCRD)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		ret = append(ret, currResults)
	}

	return ret, errs
}
