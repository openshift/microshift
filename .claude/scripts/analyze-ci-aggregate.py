#!/usr/bin/env python3
"""
Aggregate per-job analysis reports into a release or PR summary JSON file.

Reads per-job report files (containing STRUCTURED SUMMARY blocks and prose),
groups jobs by ERROR_SIGNATURE similarity, and produces JSON consumed by
analyze-ci-create-report.py.

Usage:
    analyze-ci-aggregate.py --release 4.22 [--workdir DIR]
    analyze-ci-aggregate.py --prs [--workdir DIR]

Output files:
    analyze-ci-release-<version>-summary.json
    analyze-ci-prs-summary.json
"""

import json
import sys
import os
import re
import glob as glob_mod
from datetime import datetime, timezone


# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------

STOP_WORDS = frozenset({
    "the", "a", "an", "in", "on", "at", "to", "for", "of", "with", "by",
    "is", "was", "are", "were", "be", "been", "and", "or", "not", "no",
    "but", "from", "that", "this", "all", "has", "have", "had", "do",
    "does", "did", "will", "would", "could", "should", "may", "might",
})

INFRA_LAYERS = {"aws infra", "external infrastructure"}
BUILD_LAYERS = {"build phase"}
SIMILARITY_THRESHOLD = 0.50


# ---------------------------------------------------------------------------
# Parsing per-job report files
# ---------------------------------------------------------------------------

def parse_structured_summary(filepath):
    """Extract the STRUCTURED SUMMARY block from a per-job report file."""
    with open(filepath, "r") as f:
        content = f.read()

    m = re.search(
        r"--- STRUCTURED SUMMARY ---\n(.+?)\n--- END STRUCTURED SUMMARY ---",
        content, re.DOTALL,
    )
    if not m:
        return None

    data = {}
    for line in m.group(1).strip().split("\n"):
        if ":" in line:
            key, val = line.split(":", 1)
            data[key.strip()] = val.strip()

    try:
        severity = int(data.get("SEVERITY", "3"))
    except ValueError:
        severity = 3

    return {
        "severity": severity,
        "stack_layer": data.get("STACK_LAYER", ""),
        "step_name": data.get("STEP_NAME", ""),
        "error_signature": data.get("ERROR_SIGNATURE", ""),
        "infrastructure_failure": data.get("INFRASTRUCTURE_FAILURE", "false").lower() == "true",
        "job_url": data.get("JOB_URL", ""),
        "job_name": data.get("JOB_NAME", ""),
        "release": data.get("RELEASE", ""),
        "finished": data.get("FINISHED", ""),
    }


def parse_prose_fields(filepath):
    """Extract Error: and Suggested Remediation: from report prose."""
    with open(filepath, "r") as f:
        content = f.read()

    prose = content.split("--- STRUCTURED SUMMARY ---")[0]

    error = ""
    m = re.search(
        r"^Error:\s*(.+?)(?=\nSuggested Remediation:|\nError Severity:|\nStack Layer:|\nStep Name:|\n\n|\n---|\Z)",
        prose, re.MULTILINE | re.DOTALL,
    )
    if m:
        error = " ".join(m.group(1).split())

    remediation = ""
    m = re.search(
        r"^Suggested Remediation:\s*(.+?)(?=\n\n|\n---|\nError Severity:|\nStack Layer:|\nStep Name:|\Z)",
        prose, re.MULTILINE | re.DOTALL,
    )
    if m:
        remediation = " ".join(m.group(1).split())

    return error, remediation


# ---------------------------------------------------------------------------
# Grouping
# ---------------------------------------------------------------------------

def _tokenize(text):
    words = re.findall(r"[a-z0-9][a-z0-9_.-]*[a-z0-9]|[a-z0-9]", text.lower())
    return {w for w in words if w not in STOP_WORDS and len(w) >= 2}


def signature_similarity(sig_a, sig_b):
    tokens_a = _tokenize(sig_a)
    tokens_b = _tokenize(sig_b)
    if not tokens_a or not tokens_b:
        return 0.0
    return len(tokens_a & tokens_b) / min(len(tokens_a), len(tokens_b))


def group_by_signature(jobs):
    groups = []
    for job in jobs:
        sig = job["error_signature"]
        placed = False
        for group in groups:
            if signature_similarity(sig, group[0]["error_signature"]) >= SIMILARITY_THRESHOLD:
                group.append(job)
                placed = True
                break
        if not placed:
            groups.append([job])
    return groups


def classify_severity(group):
    max_sev = max(j["severity"] for j in group)
    count = len(group)
    if max_sev >= 4:
        return "CRITICAL"
    if count >= 3:
        return "HIGH"
    if count >= 2:
        return "MEDIUM"
    return "LOW"


def classify_breakdown(stack_layer):
    lower = stack_layer.lower()
    if lower in INFRA_LAYERS:
        return "infrastructure"
    if lower in BUILD_LAYERS:
        return "build"
    return "test"


# ---------------------------------------------------------------------------
# JSON generation
# ---------------------------------------------------------------------------

def build_release_json(release, jobs, timestamp):
    """Build the release summary as a dict (ready for json.dump)."""
    groups = group_by_signature(jobs)
    groups.sort(key=lambda g: (max(j["severity"] for j in g), len(g)), reverse=True)

    breakdown = {"build": 0, "test": 0, "infrastructure": 0}
    for job in jobs:
        breakdown[classify_breakdown(job["stack_layer"])] += 1

    issues = []
    for i, group in enumerate(groups, 1):
        rep = max(group, key=lambda j: j["severity"])
        issues.append({
            "number": i,
            "title": rep["error_signature"],
            "job_count": len(group),
            "severity": classify_severity(group),
            "pattern": rep.get("error_text", ""),
            "root_cause": rep.get("error_text", ""),
            "next_steps": rep.get("remediation_text", ""),
            "affected_jobs": [
                {"name": j["job_name"], "date": j["finished"], "url": j["job_url"]}
                for j in group
            ],
        })

    return {
        "release": release,
        "total_failed": len(jobs),
        "date": timestamp.strftime("%Y-%m-%d"),
        "breakdown": breakdown,
        "issues": issues,
    }


def build_pr_json(pr_jobs, timestamp):
    """Build the PR summary as a dict (ready for json.dump).

    pr_jobs: dict mapping pr_number to list of job dicts.
    """
    total_failed = sum(len(jobs) for jobs in pr_jobs.values())

    prs = []
    for pr_number, jobs in sorted(pr_jobs.items()):
        if not jobs:
            continue
        first = jobs[0]
        prs.append({
            "number": pr_number,
            "title": first.get("pr_title", ""),
            "url": first.get("pr_url", ""),
            "passed": 0,
            "failed": len(jobs),
            "failed_jobs": [
                {
                    "number": j_idx,
                    "name": j["job_name"],
                    "date": j["finished"],
                    "root_cause": j.get("error_text", ""),
                    "url": j["job_url"],
                }
                for j_idx, j in enumerate(jobs, 1)
            ],
        })

    return {
        "total_prs": len(pr_jobs),
        "prs_with_failures": len(prs),
        "total_failed": total_failed,
        "date": timestamp.strftime("%Y-%m-%d"),
        "has_content": total_failed > 0,
        "prs": prs,
    }


# ---------------------------------------------------------------------------
# File discovery
# ---------------------------------------------------------------------------

def find_release_job_files(workdir, release):
    pattern = os.path.join(workdir, f"analyze-ci-release-{release}-job-*.txt")
    return sorted(glob_mod.glob(pattern))


def find_pr_job_files(workdir):
    pattern = os.path.join(workdir, "analyze-ci-prs-job-*.txt")
    return sorted(glob_mod.glob(pattern))


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    workdir = None
    release = None
    mode = None

    args = sys.argv[1:]
    i = 0
    while i < len(args):
        if args[i] == "--workdir":
            if i + 1 >= len(args):
                print("Error: --workdir requires an argument", file=sys.stderr)
                sys.exit(1)
            workdir = args[i + 1]
            i += 2
        elif args[i] == "--release":
            if i + 1 >= len(args):
                print("Error: --release requires a version", file=sys.stderr)
                sys.exit(1)
            mode = "release"
            release = args[i + 1]
            i += 2
        elif args[i] == "--prs":
            mode = "prs"
            i += 1
        elif args[i].startswith("-"):
            print(f"Unknown option: {args[i]}", file=sys.stderr)
            sys.exit(1)
        else:
            print(f"Unknown argument: {args[i]}", file=sys.stderr)
            sys.exit(1)

    if mode is None:
        print(
            "Usage:\n"
            "  analyze-ci-aggregate.py --release <version> [--workdir DIR]\n"
            "  analyze-ci-aggregate.py --prs [--workdir DIR]",
            file=sys.stderr,
        )
        sys.exit(1)

    if workdir is None:
        workdir = f"/tmp/analyze-ci-claude-workdir.{datetime.now().strftime('%y%m%d')}"

    if not os.path.isdir(workdir):
        print(f"Error: work directory does not exist: {workdir}", file=sys.stderr)
        sys.exit(1)

    timestamp = datetime.now(timezone.utc)

    if mode == "release":
        files = find_release_job_files(workdir, release)
        if not files:
            print(f"No job files found for release {release}", file=sys.stderr)
            sys.exit(1)

        print(f"Found {len(files)} job files for release {release}", file=sys.stderr)
        jobs = []
        for filepath in files:
            summary = parse_structured_summary(filepath)
            if summary is None:
                print(f"  WARNING: no STRUCTURED SUMMARY in {os.path.basename(filepath)}", file=sys.stderr)
                continue
            error_text, remediation_text = parse_prose_fields(filepath)
            summary["error_text"] = error_text
            summary["remediation_text"] = remediation_text
            jobs.append(summary)

        if not jobs:
            print("No valid job reports found", file=sys.stderr)
            sys.exit(1)

        result = build_release_json(release, jobs, timestamp)
        output_path = os.path.join(workdir, f"analyze-ci-release-{release}-summary.json")
        with open(output_path, "w") as f:
            json.dump(result, f, indent=2)
        print(f"Written: {output_path}", file=sys.stderr)
        print(json.dumps(result, indent=2))

    elif mode == "prs":
        files = find_pr_job_files(workdir)
        if not files:
            print("No PR job files found", file=sys.stderr)
            result = build_pr_json({}, timestamp)
        else:
            print(f"Found {len(files)} PR job files", file=sys.stderr)
            pr_jobs = {}
            for filepath in files:
                summary = parse_structured_summary(filepath)
                if summary is None:
                    print(f"  WARNING: no STRUCTURED SUMMARY in {os.path.basename(filepath)}", file=sys.stderr)
                    continue
                error_text, remediation_text = parse_prose_fields(filepath)
                summary["error_text"] = error_text
                summary["remediation_text"] = remediation_text
                summary["pr_title"] = ""
                summary["pr_url"] = ""

                m = re.search(r"-pr(\d+)-", os.path.basename(filepath))
                pr_number = int(m.group(1)) if m else 0
                pr_jobs.setdefault(pr_number, []).append(summary)

            result = build_pr_json(pr_jobs, timestamp)

        output_path = os.path.join(workdir, "analyze-ci-prs-summary.json")
        with open(output_path, "w") as f:
            json.dump(result, f, indent=2)
        print(f"Written: {output_path}", file=sys.stderr)
        print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
