/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubelet

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/features"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	utilstore "k8s.io/kubernetes/pkg/kubelet/util/store"
	utilfs "k8s.io/kubernetes/pkg/util/filesystem"
)

// bitsDataElement is the number of bits in a bitArray.data element.
const bitsDataElement = 32

type bitArray struct {
	data       []uint32
	firstIndex int
}

func makeBitArray(size uint32) *bitArray {
	m := bitArray{
		data:       make([]uint32, (size+bitsDataElement-1)/bitsDataElement),
		firstIndex: 0,
	}
	return &m
}

func (b *bitArray) set(index uint32) {
	b.data[index/bitsDataElement] |= (uint32(1) << (index % bitsDataElement))
}

func (b *bitArray) isSet(index uint32) bool {
	return (b.data[index/bitsDataElement]>>(index%bitsDataElement))&0x1 == 1
}

func (b *bitArray) findAvailable() (uint32, bool) {
	for i := b.firstIndex; i < len(b.data); i++ {
		// Check if all bits are used (all 1s).
		if b.data[i] == math.MaxUint32 {
			continue
		}
		for j := uint32(0); j < bitsDataElement; j++ {
			if (b.data[i]>>j)&0x1 == 0 {
				v := uint32(i)*bitsDataElement + j
				b.set(v)
				// Update firstIndex to the current
				// data element since there are no other
				// unset bits before the current index.
				b.firstIndex = int(i)
				return v, true
			}
		}
	}
	return 0, false
}

func (b *bitArray) clear(index uint32) {
	i := index / bitsDataElement
	// update firstIndex if the index found is less than
	// the current one.
	if i < uint32(b.firstIndex) {
		b.firstIndex = int(i)
	}
	// clear the bit by ANDing the data element with the
	// complement of the bitmask to be cleared.
	b.data[i] &= ^(1 << (index % bitsDataElement))
}

// length for the user namespace to create (65536).
const userNsLength = (1 << 16)

// Limit the total number of pods using userns in this node to this value.
// This is an alpha limitation that will probably be lifted later.
const maxPods = 1024

// Create a new map when we removed enough pods to avoid memory leaks
// since Go maps never free memory.
const mapReInitializeThreshold = 1000

type userNsPodsManager interface {
	getPodDir(podUID types.UID) string
	listPodsFromDisk() ([]types.UID, error)
}

type usernsManager struct {
	used         *bitArray
	usedBy       map[types.UID]uint32 // Map pod.UID to range used
	removed      int
	numAllocated int
	kl           userNsPodsManager
	// This protects all members except for kl.anager
	lock sync.Mutex
}

// UserNamespace holds the configuration for the user namespace.
type userNamespace struct {
	// UIDs mappings for the user namespace.
	UIDMappings []idMapping `json:"uidMappings"`
	// GIDs mappings for the user namespace.
	GIDMappings []idMapping `json:"gidMappings"`
}

// Pod user namespace mapping
type idMapping struct {
	// Required.
	HostId uint32 `json:"hostId"`
	// Required.
	ContainerId uint32 `json:"containerId"`
	// Required.
	Length uint32 `json:"length"`
}

// mappingsFile is the file where the user namespace mappings are persisted.
const mappingsFile = "userns"

// writeMappingsToFile writes the specified user namespace configuration to the pod
// directory.
func (m *usernsManager) writeMappingsToFile(pod types.UID, userNs userNamespace) error {
	dir := m.kl.getPodDir(pod)

	data, err := json.Marshal(userNs)
	if err != nil {
		return err
	}

	fstore, err := utilstore.NewFileStore(dir, &utilfs.DefaultFs{})
	if err != nil {
		return err
	}
	if err := fstore.Write(mappingsFile, data); err != nil {
		return err
	}

	// We need to fsync the parent dir so the file is guaranteed to be there.
	// fstore guarantees an atomic write, we need durability too.
	parentDir, err := os.Open(dir)
	if err != nil {
		return err
	}

	if err = parentDir.Sync(); err != nil {
		// Ignore return here, there is already an error reported.
		parentDir.Close()
		return err
	}

	return parentDir.Close()
}

// readMappingsFromFile reads the user namespace configuration from the pod directory.
func (m *usernsManager) readMappingsFromFile(pod types.UID) ([]byte, error) {
	dir := m.kl.getPodDir(pod)
	fstore, err := utilstore.NewFileStore(dir, &utilfs.DefaultFs{})
	if err != nil {
		return nil, err
	}
	return fstore.Read(mappingsFile)
}

func MakeUserNsManager(kl userNsPodsManager) (*usernsManager, error) {
	m := usernsManager{
		// Create a bitArray for all the UID space (2^32).
		// As a by product of that, no index param to bitArray can be out of bounds (index is uint32).
		used:   makeBitArray((math.MaxUint32 + 1) / userNsLength),
		usedBy: make(map[types.UID]uint32),
		kl:     kl,
	}
	// First block is reserved for the host.
	m.used.set(0)

	// Second block will be used for phase II. Don't assign that range for now.
	m.used.set(1)

	// do not bother reading the list of pods if user namespaces are not enabled.
	if !utilfeature.DefaultFeatureGate.Enabled(features.UserNamespacesStatelessPodsSupport) {
		return &m, nil
	}

	found, err := kl.listPodsFromDisk()
	if err != nil {
		if os.IsNotExist(err) {
			return &m, nil
		}
		return nil, fmt.Errorf("user namespace manager can't read pods from disk: %w", err)

	}
	for _, podUID := range found {
		klog.V(5).InfoS("reading pod from disk for user namespace", "podUID", podUID)
		if err := m.recordPodMappings(podUID); err != nil {
			return nil, err
		}
	}

	return &m, nil
}

// recordPodMappings registers the range used for the user namespace if the
// usernsConfFile exists in the pod directory.
func (m *usernsManager) recordPodMappings(pod types.UID) error {
	content, err := m.readMappingsFromFile(pod)
	if err != nil && err != utilstore.ErrKeyNotFound {
		return err
	}

	// If no content, it means the pod doesn't have userns. Nothing else to do
	if len(content) == 0 {
		return nil
	}

	_, err = m.parseUserNsFileAndRecord(pod, content)
	return err
}

// isSet checks if the specified index is already set.
func (m *usernsManager) isSet(v uint32) bool {
	index := v / userNsLength
	return m.used.isSet(index)
}

// allocateOne finds a free user namespace and allocate it to the specified pod.
// The first return value is the first ID in the user namespace, the second returns
// the length for the user namespace range.
func (m *usernsManager) allocateOne(pod types.UID) (firstID uint32, length uint32, err error) {
	if m.numAllocated >= maxPods {
		return 0, 0, fmt.Errorf("limit on count of pods with user namespaces exceeded (limit is %v, current pods with userns: %v)", maxPods, m.numAllocated)
	}
	m.numAllocated++
	defer func() {
		if err != nil {
			m.numAllocated--
		}
	}()

	firstZero, found := m.used.findAvailable()
	if !found {
		return 0, 0, fmt.Errorf("could not find an empty slot to allocate a user namespace")
	}

	klog.V(5).InfoS("new pod user namespace allocation", "podUID", pod)

	firstID = firstZero * userNsLength
	m.usedBy[pod] = firstID
	return firstID, userNsLength, nil
}

// record stores the user namespace [from; from+length] to the specified pod.
func (m *usernsManager) record(pod types.UID, from, length uint32) (err error) {
	if length != userNsLength {
		return fmt.Errorf("wrong user namespace length %v", length)
	}
	if from%userNsLength != 0 {
		return fmt.Errorf("wrong user namespace offset specified %v", from)
	}
	prevFrom, found := m.usedBy[pod]
	if found && prevFrom != from {
		return fmt.Errorf("different user namespace range already used by pod %q", pod)
	}
	index := from / userNsLength
	// if the pod wasn't found then verify the range is free.
	if !found && m.used.isSet(index) {
		return fmt.Errorf("range picked for pod %q already taken", pod)
	}
	// The pod is already registered, nothing to do.
	if found && prevFrom == from {
		return nil
	}
	if m.numAllocated >= maxPods {
		return fmt.Errorf("limit on count of pods with user namespaces exceeded (limit is %v, current pods with userns: %v)", maxPods, m.numAllocated)
	}
	m.numAllocated++
	defer func() {
		if err != nil {
			m.numAllocated--
		}
	}()

	klog.V(5).InfoS("new pod user namespace allocation", "podUID", pod)

	// "from" is a ID (UID/GID), set the corresponding userns of size
	// userNsLength in the bit-array.
	m.used.set(index)
	m.usedBy[pod] = from
	return nil
}

// Release releases the user namespace allocated to the specified pod.
func (m *usernsManager) Release(podUID types.UID) {
	if !utilfeature.DefaultFeatureGate.Enabled(features.UserNamespacesStatelessPodsSupport) {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	m.releaseWithLock(podUID)
}

func (m *usernsManager) releaseWithLock(pod types.UID) {
	v, ok := m.usedBy[pod]
	if !ok {
		klog.V(5).InfoS("pod user namespace allocation not present", "podUID", pod)
		return
	}
	delete(m.usedBy, pod)

	klog.V(5).InfoS("releasing pod user namespace allocation", "podUID", pod)
	m.numAllocated--
	m.removed++

	_ = os.Remove(filepath.Join(m.kl.getPodDir(pod), mappingsFile))

	if m.removed%mapReInitializeThreshold == 0 {
		n := make(map[types.UID]uint32)
		for k, v := range m.usedBy {
			n[k] = v
		}
		m.usedBy = n
		m.removed = 0
	}
	m.used.clear(v / userNsLength)
}

func (m *usernsManager) parseUserNsFileAndRecord(pod types.UID, content []byte) (userNs userNamespace, err error) {
	if err = json.Unmarshal([]byte(content), &userNs); err != nil {
		err = fmt.Errorf("can't parse file: %w", err)
		return
	}

	if len(userNs.UIDMappings) != 1 {
		err = fmt.Errorf("invalid user namespace configuration: no more than one mapping allowed.")
		return
	}

	if len(userNs.UIDMappings) != len(userNs.GIDMappings) {
		err = fmt.Errorf("invalid user namespace configuration: GID and UID mappings should be identical.")
		return
	}

	if userNs.UIDMappings[0] != userNs.GIDMappings[0] {
		err = fmt.Errorf("invalid user namespace configuration: GID and UID mapping should be identical")
		return
	}

	// We don't produce configs without root mapped and some runtimes assume it is mapped.
	// Validate the file has something we produced and can digest.
	if userNs.UIDMappings[0].ContainerId != 0 {
		err = fmt.Errorf("invalid user namespace configuration: UID 0 must be mapped")
		return
	}

	if userNs.GIDMappings[0].ContainerId != 0 {
		err = fmt.Errorf("invalid user namespace configuration: GID 0 must be mapped")
		return
	}

	hostId := userNs.UIDMappings[0].HostId
	length := userNs.UIDMappings[0].Length

	err = m.record(pod, hostId, length)
	return
}

func (m *usernsManager) createUserNs(pod *v1.Pod) (userNs userNamespace, err error) {
	firstID, length, err := m.allocateOne(pod.UID)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			m.releaseWithLock(pod.UID)
		}
	}()

	userNs = userNamespace{
		UIDMappings: []idMapping{
			{
				ContainerId: 0,
				HostId:      firstID,
				Length:      length,
			},
		},
		GIDMappings: []idMapping{
			{
				ContainerId: 0,
				HostId:      firstID,
				Length:      length,
			},
		},
	}

	return userNs, m.writeMappingsToFile(pod.UID, userNs)
}

// GetOrCreateUserNamespaceMappings returns the configuration for the sandbox user namespace
func (m *usernsManager) GetOrCreateUserNamespaceMappings(pod *v1.Pod) (*runtimeapi.UserNamespace, error) {
	if !utilfeature.DefaultFeatureGate.Enabled(features.UserNamespacesStatelessPodsSupport) {
		return nil, nil
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if pod.Spec.HostUsers == nil || *pod.Spec.HostUsers == true {
		return &runtimeapi.UserNamespace{
			Mode: runtimeapi.NamespaceMode_NODE,
		}, nil
	}

	content, err := m.readMappingsFromFile(pod.UID)
	if err != nil && err != utilstore.ErrKeyNotFound {
		return nil, err
	}

	var userNs userNamespace
	if string(content) != "" {
		userNs, err = m.parseUserNsFileAndRecord(pod.UID, content)
		if err != nil {
			return nil, err
		}
	} else {
		userNs, err = m.createUserNs(pod)
		if err != nil {
			return nil, err
		}
	}

	var uids []*runtimeapi.IDMapping
	var gids []*runtimeapi.IDMapping

	for _, u := range userNs.UIDMappings {
		uids = append(uids, &runtimeapi.IDMapping{
			HostId:      u.HostId,
			ContainerId: u.ContainerId,
			Length:      u.Length,
		})
	}
	for _, g := range userNs.GIDMappings {
		gids = append(gids, &runtimeapi.IDMapping{
			HostId:      g.HostId,
			ContainerId: g.ContainerId,
			Length:      g.Length,
		})
	}

	return &runtimeapi.UserNamespace{
		Mode: runtimeapi.NamespaceMode_POD,
		Uids: uids,
		Gids: gids,
	}, nil
}

// CleanupOrphanedPodUsernsAllocations reconciliates the state of user namespace
// allocations with the pods actually running. It frees any user namespace
// allocation for orphaned pods.
func (m *usernsManager) CleanupOrphanedPodUsernsAllocations(pods []*v1.Pod, runningPods []*kubecontainer.Pod) error {
	if !utilfeature.DefaultFeatureGate.Enabled(features.UserNamespacesStatelessPodsSupport) {
		return nil
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	allPods := sets.NewString()
	for _, pod := range pods {
		allPods.Insert(string(pod.UID))
	}
	for _, pod := range runningPods {
		allPods.Insert(string(pod.ID))
	}

	allFound := sets.NewString()
	found, err := m.kl.listPodsFromDisk()
	if err != nil {
		return err
	}

	for _, podUID := range found {
		allFound.Insert(string(podUID))
	}

	// Lets remove all the pods "found" that are not known.
	for _, podUID := range found {
		if allPods.Has(string(podUID)) {
			continue
		}

		klog.V(5).InfoS("Clean up orphaned pod user namespace possible allocation", "podUID", podUID)
		m.releaseWithLock(podUID)
	}

	// Lets remove any existing allocation for a pod that is not "found".
	for podUID := range m.usedBy {
		if allFound.Has(string(podUID)) {
			continue
		}

		klog.V(5).InfoS("Clean up orphaned pod user namespace possible allocation", "podUID", podUID)
		m.releaseWithLock(podUID)
	}

	return nil
}

// getHostIDsForPod if the pod uses user namespaces, takes the uid and gid
// inside the container and returns the host UID and GID those are mapped to on
// the host. If containerUID/containerGID is nil, then it returns the host
// UID/GID for ID 0 inside the container.
// If the pod is not using user namespaces, as there is no mapping needed, the
// same containerUID and containerGID params are returned.
func (m *usernsManager) getHostIDsForPod(pod *v1.Pod, containerUID, containerGID *int64) (hostUID, hostGID *int64, err error) {
	if !utilfeature.DefaultFeatureGate.Enabled(features.UserNamespacesStatelessPodsSupport) {
		return containerUID, containerGID, nil
	}

	if pod == nil || pod.Spec.HostUsers == nil || *pod.Spec.HostUsers == true {
		return containerUID, containerGID, nil
	}

	mapping, err := m.GetOrCreateUserNamespaceMappings(pod)
	if err != nil {
		err = fmt.Errorf("Error getting pod user namespace mapping: %w", err)
		return
	}

	uid, err := hostIDFromMapping(mapping.Uids, containerUID)
	if err != nil {
		err = fmt.Errorf("Error getting host UID: %w", err)
		return
	}

	gid, err := hostIDFromMapping(mapping.Gids, containerGID)
	if err != nil {
		err = fmt.Errorf("Error getting host GID: %w", err)
		return
	}

	return &uid, &gid, nil
}

func hostIDFromMapping(mapping []*runtimeapi.IDMapping, containerId *int64) (int64, error) {
	if len(mapping) == 0 {
		return 0, fmt.Errorf("can't use empty user namespace mapping")
	}

	// If none is requested, root inside the container is used
	id := int64(0)
	if containerId != nil {
		id = *containerId
	}

	for _, m := range mapping {
		if m == nil {
			continue
		}

		firstId := int64(m.ContainerId)
		lastId := firstId + int64(m.Length) - 1

		// The id we are looking for is in the range
		if id >= firstId && id <= lastId {
			// Return the host id for this container id
			return int64(m.HostId) + id - firstId, nil
		}
	}

	return 0, fmt.Errorf("ID: %v not present in pod user namespace", id)
}
