"""Brew (brewweb) scraping for MicroShift RPM builds."""

import logging
import re

import requests
import urllib3

BREW_PACKAGE_URL = "https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=82827"
ERRATA_PROBE_URL = "https://errata.devel.redhat.com/"

logger = logging.getLogger(__name__)

# Internal Red Hat services use certs not in the system trust store;
# verify=False is required when connecting via VPN.
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

# Module-level cache for the Brew HTML page
_cached_html = None


def check_vpn():
    """Check VPN connectivity by probing errata.devel.redhat.com.

    Returns:
        bool: True if VPN is connected.
    """
    try:
        response = requests.get(
            ERRATA_PROBE_URL, verify=False, timeout=5,
            allow_redirects=False,
        )
        return response.status_code < 500
    except requests.RequestException:
        return False


def _fetch_brew_page():
    """Fetch the Brew package page HTML, cached for the process lifetime.

    Returns:
        str: HTML content of the Brew page.

    Raises:
        requests.RequestException: On network failure.
    """
    global _cached_html
    if _cached_html is not None:
        return _cached_html

    response = requests.get(BREW_PACKAGE_URL, verify=False, timeout=30)
    response.raise_for_status()
    _cached_html = response.text
    return _cached_html


def fetch_latest_nightly_builds():
    """Parse Brew package page for latest MicroShift nightly builds per stream.

    Returns:
        dict: Keyed by stream (e.g., "4.21"), each value is a dict with:
            - nvr: Full NVR string
            - ocp_nightly_name: Mapped OCP nightly name
            - timestamp: ISO format timestamp string
    """
    html = _fetch_brew_page()

    # Pattern: microshift-X.Y.0~0.nightly_YYYY_MM_DD_HHMMSS
    pattern = re.compile(
        r"microshift-(\d+\.\d+)\.0~0\.nightly_(\d{4})_(\d{2})_(\d{2})_(\d{6})"
    )

    builds = {}
    for match in pattern.finditer(html):
        stream = match.group(1)
        year, month, day, time_part = match.group(2), match.group(3), match.group(4), match.group(5)
        hours = time_part[:2]
        minutes = time_part[2:4]
        seconds = time_part[4:6]

        timestamp = f"{year}-{month}-{day}T{hours}:{minutes}:{seconds}"
        ocp_name = f"{stream}.0-0.nightly-{year}-{month}-{day}-{time_part}"

        # Keep only the latest build per stream
        if stream not in builds or timestamp > builds[stream]["timestamp"]:
            builds[stream] = {
                "nvr": match.group(0),
                "ocp_nightly_name": ocp_name,
                "timestamp": timestamp,
            }

    return builds


def _find_rpms(brew_version):
    """Search Brew page for RPMs matching a Brew-format version string.

    Args:
        brew_version: Brew-format version, e.g., "4.22.0~ec.5" or "4.21.8".

    Returns:
        dict: {"found": True, "nvr": "...", "build_date": "..."} or {"found": False}.
    """
    html = _fetch_brew_page()
    escaped = re.escape(brew_version)
    pattern = re.compile(rf"(microshift-{escaped}-(\d{{12}})\.p0\.[^\s\"<]+)")

    match = pattern.search(html)
    if match:
        nvr = match.group(1)
        date_str = match.group(2)  # YYYYMMDDHHmm
        build_date = f"{date_str[:4]}-{date_str[4:6]}-{date_str[6:8]}"
        return {"found": True, "nvr": nvr, "build_date": build_date}

    return {"found": False}


def find_ecrc_rpms(version):
    """Search Brew package page for EC/RC RPMs matching a version.

    Args:
        version: e.g., "4.22.0-ec.5" or "4.22.0-rc.1".

    Returns:
        dict: {"found": True, "nvr": "...", "build_date": "..."} or {"found": False}.
    """
    # Normalize: 4.22.0-ec.5 -> 4.22.0~ec.5 (Brew uses tilde)
    brew_version = version.replace("-", "~")
    return _find_rpms(brew_version)


def find_zstream_rpms(version):
    """Search Brew package page for Z-stream RPMs matching a version.

    Args:
        version: e.g., "4.21.8".

    Returns:
        dict: {"found": True, "nvr": "...", "build_date": "..."} or {"found": False}.
    """
    return _find_rpms(version)
