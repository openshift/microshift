package registries

import (
	"sort"
	"strings"

	"github.com/containers/image/pkg/sysregistriesv2"
	apioperatorsv1alpha1 "github.com/openshift/api/operator/v1alpha1"
)

// scopeMatchesRegistry returns true if a scope value (as in sysregistriesv2.Registry.Prefix / sysregistriesv2.Endpoint.Location)
// matches a host[:port] value in reg.
func scopeMatchesRegistry(scope, reg string) bool {
	if reg == scope {
		return true
	}
	if len(scope) > len(reg) {
		return strings.HasPrefix(scope, reg) && scope[len(reg)] == '/'
	}
	return false
}

// rdmContainsARealMirror returns true if set.Mirrors contains at least one entry that is not set.Source.
func rdmContainsARealMirror(set *apioperatorsv1alpha1.RepositoryDigestMirrors) bool {
	for _, mirror := range set.Mirrors {
		if mirror != set.Source {
			return true
		}
	}
	return false
}

// mergedMirrorSets processes icspRules and returns a set of RepositoryDigestMirrors, one for each Source value,
// ordered consistently with the preference order of the individual entries (if possible)
// E.g. given mirror sets (B, C) and (A, B), it will combine them into a single (A, B, C) set.
func mergedMirrorSets(icspRules []*apioperatorsv1alpha1.ImageContentSourcePolicy) ([]apioperatorsv1alpha1.RepositoryDigestMirrors, error) {
	disjointSets := map[string]*[]*apioperatorsv1alpha1.RepositoryDigestMirrors{} // Key == Source
	for _, icsp := range icspRules {
		for i := range icsp.Spec.RepositoryDigestMirrors {
			set := &icsp.Spec.RepositoryDigestMirrors[i]
			if !rdmContainsARealMirror(set) {
				continue // No mirrors (or mirrors that only repeat the authoritative source) is not really a mirror set.
			}
			ds, ok := disjointSets[set.Source]
			if !ok {
				ds = &[]*apioperatorsv1alpha1.RepositoryDigestMirrors{}
				disjointSets[set.Source] = ds
			}
			*ds = append(*ds, set)
		}
	}

	// Sort the sets of mirrors by Source to ensure deterministic output
	sources := []string{}
	for key := range disjointSets {
		sources = append(sources, key)
	}
	sort.Strings(sources)
	// Convert the sets of mirrors
	res := []apioperatorsv1alpha1.RepositoryDigestMirrors{}
	for _, source := range sources {
		ds := disjointSets[source]
		topoGraph := newTopoGraph()
		for _, set := range *ds {
			for i := 0; i+1 < len(set.Mirrors); i++ {
				topoGraph.AddEdge(set.Mirrors[i], set.Mirrors[i+1])
			}
			sourceInGraph := false
			for _, m := range set.Mirrors {
				if m == source {
					sourceInGraph = true
					break
				}
			}
			if !sourceInGraph {
				// The build of mirrorSets guarantees len(set.Mirrors) > 0.
				topoGraph.AddEdge(set.Mirrors[len(set.Mirrors)-1], source)
			}
			// Every node in topoGraph, including source, is implicitly added by topoGraph.AddEdge (every mirror set contains at least one non-source mirror,
			// so there are no unconnected nodes that we would have to add separately from the edges).
		}
		sortedRepos, err := topoGraph.Sorted()
		if err != nil {
			return nil, err
		}
		if sortedRepos[len(sortedRepos)-1] == source {
			// We don't need to explicitly include source in the list, it will be automatically tried last per the semantics of sysregistriesv2. Mirrors.
			sortedRepos = sortedRepos[:len(sortedRepos)-1]
		}
		res = append(res, apioperatorsv1alpha1.RepositoryDigestMirrors{
			Source:  source,
			Mirrors: sortedRepos,
		})
	}
	return res, nil
}

// EditRegistriesConfig edits, IN PLACE, the /etc/containers/registries.conf configuration provided in config, to:
// - Mark whole registries in insecureRegistries as insecure (TLS is not required, and TLS certificate verification is not required when TLS is used)
// - Mark whole registries in blockedRegistries as blocked (any attempts to access them fail)
// - Implement ImageContentSourcePolicy rules in icspRules.
// “Whole registries” above means that the configuration applies to everything on that registry, including any possible separately-configured
// namespaces/repositories within that registry.
func EditRegistriesConfig(config *sysregistriesv2.V2RegistriesConf, insecureRegistries, blockedRegistries []string, icspRules []*apioperatorsv1alpha1.ImageContentSourcePolicy) error {
	// getRegistryEntry returns a pointer to a modifiable Registry object corresponding to scope,
	// creating it if necessary.
	// NOTE: We never generate entries with Prefix != Location, so everything in updateRegistriesConfig
	// only checks Location.
	// NOTE: The pointer is valid only until the next getRegistryEntry call.
	getRegistryEntry := func(scope string) *sysregistriesv2.Registry {
		for i := range config.Registries {
			if config.Registries[i].Location == scope {
				return &config.Registries[i]
			}
		}
		config.Registries = append(config.Registries, sysregistriesv2.Registry{
			Endpoint: sysregistriesv2.Endpoint{Location: scope},
		})
		return &config.Registries[len(config.Registries)-1]
	}

	mirrorSets, err := mergedMirrorSets(icspRules)
	if err != nil {
		return err
	}
	for _, mirrorSet := range mirrorSets {
		reg := getRegistryEntry(mirrorSet.Source)
		reg.MirrorByDigestOnly = true
		for _, mirror := range mirrorSet.Mirrors {
			reg.Mirrors = append(reg.Mirrors, sysregistriesv2.Endpoint{Location: mirror})
		}
	}

	// insecureRegistries and blockedRegistries are lists of registries; now that mirrors can be configured at a namespace/repo level,
	// configuration at the namespace/repo level would shadow the registry-level entries; so, propagate the insecure/blocked
	// flags to the child namespaces as well.
	for _, insecureReg := range insecureRegistries {
		reg := getRegistryEntry(insecureReg)
		reg.Insecure = true
		for i := range config.Registries {
			reg := &config.Registries[i]
			if scopeMatchesRegistry(reg.Location, insecureReg) {
				reg.Insecure = true
			}
			for j := range reg.Mirrors {
				mirror := &reg.Mirrors[j]
				if scopeMatchesRegistry(mirror.Location, insecureReg) {
					mirror.Insecure = true
				}
			}
		}
	}
	for _, blockedReg := range blockedRegistries {
		reg := getRegistryEntry(blockedReg)
		reg.Blocked = true
		for i := range config.Registries {
			reg := &config.Registries[i]
			if scopeMatchesRegistry(reg.Location, blockedReg) {
				reg.Blocked = true
			}
		}
	}
	return nil
}
