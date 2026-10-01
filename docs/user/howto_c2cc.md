# Cluster-to-Cluster Connectivity (C2CC)

Cluster-to-Cluster Connectivity (C2CC) enables direct Pod-to-Pod and
Pod-to-Service communication between independent MicroShift clusters.
It targets edge deployments where multiple single-node MicroShift instances
on the same network segment (or reachable via routable next-hops) need to
consume each other's workloads without an 3rd party interconnect solution.

C2CC provides:

- **Pod-to-Pod communication** across clusters using pod IPs.
- **Pod-to-Service communication** across clusters (ClusterIP and headless
  Services) using DNS names such as
  `myservice.mynamespace.svc.cluster-b.remote`.
- **Source pod IP preservation** — cross-cluster traffic bypasses SNAT, so
  NetworkPolicies on the remote cluster can enforce access control based on
  the originating pod IP.
- **Health monitoring** — per-remote-cluster health and latency reporting
  through the `RemoteCluster` custom resource.
- IPv4, IPv6, and dual-stack support.

C2CC is built into MicroShift and driven entirely by the MicroShift
configuration file.
There is no separate service to install: adding entries to
`clusterToCluster.remoteClusters` enables the feature, removing them disables
it and cleans up all C2CC-owned state.

Cross-cluster traffic travels as plain routed IP over the underlay network —
there is no tunnel between the hosts and no encryption by default.
For production deployments, encrypting and authenticating this traffic with
IPsec is strongly recommended.
See [Encrypting C2CC Traffic with IPsec](./howto_c2cc_ipsec.md).

## How It Works

Once enabled, the C2CC controller inside MicroShift automatically manages
several routing and NAT subsystems on the host:

- **OVN static routes** — routes on the OVN gateway router (`GR_<node>`)
  send traffic destined to remote pod and service CIDRs to the configured
  next-hop instead of the default gateway. C2CC-owned routes are tagged with
  `external_ids:k8s.ovn.org/owner-controller=microshift-c2cc` (visible via
  `ovn-nbctl lr-route-list GR_<node>`).
- **SNAT bypass** — nftables rules (commented `c2cc-no-masq`) in
  OVN-Kubernetes' `ovn-kube-pod-subnet-masq` chain and the node annotation
  `k8s.ovn.org/node-ingress-snat-exclude-subnets` prevent masquerading of
  cross-cluster traffic, preserving original pod source IPs end-to-end.
- **Linux policy routing** — routes to remote CIDRs live in a dedicated
  routing table (default 200, see `ip route show table 200`). A second table
  (default 201) reroutes inbound remote-to-local-service traffic through the
  OVN management port so that service load balancing does not SNAT the
  source.
- **CoreDNS forwarding** — for each remote cluster with a `domain`
  configured, a CoreDNS server block rewrites the remote domain to
  `cluster.local` and forwards the query to the remote cluster's DNS
  service.
- **Health probes** — a lightweight `c2cc-probe` Deployment in the
  `openshift-c2cc` namespace probes each remote cluster's probe Service
  through the full C2CC data path and reports the results in
  `RemoteCluster` custom resources.

The controller reconciles continuously (event-driven, with a periodic
fallback every few seconds).
If routes, rules, or annotations are removed — for example by an
OVN-Kubernetes restart, a firewall reload, or a host reboot — they are
restored automatically within seconds.

## Prerequisites

- Two or more hosts with MicroShift installed and the default OVN-Kubernetes
  CNI (C2CC requires OVN-Kubernetes; it is not available with
  `network.cniPlugin: none`). Ideally, MicroShift has not been started yet so
  that cluster and service CIDRs can be set before the first boot. If
  MicroShift has already run, wipe its state first with
  `microshift-cleanup-data --all` (this deletes all workloads).
- IP connectivity between the hosts on the underlay network (same L2
  segment, or routable next-hops).
- **Non-overlapping pod and service CIDRs across all clusters.**
  MicroShift's defaults are `10.42.0.0/16` (pods) and `10.43.0.0/16`
  (services), so all clusters except one must override them. Plan the CIDRs
  up front: changing `network.clusterNetwork` or `network.serviceNetwork` on
  a cluster that has already started requires wiping MicroShift data
  (`microshift-cleanup-data --all`), which deletes all workloads.
- **Bidirectional configuration.** Both clusters must configure each other
  as remotes. One-way setups are not supported.

## Configure C2CC

The examples below assume a two-cluster setup:

| Host    | Underlay IP   | Pod CIDR       | Service CIDR   | Domain             |
|---------|---------------|----------------|----------------|--------------------|
| Host A  | 192.168.1.10  | 10.42.0.0/16   | 10.43.0.0/16   | cluster-a.remote   |
| Host B  | 192.168.1.20  | 10.45.0.0/16   | 10.46.0.0/16   | cluster-b.remote   |

### Override cluster and service networks

Host A uses the MicroShift defaults.
On **Host B**, override the pod and service networks before MicroShift
starts for the first time, for example in `/etc/microshift/config.yaml`:

```yaml
network:
  clusterNetwork:
    - 10.45.0.0/16
  serviceNetwork:
    - 10.46.0.0/16
```

### Add the remote cluster configuration

On each host, create a configuration drop-in pointing at the remote cluster.

On **Host A**, create `/etc/microshift/config.d/50-c2cc.yaml`:

```yaml
clusterToCluster:
  remoteClusters:
  - nextHop:
    - 192.168.1.20
    clusterNetwork:
    - 10.45.0.0/16
    serviceNetwork:
    - 10.46.0.0/16
    domain: cluster-b.remote
```

On **Host B**, create the same file describing Host A:

```yaml
clusterToCluster:
  remoteClusters:
  - nextHop:
    - 192.168.1.10
    clusterNetwork:
    - 10.42.0.0/16
    serviceNetwork:
    - 10.43.0.0/16
    domain: cluster-a.remote
```

For each remote cluster:

- **`nextHop`** — IP addresses of the remote cluster's node, used as the
  routing next-hop. At most one IPv4 and one IPv6 address; dual-stack
  clusters need both.
- **`clusterNetwork` / `serviceNetwork`** — the remote cluster's pod and
  service CIDRs. They must not overlap with the local networks or with any
  other remote.
- **`domain`** — optional DNS suffix for the remote cluster. When set,
  Services on the remote cluster are resolvable as `<svc>.<ns>.svc.<domain>`.
  When empty, no DNS forwarding is configured for this remote.

To connect more than two clusters, add one `remoteClusters` entry per remote
on every host (full mesh, N×(N-1) entries in total).

### Configure the firewall

MicroShift intentionally does not manage firewall rules — edge
deployments often have site-specific firewall policies.
On each host, add the *remote* cluster's pod and service CIDRs to the
trusted zone:

```bash
# On Host A — trust Host B's pod and service CIDRs
sudo firewall-cmd --permanent --zone=trusted --add-source=10.45.0.0/16
sudo firewall-cmd --permanent --zone=trusted --add-source=10.46.0.0/16
sudo firewall-cmd --reload
```

```bash
# On Host B — trust Host A's pod and service CIDRs
sudo firewall-cmd --permanent --zone=trusted --add-source=10.42.0.0/16
sudo firewall-cmd --permanent --zone=trusted --add-source=10.43.0.0/16
sudo firewall-cmd --reload
```

Do **not** add the remote host's own IP to the trusted zone.
C2CC is designed for pod-to-pod and pod-to-service traffic only; leaving the
host IP untrusted blocks host-originated traffic from reaching local pods.
(The IPsec setup requires opening UDP 500/4500 and ESP separately — see the
[IPsec guide](./howto_c2cc_ipsec.md).)

### Start MicroShift

```bash
sudo systemctl start microshift
```

> **Note:** If MicroShift was already running and you changed
> `network.clusterNetwork` or `network.serviceNetwork`, you must wipe its
> data first with `microshift-cleanup-data --all` (this deletes all
> workloads) and then start the service.

MicroShift validates the C2CC configuration on startup.
If validation fails (overlapping CIDRs, a routing loop, an invalid next-hop,
and so on), MicroShift logs the errors and does not start.
All C2CC configuration changes — including DNS cache settings — require a
MicroShift restart to take effect.

## Verify

Check that C2CC is active in the journal:

```bash
journalctl -u microshift | grep C2CC
# ... "C2CC is enabled with 1 remote cluster(s)"
```

Check the health of each remote cluster.
Each remote gets a cluster-scoped `RemoteCluster` resource named after its
next-hop IP:

```bash
$ oc get remoteclusters.microshift.io
NAME                AGE
c2cc-192-168-1-20   5m

$ oc get remoteclusters.microshift.io c2cc-192-168-1-20 -o yaml
apiVersion: microshift.io/v1alpha1
kind: RemoteCluster
metadata:
  name: c2cc-192-168-1-20
  labels:
    app.kubernetes.io/managed-by: c2cc-route-manager
spec:
  probeInterval: 10s
  probeTargets:
  - 10.46.0.11:8080
status:
  state: Healthy
  lastProbeTime: "2026-07-07T12:12:09Z"
  lastSuccessfulProbe: "2026-07-07T12:12:09Z"
  targetResults:
  - target: 10.46.0.11:8080
    state: Healthy
    latency:
      avg: 1.579596ms
      last: 1.21661ms
      max: 3.621629ms
      min: 959.724µs
      stddev: 512.887µs
```

`status.state` transitions from `NeverProbed` to `Healthy` once end-to-end
probes succeed.
`status.targetResults` contains per-target latency statistics measured
through the full C2CC data path, and `status.errors` explains failures.

Inspect the routes C2CC manages (on Host A, with Host B configured as the
remote):

```bash
$ ip route show table 200
10.45.0.0/16 via 192.168.1.20 dev enp1s0 proto 200
10.46.0.0/16 via 192.168.1.20 dev enp1s0 proto 200

$ ip route show table 201
10.43.0.0/16 via 10.42.0.1 dev ovn-k8s-mp0 proto 201

$ ip rule
0:	from all lookup local
30:	from all fwmark 0x1745ec lookup 7
99:	from 10.45.0.0/16 to 10.43.0.0/16 lookup 201
99:	from 10.46.0.0/16 to 10.43.0.0/16 lookup 201
100:	to 10.45.0.0/16 lookup 200
100:	to 10.46.0.0/16 lookup 200
32766:	from all lookup main
32767:	from all lookup default
```

The SNAT bypass state is visible in the OVN-Kubernetes nftables chain and
on the Node object:

```bash
$ sudo nft list chain inet ovn-kubernetes ovn-kube-pod-subnet-masq
table inet ovn-kubernetes {
	chain ovn-kube-pod-subnet-masq {
		ip daddr 10.45.0.0/16 return comment "c2cc-no-masq:10.45.0.0/16"
		ip daddr 10.46.0.0/16 return comment "c2cc-no-masq:10.46.0.0/16"
		ip saddr 10.42.0.0/24 masquerade
	}
}

$ oc get node -o jsonpath='{.items[0].metadata.annotations.k8s\.ovn\.org/node-ingress-snat-exclude-subnets}'
["10.45.0.0/16", "10.46.0.0/16"]
```

The OVN static routes can be listed through the `ovnkube-master` pod; each
C2CC-owned route carries the `microshift-c2cc` owner tag:

```bash
$ oc exec -n openshift-ovn-kubernetes daemonset/ovnkube-master -- \
    ovn-nbctl find Logical_Router_Static_Route \
    external_ids:k8s.ovn.org/owner-controller=microshift-c2cc
_uuid               : d6708885-c1ac-4001-b49e-d70a3652b0d0
external_ids        : {"k8s.ovn.org/owner-controller"=microshift-c2cc}
ip_prefix           : "10.45.0.0/16"
nexthop             : "192.168.1.20"
...

_uuid               : 5fd91bce-ab5d-483a-a8b9-50465dad4a96
external_ids        : {"k8s.ovn.org/owner-controller"=microshift-c2cc}
ip_prefix           : "10.46.0.0/16"
nexthop             : "192.168.1.20"
...
```

Finally, test connectivity from a pod — for example, resolve and reach a
Service on the remote cluster:

```bash
oc exec <pod> -- curl -s http://myservice.mynamespace.svc.cluster-b.remote:8080
```

## Cross-Cluster DNS

When a remote cluster has a `domain` configured, MicroShift's CoreDNS
forwards queries for that domain to the remote cluster's DNS service.
Any Service on the remote cluster is resolvable as:

```text
<service>.<namespace>.svc.<domain>
```

Both ClusterIP and headless Services work: ClusterIP Services resolve to the
remote service IP, headless Services resolve directly to remote pod IPs.

The forwarding is implemented as a per-remote server block in the CoreDNS
Corefile, which you can inspect in the `dns-default` ConfigMap:

```bash
$ oc get configmap -n openshift-dns dns-default -o jsonpath='{.data.Corefile}'
cluster-b.remote:5353 {
    bufsize 1232
    errors
    log . {
        class error
    }
    rewrite stop name suffix .cluster-b.remote .cluster.local answer auto
    forward . 10.46.0.10
    cache 10 {
        denial 9984 10
    }
}
.:5353 {
    ...
```

DNS responses are cached locally.
The cache TTLs are tunable via `clusterToCluster.dns.cacheTTL` (positive
answers) and `clusterToCluster.dns.cacheNegativeTTL` (NXDOMAIN/NODATA), both
defaulting to 10 seconds.
Lower them if remote service endpoints change rapidly; raise them to reduce
DNS load on the remote cluster.

## Configuration Reference

All settings live under the `clusterToCluster` section:

| Option | Default | Description |
|--------|---------|-------------|
| `remoteClusters` | `[]` | List of remote clusters. C2CC is disabled when empty. |
| `remoteClusters[].nextHop` | — | Next-hop IPs of the remote node. At most one IPv4 and one IPv6 address. Must not equal the local node IP or duplicate another remote. |
| `remoteClusters[].clusterNetwork` | — | Remote pod CIDRs. Must not overlap with local or other remote CIDRs. Minimum mask /8 (IPv4) or /32 (IPv6). |
| `remoteClusters[].serviceNetwork` | — | Remote service CIDRs. Same constraints as `clusterNetwork`; must have the same number of entries with matching IP families. |
| `remoteClusters[].domain` | `""` | Optional DNS suffix for the remote cluster. Must be a valid DNS subdomain, unique across remotes, and not `cluster.local`. |
| `probeInterval` | `10s` | Interval between health probes to each remote. Go duration string between `1s` and `5m`. |
| `dns.cacheTTL` | `10` | Max TTL (seconds) for positive DNS cache entries for remote domains. `0` disables caching. |
| `dns.cacheNegativeTTL` | `10` | Max TTL (seconds) for negative (NXDOMAIN/NODATA) DNS cache entries. `0` disables. |
| `routing.routeTableID` | `200` | Linux policy routing table for routes to remote CIDRs. Range 1–252; must differ from `serviceRouteTableID`. |
| `routing.serviceRouteTableID` | `201` | Linux policy routing table for service routes via the OVN management port. Range 1–252; must differ from `routeTableID`. |

The routing table IDs only need to be changed if other software on the host
(for example, NetworkManager dispatcher scripts or VPN tooling) already uses
tables 200/201.

## IPv6 and Dual-Stack

C2CC supports IPv6-only and dual-stack clusters.
The local and remote clusters must have compatible IP families: for every
remote CIDR family there must be a matching-family `nextHop` entry.

For a dual-stack remote, list both families:

```yaml
clusterToCluster:
  remoteClusters:
  - nextHop:
    - 192.168.1.20
    - 2001:db8:0:1::20
    clusterNetwork:
    - 10.45.0.0/16
    - fd01:0:0:1::/64
    serviceNetwork:
    - 10.46.0.0/16
    - fd02:0:0:1::/112
    domain: cluster-b.remote
```

`clusterNetwork` and `serviceNetwork` must have the same number of entries
with matching IP families at each position (at most one IPv4 and one IPv6
CIDR each).

## Disabling C2CC

Remove the C2CC drop-in (or the `remoteClusters` entries) and restart
MicroShift:

```bash
sudo rm /etc/microshift/config.d/50-c2cc.yaml
sudo systemctl restart microshift
```

On startup, the controller detects that C2CC is disabled and cleans up all
C2CC-owned state: OVN static routes, kernel routes and rules, nftables
rules, node annotations, the probe deployment, and the `RemoteCluster`
resources.
The journal logs `C2CC is disabled - attempting best effort cleanup`.

If you overrode `routing.routeTableID` or `routing.serviceRouteTableID`,
disable in two stages so the cleanup targets the right tables: first remove
only the `remoteClusters` entries while keeping the `routing` overrides and
restart, then remove the rest of the `clusterToCluster` section and restart
again.

Remember to remove the remote CIDRs from the firewall trusted zone as well.

## Considerations

### No authentication or encryption by default.
Any host that can reach (or spoof) the configured next-hop IP can inject traffic
into the cluster, and because SNAT is bypassed, that traffic can carry an
attacker-chosen source IP that matches NetworkPolicy allow rules. Use
[IPsec](./howto_c2cc_ipsec.md) together with the nftables enforcement described
there for production deployments — that combination provides encryption and
mutual authentication, and drops unauthenticated traffic (including
host-originated traffic to pods) at the network layer even when the IPsec
service itself is stopped or misconfigured.

### NetworkPolicies are your responsibility.
C2CC does not create NetworkPolicy resources. Namespaces with default-deny
ingress must explicitly allow the remote pod CIDRs (`ipBlock` selectors work,
since original pod source IPs are preserved).

### Restart required for changes.
The C2CC configuration is read at startup; any change (adding/removing remotes,
DNS TTLs, routing tables) requires a MicroShift restart.

### Changing routing table IDs leaves stale state.
MicroShift does not clean up C2CC routes on shutdown (by design, so workloads
keep network connectivity if the MicroShift process stops). If you change
`routing.*` table IDs while C2CC is enabled, the routes and rules in the old
tables remain until a reboot or manual cleanup (`ip route flush table <old-id>`).

### MTU
C2CC does not adjust MTU, and cross-cluster traffic itself adds no
encapsulation — it leaves the host as plain IP at the pod MTU.  MicroShift
defaults the pod MTU to the MTU of the physical interface; on jumbo-frame
networks it can be set explicitly by creating `/etc/microshift/ovn.yaml` with an
`mtu:` value (changing it requires a node reboot to take effect). When
encrypting the traffic with IPsec, leave headroom for the ESP overhead when
sizing the pod MTU; see the [IPsec guide](./howto_c2cc_ipsec.md).

### Scale
Configuration is static and full-mesh (each cluster lists every other cluster),
so it does not scale to large fleets without external configuration automation.
C2CC is validated with up to 3 interconnected clusters.

### Validation failures block startup
Invalid C2CC configuration (overlapping CIDRs, routing loops, bad masks,
duplicate next-hops or domains) prevents MicroShift from starting; check
`journalctl -u microshift` for the specific error.
