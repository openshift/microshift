package config

import (
	"path/filepath"
	"time"
)

const (
	// Etcd performance degrades significantly if the memory available
	// is less than 128MB, enforce this minimum.
	EtcdMinimumMemoryLimit = 128
)

type EtcdConfig struct {
	// Set a memory limit on the etcd process; etcd will begin paging
	// memory when it gets to this value. 0 means no limit.
	MemoryLimitMB uint64 `json:"memoryLimitMB"`

	// The limit on the size of the etcd database; etcd will start
	// failing writes if its size on disk reaches this value
	QuotaBackendBytes int64 `json:"-"`

	// If the backend is fragmented more than
	// `maxFragmentedPercentage` and the database size is greater than
	// `minDefragBytes`, do a defrag.
	MinDefragBytes          int64   `json:"-"`
	MaxFragmentedPercentage float64 `json:"-"`

	// How often to check the conditions for defragging (0 means no
	// defrags, except for a single on startup).
	DefragCheckFreq time.Duration `json:"-"`
}

func (cfg *Config) EtcdConfigPath() string {
	return filepath.Join(DataDir, "etcd", "config")
}
