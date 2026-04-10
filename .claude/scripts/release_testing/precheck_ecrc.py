#!/usr/bin/env python3
"""EC/RC discovery for MicroShift.

Auto-discovers the latest Engineering Candidate (EC) or Release Candidate (RC)
from the OCP release controller API and verifies MicroShift RPMs exist in Brew.

Usage: precheck_ecrc.py <EC|RC> [version]
"""

import argparse
import json
import logging
import re
import sys
from datetime import datetime

from lib import art_jira, brew, release_controller

logging.basicConfig(
    level=logging.INFO,
    format="%(levelname)s: %(message)s",
    stream=sys.stderr,
)
logger = logging.getLogger(__name__)


def parse_ecrc_version(name):
    """Parse an EC/RC version name.

    Args:
        name: e.g., "4.22.0-ec.5" or "4.22.0-rc.1".

    Returns:
        dict: {"type": "EC"|"RC", "base": "4.22.0", "num": 5, "minor": "4.22"} or None.
    """
    match = re.match(r"(4\.\d+\.\d+)-(ec|rc)\.(\d+)", name)
    if not match:
        return None
    return {
        "type": match.group(2).upper(),
        "base": match.group(1),
        "num": int(match.group(3)),
        "minor": ".".join(match.group(1).split(".")[:2]),
    }


def probe_next_versions(current_name):
    """Check if the next EC/RC exists on the release controller.

    Args:
        current_name: e.g., "4.22.0-ec.4".

    Returns:
        list[dict]: Each dict has keys: version, exists, phase.
    """
    parsed = parse_ecrc_version(current_name)
    if not parsed:
        return []

    candidates = []

    # Next increment of same type
    next_same = f"{parsed['base']}-{parsed['type'].lower()}.{parsed['num'] + 1}"
    candidates.append(next_same)

    # If EC, also check rc.1
    if parsed["type"] == "EC":
        candidates.append(f"{parsed['base']}-rc.1")

    results = []
    for candidate in candidates:
        release = release_controller.get_specific_release("4-dev-preview", candidate)
        if release:
            results.append({
                "version": candidate,
                "exists": True,
                "phase": release.get("phase", "unknown"),
            })
        else:
            results.append({
                "version": candidate,
                "exists": False,
            })

    return results


STATUS_DISPLAY = {
    "READY": "READY",
    "RPMS_NOT_BUILT": "RPMS NOT BUILT",
    "OCP_PENDING": "OCP PENDING",
    "NOT_FOUND": "NOT FOUND",
    "UNKNOWN": "UNKNOWN",
}


def format_text(data, verbose=False):
    """Format EC/RC result as pre-formatted text.

    Args:
        data: Output dict from the evaluation.
        verbose: If True, show extra detail lines (Next versions, etc.).

    Returns:
        str: Pre-formatted text output.
    """
    requested = data.get("type", "EC")
    actual = data.get("actual_type", requested)
    version = data.get("version", "")
    status = data.get("status", "UNKNOWN")
    ocp_phase = data.get("ocp_phase", "unknown")
    brew_data = data.get("brew")

    # Determine brew display
    if brew_data is None:
        brew_str = "--"
    elif brew_data.get("error"):
        brew_str = brew_data["error"]
    elif brew_data.get("found"):
        brew_str = brew_data.get("build_date", "built")
    else:
        brew_str = "not found"

    # Type mismatch: no active EC/RC
    if data.get("type_mismatch"):
        return f"OK      [{version}] (No active {requested}, latest is {actual})"

    # Determine status label
    if status == "READY":
        label = "OK"
    elif status == "ERROR":
        label = "ERROR"
    else:
        label = "ASK ART"

    STATUS_WIDTH = 7
    if status == "ERROR":
        detail = data.get("detail", "unknown error")
        line = f"{label:<{STATUS_WIDTH}} [{version}] ({detail})"
    elif status == "OCP_PENDING":
        line = f"{label:<{STATUS_WIDTH}} [{version}] [OCP: {ocp_phase}]"
    elif status == "NOT_FOUND":
        line = f"{label:<{STATUS_WIDTH}} [{version}] (not on release controller)"
    else:
        line = f"{label:<{STATUS_WIDTH}} [{version}] [OCP: {ocp_phase} | Brew: {brew_str}]"

    if not verbose:
        return line

    # Verbose: append extra detail lines
    lines = [line]
    nexts = data.get("next_versions", [])
    if nexts:
        next_parts = []
        for nv in nexts:
            if nv.get("exists"):
                next_parts.append(f"{nv['version']} ({nv.get('phase', 'found')})")
            else:
                next_parts.append(f"{nv['version']} (not found)")
        lines.append(f"  Next: {', '.join(next_parts)}")

    return "\n".join(lines)


def main():
    parser = argparse.ArgumentParser(description="MicroShift EC/RC discovery")
    parser.add_argument("type", choices=["EC", "RC", "ec", "rc"], help="EC or RC")
    parser.add_argument("version", nargs="?", help="Specific version, e.g., 4.22.0-ec.4")
    parser.add_argument("--json", action="store_true", dest="json_output",
                        help="Output raw JSON instead of formatted text")
    parser.add_argument("--verbose", action="store_true",
                        help="Show extra detail lines")
    args = parser.parse_args()

    release_type = args.type.upper()

    # Step 1: Discover latest EC/RC
    if args.version:
        logger.info("Fetching specific release: %s", args.version)
        api_error = None
        try:
            release = release_controller.get_specific_release("4-dev-preview", args.version)
        except Exception as e:
            logger.warning("Release controller API error: %s", e)
            release = None
            api_error = str(e)
        if not release:
            status = "ERROR" if api_error else "NOT_FOUND"
            detail = api_error or f"{args.version} not found on release controller"
            not_found = {
                "command": "precheck_ecrc",
                "type": release_type,
                "version": args.version,
                "status": status,
                "detail": detail,
                "timestamp": datetime.now().isoformat(),
            }
            if args.json_output:
                print(json.dumps(not_found, indent=2))
            else:
                print(format_text(not_found, verbose=args.verbose))
            return
    else:
        logger.info("Discovering latest from 4-dev-preview stream...")
        try:
            release = release_controller.get_latest_dev_preview()
        except Exception as e:
            logger.error("Failed to query release controller: %s", e)
            error_data = {
                "command": "precheck_ecrc",
                "type": release_type,
                "version": "",
                "status": "ERROR",
                "detail": f"Release controller unavailable: {e}",
                "timestamp": datetime.now().isoformat(),
            }
            if args.json_output:
                print(json.dumps(error_data, indent=2))
            else:
                print(format_text(error_data, verbose=args.verbose))
            sys.exit(1)

    version_name = release.get("name", "")
    phase = release.get("phase", "unknown")
    parsed = parse_ecrc_version(version_name)

    # Detect type mismatch
    actual_type = parsed["type"] if parsed else "unknown"
    type_mismatch = actual_type != release_type
    if type_mismatch:
        logger.info(
            "Requested %s but latest is %s (%s)",
            release_type, actual_type, version_name,
        )

    # Step 2: ART Jira lookup (optional)
    logger.info("Looking up ART Jira ticket...")
    art_ticket = art_jira.query_art_ecrc(version_name)

    # Step 3: Verify Brew RPMs
    brew_result = None
    if phase == "Accepted":
        logger.info("Checking Brew for MicroShift RPMs...")
        if brew.check_vpn():
            brew_result = brew.find_ecrc_rpms(version_name)
        else:
            brew_result = {"found": False, "error": "VPN not connected"}

    # Step 4: Probe next versions
    logger.info("Probing for next versions...")
    next_versions = probe_next_versions(version_name)

    # Determine overall status
    if phase != "Accepted":
        status = "OCP_PENDING"
    elif brew_result and brew_result.get("found"):
        status = "READY"
    elif brew_result and not brew_result.get("found"):
        status = "RPMS_NOT_BUILT"
    else:
        status = "UNKNOWN"

    output = {
        "command": "precheck_ecrc",
        "type": release_type,
        "actual_type": actual_type,
        "type_mismatch": type_mismatch,
        "version": version_name,
        "minor": parsed["minor"] if parsed else "",
        "ocp_phase": phase,
        "art_ticket": art_ticket,
        "brew": brew_result,
        "next_versions": next_versions,
        "status": status,
        "timestamp": datetime.now().isoformat(),
    }
    if args.json_output:
        print(json.dumps(output, indent=2))
    else:
        print(format_text(output, verbose=args.verbose))


if __name__ == "__main__":
    main()
