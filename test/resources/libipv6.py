import ipaddress
import socket
from robot.libraries.BuiltIn import BuiltIn

_log = BuiltIn().log


def is_ipv6(addr: str) -> bool:
    """
    Determine if the provided address is IPv6. If addr is a hostname then resolve it
    to check if the resulting IP address is IPv6.
    """
    try:
        parsed_addr = ipaddress.ip_network(addr)
        return parsed_addr.version == 6
    except ValueError:
        pass
    try:
        socket.getaddrinfo(addr, None, socket.AF_INET6)
    except socket.gaierror:
        return False
    return True


def must_be_ipv6(addr: str) -> None:
    BuiltIn().should_be_true(is_ipv6(addr))


def must_not_be_ipv6(addr: str) -> None:
    BuiltIn().should_not_be_true(is_ipv6(addr))


def add_brackets_if_ipv6(ip: str) -> str:
    """
    Add square brackets to the given IP if its ipv6 for later use in other tools (e.g. curl)
    """

    # If it's a hostname (not a valid IP address), return it as-is
    try:
        ipaddress.ip_address(ip)
    except ValueError:
        return ip

    if is_ipv6(ip):
        return f"[{ip}]"
    return ip
