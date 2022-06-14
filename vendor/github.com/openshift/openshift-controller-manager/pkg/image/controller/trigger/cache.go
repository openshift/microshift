package trigger

import (
	"fmt"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/openshift/library-go/pkg/image/imageutil"
	triggerutil "github.com/openshift/library-go/pkg/image/trigger"
	"github.com/openshift/openshift-controller-manager/pkg/image/trigger"
	"k8s.io/apimachinery/pkg/api/meta"
)

// NewTriggerCache constructs a cacher that expects objects of type *trigger.CacheEntry
// and converts those triggers into entries in the thread safe cache by image stream namespace
// and name.
func NewTriggerCache() cache.ThreadSafeStore {
	return cache.NewThreadSafeStore(
		cache.Indexers{
			"images": triggerCacheIndexer,
		},
		cache.Indices{},
	)
}

// triggerCacheIndexer converts a trigger cache entry into a set of image stream keys.
func triggerCacheIndexer(obj interface{}) ([]string, error) {
	entry := obj.(*trigger.CacheEntry)
	var keys []string
	for _, t := range entry.Triggers {
		if t.From.Kind != "ImageStreamTag" || len(t.From.APIVersion) != 0 || t.Paused {
			continue
		}
		name, _, ok := imageutil.SplitImageStreamTag(t.From.Name)
		if !ok {
			continue
		}
		namespace := t.From.Namespace
		if len(namespace) == 0 {
			namespace = entry.Namespace
		}
		keys = append(keys, namespace+"/"+name)
	}
	return keys, nil
}

// ProcessEvents returns a ResourceEventHandler suitable for use with an Informer to maintain the cache.
// indexer is responsible for calculating triggers and any pending changes. Operations are added to
// the operation queue if a change is required.
func ProcessEvents(c cache.ThreadSafeStore, indexer trigger.Indexer, queue workqueue.RateLimitingInterface, tags triggerutil.TagRetriever) cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, entry, _, err := indexer.Index(obj, nil)
			if err != nil {
				utilruntime.HandleError(extractErrorForObj(obj, err))
				return
			}
			if entry != nil {
				c.Add(key, entry)
				queue.Add(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, entry, change, err := indexer.Index(newObj, oldObj)
			if err != nil {
				utilruntime.HandleError(extractErrorForObj(newObj, err))
				return
			}
			switch {
			case entry == nil:
				c.Delete(key)
			case change == cache.Added:
				c.Add(key, entry)
				queue.Add(key)
			case change == cache.Updated:
				c.Update(key, entry)
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, entry, _, err := indexer.Index(nil, obj)
			if err != nil {
				utilruntime.HandleError(extractErrorForObj(obj, err))
				return
			}
			if entry != nil {
				c.Delete(key)
			}
		},
	}
}

func extractErrorForObj(obj interface{}, err error) error {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return fmt.Errorf("unable to extract cache data from %T: %v", obj, err)
	}
	return fmt.Errorf("unable to extract cache data from %T %s/%s: %v", obj, accessor.GetNamespace(), accessor.GetName(), err)
}
