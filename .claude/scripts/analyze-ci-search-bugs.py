#!/usr/bin/env python3
"""
Prepare bug candidates from per-job analysis reports.

Parses STRUCTURED SUMMARY blocks, groups by ERROR_SIGNATURE similarity,
extracts Jira search keywords, and writes a candidates JSON file for
the create-bugs skill to search Jira against.

Usage:
    analyze-ci-search-bugs.py <source> [--workdir DIR]

    <source> is one of:
      - Release version: 4.22, main
      - PR number: pr-6396, pr6396
      - Rebase shorthand: rebase-release-4.22

Output:
    ${WORKDIR}/analyze-ci-bug-candidates-<source>.json
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

# Additional stop words filtered only during keyword extraction for Jira search,
# not during signature grouping (which must match aggregate.py's tokenization).
KEYWORD_STOP_WORDS = STOP_WORDS | frozenset({
    "ci", "microshift", "failure", "failed", "error", "test", "tests",
    "job", "jobs", "step", "periodic",
})

SIMILARITY_THRESHOLD = 0.50


# ---------------------------------------------------------------------------
# Parsing
# ---------------------------------------------------------------------------

def parse_structured_summary(filepath):
    """Extract STRUCTURED SUMMARY block from a per-job report file."""
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

    # Get the analysis text (everything before STRUCTURED SUMMARY)
    analysis_text = content.split("--- STRUCTURED SUMMARY ---")[0].strip()

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
        "analysis_text": analysis_text,
        "source_file": filepath,
    }


# ---------------------------------------------------------------------------
# Grouping
# ---------------------------------------------------------------------------

def _normalize_step_name(step_name):
    """Extract the step ref from a fully-qualified Prow step name.

    Prow step names follow ``<test-variant>-<step-ref>`` where the
    step ref typically starts with ``openshift-microshift-``.
    """
    m = re.search(r"(openshift-microshift-\S+)", step_name)
    return m.group(1) if m else step_name


def _tokenize(text):
    words = re.findall(r"[a-z0-9][a-z0-9_.-]*[a-z0-9]|[a-z0-9]", text.lower())
    return {w for w in words if w not in STOP_WORDS and len(w) >= 2}


def _signature_similarity(sig_a, sig_b):
    tokens_a = _tokenize(sig_a)
    tokens_b = _tokenize(sig_b)
    if not tokens_a or not tokens_b:
        return 0.0
    return len(tokens_a & tokens_b) / min(len(tokens_a), len(tokens_b))


def _group_by_similarity(jobs):
    """Group jobs by ERROR_SIGNATURE token similarity.

    A new job is compared against ALL existing members of each group,
    not just the first.  If any member exceeds the similarity threshold
    the job joins that group.  This makes grouping less sensitive to
    insertion order and to signature phrasing variation.
    """
    groups = []
    for job in jobs:
        sig = job["error_signature"]
        placed = False
        for group in groups:
            if any(
                _signature_similarity(sig, member["error_signature"]) >= SIMILARITY_THRESHOLD
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
    from being merged together even when their error signatures share
    enough tokens to exceed the similarity threshold.
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


# ---------------------------------------------------------------------------
# Keyword extraction
# ---------------------------------------------------------------------------

def _tokenize_for_keywords(text):
    """Tokenize with extra stop words filtered for Jira keyword extraction."""
    words = re.findall(r"[a-z0-9][a-z0-9_.-]*[a-z0-9]|[a-z0-9]", text.lower())
    return {w for w in words if w not in KEYWORD_STOP_WORDS and len(w) >= 2}


def extract_keywords(error_signature):
    """Extract distinctive search keywords from an error signature.

    Returns a list of 2-4 keywords ranked by specificity.
    Uses KEYWORD_STOP_WORDS (broader filtering) so generic CI terms
    like "test", "failed", "microshift" don't pollute Jira searches.
    """
    tokens = _tokenize_for_keywords(error_signature)
    if not tokens:
        return []

    def specificity(token):
        score = len(token)
        if "-" in token or "." in token:
            score += 10
        if any(c.isdigit() for c in token):
            score += 5
        return score

    ranked = sorted(tokens, key=specificity, reverse=True)
    return ranked[:4]


def extract_test_ids(error_signature):
    """Extract numeric test case IDs (4-6 digits) from error signature."""
    return re.findall(r"\b(\d{4,6})\b", error_signature)


# ---------------------------------------------------------------------------
# Candidate building
# ---------------------------------------------------------------------------

def build_candidates(groups):
    """Build bug candidate list from grouped jobs."""
    candidates = []

    for group in groups:
        rep = max(group, key=lambda j: j["severity"])
        keywords = extract_keywords(rep["error_signature"])
        test_ids = extract_test_ids(rep["error_signature"])

        step_names = list({j["step_name"] for j in group if j["step_name"]})

        candidates.append({
            "error_signature": rep["error_signature"],
            "severity": max(j["severity"] for j in group),
            "step_name": ", ".join(step_names),
            "affected_jobs": len(group),
            "keywords": keywords,
            "test_ids": test_ids,
            "jobs": [
                {
                    "job_name": j["job_name"],
                    "job_url": j["job_url"],
                    "finished": j["finished"],
                }
                for j in group
            ],
            "analysis_text": rep["analysis_text"],
        })

    # Sort by severity desc, then job count desc
    candidates.sort(key=lambda c: (c["severity"], c["affected_jobs"]), reverse=True)
    return candidates


# ---------------------------------------------------------------------------
# File discovery
# ---------------------------------------------------------------------------

def find_job_files(workdir, source):
    """Find per-job report files for a given source.

    Returns (files, source_label) tuple.
    """
    # Release version
    if re.match(r"^(\d+\.\d+|main)$", source):
        pattern = os.path.join(workdir, f"analyze-ci-release-{source}-job-*.txt")
        files = sorted(glob_mod.glob(pattern))
        return files, f"release {source}"

    # PR number
    m = re.match(r"^pr-?(\d+)$", source)
    if m:
        pr_num = m.group(1)
        pattern = os.path.join(workdir, f"analyze-ci-prs-job-*-pr{pr_num}-*.txt")
        files = sorted(glob_mod.glob(pattern))
        return files, f"PR #{pr_num}"

    # Rebase PR shorthand
    m = re.match(r"^rebase-release-(.+)$", source)
    if m:
        release = m.group(1)
        # Scan all PR job files for ones matching this release
        pattern = os.path.join(workdir, "analyze-ci-prs-job-*.txt")
        all_files = sorted(glob_mod.glob(pattern))
        files = []
        for filepath in all_files:
            summary = parse_structured_summary(filepath)
            if summary and (
                f"release-{release}" in summary.get("job_name", "")
                or summary.get("release", "") == release
            ):
                files.append(filepath)
        return files, f"rebase PR for {release}"

    return [], source


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    workdir = None
    source = None

    args = sys.argv[1:]
    i = 0
    while i < len(args):
        if args[i] == "--workdir":
            if i + 1 >= len(args):
                print("Error: --workdir requires an argument", file=sys.stderr)
                sys.exit(1)
            workdir = args[i + 1]
            i += 2
        elif args[i].startswith("-"):
            print(f"Unknown option: {args[i]}", file=sys.stderr)
            sys.exit(1)
        else:
            source = args[i]
            i += 1

    if not source:
        print(
            "Usage: analyze-ci-search-bugs.py <source> [--workdir DIR]\n"
            "  <source>: release version (4.22), PR (pr-6396), or rebase (rebase-release-4.22)",
            file=sys.stderr,
        )
        sys.exit(1)

    if workdir is None:
        workdir = f"/tmp/analyze-ci-claude-workdir.{datetime.now().strftime('%y%m%d')}"

    if not os.path.isdir(workdir):
        print(f"Error: work directory does not exist: {workdir}", file=sys.stderr)
        sys.exit(1)

    files, source_label = find_job_files(workdir, source)
    if not files:
        print(f"No job files found for {source_label} in {workdir}", file=sys.stderr)
        sys.exit(1)

    print(f"Found {len(files)} job files for {source_label}", file=sys.stderr)

    # Parse all files
    jobs = []
    skipped = 0
    for filepath in files:
        summary = parse_structured_summary(filepath)
        if summary is None:
            print(f"  WARNING: no STRUCTURED SUMMARY in {os.path.basename(filepath)}", file=sys.stderr)
            skipped += 1
            continue
        jobs.append(summary)

    if not jobs:
        print("No valid job reports found", file=sys.stderr)
        sys.exit(1)

    print(f"Parsed {len(jobs)} jobs ({skipped} skipped)", file=sys.stderr)

    # Group and build candidates
    groups = group_by_signature(jobs)
    candidates = build_candidates(groups)

    print(f"Deduplicated to {len(candidates)} bug candidates", file=sys.stderr)

    # Build output
    result = {
        "source": source,
        "source_label": source_label,
        "date": datetime.now(timezone.utc).strftime("%Y-%m-%d"),
        "job_files_found": len(files),
        "job_files_parsed": len(jobs),
        "job_files_skipped": skipped,
        "candidates": candidates,
    }

    output_path = os.path.join(workdir, f"analyze-ci-bug-candidates-{source}.json")
    with open(output_path, "w") as f:
        json.dump(result, f, indent=2)

    print(f"Written: {output_path}", file=sys.stderr)
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
