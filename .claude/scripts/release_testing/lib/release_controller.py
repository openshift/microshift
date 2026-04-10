"""OCP Release Controller API client."""

import logging

import requests

RC_BASE_URL = "https://openshift-release.apps.ci.l2s4.p1.openshiftapps.com"

logger = logging.getLogger(__name__)


def get_latest_nightly(stream):
    """Fetch latest accepted nightly for a stream.

    Args:
        stream: Minor version, e.g., "4.21".

    Returns:
        dict: Release info with keys: name, phase, pullSpec, downloadURL.
    """
    url = f"{RC_BASE_URL}/api/v1/releasestream/{stream}.0-0.nightly/latest"
    response = requests.get(url, timeout=30)
    response.raise_for_status()
    return response.json()


def get_latest_dev_preview():
    """Fetch latest from 4-dev-preview stream (EC/RC).

    Returns:
        dict: Release info with keys: name, phase, pullSpec, downloadURL.
    """
    url = f"{RC_BASE_URL}/api/v1/releasestream/4-dev-preview/latest"
    response = requests.get(url, timeout=30)
    response.raise_for_status()
    return response.json()


def get_specific_release(stream, version):
    """Fetch a specific release from a stream.

    Args:
        stream: Stream name, e.g., "4-dev-preview".
        version: Release version, e.g., "4.22.0-ec.5".

    Returns:
        dict or None: Release info, or None if not found.
    """
    url = f"{RC_BASE_URL}/api/v1/releasestream/{stream}/release/{version}"
    response = requests.get(url, timeout=30)
    if response.status_code == 404:
        return None
    response.raise_for_status()
    return response.json()


def check_ocp_payload_accepted(version):
    """Check if OCP payload for X.Y.Z is accepted on the release controller.

    Uses the 4-stable stream API to look up the exact version.

    Args:
        version: Full version, e.g., "4.21.8".

    Returns:
        str: "available", "not_available", or "failed".
    """
    parts = version.split(".")
    if len(parts) < 2:
        return "not_available"

    minor = f"{parts[0]}.{parts[1]}"

    # Try the stable stream API for an exact match
    for stream in [f"{minor}.0-0.nightly", "4-stable"]:
        release = get_specific_release(stream, version)
        if release:
            phase = release.get("phase", "")
            if phase == "Accepted":
                return "available"
            if phase == "Rejected":
                return "failed"
            return "pending"

    return "not_available"
