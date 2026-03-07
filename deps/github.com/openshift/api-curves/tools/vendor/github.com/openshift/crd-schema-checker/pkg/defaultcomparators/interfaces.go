package defaultcomparators

import "github.com/openshift/crd-schema-checker/pkg/manifestcomparators"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func NewDefaultComparators() manifestcomparators.CRDComparatorRegistry {
	ret := manifestcomparators.NewRegistry()
	must(ret.AddComparator(manifestcomparators.NoBools()))
	must(ret.AddComparator(manifestcomparators.NoFloats()))
	must(ret.AddComparator(manifestcomparators.NoUints()))
	must(ret.AddComparator(manifestcomparators.NoFieldRemoval()))
	must(ret.AddComparator(manifestcomparators.NoEnumRemoval()))
	must(ret.AddComparator(manifestcomparators.NoMaps()))
	must(ret.AddComparator(manifestcomparators.NoDataTypeChange()))
	must(ret.AddComparator(manifestcomparators.MustHaveStatus()))
	must(ret.AddComparator(manifestcomparators.ListsMustHaveSSATags()))
	must(ret.AddComparator(manifestcomparators.ConditionsMustHaveProperSSATags()))
	must(ret.AddComparator(manifestcomparators.NoNewRequiredFields()))
	must(ret.AddComparator(manifestcomparators.MustNotExceedCostBudget()))

	/*
		other useful comparators

		2. don't change field types
		3. don't tighten validation rules
		4. don't loosen validation rules (warning)
		6. conditions must match metav1.Conditions with proper SSA tags
		7. all lists must have SSA tags
		9. don't use floats
		10. don't use unsigned ints
		11. no durations (for kube, openshift configuration API allowed)
		13. enumerated values should use CamelCase
		14. optional should be pointers (for kube, openshift configuration API allowed)
		15. no new fields can be required
		17. defaulted values have to be optional to pass CR validation
		18. stable versions cannot coexist with unstable versions
		19. booleans cannot be defaulted
		20. no uses of corev1.ObjectReference
		21. no uses of corev1.LocalObjectReference
		22. no new enumerated values (warning)
		23. no removed enumerated values (error)
		24. no replace in list (warning)

		On a risky update we could add a live-lookup mode that runs existing data through the new validation rules as
		though it was a create and runs it a second time as though it was a "from" level that had dummy annotation change.

		.annotations[version.path.to.field/compratorname] = github.com/openshift

		Require cryptographically signed hash of offensive change to allow?
	*/

	return ret
}
