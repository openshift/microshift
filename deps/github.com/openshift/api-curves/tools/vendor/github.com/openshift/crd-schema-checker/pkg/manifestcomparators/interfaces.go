package manifestcomparators

import apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

type ComparisonResults struct {
	Name         string `yaml:"name"`
	WhyItMatters string `yaml:"whyItMatters"`

	Errors   []string `yaml:"errors"`
	Warnings []string `yaml:"warnings"`
	Infos    []string `yaml:"infos"`
}

type CRDComparator interface {
	Name() string
	WhyItMatters() string
	Compare(existingCRD, newCRD *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error)
}

type SingleCRDValidator interface {
	Validate(crd *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error)
}

type CRDComparatorRegistry interface {
	AddComparator(comparator CRDComparator) error
	GetComparator(name string) (CRDComparator, error)

	KnownComparators() []string
	AllComparators() []CRDComparator

	Compare(existingCRD, newCRD *apiextensionsv1.CustomResourceDefinition, names ...string) ([]ComparisonResults, []error)
}
