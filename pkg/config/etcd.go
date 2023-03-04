package config

import "time"

type EtcdConfig struct {
	// The limit on the size of the etcd database; etcd will start failing writes if its size on disk reaches this value
	QuotaBackendSize  string `json:"quotaBackendSize"`
	QuotaBackendBytes int64  `json:"-"`

	// If the backend is fragmented more than `maxFragmentedPercentage`
	//		and the database size is greater than `minDefragSize`, do a defrag.
	MinDefragSize           string  `json:"minDefragSize"`
	MinDefragBytes          int64   `json:"-"`
	MaxFragmentedPercentage float64 `json:"maxFragmentedPercentage"`

	// How often to check the conditions for defragging (0 means no defrags, except for a single on startup if `doStartupDefrag` is set).
	DefragCheckFreq     string        `json:"defragCheckFreq"`
	DefragCheckDuration time.Duration `json:"-"`

	// Whether or not to do a defrag when the server finishes starting
	DoStartupDefrag bool `json:"doStartupDefrag"`
}
