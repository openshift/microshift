"""Red Hat Catalog (Pyxis) API client for checking published MicroShift versions."""

import logging
import re
from concurrent.futures import ThreadPoolExecutor

import requests

PYXIS_BASE_URL = "https://catalog.redhat.com/api/containers/v1"
BOOTC_REPO_PATH = (
    "repositories/registry/registry.access.redhat.com"
    "/repository/openshift4/microshift-bootc-rhel9/images"
)

logger = logging.getLogger(__name__)


def _fetch_page(page):
    """Fetch a single page of bootc images from Pyxis.

    Args:
        page: Page number (0-indexed).

    Returns:
        str: Response text.
    """
    url = f"{PYXIS_BASE_URL}/{BOOTC_REPO_PATH}"
    params = {
        "filter": "architecture==amd64",
        "page_size": 100,
        "page": page,
    }
    response = requests.get(url, params=params, timeout=30)
    response.raise_for_status()
    return response.text


def _scan_pages_for_versions(minor_version, pages=5):
    """Scan Pyxis pages for all published z-stream versions of a minor version.

    Args:
        minor_version: e.g., "4.21".
        pages: Number of pages to scan.

    Returns:
        set[int]: Set of z-stream numbers found (e.g., {0, 1, 7, 13}).
    """
    pattern = re.compile(rf"assembly\.{re.escape(minor_version)}\.(\d+)")
    found_z = set()

    with ThreadPoolExecutor(max_workers=pages) as executor:
        futures = [executor.submit(_fetch_page, p) for p in range(pages)]
        for future in futures:
            try:
                text = future.result()
                for match in pattern.finditer(text):
                    found_z.add(int(match.group(1)))
            except requests.RequestException as e:
                logger.warning("Pyxis page fetch failed: %s", e)

    return found_z


def is_version_published(version, pages=5):
    """Check if MicroShift version X.Y.Z has a published bootc image.

    Looks for 'assembly.X.Y.Z' tag in the Pyxis catalog API.

    Args:
        version: Full version string, e.g., "4.21.7".
        pages: Number of pages to paginate (page_size=100).

    Returns:
        bool: True if 'assembly.X.Y.Z' found in any page.
    """
    pattern = f"assembly.{version}"

    with ThreadPoolExecutor(max_workers=pages) as executor:
        futures = [executor.submit(_fetch_page, p) for p in range(pages)]
        for future in futures:
            try:
                text = future.result()
                if pattern in text:
                    return True
            except requests.RequestException as e:
                logger.warning("Pyxis page fetch failed: %s", e)

    return False


def find_latest_published_zstream(minor_version, pages=5):
    """Find the highest published z-stream for a minor version.

    Args:
        minor_version: e.g., "4.21".
        pages: Number of pages to scan.

    Returns:
        dict or None: {"version": "4.21.13", "z": 13} or None if not found.
    """
    found_z = _scan_pages_for_versions(minor_version, pages)
    if not found_z:
        return None

    highest_z = max(found_z)
    return {
        "version": f"{minor_version}.{highest_z}",
        "z": highest_z,
    }


def find_all_published_versions(minor_version, pages=5):
    """Find all published z-stream versions for a minor version.

    Args:
        minor_version: e.g., "4.21".
        pages: Number of pages to scan.

    Returns:
        list[str]: Sorted list of published versions, e.g., ["4.21.0", "4.21.1", ...].
    """
    found_z = _scan_pages_for_versions(minor_version, pages)
    return sorted(f"{minor_version}.{z}" for z in found_z)
