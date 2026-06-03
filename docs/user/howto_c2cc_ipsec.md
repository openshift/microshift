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
sudo ipsec trafficstatus
```

You should see output containing `type=ESP` entries for each subnet pair.
For a two-cluster setup with 2 local CIDRs and 2 remote CIDRs, expect 4 child SAs.

Verify XFRM state is populated:

```bash
ip xfrm state
```

You can also capture packets on the wire to confirm ESP encapsulation:

```bash
sudo tcpdump -i enp1s0 -c 10 esp
```

## Considerations

- **IPsec adds overhead.** ESP tunnel mode adds approximately 36-52 bytes per packet. If you experience MTU issues, verify that path MTU discovery is working or adjust MTU settings accordingly.
- **Tunnel recovery.** If IPsec is restarted on one host, tunnels renegotiate automatically when `auto=start` is set. No MicroShift restart is required.
- **Certificates.** This guide uses pre-shared keys for simplicity. For production deployments, consider certificate-based authentication. See the [RHEL VPN documentation](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/9/html/configuring_and_managing_networking/setting-up-an-ipsec-vpn_configuring-and-managing-networking) for details.
- **Policy enforcement.** The example connection definitions include `failureshunt=drop` and `negotiationshunt=drop` to prevent traffic from falling back to plaintext when the tunnel is down or still negotiating. If you remove these options, traffic matching the tunnel selectors will be sent unencrypted whenever the SA is unavailable.
