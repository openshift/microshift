package psalabelsyncer

import (
	"fmt"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/authentication/serviceaccount"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	rbacv1informers "k8s.io/client-go/informers/rbac/v1"
	rbacv1listers "k8s.io/client-go/listers/rbac/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/plugin/pkg/auth/authorizer/rbac"

	securityv1 "github.com/openshift/api/security/v1"
	securityv1informers "github.com/openshift/client-go/security/informers/externalversions/security/v1"
	securityv1listers "github.com/openshift/client-go/security/listers/security/v1"
)

// The index name to be used along with the BySAIndexKeys indexing function
const BySAIndexName = "ByServiceAccount"

// saToSCCCache is a construct that caches roles and rolebindings
// (and their cluster variants) and based on that and on SCCs present in the cluster
// it allows retrieving a set of SCCs for a given ServiceAccount
type saToSCCCache struct {
	roleLister                rbacv1listers.RoleLister
	clusterRoleLister         rbacv1listers.ClusterRoleLister
	roleBindingIndexer        cache.Indexer
	clusterRoleBindingIndexer cache.Indexer

	sccLister securityv1listers.SecurityContextConstraintsLister

	usefulRolesLock sync.Mutex
	usefulRoles     map[string]sets.String

	externalQueueEnqueueFunc func(interface{})
}

type SAToSCCCache interface {
	SCCsFor(serviceAccount *corev1.ServiceAccount) (sets.String, error)
	IsRoleBindingRelevant(obj interface{}) bool

	AddEventHandlers(
		rbacv1informers rbacv1informers.Interface,
		sccInformer securityv1informers.SecurityContextConstraintsInformer,
	)
	WithExternalQueueEnqueue(enqueueFunc func(interface{})) SAToSCCCache
}

type roleBindingInterface interface {
	Namespace() string
	RoleRef() rbacv1.RoleRef
	Subjects() []rbacv1.Subject
}

// RoleInterface is an interface for generic access to role-like object, such
// as rbac.Role and rbac.ClusterRole
type RoleInterface interface {
	metav1.ObjectMetaAccessor
	Name() string
	Namespace() string
	Rules() []rbacv1.PolicyRule
}

func usefulRolesKey(r RoleInterface) string {
	return fmt.Sprintf("%s/%s", r.Namespace(), r.Name())
}

// role and clusterrolebinding object for generic handling, assumes one and
// at most one of role/clusterrole is always non-nil
type roleBindingObj struct {
	roleBinding        *rbacv1.RoleBinding
	clusterRoleBinding *rbacv1.ClusterRoleBinding
}

func newRoleBindingObj(obj interface{}) (roleBindingInterface, error) {
	if o, ok := obj.(*rbacv1.ClusterRoleBinding); ok {
		return &roleBindingObj{
			clusterRoleBinding: o,
		}, nil
	} else if o, ok := obj.(*rbacv1.RoleBinding); ok {
		return &roleBindingObj{
			roleBinding: o,
		}, nil
	}

	return nil, fmt.Errorf("the object is neither a RoleBinding, nor a ClusterRoleBinding: %T", obj)
}

func (r *roleBindingObj) RoleRef() rbacv1.RoleRef {
	if binding := r.clusterRoleBinding; binding != nil {
		return binding.RoleRef
	}
	return r.roleBinding.RoleRef
}

func (r *roleBindingObj) Subjects() []rbacv1.Subject {
	if binding := r.clusterRoleBinding; binding != nil {
		return binding.Subjects
	}
	return r.roleBinding.Subjects
}

func (r *roleBindingObj) Namespace() string {
	if r.clusterRoleBinding != nil {
		return ""
	}
	return r.roleBinding.Namespace
}

// roleObj helps to handle roles and clusterroles in a generic manner
type roleObj struct {
	role        *rbacv1.Role
	clusterRole *rbacv1.ClusterRole
}

// NewRoleObj expects either a Role or a ClusterRole as its `obj` input argument,
// it returns an object that allows generic access to the role-like object
func NewRoleObj(obj interface{}) (RoleInterface, error) {
	switch r := obj.(type) {
	case *rbacv1.ClusterRole:
		return &roleObj{
			clusterRole: r,
		}, nil
	case *rbacv1.Role:
		return &roleObj{
			role: r,
		}, nil
	case *roleObj:
		return r, nil
	default:
		return nil, fmt.Errorf("the object is neither a Role, nor a ClusterRole: %T", obj)
	}
}

func (r *roleObj) Rules() []rbacv1.PolicyRule {
	if role := r.clusterRole; role != nil {
		return role.Rules
	}
	return r.role.Rules
}

func (r *roleObj) Name() string {
	if role := r.clusterRole; role != nil {
		return role.Name
	}
	return r.role.Name
}

func (r *roleObj) Namespace() string {
	if role := r.clusterRole; role != nil {
		return role.Namespace
	}
	return r.role.Namespace
}

func (r *roleObj) GetObjectMeta() metav1.Object {
	if role := r.clusterRole; role != nil {
		return role.GetObjectMeta()
	}
	return r.role.GetObjectMeta()
}

// BySAIndexKeys is a cache.IndexFunc indexing function that shall be used on
// rolebinding and clusterrolebinding informer caches.
// It retrieves the subjects of the incoming object and if there are SA, SA groups
// or the system:authenticated group subjects, these will all be returned as a slice
// of strings to create an index for the SA or SA group.
func BySAIndexKeys(obj interface{}) ([]string, error) {
	roleBinding, err := newRoleBindingObj(obj)
	if err != nil {
		return nil, err
	}

	serviceAccounts := []string{}
	for _, subject := range roleBinding.Subjects() {
		switch {
		case subject.APIGroup == "" && subject.Kind == "ServiceAccount":
			serviceAccounts = append(serviceAccounts, serviceaccount.MakeUsername(subject.Namespace, subject.Name))
		case subject.APIGroup == rbacv1.GroupName && subject.Kind == "Group":
			if subject.Name == serviceaccount.AllServiceAccountsGroup ||
				subject.Name == user.AllAuthenticated ||
				strings.HasPrefix(subject.Name, serviceaccount.ServiceAccountGroupPrefix) {
				serviceAccounts = append(serviceAccounts, subject.Name)
			}
		}
	}

	return serviceAccounts, nil
}

func NewSAToSCCCache(
	rbacInformers rbacv1informers.Interface,
	sccInfomer securityv1informers.SecurityContextConstraintsInformer,
) SAToSCCCache {
	c := &saToSCCCache{
		roleLister:                rbacInformers.Roles().Lister(),
		clusterRoleLister:         rbacInformers.ClusterRoles().Lister(),
		roleBindingIndexer:        rbacInformers.RoleBindings().Informer().GetIndexer(),
		clusterRoleBindingIndexer: rbacInformers.ClusterRoleBindings().Informer().GetIndexer(),

		sccLister: sccInfomer.Lister(),

		usefulRoles: make(map[string]sets.String),
	}
	c.AddEventHandlers(rbacInformers, sccInfomer)

	return c
}

func (c *saToSCCCache) AddEventHandlers(
	rbacv1informers rbacv1informers.Interface,
	sccInformer securityv1informers.SecurityContextConstraintsInformer,
) {

	roleHandlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleRoleModified,
		UpdateFunc: func(_, new interface{}) { c.handleRoleModified(new) },
		DeleteFunc: c.handleRoleRemoved,
	}
	rbacv1informers.Roles().Informer().AddEventHandler(roleHandlerFuncs)
	rbacv1informers.ClusterRoles().Informer().AddEventHandler(roleHandlerFuncs)

	sccInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.handleSCCAdded,
		// we don't care about SCCs being updated, SCCs get introspected in the SCCsFor directly
		DeleteFunc: c.handleSCCDeleted,
	})

	sccInformer.Informer().HasSynced()
}

func (c *saToSCCCache) WithExternalQueueEnqueue(enqueueFunc func(interface{})) SAToSCCCache {
	c.externalQueueEnqueueFunc = enqueueFunc
	return c
}

func (c *saToSCCCache) addToExternalJobQueue(obj interface{}) {
	if c.externalQueueEnqueueFunc != nil {
		c.externalQueueEnqueueFunc(obj)
	}
}

func (c *saToSCCCache) handleRoleModified(obj interface{}) {
	sccs, err := c.sccLister.List(labels.Everything())
	if err != nil {
		klog.Warning("failed to list SCCs: %v", err)
		return
	}

	role, err := NewRoleObj(obj)
	if err != nil {
		klog.Warningf("unexpected error, this may be a bug: %v", err)
		return
	}

	c.usefulRolesLock.Lock()
	defer c.usefulRolesLock.Unlock()
	currentRoleSCCs := c.usefulRoles[usefulRolesKey(role)]
	c.syncRoleCache(role, sccs)
	updatedRoleSCCs := c.usefulRoles[usefulRolesKey(role)]

	// must be order-independent comparisons
	if !currentRoleSCCs.Equal(updatedRoleSCCs) {
		c.addToExternalJobQueue(role)
	}
}

func (c *saToSCCCache) handleRoleRemoved(obj interface{}) {
	role, err := NewRoleObj(obj)
	if err != nil {
		klog.Warningf("unexpected error, this may be a bug: %v", err)
		return
	}

	c.usefulRolesLock.Lock()
	defer c.usefulRolesLock.Unlock()

	if _, roleExists := c.usefulRoles[usefulRolesKey(role)]; roleExists {
		delete(c.usefulRoles, usefulRolesKey(role))

		c.addToExternalJobQueue(obj)
	}
}

func (c *saToSCCCache) handleSCCAdded(obj interface{}) {
	scc, ok := obj.(*securityv1.SecurityContextConstraints)
	if !ok {
		klog.Errorf("expected object to be of type SecurityContextConstraints but it is %T", obj)
		return
	}

	roles, err := c.roleLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("failed to handle %q SCC addition: %v", scc.Name, err)
		return
	}

	clusterRoles, err := c.clusterRoleLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("failed to handle %q SCC addition: %v", scc.Name, err)
		return
	}

	c.usefulRolesLock.Lock()
	defer c.usefulRolesLock.Unlock()

	for _, r := range roles {
		role, err := NewRoleObj(r)
		if err != nil {
			panic(err)
		}
		c.syncRoleCacheSingleSCCAdded(role, scc)
	}

	for _, r := range clusterRoles {
		role, err := NewRoleObj(r)
		if err != nil {
			panic(err)
		}
		c.syncRoleCacheSingleSCCAdded(role, scc)
	}

	c.addToExternalJobQueue(obj)
}

func (c *saToSCCCache) handleSCCDeleted(obj interface{}) {
	scc, ok := obj.(*securityv1.SecurityContextConstraints)
	if !ok {
		klog.Error("expected obj to be of type SecurityContextConstraints, got %T", obj)
		return
	}

	c.usefulRolesLock.Lock()
	defer c.usefulRolesLock.Unlock()

	sccName := scc.Name
	for role := range c.usefulRoles {
		c.usefulRoles[role].Delete(sccName)
	}

	c.addToExternalJobQueue(obj)
}

// SCCsFor returns a slice of all the SCCs that a given service account can use
// to run pods in its namespace.
// It expects the serviceAccount name in the system:serviceaccount:<ns>:<name> form.
func (c *saToSCCCache) SCCsFor(serviceAccount *corev1.ServiceAccount) (sets.String, error) {
	saUserInfo := serviceaccount.UserInfo(
		serviceAccount.Namespace,
		serviceAccount.Name,
		string(serviceAccount.UID),
	)

	// realSAUserInfo adds the "system:authenticated" group to SA UserInfo groups
	// because that's the group all authenticated entities have, including SAs
	realSAUserInfo := &user.DefaultInfo{
		Name:   saUserInfo.GetName(),
		Groups: append(saUserInfo.GetGroups(), user.AllAuthenticated),
		UID:    saUserInfo.GetUID(),
		Extra:  saUserInfo.GetExtra(),
	}

	bindingObjs, err := getIndexedRolebindings(c.roleBindingIndexer, realSAUserInfo)
	if err != nil {
		return nil, fmt.Errorf("failed retrieve rolebindings: %w", err)
	}

	clusterBindingObjs, err := getIndexedRolebindings(c.clusterRoleBindingIndexer, realSAUserInfo)
	if err != nil {
		return nil, fmt.Errorf("failed retrieve clusterrolebindings: %w", err)
	}
	objs := append(bindingObjs, clusterBindingObjs...)

	sccs, err := c.sccLister.List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("couldn't list cluster SCCs: %w", err)
	}

	allowedSCCs := sets.NewString()
	for _, scc := range sccs {
		if sccAllowsSA(scc, realSAUserInfo) {
			allowedSCCs.Insert(scc.Name)
		}
	}

	for _, o := range objs {
		rb, err := newRoleBindingObj(o)
		if err != nil {
			// this would be rather weird
			return nil, fmt.Errorf("failed to create internal rolebinding representation: %v", err)
		}

		roleCachedKey := fmt.Sprintf("/%s", rb.RoleRef().Name)
		if rb.RoleRef().Kind == "Role" {
			roleCachedKey = rb.Namespace() + roleCachedKey
		}

		c.usefulRolesLock.Lock()
		// this role does not have SCC-related rules
		cachedAllowedSCCs := c.usefulRoles[roleCachedKey]
		c.usefulRolesLock.Unlock()
		if len(cachedAllowedSCCs) == 0 {
			continue
		}

		// we particularly care only about Roles in the SA NS
		if roleRef := rb.RoleRef(); roleBindingAppliesToNS(rb, serviceAccount.Namespace) && roleRef.APIGroup == rbacv1.GroupName {
			allowedSCCs = allowedSCCs.Union(cachedAllowedSCCs)
		}
	}

	return allowedSCCs, nil
}

// getRoleFromRoleRef tries to retrieve the role or clusterrole from roleRef.
func (c *saToSCCCache) getRoleFromRoleRef(ns string, roleRef rbacv1.RoleRef) (RoleInterface, error) {
	var err error
	var role interface{}
	switch kind := roleRef.Kind; kind {
	case "Role":
		role, err = c.roleLister.Roles(ns).Get(roleRef.Name)
		if err != nil {
			return nil, fmt.Errorf("couldn't retrieve role from role ref: %w", err)
		}

	case "ClusterRole":
		role, err = c.clusterRoleLister.Get(roleRef.Name)
		if err != nil {
			return nil, fmt.Errorf("couldn't retrieve clusterrole from role ref: %w", err)
		}

	default:
		return nil, fmt.Errorf("unknown kind in roleRef: %s", kind)
	}

	return NewRoleObj(role)
}

// IsRoleBindingRelevant returns true if the cluster/rolebinding supplied binds
// to a Role that provides access to at least one of the SCCs available in the
// cluster.
func (c *saToSCCCache) IsRoleBindingRelevant(obj interface{}) bool {
	rb, err := newRoleBindingObj(obj)
	if err != nil {
		klog.Warningf("unexpected error, this may be a bug: %v", err)
		return false
	}

	role, err := c.getRoleFromRoleRef(rb.Namespace(), rb.RoleRef())
	if err != nil {
		klog.Infof("failed to retrieve a role for a rolebinding ref: %v", err)
		return false
	}

	c.usefulRolesLock.Lock()
	defer c.usefulRolesLock.Unlock()

	return c.usefulRoles[usefulRolesKey(role)].Len() != 0
}

// syncRoleCacheSingleSCCAdded will attempt to check whether the provided role
// allows access to the given SCC and if it does, it adds the name of the SCC to
// the slice at the usefulRoles[role] cache
// CAREFUL: The c.usefulRolesLock MUST be locked before calling this function!
func (c *saToSCCCache) syncRoleCacheSingleSCCAdded(role RoleInterface, scc *securityv1.SecurityContextConstraints) {
	if sccAllowedByPolicyRules(scc, role.Rules()) {
		cacheKey := usefulRolesKey(role)
		if c.usefulRoles[cacheKey] == nil {
			c.usefulRoles[cacheKey] = sets.NewString()
		}
		c.usefulRoles[usefulRolesKey(role)].Insert(scc.Name)
	}
}

// syncRoleCache will write the current mapping of "role->SCCs it allows" to the cache.
// CAREFUL: The c.usefulRolesLock MUST be locked before calling this function!
func (c *saToSCCCache) syncRoleCache(role RoleInterface, sccs []*securityv1.SecurityContextConstraints) {
	allowedSCCs := sets.NewString()
	for _, scc := range sccs {
		if sccAllowedByPolicyRules(scc, role.Rules()) {
			allowedSCCs.Insert(scc.Name)
		}
	}

	c.usefulRoles[usefulRolesKey(role)] = allowedSCCs
}

func sccAllowedByPolicyRules(scc *securityv1.SecurityContextConstraints, rules []rbacv1.PolicyRule) bool {
	ar := authorizer.AttributesRecord{
		User: &user.DefaultInfo{
			Name: "dummyUser",
		},
		APIGroup:        securityv1.GroupName,
		Resource:        "securitycontextconstraints",
		Name:            scc.Name,
		Verb:            "use",
		ResourceRequest: true,
	}

	return rbac.RulesAllow(ar, rules...)
}

// roleBindingAppliesToNS returns true if:
// - the rb object is a cluster role binding (zero-len namespace)
// - the namespace of the rb matches the namespace supplied in `ns`
func roleBindingAppliesToNS(rb roleBindingInterface, ns string) bool {
	roleBindingNS := rb.Namespace()
	if len(roleBindingNS) == 0 {
		return true
	}
	return ns == roleBindingNS
}

func getIndexedRolebindings(indexer cache.Indexer, saUserInfo user.Info) ([]interface{}, error) {
	objs, err := indexer.ByIndex(BySAIndexName, saUserInfo.GetName())
	if err != nil {
		return nil, fmt.Errorf("retrieving rolebindings for serviceaccount %q from the %q index failed: %w", saUserInfo.GetName(), BySAIndexName, err)
	}

	for _, g := range saUserInfo.GetGroups() {
		groupObjs, err := indexer.ByIndex(BySAIndexName, g)
		if err != nil {
			return nil, fmt.Errorf("retrieving rolebindings for group %q from the %q index failed: %w", g, BySAIndexName, err)
		}
		objs = append(objs, groupObjs...)
	}

	return objs, nil
}

func sccAllowsSA(scc *securityv1.SecurityContextConstraints, saUserInfo user.Info) bool {
	for _, u := range scc.Users {
		if u == saUserInfo.GetName() {
			return true
		}
	}

	saNSGroups := sets.NewString(saUserInfo.GetGroups()...)
	for _, g := range scc.Groups {
		if saNSGroups.Has(g) {
			return true
		}
	}

	return false
}
