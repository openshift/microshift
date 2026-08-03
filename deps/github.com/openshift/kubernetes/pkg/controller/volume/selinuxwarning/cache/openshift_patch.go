package cache

type ConflictCounter interface {
	GetConflictCount() int
}

var _ ConflictCounter = &volumeCache{}

func (c *volumeCache) GetConflictCount() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	conflictCount := 0
	for _, conflicts := range c.conflicts {
		conflictCount += len(conflicts)
	}
	return conflictCount
}
