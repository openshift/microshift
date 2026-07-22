# Encrypting C2CC Traffic with IPsec

MicroShift Cluster-to-Cluster Connectivity (C2CC) routes cross-cluster pod and service traffic as raw IP between nodes.
This traffic traverses the physical network unencrypted by default.
You can use IPsec to protect it using standard Linux tools.

MicroShift does not configure or manage IPsec.
Setting up and maintaining the IPsec tunnels is the responsibility of the system administrator.
This guide provides a minimal working example using Libreswan in tunnel mode to help you get started.

For comprehensive IPsec/VPN documentation, see [Setting up an IPsec VPN](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/9/html/configuring_and_managing_networking/setting-up-an-ipsec-vpn_configuring-and-managing-networking) in the RHEL documentation.

## Prerequisites

- Two or more RHEL hosts running MicroShift with C2CC configured (non-overlapping pod and service CIDRs, `clusterToCluster.remoteClusters` populated in each node's config).
  See [Cluster-to-Cluster Connectivity (C2CC)](./howto_c2cc.md) for the C2CC setup.
- IP connectivity between the hosts on the underlay network.
- Libreswan installed on every host:

```bash
sudo dnf install -y libreswan
```

## Firewall

Open the firewall for IKE negotiation and ESP:

```bash
sudo firewall-cmd --permanent --zone=public --add-service=ipsec
sudo firewall-cmd --reload
```

This allows UDP ports 500 and 4500 (IKE/NAT-T) and IP protocol 50 (ESP).

## Generate a Pre-Shared Key

Generate a shared secret on one host and distribute it to all others through a secure channel:

```bash
openssl rand -hex 32
```

## Configure Libreswan

The examples below assume a two-cluster setup:

| Host    | Underlay IP   | Pod CIDR       | Service CIDR   |
|---------|---------------|----------------|----------------|
| Host A  | 192.168.1.10  | 10.42.0.0/16   | 10.43.0.0/16   |
| Host B  | 192.168.1.20  | 10.45.0.0/16   | 10.46.0.0/16   |

### Secrets

On **each** host, create `/etc/ipsec.d/c2cc.secrets`:

```conf
192.168.1.10 192.168.1.20 : PSK "<your-hex-key>"
```

Set permissions:

```bash
sudo chmod 600 /etc/ipsec.d/c2cc.secrets
sudo restorecon -v /etc/ipsec.d/c2cc.secrets
```

### Connection definition

On **Host A**, create `/etc/ipsec.d/c2cc-tunnel.conf`:

```conf
conn c2cc-to-host-b
    type=tunnel
    authby=secret
    left=192.168.1.10
    right=192.168.1.20
    leftsubnets={10.42.0.0/16 10.43.0.0/16}
    rightsubnets={10.45.0.0/16 10.46.0.0/16}
    auto=start
    ike=aes256-sha2_256-modp2048
    esp=aes256-sha2_256
    failureshunt=drop
    negotiationshunt=drop
    ikev2=insist
```

On **Host B**, create the same file with `left`/`right` and subnet values swapped:

```conf
conn c2cc-to-host-a
    type=tunnel
    authby=secret
    left=192.168.1.20
    right=192.168.1.10
    leftsubnets={10.45.0.0/16 10.46.0.0/16}
    rightsubnets={10.42.0.0/16 10.43.0.0/16}
    auto=start
    ike=aes256-sha2_256-modp2048
    esp=aes256-sha2_256
    failureshunt=drop
    negotiationshunt=drop
    ikev2=insist
```

Key parameters:

- **`type=tunnel`** -- Tunnel mode encrypts the original IP packet and wraps it in a new IP header. This is required because C2CC traffic uses pod/service CIDRs as source and destination, which are not routable on the underlay.
- **`leftsubnets` / `rightsubnets`** -- Must match the pod and service CIDRs configured in MicroShift. Each `{cidr1 cidr2}` pair creates one child SA per local/remote CIDR combination.
- **`auto=start`** -- Bring the tunnel up automatically when the IPsec service starts.
- **`failureshunt=drop` / `negotiationshunt=drop`** -- Drop traffic that matches the tunnel selectors if the SA fails or is still negotiating, preventing fallback to plaintext.
- **`ikev2=insist`** -- Require IKEv2. IKEv1 is not recommended.

### Three or more clusters

For a full mesh of N clusters, each host needs a connection definition and a secrets entry for every remote host.
For example, with three hosts, Host A would have two `conn` blocks (one for Host B, one for Host C) and two secrets entries.

## Start IPsec

Initialize the NSS database (first time only) and start the service:

```bash
sudo ipsec checknss
sudo systemctl enable --now ipsec
```

## Verify the Tunnels

Check that tunnel SAs are established:

```bash
$ sudo ipsec trafficstatus
#8: "c2cc-to-host-b/1x1", type=ESP, add_time=1783531540, inBytes=1842, outBytes=1842, maxBytes=2^63B, id='192.168.1.20'
#10: "c2cc-to-host-b/1x2", type=ESP, add_time=1783531540, inBytes=3386, outBytes=3707, maxBytes=2^63B, id='192.168.1.20'
#9: "c2cc-to-host-b/2x1", type=ESP, add_time=1783531540, inBytes=3287, outBytes=3052, maxBytes=2^63B, id='192.168.1.20'
#11: "c2cc-to-host-b/2x2", type=ESP, add_time=1783531540, inBytes=0, outBytes=0, maxBytes=2^63B, id='192.168.1.20'
```

You should see `type=ESP` entries for each subnet pair — the `/1x1`, `/1x2`, `/2x1`, `/2x2` suffixes are the combinations of the `leftsubnets`/`rightsubnets` selectors.
Each remote produces 4 child SAs (2 local CIDRs × 2 remote CIDRs), so a two-cluster setup shows 4 and a three-cluster full mesh shows 8 per host.
On RHEL 9 (Libreswan 4.x), each line is additionally prefixed with a `006` status code.

Verify XFRM state is populated and that traffic actually flows through the SAs — the byte counters must increase while cross-cluster traffic is running:

```bash
$ sudo ip -s xfrm state
src 192.168.1.10 dst 192.168.1.20
	proto esp spi 0xca56d890(3394689168) reqid 16409(0x00004019) mode tunnel
	replay-window 0 seq 0x00000000 flag af-unspec esn (0x10100000)
	auth-trunc hmac(sha256) 0x... (256 bits) 128
	enc cbc(aes) 0x... (256 bits)
	lastused 2026-07-08 17:25:53
	anti-replay esn context:
	 seq-hi 0x0, seq 0x0, oseq-hi 0x0, oseq 0xc
	 replay_window 0, bitmap-length 0
	dir out
	lifetime config:
	  ...
	lifetime current:
	  832(bytes), 12(packets)
	  add 2026-07-08 17:25:40 use 2026-07-08 17:25:43
	stats:
	  replay-window 0 replay 0 failed 0
src 192.168.1.20 dst 192.168.1.10
	proto esp spi 0xa1a985b4(2712241588) reqid 16409(0x00004019) mode tunnel
	...
```

There is one state entry per direction per child SA; the keys are shown in full in the real output (truncated to `0x...` here).
To watch the total encrypted byte count as a single number (useful for before/after comparison around a test transfer):

```bash
$ sudo ip -s xfrm state | awk '/bytes/{gsub(/[^0-9]/,"",$1); sum+=$1} END{print sum+0}'
6231
```

You can also capture packets on the wire to confirm ESP encapsulation — cross-cluster traffic must appear as ESP, never as plaintext TCP/UDP between pod IPs:

```bash
$ sudo tcpdump -ni enp1s0 -c 10 esp
17:25:53.766575 IP 192.168.1.10 > 192.168.1.20: ESP(spi=0xca56d890,seq=0x7), length 104
17:25:53.766989 IP 192.168.1.20 > 192.168.1.10: ESP(spi=0xa1a985b4,seq=0x7), length 104
...
```

## Enforce Encryption with nftables

The `failureshunt=drop` and `negotiationshunt=drop` options prevent plaintext fallback, but they are enforced by Libreswan itself — they offer no protection if the IPsec service is stopped or a connection definition is removed.
For defense in depth, add kernel-level nftables rules that drop any traffic reaching local pod or service CIDRs without IPsec protection (requires kernel 5.10+ for the `meta ipsec missing` match).
The rules must be attached to the `forward` hook: inbound cross-cluster packets are not addressed to the host itself but forwarded by the host into the OVN network, so an `input`-hook chain would never see them.

On **Host A**:

```bash
sudo nft add table inet c2cc_ipsec
sudo nft 'add chain inet c2cc_ipsec enforce { type filter hook forward priority -150; policy accept; }'
sudo nft add rule inet c2cc_ipsec enforce ip daddr 10.42.0.0/16 meta ipsec missing counter drop
sudo nft add rule inet c2cc_ipsec enforce ip daddr 10.43.0.0/16 meta ipsec missing counter drop
```

On **Host B**, use its own local CIDRs (`10.45.0.0/16`, `10.46.0.0/16`).

With these rules in place, inbound cross-cluster traffic is accepted only when it arrived through an IPsec SA — even if IPsec is misconfigured or stopped on the sending side, unencrypted packets are dropped rather than delivered.
The `counter` keyword lets you watch the drop count — it should stay at zero while the tunnels are healthy:

```bash
$ sudo nft list table inet c2cc_ipsec
table inet c2cc_ipsec {
	chain enforce {
		type filter hook forward priority -150; policy accept;
		ip daddr 10.42.0.0/16 meta ipsec missing counter packets 0 bytes 0 drop
		ip daddr 10.43.0.0/16 meta ipsec missing counter packets 0 bytes 0 drop
	}
}
```

Note that these rules are not persistent across reboots.
To make them permanent, add them to `/etc/sysconfig/nftables.conf` and enable the `nftables` service, or ship them via your image build.

## Considerations

- **IPsec adds overhead.** ESP tunnel mode adds roughly 73-93 bytes per packet (ESP header/trailer plus the outer IP header, depending on the cipher suite). Cross-cluster C2CC traffic travels as plain routed IP — it is not Geneve-encapsulated between the hosts — so the largest packet entering the tunnel is the pod MTU, which MicroShift by default sets equal to the MTU of the physical interface. Size the pod MTU so that it plus the ESP overhead fits within the physical MTU. For example, on a jumbo-frame network (physical MTU 9000), set the pod MTU to 8900 by creating `/etc/microshift/ovn.yaml` with `mtu: 8900` — changing this value requires a node reboot to take effect. Also verify that path MTU discovery works between the pods and that near-MTU-sized payloads pass through the tunnel.
- **IPsec also blocks unauthenticated hosts.** Because the tunnel selectors are scoped to pod and service CIDRs and the shunt policies drop unmatched traffic, a host without valid IPsec credentials — including the peer host itself — cannot reach pods on a remote cluster. This closes the host-to-pod path that plain C2CC leaves open at the routing level.
- **Tunnel recovery.** If IPsec is restarted on one host, tunnels renegotiate automatically when `auto=start` is set. No MicroShift restart is required.
- **Certificates.** This guide uses pre-shared keys for simplicity. For production deployments, consider certificate-based authentication. See the [RHEL VPN documentation](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/9/html/configuring_and_managing_networking/setting-up-an-ipsec-vpn_configuring-and-managing-networking) for details.
- **Policy enforcement.** The example connection definitions include `failureshunt=drop` and `negotiationshunt=drop` to prevent traffic from falling back to plaintext when the tunnel is down or still negotiating. If you remove these options, traffic matching the tunnel selectors will be sent unencrypted whenever the SA is unavailable.
