#!/usr/bin/env python3
"""Send a UDP datagram with the DF (Don't Fragment) bit set.

Uses IP_PMTUDISC_DO (IPv4) or IPV6_DONTFRAG (IPv6) so the kernel
rejects datagrams that exceed the path MTU with EMSGSIZE instead of
fragmenting them.  Auto-detects the address family from the
destination address.

Uses UDP instead of ICMP to avoid requiring NET_RAW capability.
UDP overhead (IP+UDP) matches ICMP overhead (IP+ICMP): 28B for IPv4,
48B for IPv6, so payload sizes are equivalent to ping -s values
within each address family.

Usage: python3 df-bit-send.py <dest_ip> <payload_size>
Exit:  prints "OK" on success, raises OSError on rejection.
"""

import socket
import sys

IPPROTO_IPV6 = 41
IPV6_DONTFRAG = 62
IP_MTU_DISCOVER = 10
IP_PMTUDISC_DO = 2

dest = sys.argv[1]
size = int(sys.argv[2])

family = socket.AF_INET6 if ":" in dest else socket.AF_INET
sock = socket.socket(family, socket.SOCK_DGRAM)

if family == socket.AF_INET6:
    sock.setsockopt(IPPROTO_IPV6, IPV6_DONTFRAG, 1)
else:
    sock.setsockopt(socket.IPPROTO_IP, IP_MTU_DISCOVER, IP_PMTUDISC_DO)

try:
    sock.sendto(b"A" * size, (dest, 9999))
    print("OK")
except OSError as e:
    print(f"FAIL: {e}")
finally:
    sock.close()
