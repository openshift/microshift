"""Red Hat Product Lifecycle API client."""

import logging

import requests

LIFECYCLE_URL = (
    "https://access.redhat.com/product-life-cycles/api/v1/products/"
    "?name=Red%20Hat%20build%20of%20MicroShift"
)

ACTIVE_TYPES = {"Full Support", "Maintenance Support"}

logger = logging.getLogger(__name__)


def _find_end_date(phases, phase_type):
    """Find the end date for a lifecycle phase type.

    Args:
        phases: List of phase dicts from the API.
        phase_type: The version's current type, e.g., "Full Support".

    Returns:
        str: End date in YYYY-MM-DD format, or empty string.
    """
    # Map the type field to the phase name in the phases array
    type_to_phase = {
        "Full Support": "Full support",
        "Maintenance Support": "Maintenance support",
        "Extended Support": "Extended update support",
        "End of life": "Maintenance support",
    }
    target_phase = type_to_phase.get(phase_type, "")

    for phase in phases:
        if phase.get("name", "").lower() == target_phase.lower():
            end_date = phase.get("end_date", "")
            if end_date and "T" in str(end_date):
                return str(end_date).split("T")[0]
            return str(end_date) if end_date != "N/A" else ""
    return ""


def fetch_lifecycle_data():
    """Fetch product lifecycle data from access.redhat.com.

    Returns:
        list[dict]: Each dict has keys: version, phase, end_date, active.
            Phase is the current lifecycle status from the API's 'type' field:
            "Full Support", "Maintenance Support", "Extended Support", "End of life".
    """
    response = requests.get(LIFECYCLE_URL, timeout=30)
    response.raise_for_status()

    data = response.json()
    results = []
    for product in data.get("data", []):
        for version_entry in product.get("versions", []):
            # The 'type' field contains the current lifecycle status
            raw_version = version_entry.get("name", "")
            phase_type = version_entry.get("type", "")
            phases = version_entry.get("phases", [])
            end_date = _find_end_date(phases, phase_type)

            results.append({
                "version": raw_version,
                "phase": phase_type,
                "end_date": end_date,
                "active": phase_type in ACTIVE_TYPES,
            })

    return results


def is_version_active(minor, lifecycle_data):
    """Check if a minor version (e.g., '4.21') is in Full or Maintenance Support.

    Args:
        minor: Minor version string, e.g., "4.21".
        lifecycle_data: Output from fetch_lifecycle_data().

    Returns:
        bool: True if the version is active.
    """
    entry = get_lifecycle_status(minor, lifecycle_data)
    return entry["active"] if entry else False


def get_lifecycle_status(minor, lifecycle_data):
    """Return the lifecycle entry for a minor version.

    Args:
        minor: Minor version string, e.g., "4.21".
        lifecycle_data: Output from fetch_lifecycle_data().

    Returns:
        dict or None: The matching lifecycle entry, or None if not found.
    """
    for entry in lifecycle_data:
        if entry["version"] == minor:
            return entry
    return None
