// Copyright 2015 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.etcd.io/etcd/client/pkg/v3/transport"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/pkg/v3/netutil"
	"go.etcd.io/etcd/server/v3/datadir"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
)

// ServerConfig holds the configuration of etcd as taken from the command line or discovery.
type ServerConfig struct {
	Name           string
	DiscoveryURL   string
	DiscoveryProxy string
	ClientURLs     types.URLs
	PeerURLs       types.URLs
	DataDir        string
	// DedicatedWALDir config will make the etcd to write the WAL to the WALDir
	// rather than the dataDir/member/wal.
	DedicatedWALDir string

	SnapshotCount uint64

	// SnapshotCatchUpEntries is the number of entries for a slow follower
	// to catch-up after compacting the raft storage entries.
	// We expect the follower has a millisecond level latency with the leader.
	// The max throughput is around 10K. Keep a 5K entries is enough for helping
	// follower to catch up.
	// WARNING: only change this for tests. Always use "DefaultSnapshotCatchUpEntries"
	SnapshotCatchUpEntries uint64

	MaxSnapFiles uint
	MaxWALFiles  uint

	// BackendBatchInterval is the maximum time before commit the backend transaction.
	BackendBatchInterval time.Duration
	// BackendBatchLimit is the maximum operations before commit the backend transaction.
	BackendBatchLimit int

	// BackendFreelistType is the type of the backend boltdb freelist.
	BackendFreelistType bolt.FreelistType

	InitialPeerURLsMap  types.URLsMap
	InitialClusterToken string
	NewCluster          bool
	PeerTLSInfo         transport.TLSInfo

	CORS map[string]struct{}

	// HostWhitelist lists acceptable hostnames from client requests.
	// If server is insecure (no TLS), server only accepts requests
	// whose Host header value exists in this white list.
	HostWhitelist map[string]struct{}

	TickMs        uint
	ElectionTicks int

	// InitialElectionTickAdvance is true, then local member fast-forwards
	// election ticks to speed up "initial" leader election trigger. This
	// benefits the case of larger election ticks. For instance, cross
	// datacenter deployment may require longer election timeout of 10-second.
	// If true, local node does not need wait up to 10-second. Instead,
	// forwards its election ticks to 8-second, and have only 2-second left
	// before leader election.
	//
	// Major assumptions are that:
	//  - cluster has no active leader thus advancing ticks enables faster
	//    leader election, or
	//  - cluster already has an established leader, and rejoining follower
	//    is likely to receive heartbeats from the leader after tick advance
	//    and before election timeout.
	//
	// However, when network from leader to rejoining follower is congested,
	// and the follower does not receive leader heartbeat within left election
	// ticks, disruptive election has to happen thus affecting cluster
	// availabilities.
	//
	// Disabling this would slow down initial bootstrap process for cross
	// datacenter deployments. Make your own tradeoffs by configuring
	// --initial-election-tick-advance at the cost of slow initial bootstrap.
	//
	// If single-node, it advances ticks regardless.
	//
	// See https://github.com/etcd-io/etcd/issues/9333 for more detail.
	InitialElectionTickAdvance bool

	BootstrapTimeout time.Duration

	AutoCompactionRetention time.Duration
	AutoCompactionMode      string
	CompactionBatchLimit    int
	CompactionSleepInterval time.Duration
	QuotaBackendBytes       int64
	MaxTxnOps               uint

	// MaxRequestBytes is the maximum request size to send over raft.
	MaxRequestBytes uint

	// MaxConcurrentStreams specifies the maximum number of concurrent
	// streams that each client can open at a time.
	MaxConcurrentStreams uint32

	WarningApplyDuration time.Duration

	StrictReconfigCheck bool

	// ClientCertAuthEnabled is true when cert has been signed by the client CA.
	ClientCertAuthEnabled bool

	AuthToken  string
	BcryptCost uint
	TokenTTL   uint

	// InitialCorruptCheck is true to check data corruption on boot
	// before serving any peer/client traffic.
	InitialCorruptCheck     bool
	CorruptCheckTime        time.Duration
	CompactHashCheckEnabled bool
	CompactHashCheckTime    time.Duration

	// PreVote is true to enable Raft Pre-Vote.
	PreVote bool

	// SocketOpts are socket options passed to listener config.
	SocketOpts transport.SocketOpts

	// Logger logs server-side operations.
	Logger *zap.Logger

	ForceNewCluster           bool
	ForceNewClusterBumpAmount uint64

	// EnableLeaseCheckpoint enables leader to send regular checkpoints to other members to prevent reset of remaining TTL on leader change.
	EnableLeaseCheckpoint bool
	// LeaseCheckpointInterval time.Duration is the wait duration between lease checkpoints.
	LeaseCheckpointInterval time.Duration
	// LeaseCheckpointPersist enables persisting remainingTTL to prevent indefinite auto-renewal of long lived leases. Always enabled in v3.6. Should be used to ensure smooth upgrade from v3.5 clusters with this feature enabled.
	LeaseCheckpointPersist bool

	EnableGRPCGateway bool

	// ExperimentalEnableDistributedTracing enables distributed tracing using OpenTelemetry protocol.
	ExperimentalEnableDistributedTracing bool
	// ExperimentalTracerOptions are options for OpenTelemetry gRPC interceptor.
	ExperimentalTracerOptions []otelgrpc.Option

	WatchProgressNotifyInterval time.Duration

	// UnsafeNoFsync disables all uses of fsync.
	// Setting this is unsafe and will cause data loss.
	UnsafeNoFsync bool `json:"unsafe-no-fsync"`

	DowngradeCheckTime time.Duration

	// ExperimentalMemoryMlock enables mlocking of etcd owned memory pages.
	// The setting improves etcd tail latency in environments were:
	//   - memory pressure might lead to swapping pages to disk
	//   - disk latency might be unstable
	// Currently all etcd memory gets mlocked, but in future the flag can
	// be refined to mlock in-use area of bbolt only.
	ExperimentalMemoryMlock bool `json:"experimental-memory-mlock"`

	// ExperimentalTxnModeWriteWithSharedBuffer enable write transaction to use
	// a shared buffer in its readonly check operations.
	ExperimentalTxnModeWriteWithSharedBuffer bool `json:"experimental-txn-mode-write-with-shared-buffer"`

	// ExperimentalStopGRPCServiceOnDefrag enables etcd gRPC service to stop serving client requests on defragmentation.
	ExperimentalStopGRPCServiceOnDefrag bool `json:"experimental-stop-grpc-service-on-defrag"`

	// ExperimentalBootstrapDefragThresholdMegabytes is the minimum number of megabytes needed to be freed for etcd server to
	// consider running defrag during bootstrap. Needs to be set to non-zero value to take effect.
	ExperimentalBootstrapDefragThresholdMegabytes uint `json:"experimental-bootstrap-defrag-threshold-megabytes"`

	// ExperimentalMaxLearners sets a limit to the number of learner members that can exist in the cluster membership.
	ExperimentalMaxLearners int `json:"experimental-max-learners"`

	// V2Deprecation defines a phase of v2store deprecation process.
	V2Deprecation V2DeprecationEnum `json:"v2-deprecation"`

	// openshiftHardwareDelaySeconds sets the duration, in seconds, to extend the standard etcd
	// request timeout by based on hardware speed. OpenShift currently has a notion of differentiating
	// between standard and slower hardware: https://github.com/openshift/cluster-etcd-operator/blob/5bbe49442101475febb89bba3be808aa121f5c0c/pkg/hwspeedhelpers/hwhelper.go#L21-L33
	openshiftHardwareDelaySeconds int
}

// VerifyBootstrap sanity-checks the initial config for bootstrap case
// and returns an error for things that should never happen.
func (c *ServerConfig) VerifyBootstrap() error {
	if err := c.hasLocalMember(); err != nil {
		return err
	}
	if err := c.advertiseMatchesCluster(); err != nil {
		return err
	}
	if CheckDuplicateURL(c.InitialPeerURLsMap) {
		return fmt.Errorf("initial cluster %s has duplicate url", c.InitialPeerURLsMap)
	}
	if c.InitialPeerURLsMap.String() == "" && c.DiscoveryURL == "" {
		return fmt.Errorf("initial cluster unset and no discovery URL found")
	}
	return nil
}

// VerifyJoinExisting sanity-checks the initial config for join existing cluster
// case and returns an error for things that should never happen.
func (c *ServerConfig) VerifyJoinExisting() error {
	// The member has announced its peer urls to the cluster before starting; no need to
	// set the configuration again.
	if err := c.hasLocalMember(); err != nil {
		return err
	}
	if CheckDuplicateURL(c.InitialPeerURLsMap) {
		return fmt.Errorf("initial cluster %s has duplicate url", c.InitialPeerURLsMap)
	}
	if c.DiscoveryURL != "" {
		return fmt.Errorf("discovery URL should not be set when joining existing initial cluster")
	}
	return nil
}

// hasLocalMember checks that the cluster at least contains the local server.
func (c *ServerConfig) hasLocalMember() error {
	if urls := c.InitialPeerURLsMap[c.Name]; urls == nil {
		return fmt.Errorf("couldn't find local name %q in the initial cluster configuration", c.Name)
	}
	return nil
}

// advertiseMatchesCluster confirms peer URLs match those in the cluster peer list.
func (c *ServerConfig) advertiseMatchesCluster() error {
	urls, apurls := c.InitialPeerURLsMap[c.Name], c.PeerURLs.StringSlice()
	urls.Sort()
	sort.Strings(apurls)
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()
	ok, err := netutil.URLStringsEqual(ctx, c.Logger, apurls, urls.StringSlice())
	if ok {
		return nil
	}

	initMap, apMap := make(map[string]struct{}), make(map[string]struct{})
	for _, url := range c.PeerURLs {
		apMap[url.String()] = struct{}{}
	}
	for _, url := range c.InitialPeerURLsMap[c.Name] {
		initMap[url.String()] = struct{}{}
	}

	missing := []string{}
	for url := range initMap {
		if _, ok := apMap[url]; !ok {
			missing = append(missing, url)
		}
	}
	if len(missing) > 0 {
		for i := range missing {
			missing[i] = c.Name + "=" + missing[i]
		}
		mstr := strings.Join(missing, ",")
		apStr := strings.Join(apurls, ",")
		return fmt.Errorf("--initial-cluster has %s but missing from --initial-advertise-peer-urls=%s (%v)", mstr, apStr, err)
	}

	for url := range apMap {
		if _, ok := initMap[url]; !ok {
			missing = append(missing, url)
		}
	}
	if len(missing) > 0 {
		mstr := strings.Join(missing, ",")
		umap := types.URLsMap(map[string]types.URLs{c.Name: c.PeerURLs})
		return fmt.Errorf("--initial-advertise-peer-urls has %s but missing from --initial-cluster=%s", mstr, umap.String())
	}

	// resolved URLs from "--initial-advertise-peer-urls" and "--initial-cluster" did not match or failed
	apStr := strings.Join(apurls, ",")
	umap := types.URLsMap(map[string]types.URLs{c.Name: c.PeerURLs})
	return fmt.Errorf("failed to resolve %s to match --initial-cluster=%s (%v)", apStr, umap.String(), err)
}

func (c *ServerConfig) MemberDir() string { return datadir.ToMemberDir(c.DataDir) }

func (c *ServerConfig) WALDir() string {
	if c.DedicatedWALDir != "" {
		return c.DedicatedWALDir
	}
	return datadir.ToWalDir(c.DataDir)
}

func (c *ServerConfig) SnapDir() string { return filepath.Join(c.MemberDir(), "snap") }

func (c *ServerConfig) ShouldDiscover() bool { return c.DiscoveryURL != "" }

// ReqTimeout returns timeout for request to finish.
func (c *ServerConfig) ReqTimeout() time.Duration {
	// 5s for queue waiting, computation and disk IO delay
	// + 2 * election timeout for possible leader election
	// return 5*time.Second + 2*time.Duration(c.ElectionTicks*int(c.TickMs))*time.Millisecond

	// OCP PATCH: use an environment variable to configure the non election-related timeout
	// value used. If it isn't configured, default to 25 seconds.
	// Default of 25 seconds gets us to a total of 27s on "standard hardware" and 30s on "slow hardware"
	// due to the configurations of election timeouts in https://github.com/openshift/cluster-etcd-operator/blob/5bbe49442101475febb89bba3be808aa121f5c0c/pkg/hwspeedhelpers/hwhelper.go#L21-L33
	// This gets us closer to the kube-apiserver's upper bound request timeout of 34 seconds while leaving some time
	// for other things to have on the kube-apiserver side of things.
	// See: https://issues.redhat.com/browse/OCPBUGS-50510
	// TODO(everettraven): pursue an upstream fix for this so we don't have to carry this patch for forever
	if c.openshiftHardwareDelaySeconds == 0 {
		hardwareDelayTimeout := 25
		if hardwareDelayOverrideEnv := os.Getenv("OPENSHIFT_ETCD_HARDWARE_DELAY_TIMEOUT"); hardwareDelayOverrideEnv != "" {
			hardwareDelayOverride, err := strconv.Atoi(hardwareDelayOverrideEnv)
			if err != nil {
				c.Logger.Sugar().Infof("OPENSHIFT_ETCD_HARDWARE_DELAY_TIMEOUT specified but could not be parsed. falling back to default of %s seconds. parse error: %v", hardwareDelayTimeout, err)
			} else {
				hardwareDelayTimeout = hardwareDelayOverride
			}
		}

		c.openshiftHardwareDelaySeconds = hardwareDelayTimeout
	}

	return time.Duration(c.openshiftHardwareDelaySeconds)*time.Second + 2*time.Duration(c.ElectionTicks*int(c.TickMs))*time.Millisecond
}

func (c *ServerConfig) ElectionTimeout() time.Duration {
	return time.Duration(c.ElectionTicks*int(c.TickMs)) * time.Millisecond
}

func (c *ServerConfig) PeerDialTimeout() time.Duration {
	// 1s for queue wait and election timeout
	return time.Second + time.Duration(c.ElectionTicks*int(c.TickMs))*time.Millisecond
}

func CheckDuplicateURL(urlsmap types.URLsMap) bool {
	um := make(map[string]bool)
	for _, urls := range urlsmap {
		for _, url := range urls {
			u := url.String()
			if um[u] {
				return true
			}
			um[u] = true
		}
	}
	return false
}

func (c *ServerConfig) BootstrapTimeoutEffective() time.Duration {
	if c.BootstrapTimeout != 0 {
		return c.BootstrapTimeout
	}
	return time.Second
}

func (c *ServerConfig) BackendPath() string { return datadir.ToBackendFileName(c.DataDir) }
