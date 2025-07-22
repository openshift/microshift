package psalabelsyncer

import (
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	rbacv1listers "k8s.io/client-go/listers/rbac/v1"
	"k8s.io/client-go/tools/cache"

	securityv1 "github.com/openshift/api/security/v1"
	securityv1listers "github.com/openshift/client-go/security/listers/security/v1"
)

const (
	NS1     = "mambonumberfive"
	NS2     = "mumbonumbertwo"
	SA1Name = "one"
	SA2Name = "two"
)

var (
	basicSCCs = []string{"scc_authenticated", "scc_allsa"}
	allSCCs   = append(basicSCCs, "scc_sa1", "scc_sa1group_sa2", "scc_none", "scc_none2")

	NS1SA1 = corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SA1Name,
			Namespace: NS1,
		},
	}
	NS2SA1 = corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SA1Name,
			Namespace: NS2,
		},
	}
	NS2SA2 = corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SA2Name,
			Namespace: NS2,
		},
	}
	NSDontCareSA1 = corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SA1Name,
			Namespace: "randomns",
		},
	}
)

func sccLister(t *testing.T) securityv1listers.SecurityContextConstraintsLister {
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	for _, scc := range []*securityv1.SecurityContextConstraints{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "scc_authenticated",
			},
			Groups: []string{"system:authenticated"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "scc_allsa",
			},
			Groups: []string{"system:serviceaccounts"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "scc_sa1",
			},
			Users: []string{"system:serviceaccount:" + NS1 + ":" + SA1Name},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "scc_sa1group_sa2",
			},
			Groups: []string{"system:serviceaccounts:" + NS1},
			Users:  []string{"system:serviceaccount:" + NS2 + ":" + SA1Name},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "scc_none",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "scc_none2",
			},
		},
	} {
		require.NoError(t, indexer.Add(scc))
	}

	return securityv1listers.NewSecurityContextConstraintsLister(indexer)
}

func TestSCCRoleCache_SCCsFor(t *testing.T) {

	tests := []struct {
		name         string
		roles        []*rbacv1.Role
		clusterRoles []*rbacv1.ClusterRole

		roleBindings        []*rbacv1.RoleBinding
		clusterRoleBindings []*rbacv1.ClusterRoleBinding

		serviceAccount *corev1.ServiceAccount
		sccsOverrides  []*securityv1.SecurityContextConstraints

		want    sets.String
		wantErr bool
	}{
		{
			name:           "no SCCs found",
			serviceAccount: &NSDontCareSA1,
			sccsOverrides:  []*securityv1.SecurityContextConstraints{},
		},
		{
			name:           "only SCCs with authenticated/SAs groups match",
			serviceAccount: &NSDontCareSA1,
			want:           sets.NewString(basicSCCs...),
		},
		{
			name:           "SCC with specific SA username matches",
			serviceAccount: &NS2SA1,
			want:           sets.NewString(basicSCCs...).Insert("scc_sa1group_sa2"),
		},
		{
			name:           "SCC with specific SA user and NS group matches",
			serviceAccount: &NS1SA1,
			want:           sets.NewString(basicSCCs...).Insert("scc_sa1", "scc_sa1group_sa2"),
		},
		{
			name:           "all SCCs assigned via a broad cluster rolebinding",
			serviceAccount: &NS2SA1,
			clusterRoles: []*rbacv1.ClusterRole{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "clusterrole",
					},
					Rules: []rbacv1.PolicyRule{
						{
							APIGroups: []string{"*"},
							Resources: []string{"*"},
							Verbs:     []string{"*"},
						},
					},
				},
			},
			clusterRoleBindings: []*rbacv1.ClusterRoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "clusterrolebinding",
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.GroupName,
						Kind:     "ClusterRole",
						Name:     "clusterrole",
					},
					Subjects: []rbacv1.Subject{
						{
							APIGroup:  corev1.GroupName,
							Kind:      "ServiceAccount",
							Name:      SA1Name,
							Namespace: NS2,
						},
					},
				},
			},
			want: sets.NewString(allSCCs...),
		},

		{
			name:           "single SCCs assigned via a cluster rolebinding",
			serviceAccount: &NS2SA2,
			clusterRoles: []*rbacv1.ClusterRole{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "clusterrole",
					},
					Rules: []rbacv1.PolicyRule{
						{
							APIGroups:     []string{securityv1.GroupName},
							Resources:     []string{"securitycontextconstraints"},
							ResourceNames: []string{"scc_none"},
							Verbs:         []string{"use"},
						},
					},
				},
			},
			clusterRoleBindings: []*rbacv1.ClusterRoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "clusterrolebinding",
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.GroupName,
						Kind:     "ClusterRole",
						Name:     "clusterrole",
					},
					Subjects: []rbacv1.Subject{
						{
							APIGroup:  corev1.GroupName,
							Kind:      "ServiceAccount",
							Name:      SA2Name,
							Namespace: NS2,
						},
					},
				},
			},
			want: sets.NewString(basicSCCs...).Insert("scc_none"),
		},
		{
			name:           "all SCCs assigned via a rolebinding to a broad role",
			serviceAccount: &NS2SA1,
			roles: []*rbacv1.Role{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "role",
						Namespace: NS2,
					},
					Rules: []rbacv1.PolicyRule{
						{
							APIGroups: []string{"*"},
							Resources: []string{"*"},
							Verbs:     []string{"*"},
						},
					},
				},
			},
			roleBindings: []*rbacv1.RoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "rolebinding",
						Namespace: NS2,
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.GroupName,
						Kind:     "Role",
						Name:     "role",
					},
					Subjects: []rbacv1.Subject{
						{
							APIGroup:  corev1.GroupName,
							Kind:      "ServiceAccount",
							Name:      SA1Name,
							Namespace: NS2,
						},
					},
				},
			},
			want: sets.NewString(allSCCs...),
		},
		{
			name:           "specific SCC assigned via a rolebinding to a clusterrole",
			serviceAccount: &NS2SA2,
			clusterRoles: []*rbacv1.ClusterRole{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "role",
					},
					Rules: []rbacv1.PolicyRule{
						{
							APIGroups:     []string{securityv1.GroupName},
							Resources:     []string{"securitycontextconstraints"},
							ResourceNames: []string{"scc_none"},
							Verbs:         []string{"use"},
						},
					},
				},
			},
			roleBindings: []*rbacv1.RoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "rolebinding",
						Namespace: NS2,
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.GroupName,
						Kind:     "ClusterRole",
						Name:     "role",
					},
					Subjects: []rbacv1.Subject{
						{
							APIGroup:  corev1.GroupName,
							Kind:      "ServiceAccount",
							Name:      SA2Name,
							Namespace: NS2,
						},
					},
				},
			},
			want: sets.NewString(basicSCCs...).Insert("scc_none"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roles := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
			for _, r := range tt.roles {
				require.NoError(t, roles.Add(r))
			}

			clusterRoles := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
			for _, cr := range tt.clusterRoles {
				require.NoError(t, clusterRoles.Add(cr))
			}

			roleBindings := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{BySAIndexName: BySAIndexKeys})
			for _, rb := range tt.roleBindings {
				require.NoError(t, roleBindings.Add(rb))
			}

			clusterRoleBindings := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{BySAIndexName: BySAIndexKeys})
			for _, crb := range tt.clusterRoleBindings {
				require.NoError(t, clusterRoleBindings.Add(crb))
			}

			roleLister := rbacv1listers.NewRoleLister(roles)
			clusterRoleLister := rbacv1listers.NewClusterRoleLister(clusterRoles)

			c := &saToSCCCache{
				roleLister:                roleLister,
				clusterRoleLister:         clusterRoleLister,
				roleBindingIndexer:        roleBindings,
				clusterRoleBindingIndexer: clusterRoleBindings,
				sccLister:                 sccLister(t),

				usefulRoles: make(map[string]sets.String),
			}

			if tt.sccsOverrides != nil {
				sccs := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
				for _, scc := range tt.sccsOverrides {
					require.NoError(t, sccs.Add(scc))
				}
				c.sccLister = securityv1listers.NewSecurityContextConstraintsLister(sccs)
			}

			// need to init the cache first
			testSCCs, err := c.sccLister.List(labels.Everything())
			require.NoError(t, err)
			for _, scc := range testSCCs {
				c.handleSCCAdded(scc)
			}

			got, err := c.SCCsFor(tt.serviceAccount)
			if (err != nil) != tt.wantErr {
				t.Errorf("SCCRoleCache.SCCsFor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.want.Equal(got) {
				t.Errorf("SCCRoleCache.SCCsFor() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_saToSCCCache_getRoleFromRoleRef(t *testing.T) {
	roleRef := func(name string) rbacv1.RoleRef {
		return rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     name,
		}
	}
	clusterRoleRef := func(name string) rbacv1.RoleRef {
		return rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     name,
		}
	}

	testRoles := []*rbacv1.Role{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testrole",
				Namespace: "testns",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testrole2",
				Namespace: "testns2",
			},
		},
	}

	testClusterRoles := []*rbacv1.ClusterRole{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "clusterrole",
			},
		},
	}

	roleObjOrDie := func(obj interface{}) RoleInterface {
		roleObj, err := NewRoleObj(obj)
		require.NoError(t, err)
		return roleObj
	}

	roleCache := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	for _, r := range testRoles {
		require.NoError(t, roleCache.Add(r))
	}
	roleLister := rbacv1listers.NewRoleLister(roleCache)

	clusterRoleCache := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	for _, r := range testClusterRoles {
		require.NoError(t, clusterRoleCache.Add(r))
	}
	clusterRoleLister := rbacv1listers.NewClusterRoleLister(clusterRoleCache)

	tests := []struct {
		name      string
		namespace string
		roleRef   rbacv1.RoleRef
		want      RoleInterface
		wantErr   bool
	}{
		{
			name:      "empty role ref",
			namespace: "testns",
			roleRef:   rbacv1.RoleRef{},
			wantErr:   true,
		},
		{
			name:      "role ref to a non-existent role",
			namespace: "testns",
			roleRef:   roleRef("nonexistent"),
			wantErr:   true,
		},
		{
			name:      "role ref to a role in a diferent NS",
			namespace: "testns",
			roleRef:   roleRef("testrole2"),
			wantErr:   true,
		},
		{
			name:      "role ref to an existing role",
			namespace: "testns",
			roleRef:   roleRef("testrole"),
			want:      roleObjOrDie(testRoles[0]),
		},
		{
			name:    "role ref to a non-existent cluster role",
			roleRef: clusterRoleRef("randomrole"),
			wantErr: true,
		},
		{
			name:    "role ref to a cluster role",
			roleRef: clusterRoleRef("clusterrole"),
			want:    roleObjOrDie(testClusterRoles[0]),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			c := &saToSCCCache{
				roleLister:        roleLister,
				clusterRoleLister: clusterRoleLister,
			}
			got, err := c.getRoleFromRoleRef(tt.namespace, tt.roleRef)
			if (err != nil) != tt.wantErr {
				t.Errorf("saToSCCCache.getRoleFromRoleRef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("saToSCCCache.getRoleFromRoleRef() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_saToSCCCache_IsRoleBindingRelevant(t *testing.T) {
	testUsefulRoles := map[string]sets.String{
		"ns/testrole":      sets.NewString("a, b"),
		"/testclusterrole": sets.NewString("c"),
	}

	bToRole := func(ns, name string) *rbacv1.RoleBinding {
		return &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rbname",
				Namespace: ns,
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "Role",
				Name:     name,
			},
		}
	}

	bToCRole := func(ns, name string) *rbacv1.RoleBinding {
		return &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rbname",
				Namespace: ns,
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "ClusterRole",
				Name:     name,
			},
		}
	}

	cbToCRole := func(name string) *rbacv1.ClusterRoleBinding {
		return &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: "crbname",
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "ClusterRole",
				Name:     name,
			},
		}
	}

	roleCache := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	require.NoError(t, roleCache.Add(&rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: "testrole", Namespace: "ns"}}))
	require.NoError(t, roleCache.Add(&rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: "uncached", Namespace: "ns"}}))
	roleLister := rbacv1listers.NewRoleLister(roleCache)

	clusterRoleCache := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	require.NoError(t, clusterRoleCache.Add(&rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "testclusterrole"}}))
	require.NoError(t, clusterRoleCache.Add(&rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "uncachedclusterrole"}}))
	clusterRoleLister := rbacv1listers.NewClusterRoleLister(clusterRoleCache)

	tests := []struct {
		name     string
		bindings interface{}
		want     bool
	}{

		{
			name:     "rolebinding to nonexistent role",
			bindings: bToRole("ns", "nonexistent"),
			want:     false,
		},
		{
			name:     "rolebinding to an uncached role",
			bindings: bToRole("ns", "uncached"),
			want:     false,
		},
		{
			name:     "rolebinding to cached role",
			bindings: bToRole("ns", "testrole"),
			want:     true,
		},
		{
			name:     "rolebinding to a nonexistent cluster role",
			bindings: bToCRole("ns", "nonexistentclusterrole"),
			want:     false,
		},
		{
			name:     "rolebinding to an uncached cluster role",
			bindings: bToCRole("ns", "uncachedclusterrole"),
			want:     false,
		},
		{
			name:     "rolebinding to a cached cluster role",
			bindings: bToCRole("ns", "testclusterrole"),
			want:     true,
		},
		{
			name:     "clusterrolebinding to a nonexistent cluster role",
			bindings: cbToCRole("testrole"),
			want:     false,
		},
		{
			name:     "clusterrolebinding to an uncached cluster role",
			bindings: cbToCRole("uncachedclusterrole"),
			want:     false,
		},
		{
			name:     "clusterrolebinding to a cached cluster role",
			bindings: cbToCRole("testclusterrole"),
			want:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &saToSCCCache{
				roleLister:                roleLister,
				clusterRoleLister:         clusterRoleLister,
				roleBindingIndexer:        nil, // should not be necessary
				clusterRoleBindingIndexer: nil, // should not be necessary
				usefulRoles:               testUsefulRoles,
			}
			if got := c.IsRoleBindingRelevant(tt.bindings); got != tt.want {
				t.Errorf("saToSCCCache.IsRoleBindingRelevant() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_saToSCCCache_handleRoleModified(t *testing.T) {
	const (
		testNSName              = "testns"
		testRoleCacheKey        = testNSName + "/testrole"
		testClusterRoleCacheKey = "/testclusterrole"
		allNSConstant           = "ALL_NAMESPACES"
	)

	createRole := func(objKind interface{}, rules []rbacv1.PolicyRule) interface{} {
		switch objKind.(type) {
		case *rbacv1.Role:
			return &rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testrole",
					Namespace: testNSName,
				},
				Rules: rules,
			}
		case *rbacv1.ClusterRole:
			return &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testclusterrole",
				},
				Rules: rules,
			}
		}

		return nil
	}

	tests := []struct {
		name                string
		priorUsefulRoles    map[string]sets.String
		expectedUsefulRoles map[string]sets.String
		roleObj             interface{}
		expectedQueueKey    string
	}{
		{
			name: "role without any SCC rules",
			roleObj: createRole(
				&rbacv1.Role{},
				[]rbacv1.PolicyRule{{APIGroups: []string{""}, Verbs: []string{"get", "update"}, Resources: []string{"pods"}}},
			),
			expectedUsefulRoles: map[string]sets.String{"testns/testrole": {}},
		},
		{
			name: "role with irrelevant SCC rules",
			roleObj: createRole(
				&rbacv1.Role{},
				[]rbacv1.PolicyRule{{APIGroups: []string{securityv1.GroupName}, Verbs: []string{"get", "update"}, Resources: []string{"securitycontextconstraints"}}},
			),
			expectedUsefulRoles: map[string]sets.String{"testns/testrole": {}},
		},
		{
			name: "role with a relevant rule to an unknown SCC",
			roleObj: createRole(
				&rbacv1.Role{},
				[]rbacv1.PolicyRule{{APIGroups: []string{securityv1.GroupName}, Verbs: []string{"use"}, Resources: []string{"securitycontextconstraints"}, ResourceNames: []string{"unknown"}}},
			),
			expectedUsefulRoles: map[string]sets.String{"testns/testrole": {}},
		},
		{
			name: "role with a relevant rule to a real SCC",
			roleObj: createRole(
				&rbacv1.Role{},
				[]rbacv1.PolicyRule{{APIGroups: []string{securityv1.GroupName}, Verbs: []string{"use"}, Resources: []string{"securitycontextconstraints"}, ResourceNames: []string{"scc_none"}}},
			),
			expectedUsefulRoles: map[string]sets.String{
				testRoleCacheKey: sets.NewString("scc_none"),
			},
			expectedQueueKey: testNSName,
		},
		{
			name: "role with a broad rule to all SCCs",
			roleObj: createRole(
				&rbacv1.Role{},
				[]rbacv1.PolicyRule{{APIGroups: []string{securityv1.GroupName}, Verbs: []string{"use"}, Resources: []string{"securitycontextconstraints"}}},
			),
			expectedUsefulRoles: map[string]sets.String{
				testRoleCacheKey: sets.NewString(allSCCs...),
			},
			expectedQueueKey: testNSName,
		},
		{
			name: "clusterrole without any SCC rules",
			roleObj: createRole(
				&rbacv1.ClusterRole{},
				[]rbacv1.PolicyRule{{APIGroups: []string{""}, Verbs: []string{"get", "update"}, Resources: []string{"pods"}}},
			),
			expectedUsefulRoles: map[string]sets.String{testClusterRoleCacheKey: {}},
		},
		{
			name: "clusterrole with irrelevant SCC rules",
			roleObj: createRole(
				&rbacv1.ClusterRole{},
				[]rbacv1.PolicyRule{{APIGroups: []string{securityv1.GroupName}, Verbs: []string{"get", "update"}, Resources: []string{"securitycontextconstraints"}}},
			),
			expectedUsefulRoles: map[string]sets.String{testClusterRoleCacheKey: {}},
		},
		{
			name: "clusterrole with a relevant rule to an unknown SCC",
			roleObj: createRole(
				&rbacv1.ClusterRole{},
				[]rbacv1.PolicyRule{{APIGroups: []string{securityv1.GroupName}, Verbs: []string{"use"}, Resources: []string{"securitycontextconstraints"}, ResourceNames: []string{"unknown"}}},
			),
			expectedUsefulRoles: map[string]sets.String{testClusterRoleCacheKey: {}},
		},
		{
			name: "clusterrole with a relevant rule to a real SCC",
			roleObj: createRole(
				&rbacv1.ClusterRole{},
				[]rbacv1.PolicyRule{{APIGroups: []string{securityv1.GroupName}, Verbs: []string{"use"}, Resources: []string{"securitycontextconstraints"}, ResourceNames: []string{"scc_none", "scc_none2"}}},
			),
			expectedUsefulRoles: map[string]sets.String{
				testClusterRoleCacheKey: sets.NewString("scc_none", "scc_none2"),
			},
			expectedQueueKey: allNSConstant,
		},
		{
			name: "clusterrole with a broad rule to all SCCs",
			roleObj: createRole(
				&rbacv1.ClusterRole{},
				[]rbacv1.PolicyRule{{APIGroups: []string{securityv1.GroupName}, Verbs: []string{"use"}, Resources: []string{"securitycontextconstraints"}}},
			),
			expectedUsefulRoles: map[string]sets.String{
				testClusterRoleCacheKey: sets.NewString(allSCCs...),
			},
			expectedQueueKey: allNSConstant,
		},
		{
			name: "role with a relevant rule to a real SCC modified to not contain a relevant SCC rule",
			roleObj: createRole(
				&rbacv1.Role{},
				[]rbacv1.PolicyRule{{APIGroups: []string{securityv1.GroupName}, Verbs: []string{"get", "list", "watch"}, Resources: []string{"securitycontextconstraints"}, ResourceNames: []string{"scc_none"}}},
			),
			priorUsefulRoles: map[string]sets.String{
				testRoleCacheKey: sets.NewString("scc_none"),
			},
			expectedUsefulRoles: map[string]sets.String{testRoleCacheKey: {}},
			expectedQueueKey:    testNSName,
		},
		{
			name: "clusterrole with a relevant rule to a real SCC modified to not contain a relevant SCC rule",
			roleObj: createRole(
				&rbacv1.ClusterRole{},
				[]rbacv1.PolicyRule{{APIGroups: []string{securityv1.GroupName}, Verbs: []string{"do-magic"}, Resources: []string{"securitycontextconstraints"}, ResourceNames: []string{"scc_none"}}},
			),
			priorUsefulRoles: map[string]sets.String{
				testClusterRoleCacheKey: sets.NewString("scc_none"),
			},
			expectedUsefulRoles: map[string]sets.String{testClusterRoleCacheKey: {}},
			expectedQueueKey:    allNSConstant,
		},
	}

	for _, tt := range tests {
		var itemAdded string
		enqueueFunc := func(obj interface{}) {
			switch o := obj.(type) {
			case metav1.ObjectMetaAccessor:
				itemAdded = o.GetObjectMeta().GetNamespace()
			default:
				t.Errorf("what is this? %T", obj)
			}

			if len(itemAdded) == 0 {
				itemAdded = allNSConstant
			}
		}

		roleCache := tt.priorUsefulRoles
		if roleCache == nil {
			roleCache = make(map[string]sets.String)
		}
		t.Run(tt.name, func(t *testing.T) {
			c := &saToSCCCache{
				sccLister:                sccLister(t),
				usefulRolesLock:          sync.Mutex{},
				usefulRoles:              roleCache,
				externalQueueEnqueueFunc: enqueueFunc,
			}
			c.handleRoleModified(tt.roleObj)

			require.Equal(t, tt.expectedQueueKey, itemAdded)

			if !reflect.DeepEqual(roleCache, tt.expectedUsefulRoles) {
				t.Errorf("the expected cache is different from the one retrieved: %s", cmp.Diff(tt.expectedUsefulRoles, c.usefulRoles))
			}
		})
	}
}

func Test_saToSCCCache_handleRoleRemoved(t *testing.T) {
	const (
		testNSName              = "testns"
		testRoleName            = "testrole"
		testRoleCacheKey        = testNSName + "/" + testRoleName
		testClusterRoleName     = "testclusterrole"
		testClusterRoleCacheKey = "/testclusterrole"
		allNSConstant           = "ALL_NAMESPACES"
	)

	createRole := func(objKind interface{}, name string) interface{} {
		switch objKind.(type) {
		case *rbacv1.Role:
			return &rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: testNSName,
				},
				Rules: []rbacv1.PolicyRule{{APIGroups: []string{securityv1.GroupName}, Resources: []string{"securitycontextconstraints"}, Verbs: []string{"use"}}},
			}
		case *rbacv1.ClusterRole:
			return &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Rules: []rbacv1.PolicyRule{{APIGroups: []string{securityv1.GroupName}, Resources: []string{"securitycontextconstraints"}, Verbs: []string{"use"}}},
			}
		}

		return nil
	}

	priorUsefulRoles := map[string]sets.String{
		testRoleCacheKey:        sets.NewString("great", "content"),
		testClusterRoleCacheKey: sets.NewString("this", "is", "pretty", "good", "too"),
	}

	tests := []struct {
		name                string
		expectedUsefulRoles map[string]sets.String
		roleObj             interface{}
		expectedQueueKey    string
	}{
		{
			name:                "cached role removed",
			roleObj:             createRole(&rbacv1.Role{}, testRoleName),
			expectedUsefulRoles: copyCacheMap(priorUsefulRoles, testRoleCacheKey),
			expectedQueueKey:    testNSName,
		},
		{
			name:                "cached clusterrole removed",
			roleObj:             createRole(&rbacv1.ClusterRole{}, testClusterRoleName),
			expectedUsefulRoles: copyCacheMap(priorUsefulRoles, testClusterRoleCacheKey),
			expectedQueueKey:    allNSConstant,
		},
		{
			name:                "uncached role removed",
			roleObj:             createRole(&rbacv1.Role{}, "uncached"),
			expectedUsefulRoles: copyCacheMap(priorUsefulRoles),
		},
		{
			name:                "uncached clusterrole removed",
			roleObj:             createRole(&rbacv1.ClusterRole{}, "uncached"),
			expectedUsefulRoles: copyCacheMap(priorUsefulRoles),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var itemAdded string
			enqueueFunc := func(obj interface{}) {
				switch o := obj.(type) {
				case metav1.ObjectMetaAccessor:
					itemAdded = o.GetObjectMeta().GetNamespace()
				default:
					t.Errorf("what is this? %T", obj)
				}

				if len(itemAdded) == 0 {
					itemAdded = allNSConstant
				}
			}

			c := &saToSCCCache{
				usefulRoles:              copyCacheMap(priorUsefulRoles),
				externalQueueEnqueueFunc: enqueueFunc,
			}
			c.handleRoleRemoved(tt.roleObj)

			require.Equal(t, tt.expectedQueueKey, itemAdded)

			if !reflect.DeepEqual(c.usefulRoles, tt.expectedUsefulRoles) {
				t.Errorf("the expected cache is different from the one retrieved: %s", cmp.Diff(tt.expectedUsefulRoles, c.usefulRoles))
			}

		})
	}
}

func Test_saToSCCCache_handleSCCDeleted(t *testing.T) {
	const (
		testNSName                      = "testns"
		testRoleNameHasSCC              = "testrole"
		testRoleNameNoHazSCC            = "testrole-nohaz"
		testRoleHasSCCCacheKey          = testNSName + "/" + testRoleNameHasSCC
		testRoleNoHazSCCCacheKey        = testNSName + "/" + testRoleNameNoHazSCC
		testClusterRoleHasSCCName       = "testclusterrole"
		testClusterRoleNoHazSCCName     = "testclusterrole-nohaz"
		testClusterRoleHasSCCCacheKey   = "/" + testClusterRoleHasSCCName
		testClusterRoleNoHazSCCCacheKey = "/" + testClusterRoleNoHazSCCName
		allNSConstant                   = "ALL_NAMESPACES"
		testSCCName                     = "testscc"
	)

	hasSCCNames := sets.NewString("scc_none", testSCCName, "scc_none2")
	noHazSCCNames := sets.NewString("scc_one", "scc_two")
	cHasSCCNames := sets.NewString(testSCCName)
	cNoHazSCCNames := sets.NewString("scc_three")
	priorUsefulRoles := map[string]sets.String{
		testRoleHasSCCCacheKey:        hasSCCNames,
		testRoleNameNoHazSCC:          noHazSCCNames,
		testClusterRoleHasSCCCacheKey: cHasSCCNames,
		testClusterRoleNoHazSCCName:   cNoHazSCCNames,
	}

	roles := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	require.NoError(t, roles.Add(&rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: testRoleNameHasSCC, Namespace: testNSName},
		Rules:      []rbacv1.PolicyRule{{APIGroups: []string{securityv1.GroupName}, Resources: []string{"securitycontextconstraints"}, Verbs: []string{"use"}, ResourceNames: hasSCCNames.UnsortedList()}},
	}))
	require.NoError(t, roles.Add(&rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: testRoleNameNoHazSCC, Namespace: testNSName},
		Rules:      []rbacv1.PolicyRule{{APIGroups: []string{securityv1.GroupName}, Resources: []string{"securitycontextconstraints"}, Verbs: []string{"use"}, ResourceNames: noHazSCCNames.UnsortedList()}},
	}))
	require.NoError(t, roles.Add(&rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: "idling-unrelated-role", Namespace: testNSName},
		Rules:      []rbacv1.PolicyRule{{APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"create"}}},
	}))
	clusterRoles := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	require.NoError(t, clusterRoles.Add(&rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: testClusterRoleHasSCCName},
		Rules:      []rbacv1.PolicyRule{{APIGroups: []string{securityv1.GroupName}, Resources: []string{"securitycontextconstraints"}, Verbs: []string{"use"}, ResourceNames: cHasSCCNames.UnsortedList()}},
	}))
	require.NoError(t, clusterRoles.Add(&rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: testClusterRoleNoHazSCCName},
		Rules:      []rbacv1.PolicyRule{{APIGroups: []string{securityv1.GroupName}, Resources: []string{"securitycontextconstraints"}, Verbs: []string{"use"}, ResourceNames: cNoHazSCCNames.UnsortedList()}},
	}))
	require.NoError(t, clusterRoles.Add(&rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: "idling-unrelated-clusterrole"},
		Rules:      []rbacv1.PolicyRule{{APIGroups: []string{""}, Resources: []string{"configmaps"}, Verbs: []string{"get", "list", "watch"}, ResourceNames: []string{"bunch", "of", "names"}}},
	}))

	tests := []struct {
		name                string
		scc                 *securityv1.SecurityContextConstraints
		expectedUsefulRoles map[string]sets.String
	}{
		{
			name:                "remove uncached SCC",
			scc:                 &securityv1.SecurityContextConstraints{ObjectMeta: metav1.ObjectMeta{Name: "uncached"}},
			expectedUsefulRoles: copyCacheMap(priorUsefulRoles),
		},
		{
			name: "remove cached SCC",
			scc:  &securityv1.SecurityContextConstraints{ObjectMeta: metav1.ObjectMeta{Name: testSCCName}},
			expectedUsefulRoles: map[string]sets.String{
				testRoleHasSCCCacheKey:        sets.NewString("scc_none", "scc_none2"),
				testRoleNameNoHazSCC:          sets.NewString("scc_one", "scc_two"),
				testClusterRoleHasSCCCacheKey: {},
				testClusterRoleNoHazSCCName:   sets.NewString("scc_three"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var itemAdded string
			enqueueFunc := func(obj interface{}) {
				switch o := obj.(type) {
				case metav1.ObjectMetaAccessor:
					itemAdded = o.GetObjectMeta().GetNamespace()
				default:
					t.Errorf("what is this? %T", obj)
				}

				if len(itemAdded) == 0 {
					itemAdded = allNSConstant
				}
			}

			c := &saToSCCCache{
				roleLister:               rbacv1listers.NewRoleLister(roles),
				clusterRoleLister:        rbacv1listers.NewClusterRoleLister(clusterRoles),
				usefulRoles:              copyCacheMap(priorUsefulRoles),
				externalQueueEnqueueFunc: enqueueFunc,
			}
			c.handleSCCDeleted(tt.scc)

			require.Equal(t, allNSConstant, itemAdded)

			if !reflect.DeepEqual(c.usefulRoles, tt.expectedUsefulRoles) {
				t.Errorf("the expected cache is different from the one retrieved: %s", cmp.Diff(tt.expectedUsefulRoles, c.usefulRoles))
			}
		})
	}
}

func TestSCCRoleCache_BySAIndexKeys(t *testing.T) {
	type subjectDef struct {
		name      string
		namespace string
	}

	newRBObj := func(rbNS string, subjects []subjectDef) interface{} {
		obj := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{Namespace: rbNS},
			Subjects:   []rbacv1.Subject{},
		}

		for _, subject := range subjects {
			obj.Subjects = append(obj.Subjects, rbacv1.Subject{
				Kind:      "ServiceAccount",
				Name:      subject.name,
				Namespace: subject.namespace,
			})
		}

		return obj
	}

	newCRBObj := func(subjects []subjectDef) interface{} {
		obj := &rbacv1.ClusterRoleBinding{}

		for _, subject := range subjects {
			obj.Subjects = append(obj.Subjects, rbacv1.Subject{
				Kind:      "ServiceAccount",
				Name:      subject.name,
				Namespace: subject.namespace,
			})
		}

		return obj
	}

	const (
		ns1 = "testns1"
		sa1 = "testsa1"
		sa2 = "testsa2"
	)

	tests := []struct {
		name        string
		roleBinding interface{}
		want        []string
	}{
		{
			name:        "role binding without subjects",
			roleBinding: newRBObj(ns1, nil),
			want:        []string{},
		},
		{
			name:        "role binding with subjects with namespaces",
			roleBinding: newRBObj(ns1, []subjectDef{{sa1, ns1}, {sa2, ns1}}),
			want: []string{
				fmt.Sprintf("system:serviceaccount:%s:%s", ns1, sa1),
				fmt.Sprintf("system:serviceaccount:%s:%s", ns1, sa2),
			},
		},
		{
			name:        "role binding with subjects without namespaces",
			roleBinding: newRBObj(ns1, []subjectDef{{sa1, ""}, {sa2, ""}}),
			want: []string{
				fmt.Sprintf("system:serviceaccount:%s:%s", ns1, sa1),
				fmt.Sprintf("system:serviceaccount:%s:%s", ns1, sa2),
			},
		},
		{
			name:        "cluster role binding with subjects with namespaces",
			roleBinding: newCRBObj([]subjectDef{{sa1, ns1}, {sa2, ns1}}),
			want: []string{
				fmt.Sprintf("system:serviceaccount:%s:%s", ns1, sa1),
				fmt.Sprintf("system:serviceaccount:%s:%s", ns1, sa2),
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			subjects, err := BySAIndexKeys(testCase.roleBinding)
			require.Nil(t, err)
			require.Equal(t, testCase.want, subjects)
		})
	}
}

func copyCacheMap(m map[string]sets.String, omit ...string) map[string]sets.String {
	omitSet := sets.NewString(omit...)
	newMap := make(map[string]sets.String)
	for k, v := range m {
		if !omitSet.Has(k) {
			newMap[k] = sets.NewString(v.UnsortedList()...)
		}
	}
	return newMap
}
