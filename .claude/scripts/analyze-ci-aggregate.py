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
        "raw_error": data.get("RAW_ERROR", ""),
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

def _normalize_step_name(step_name):
    """Extract the step ref from a fully-qualified Prow step name.

    Prow step names follow the pattern ``<test-variant>-<step-ref>``
    where the step ref typically starts with ``openshift-microshift-``.
    The LLM sometimes includes the test-variant prefix, sometimes not,
    which would cause identical steps to land in different buckets
    during two-pass grouping.

    Examples:
        "openshift-microshift-infra-aws-ec2"
            → "openshift-microshift-infra-aws-ec2"
        "e2e-aws-tests-bootc-arm-nightly-el10-openshift-microshift-infra-aws-ec2"
            → "openshift-microshift-infra-aws-ec2"
        "clusterbot-nightly-openshift-microshift-infra-aws-ec2"
            → "openshift-microshift-infra-aws-ec2"
    """
    m = re.search(r"(openshift-microshift-\S+)", step_name)
    return m.group(1) if m else step_name


def _tokenize(text):
    words = re.findall(r"[a-z0-9][a-z0-9_.-]*[a-z0-9]|[a-z0-9]", text.lower())
    return {w for w in words if w not in STOP_WORDS and len(w) >= 2}


def signature_similarity(sig_a, sig_b):
    tokens_a = _tokenize(sig_a)
    tokens_b = _tokenize(sig_b)
    if not tokens_a or not tokens_b:
        return 0.0
    return len(tokens_a & tokens_b) / min(len(tokens_a), len(tokens_b))


def _grouping_text(job):
    """Return the text used for similarity grouping.

    Prefers RAW_ERROR (verbatim log text, deterministic) over
    ERROR_SIGNATURE (LLM-paraphrased, variable across runs).
    Falls back to ERROR_SIGNATURE when RAW_ERROR is absent
    (backward compatibility with older report files).
    """
    return job.get("raw_error") or job.get("error_signature", "")


def _group_by_similarity(jobs):
    """Group jobs by similarity of their grouping text.

    Uses RAW_ERROR when available (deterministic log text),
    falling back to ERROR_SIGNATURE for older reports.

    A new job is compared against ALL existing members of each group,
    not just the first.  If any member exceeds the similarity threshold
    the job joins that group.  This makes grouping less sensitive to
    insertion order and to phrasing variation — each member added to
    a group acts as an additional reference point for future matches.
    """
    groups = []
    for job in jobs:
        sig = _grouping_text(job)
        placed = False
        for group in groups:
            if any(
                signature_similarity(sig, _grouping_text(member)) >= SIMILARITY_THRESHOLD
                for member in group
            ):
                group.append(job)
                placed = True
                break
        if not placed:
            groups.append([job])
    return groups


def group_by_signature(jobs):
    """Two-pass grouping: first by step_name, then by signature similarity.

    Grouping by step_name first prevents jobs from different CI steps
    (e.g. conformance vs metal-tests) from being merged together even
    when their error signatures share enough tokens to exceed the
    similarity threshold.  This makes the issue count deterministic
    across runs where only the signature wording varies.
    """
    # Pass 1: bucket by normalized step_name
    by_step = {}
    for job in jobs:
        step = _normalize_step_name(job.get("step_name", ""))
        by_step.setdefault(step, []).append(job)

    # Pass 2: within each step bucket, group by signature similarity
    all_groups = []
    for step_jobs in by_step.values():
        all_groups.extend(_group_by_similarity(step_jobs))
    return all_groups


def classify_severity(group):
    count = len(group)
    if count >= 5:
        return "CRITICAL"
    if count >= 3:
        return "HIGH"
    if count >= 2:
        return "MEDIUM"
    return "LOW"


# Patterns for deterministic breakdown classification.
# These override the LLM's STACK_LAYER, because step names and error
# signatures are deterministic while STACK_LAYER varies across runs.
INFRA_STEP_PATTERNS = ("infra-aws", "infra-gcp", "infra-setup")
BUILD_STEP_PATTERNS = ("update-origin", "build-image", "iso-build")
BUILD_SIGNATURE_PATTERNS = ("update-origin", "build-image")


def classify_breakdown(stack_layer, step_name="", error_signature=""):
    lower_step = step_name.lower()
    lower_sig = error_signature.lower()

    # Step-name overrides — more reliable than LLM's STACK_LAYER
    if any(k in lower_step for k in INFRA_STEP_PATTERNS):
        return "infrastructure"
    if any(k in lower_step for k in BUILD_STEP_PATTERNS):
        return "build"

    # Error-signature overrides — catches build operations that run
    # inside a test step (e.g. "make update-origin" in e2e-metal-tests)
    if any(k in lower_sig for k in BUILD_SIGNATURE_PATTERNS):
        return "build"

    # Fall back to LLM's classification
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
        breakdown[classify_breakdown(
            job["stack_layer"],
            job.get("step_name", ""),
            job.get("error_signature", ""),
        )] += 1

    issues = []
    for i, group in enumerate(groups, 1):
        rep = max(group, key=lambda j: j["severity"])
        issues.append({
            "number": i,
            "title": rep["error_signature"],
            "job_count": len(group),
            "severity": classify_severity(group),
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
