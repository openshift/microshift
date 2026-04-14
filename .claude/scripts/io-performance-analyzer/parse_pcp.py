#!/usr/bin/env python3
"""Parse pcp2json output with disk.dev.read, disk.dev.write, and disk.dev.await.

pcp2json produces a nested structure:
  @pcp.@hosts[].@metrics[] where each metric entry has @timestamp and
  nested disk.dev.{read,write,await}.@instances[].{name, value}.

Aggregates per-device instances: sum for read/write, max for await.
Outputs a clean JSON with arrays: timestamps, bi, bo, await.
"""

import argparse
import json
import sys
from datetime import datetime
from zoneinfo import ZoneInfo, ZoneInfoNotFoundError

FMT = "%Y-%m-%d %H:%M:%S"


def get_instances(sample, *path):
    """Walk the nested dict to reach @instances list for a metric."""
    node = sample
    for key in path:
        if not isinstance(node, dict) or key not in node:
            return []
        node = node[key]
    return node.get("@instances", [])


def aggregate_instances(instances, func):
    """Aggregate instance values, converting strings to float."""
    vals = []
    for inst in instances:
        try:
            vals.append(float(inst["value"]))
        except (KeyError, ValueError, TypeError):
            continue
    return func(vals) if vals else 0


def convert_timestamp(ts_str, target_tz):
    """Parse a pcp2json timestamp and convert to the target timezone."""
    utc = ZoneInfo("UTC")
    for fmt in (FMT, "%Y-%m-%d %H:%M:%S%z", "%Y-%m-%dT%H:%M:%S",
                "%Y-%m-%dT%H:%M:%S%z"):
        try:
            ts = datetime.strptime(ts_str, fmt)
            break
        except ValueError:
            continue
    else:
        return None
    if ts.tzinfo is None:
        ts = ts.replace(tzinfo=utc)
    return ts.astimezone(target_tz).strftime(FMT)


def main():
    parser = argparse.ArgumentParser(
        description="Parse pcp2json output into plot-ready JSON")
    parser.add_argument("input_json", help="Raw pcp2json output file")
    parser.add_argument("output_json", help="Output JSON file")
    parser.add_argument("--timezone", default="UTC",
                        help="Target timezone for timestamps (default: UTC)")
    args = parser.parse_args()

    try:
        target_tz = ZoneInfo(args.timezone)
    except (KeyError, ZoneInfoNotFoundError):
        print(f"ERROR: Unknown timezone '{args.timezone}'", file=sys.stderr)
        sys.exit(1)

    with open(args.input_json) as f:
        raw = f.read()

    # Strip pcp2json comment lines (lines starting with { "//" ... })
    lines = []
    for line in raw.splitlines():
        stripped = line.strip()
        if stripped.startswith('{ "//":') and stripped.endswith("}"):
            continue
        lines.append(line)
    raw = "\n".join(lines)

    try:
        data = json.loads(raw)
    except json.JSONDecodeError as e:
        print(f"ERROR: Failed to parse pcp2json output: {e}", file=sys.stderr)
        sys.exit(1)

    # Navigate to the metrics array
    hosts = data.get("@pcp", {}).get("@hosts", [])
    if not hosts:
        print("ERROR: No hosts found in pcp2json output", file=sys.stderr)
        sys.exit(1)

    samples = hosts[0].get("@metrics", [])
    if not samples:
        print("ERROR: No metric samples found in pcp2json output",
              file=sys.stderr)
        sys.exit(1)

    result = {"timestamps": [], "bi": [], "bo": [], "await": []}

    for sample in samples:
        ts_str = sample.get("@timestamp", "")
        ts = convert_timestamp(ts_str, target_tz)
        if ts is None:
            continue

        read_insts = get_instances(sample, "disk", "dev", "read")
        write_insts = get_instances(sample, "disk", "dev", "write")
        await_insts = get_instances(sample, "disk", "dev", "await")

        # Skip samples with no metric data (e.g. first interval)
        if not read_insts and not write_insts and not await_insts:
            continue

        result["timestamps"].append(ts)
        result["bi"].append(aggregate_instances(read_insts, sum))
        result["bo"].append(aggregate_instances(write_insts, sum))
        result["await"].append(aggregate_instances(await_insts, max))

    if not result["timestamps"]:
        print("ERROR: No valid data points found", file=sys.stderr)
        sys.exit(1)

    with open(args.output_json, "w") as f:
        json.dump(result, f, indent=2)


if __name__ == "__main__":
    main()
