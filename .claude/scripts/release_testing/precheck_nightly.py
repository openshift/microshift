#!/usr/bin/env python3
"""Nightly build gap detection for MicroShift.

Compares Brew nightly RPM timestamps against OCP accepted nightlies
to detect missing or lagging MicroShift builds.

Usage: precheck_nightly.py [minor_version]
"""

import argparse
import json
import logging
import re
import sys
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import datetime

from lib import brew, git_ops, lifecycle, release_controller

logging.basicConfig(
    level=logging.INFO,
    format="%(levelname)s: %(message)s",
    stream=sys.stderr,
)
logger = logging.getLogger(__name__)


_NIGHTLY_TS_RE = re.compile(
    r"nightly-(\d{4})-(\d{2})-(\d{2})-(\d{2})(\d{2})(\d{2})"
)


def parse_nightly_timestamp(nightly_name):
    """Extract ISO timestamp from an OCP nightly name.

    Args:
        nightly_name: e.g., "4.21.0-0.nightly-2026-03-29-021947".

    Returns:
        str: ISO timestamp, e.g., "2026-03-29T02:19:47", or "" if no match.
    """
    match = _NIGHTLY_TS_RE.search(nightly_name)
    if not match:
        return ""
    y, mo, d, h, mi, s = match.groups()
    return f"{y}-{mo}-{d}T{h}:{mi}:{s}"


def format_gap(gap_hours):
    """Format gap hours as human-readable string.

    Args:
        gap_hours: Number of hours.

    Returns:
        str: e.g., "3d 19h", "13h", "0h".
    """
    if gap_hours < 0:
        return "0h"
    days = int(gap_hours // 24)
    hours = int(gap_hours % 24)
    if days > 0:
        return f"{days}d {hours}h"
    return f"{hours}h"


def classify_gap(gap_hours):
    """Classify the build gap.

    Args:
        gap_hours: Number of hours between OCP and Brew timestamps.

    Returns:
        str: "OK" or "ASK ART".
    """
    if gap_hours <= 24:
        return "OK"
    return "ASK ART"


def check_branch(stream, brew_builds):
    """Check one branch for nightly build gaps.

    Args:
        stream: Minor version, e.g., "4.21".
        brew_builds: Output from brew.fetch_latest_nightly_builds().

    Returns:
        dict: Branch status with keys: stream, branch, status, ocp_nightly,
            ocp_timestamp, brew_build, brew_timestamp, gap_hours, gap_display.
    """
    result = {"stream": stream, "branch": f"release-{stream}"}

    # Fetch OCP nightly
    try:
        nightly = release_controller.get_latest_nightly(stream)
        result["ocp_nightly"] = nightly.get("name", "")
        result["ocp_phase"] = nightly.get("phase", "unknown")
        result["ocp_timestamp"] = parse_nightly_timestamp(nightly.get("name", ""))
    except Exception as e:
        logger.warning("Failed to fetch OCP nightly for %s: %s", stream, e)
        result["ocp_nightly"] = None
        result["ocp_error"] = str(e)
        result["status"] = "ERROR"
        return result

    # Brew build
    brew_data = brew_builds.get(stream)
    if not brew_data:
        result["brew_build"] = None
        result["brew_timestamp"] = None
        result["status"] = "ASK ART"
        result["gap_display"] = "no Brew builds"
        return result

    result["brew_build"] = brew_data["nvr"]
    result["brew_timestamp"] = brew_data["timestamp"]

    # Compare timestamps
    try:
        ocp_dt = datetime.fromisoformat(result["ocp_timestamp"])
        brew_dt = datetime.fromisoformat(brew_data["timestamp"])
        gap_hours = (ocp_dt - brew_dt).total_seconds() / 3600
        result["gap_hours"] = round(gap_hours, 1)
        result["gap_display"] = format_gap(gap_hours)
        result["status"] = classify_gap(gap_hours)
    except (ValueError, TypeError) as e:
        logger.warning("Failed to compare timestamps for %s: %s", stream, e)
        result["status"] = "ERROR"
        result["gap_display"] = "timestamp parse error"

    return result


def _format_ts(iso_ts):
    """Format ISO timestamp as 'YYYY-MM-DD HH:MM'."""
    if not iso_ts:
        return "--"
    return iso_ts.replace("T", " ")[:16]


def format_text(branches, verbose=False):
    """Format branch results as aligned text output.

    Format: STATUS [branch] [OCP: ts | Brew: ts] (gap)

    Args:
        branches: List of branch result dicts.
        verbose: If True, show extra detail lines (NVR, OCP nightly name).

    Returns:
        str: Pre-formatted text output.
    """
    STATUS_WIDTH = 7  # len("ASK ART")
    output_lines = []

    for b in branches:
        status = b.get("status", "UNKNOWN")
        branch = b.get("branch", f"release-{b['stream']}")

        if status in ("EOL", "EUS"):
            phase = b.get("lifecycle_phase", status)
            end = b.get("end_date", "N/A")
            line = f"{status:<{STATUS_WIDTH}} [{branch}] ({phase}, until {end})"
        elif status == "ERROR":
            error = b.get("ocp_error", b.get("gap_display", "unknown error"))
            line = f"{'ERROR':<{STATUS_WIDTH}} [{branch}] ({error})"
        else:
            ocp_ts = _format_ts(b.get("ocp_timestamp"))
            brew_ts = _format_ts(b.get("brew_timestamp"))
            gap = b.get("gap_display", "?")
            line = f"{status:<{STATUS_WIDTH}} [{branch}] [OCP: {ocp_ts} | Brew: {brew_ts}] ({gap})"

        output_lines.append(line)

        if verbose and status not in ("EOL", "EUS", "ERROR"):
            ocp_name = b.get("ocp_nightly", "")
            brew_nvr = b.get("brew_build", "")
            if ocp_name:
                output_lines.append(f"  OCP: {ocp_name}")
            if brew_nvr:
                output_lines.append(f"  Brew: {brew_nvr}")

    if not output_lines:
        return "No branches to check."

    return "\n".join(output_lines)


def main():
    parser = argparse.ArgumentParser(description="MicroShift nightly build gap detection")
    parser.add_argument("version", nargs="?", help="Minor version, e.g., 4.21")
    parser.add_argument("--json", action="store_true", dest="json_output",
                        help="Output raw JSON instead of formatted text")
    parser.add_argument("--verbose", action="store_true",
                        help="Show extra detail lines (NVR, OCP nightly name)")
    args = parser.parse_args()

    # Step 1: Fetch lifecycle data
    logger.info("Fetching lifecycle data...")
    try:
        lifecycle_data = lifecycle.fetch_lifecycle_data()
    except Exception as e:
        logger.error("Failed to fetch lifecycle data: %s", e)
        if args.json_output:
            print(json.dumps({
                "command": "precheck_nightly",
                "error": f"Lifecycle API unavailable: {e}",
                "timestamp": datetime.now().isoformat(),
            }, indent=2))
        else:
            print(f"ERROR: Lifecycle API unavailable: {e}")
        sys.exit(1)

    # Step 2: Determine branches from git remote
    if args.version:
        streams = [args.version]
    else:
        streams = git_ops.list_release_branches()

    # Step 3: Split into active vs inactive streams
    results = []
    active_streams = []

    for stream in streams:
        lc = lifecycle.get_lifecycle_status(stream, lifecycle_data)
        if lc and not lc["active"]:
            results.append({
                "stream": stream,
                "branch": f"release-{stream}",
                "status": "EOL" if lc["phase"] == "End of life" else "EUS",
                "lifecycle_phase": lc["phase"],
                "end_date": lc.get("end_date", ""),
            })
        else:
            active_streams.append(stream)

    # Step 4: Fetch Brew builds only if there are active streams to check
    brew_builds = {}
    if active_streams:
        logger.info("Fetching Brew nightly builds...")
        if not brew.check_vpn():
            if args.json_output:
                print(json.dumps({
                    "command": "precheck_nightly",
                    "error": "VPN not connected. Brew requires VPN access.",
                    "timestamp": datetime.now().isoformat(),
                }, indent=2))
            else:
                print("ERROR: VPN not connected. Brew requires VPN access.")
            sys.exit(1)
        brew_builds = brew.fetch_latest_nightly_builds()

    logger.info("Checking %d active branches...", len(active_streams))
    with ThreadPoolExecutor(max_workers=8) as executor:
        futures = {
            executor.submit(check_branch, s, brew_builds): s
            for s in active_streams
        }
        for future in as_completed(futures):
            results.append(future.result())

    # Sort by stream version descending
    results.sort(
        key=lambda r: [int(x) for x in r["stream"].split(".")],
        reverse=True,
    )

    output = {
        "command": "precheck_nightly",
        "timestamp": datetime.now().isoformat(),
        "branches": results,
    }

    if args.json_output:
        print(json.dumps(output, indent=2))
    else:
        print(format_text(results, verbose=args.verbose))


if __name__ == "__main__":
    main()
