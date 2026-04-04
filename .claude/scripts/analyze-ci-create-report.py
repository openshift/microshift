#!/usr/bin/env python3
"""
Generate an HTML report from analyze-ci analysis files.

Reads summary files (from analyze-ci:release and analyze-ci:pull-requests)
and bug mapping files (from analyze-ci:create-bugs) to produce a single
consolidated HTML report with tabbed navigation.

Usage:
    analyze-ci-create-report.py [--workdir DIR] <release1,release2,...>

Examples:
    analyze-ci-create-report.py 4.19,4.20,4.21,4.22
    analyze-ci-create-report.py --workdir /tmp/my-workdir 4.22
"""

import sys
import os
import re
import html as html_mod
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

CSS = """\
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; color: #333; }
        .container { max-width: 1200px; margin: 0 auto; }
        h1 { color: #1a1a2e; border-bottom: 3px solid #e94560; padding-bottom: 10px; }
        .release-section { background: white; border-radius: 8px; padding: 20px; margin: 20px 0; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .release-header { display: flex; justify-content: space-between; align-items: center; }
        .release-header h2 { color: #16213e; margin: 0; }
        .badge { padding: 4px 12px; border-radius: 12px; font-size: 0.85em; font-weight: 600; }
        .badge-ok { background: #d4edda; color: #155724; }
        .badge-issues { background: #fff3cd; color: #856404; }
        .badge-critical { background: #f8d7da; color: #721c24; }
        .badge-nodata { background: #e2e3e5; color: #383d41; }
        .severity-1, .severity-2 { color: #28a745; }
        .severity-3 { color: #856404; }
        .severity-4 { color: #dc3545; }
        .severity-5 { color: #721c24; font-weight: 700; }
        .infra-tag { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 0.8em; font-weight: 600; background: #cce5ff; color: #004085; }
        .root-cause { background: #fff8e1; border-left: 3px solid #ffc107; padding: 8px 12px; margin: 8px 0; font-size: 0.9em; }
        .status-pass { color: #28a745; }
        .status-fail { color: #dc3545; }
        .overview-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; margin: 20px 0; }
        .overview-card { background: white; border-radius: 8px; padding: 20px; text-align: center; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .overview-card .number { font-size: 2em; font-weight: 700; }
        .overview-card .label { color: #6c757d; font-size: 0.9em; }
        .collapsible { cursor: pointer; user-select: none; }
        .collapsible::before { content: '\\25B6  '; font-size: 0.8em; }
        .collapsible.active::before { content: '\\25BC  '; }
        .collapsible .job-date { font-weight: 400; color: #6c757d; font-size: 0.85em; }
        .job-date { font-weight: 400; color: #6c757d; font-size: 0.85em; }
        .collapsible-content { display: none; }
        .collapsible-content.show { display: block; }
        .bug-links { margin: 8px 0; padding: 8px 12px; background: #f0f4ff; border-left: 3px solid #0366d6; font-size: 0.9em; }
        .bug-links .bug-tag { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 0.85em; font-weight: 600; margin: 2px 4px 2px 0; text-decoration: none; }
        .bug-tag-open { background: #fff3cd; color: #856404; border: 1px solid #ffc107; }
        .bug-tag-regression { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
        .bug-links-label { font-weight: 600; color: #495057; margin-right: 6px; }
        .no-bugs { color: #6c757d; font-style: italic; font-size: 0.85em; }
        .toc { background: white; border-radius: 8px; padding: 20px; margin: 20px 0; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .toc ul { list-style: none; padding-left: 0; }
        .toc li { padding: 5px 0; }
        .toc a { color: #0366d6; text-decoration: none; }
        .toc a:hover { text-decoration: underline; }
        .timestamp { color: #6c757d; font-size: 0.9em; }
        a { color: #0366d6; }
        .tab-bar { display: flex; gap: 0; margin: 20px 0 0 0; border-bottom: 2px solid #dee2e6; }
        .tab-btn { padding: 12px 24px; border: none; background: transparent; font-size: 1em; font-weight: 600; color: #6c757d; cursor: pointer; border-bottom: 3px solid transparent; margin-bottom: -2px; transition: color 0.2s, border-color 0.2s; }
        .tab-btn:hover { color: #333; }
        .tab-btn.active { color: #e94560; border-bottom-color: #e94560; }
        .tab-content { display: none; }
        .tab-content.active { display: block; }
        .breakdown { display: flex; gap: 15px; margin: 10px 0; flex-wrap: wrap; }
        .breakdown-item { font-size: 0.9em; color: #495057; }
        .breakdown-item strong { color: #333; }"""

JS = """\
function showTab(e, name) {
    document.querySelectorAll('.tab-content').forEach(function(el) {
        el.classList.remove('active');
    });
    document.querySelectorAll('.tab-btn').forEach(function(el) {
        el.classList.remove('active');
    });
    document.getElementById('tab-' + name).classList.add('active');
    e.target.classList.add('active');
}
document.querySelectorAll('.collapsible').forEach(function(el) {
    el.addEventListener('click', function() {
        this.classList.toggle('active');
        this.nextElementSibling.classList.toggle('show');
    });
});"""


# ---------------------------------------------------------------------------
# File discovery
# ---------------------------------------------------------------------------

def discover_files(workdir, releases):
    """Find summary and bug mapping files for each release and for PRs."""
    result = {"releases": {}, "prs": {"summary": None, "bugs": {}}}

    for version in releases:
        entry = {"summary": None, "bugs": None}
        pattern = os.path.join(workdir, f"analyze-ci-release-{version}-summary.*.txt")
        matches = sorted(glob_mod.glob(pattern))
        if matches:
            entry["summary"] = matches[-1]

        bugs_path = os.path.join(workdir, f"analyze-ci-bugs-{version}.txt")
        if os.path.exists(bugs_path):
            entry["bugs"] = bugs_path

        result["releases"][version] = entry

    # PR summary
    pattern = os.path.join(workdir, "analyze-ci-prs-summary.*.txt")
    matches = sorted(glob_mod.glob(pattern))
    if matches:
        result["prs"]["summary"] = matches[-1]

    # PR bug mapping files (rebase-release-X.YY)
    for path in glob_mod.glob(os.path.join(workdir, "analyze-ci-bugs-rebase-release-*.txt")):
        m = re.match(r"analyze-ci-bugs-rebase-release-(.+)\.txt", os.path.basename(path))
        if m:
            result["prs"]["bugs"][m.group(1)] = path

    return result


# ---------------------------------------------------------------------------
# Summary parsing
# ---------------------------------------------------------------------------

def parse_release_summary(filepath):
    """Parse a release summary file into structured data.

    Returns a dict with: release, total_failed, date, breakdown, issues.
    Each issue has: number, title, job_count, severity, pattern,
    affected_jobs (list of {name, date, url}), root_cause, next_steps.
    """
    if not filepath or not os.path.exists(filepath):
        return None

    with open(filepath, "r") as f:
        lines = f.readlines()

    data = {
        "release": "",
        "total_failed": 0,
        "date": "",
        "breakdown": {"build": 0, "test": 0, "infrastructure": 0},
        "issues": [],
    }

    # Header
    for line in lines:
        m = re.match(r"MICROSHIFT (\S+) RELEASE", line)
        if m:
            data["release"] = m.group(1)
            break

    # Overview numbers
    for line in lines:
        m = re.match(r"\s+Total Failed Jobs:\s+(\d+)", line)
        if m:
            data["total_failed"] = int(m.group(1))
        m = re.match(r"\s+Analysis Date:\s+(.+)", line)
        if m:
            data["date"] = m.group(1).strip()
        m = re.match(r"\s+Build Failures:\s+(\d+)", line)
        if m:
            data["breakdown"]["build"] = int(m.group(1))
        m = re.match(r"\s+Test Failures:\s+(\d+)", line)
        if m:
            data["breakdown"]["test"] = int(m.group(1))
        m = re.match(r"\s+Infrastructure:\s+(\d+)", line)
        if m:
            data["breakdown"]["infrastructure"] = int(m.group(1))

    # --- Issue parsing (state machine) ---
    ISSUE_RE = re.compile(r"^(\d+)\.\s+(.+?)\s+\((\d+)\s+jobs?[^)]*\)")
    FIELD_RE = re.compile(r"^   (Severity|Pattern|Root Cause|Next Steps):\s*(.*)")
    JOB_RE = re.compile(r"^   \*\s+(.+?)\s+\[(\d{4}-\d{2}-\d{2})\]")
    URL_RE = re.compile(r"^\s+(https?://\S+)")

    in_issues = False
    cur = None
    field = None
    field_lines = []
    in_jobs = False

    def _flush_field():
        nonlocal field, field_lines
        if cur is not None and field and field_lines:
            cur[field] = " ".join(" ".join(field_lines).split())
        field = None
        field_lines = []

    def _flush_issue():
        nonlocal cur, in_jobs
        _flush_field()
        if cur is not None:
            data["issues"].append(cur)
        cur = None
        in_jobs = False

    for raw in lines:
        line = raw.rstrip("\n")

        if "TOP ISSUES" in line:
            in_issues = True
            continue
        if not in_issues:
            continue
        if line.startswith("===") or line.startswith("\u2550"):
            _flush_issue()
            break

        # New issue header
        m = ISSUE_RE.match(line)
        if m:
            _flush_issue()
            cur = {
                "number": int(m.group(1)),
                "title": m.group(2),
                "job_count": int(m.group(3)),
                "severity": "",
                "pattern": "",
                "affected_jobs": [],
                "root_cause": "",
                "next_steps": "",
            }
            continue

        if cur is None:
            continue

        # Known field start
        fm = FIELD_RE.match(line)
        if fm:
            _flush_field()
            in_jobs = False
            name = fm.group(1)
            val = fm.group(2).strip()
            if name == "Severity":
                cur["severity"] = val
            else:
                key = name.lower().replace(" ", "_")
                field = key
                field_lines = [val] if val else []
            continue

        # Affected Jobs header
        if re.match(r"^   Affected Jobs:", line):
            _flush_field()
            in_jobs = True
            continue

        # Inside Affected Jobs
        if in_jobs:
            jm = JOB_RE.match(line)
            if jm:
                cur["affected_jobs"].append(
                    {"name": jm.group(1), "date": jm.group(2), "url": ""}
                )
                continue
            um = URL_RE.match(line)
            if um and cur["affected_jobs"]:
                cur["affected_jobs"][-1]["url"] = um.group(1)
                continue

        # Continuation of multi-line field
        if field and line.strip():
            field_lines.append(line.strip())

    _flush_issue()
    return data


def parse_pr_summary(filepath):
    """Parse a PR summary file into structured data.

    Returns a dict with: overview, prs, has_content.
    Each PR has: number, title, url, passed, failed, failed_jobs.
    Each failed_job has: number, name, date, root_cause, url.
    """
    if not filepath or not os.path.exists(filepath):
        return None

    with open(filepath, "r") as f:
        content = f.read()

    data = {
        "overview": {
            "total_prs": 0,
            "prs_with_failures": 0,
            "total_failed": 0,
            "date": "",
        },
        "prs": [],
        "has_content": False,
    }

    for line in content.split("\n"):
        m = re.match(r"\s+Total Open PRs:\s+(\d+)", line)
        if m:
            data["overview"]["total_prs"] = int(m.group(1))
        m = re.match(r"\s+PRs with Failures:\s+(\d+)", line)
        if m:
            data["overview"]["prs_with_failures"] = int(m.group(1))
        m = re.match(r"\s+Total Failed Jobs:\s+(\d+)", line)
        if m:
            data["overview"]["total_failed"] = int(m.group(1))
        m = re.match(r"\s+Analysis Date:\s+(.+)", line)
        if m:
            data["overview"]["date"] = m.group(1).strip()

    no_content_markers = [
        "No open rebase pull requests found",
        "No open pull requests found",
        "All open PR jobs are passing",
        "All PR jobs are passing",
    ]
    if any(marker in content for marker in no_content_markers):
        return data

    data["has_content"] = True
    lines = content.split("\n")

    cur_pr = None
    cur_job = None
    field = None
    field_lines = []

    def _flush_field():
        nonlocal field, field_lines
        if cur_job is not None and field and field_lines:
            cur_job[field] = " ".join(" ".join(field_lines).split())
        field = None
        field_lines = []

    def _flush_job():
        nonlocal cur_job
        _flush_field()
        if cur_job is not None and cur_pr is not None:
            cur_pr["failed_jobs"].append(cur_job)
        cur_job = None

    def _flush_pr():
        nonlocal cur_pr
        _flush_job()
        if cur_pr is not None:
            data["prs"].append(cur_pr)
        cur_pr = None

    for raw in lines:
        line = raw.rstrip("\n")

        # PR header
        m = re.match(r"^PR #(\d+):\s+(.+)", line)
        if m:
            _flush_pr()
            cur_pr = {
                "number": int(m.group(1)),
                "title": m.group(2).strip(),
                "url": "",
                "passed": 0,
                "failed": 0,
                "failed_jobs": [],
            }
            continue

        if cur_pr is None:
            continue

        # PR URL
        m = re.match(r"^\s+(https://github\.com/\S+)", line)
        if m and not cur_pr["url"]:
            cur_pr["url"] = m.group(1)
            continue

        # Jobs count
        m = re.match(r"^\s+Jobs:\s+(\d+)\s+passed,\s+(\d+)\s+failed", line)
        if m:
            cur_pr["passed"] = int(m.group(1))
            cur_pr["failed"] = int(m.group(2))
            continue

        # Failed job entry
        m = re.match(r"^\s+(\d+)\.\s+(\S+)\s+\[(\d{4}-\d{2}-\d{2})\]", line)
        if m:
            _flush_job()
            cur_job = {
                "number": int(m.group(1)),
                "name": m.group(2),
                "date": m.group(3),
                "root_cause": "",
                "url": "",
            }
            continue

        if cur_job is None:
            continue

        # Job fields
        m = re.match(r"^\s+Root Cause:\s*(.*)", line)
        if m:
            _flush_field()
            field = "root_cause"
            val = m.group(1).strip()
            field_lines = [val] if val else []
            continue

        m = re.match(r"^\s+URL:\s+(https?://\S+)", line)
        if m:
            _flush_field()
            cur_job["url"] = m.group(1)
            continue

        if re.match(r"^\s+Status:", line):
            _flush_field()
            continue

        # Separator
        if line.startswith("===") or line.startswith("\u2550"):
            _flush_pr()
            continue

        # Continuation
        if field and line.strip():
            field_lines.append(line.strip())

    _flush_pr()
    return data


# ---------------------------------------------------------------------------
# Bug mapping parsing
# ---------------------------------------------------------------------------

def parse_bug_mapping(filepath):
    """Parse a bug mapping file into a list of bug candidates.

    Each candidate has: error_signature, severity, step_name, affected_jobs,
    duplicates (list of {key, summary, status}),
    regressions (list of {key, summary, status}).
    """
    if not filepath or not os.path.exists(filepath):
        return []

    with open(filepath, "r") as f:
        content = f.read()

    candidates = []
    blocks = re.split(r"--- BUG CANDIDATE ---", content)

    for block in blocks[1:]:
        block = block.split("--- END BUG CANDIDATE ---")[0]
        cand = {
            "error_signature": "",
            "severity": 0,
            "step_name": "",
            "affected_jobs": 0,
            "duplicates": [],
            "regressions": [],
        }

        for line in block.strip().split("\n"):
            line = line.strip()
            if line.startswith("ERROR_SIGNATURE:"):
                cand["error_signature"] = line.split(":", 1)[1].strip()
            elif line.startswith("SEVERITY:"):
                try:
                    cand["severity"] = int(line.split(":", 1)[1].strip())
                except ValueError:
                    pass
            elif line.startswith("STEP_NAME:"):
                cand["step_name"] = line.split(":", 1)[1].strip()
            elif line.startswith("AFFECTED_JOBS:"):
                try:
                    cand["affected_jobs"] = int(line.split(":", 1)[1].strip())
                except ValueError:
                    pass
            elif line.startswith("JIRA_DUPLICATE_DETAILS:"):
                details = line.split(":", 1)[1].strip()
                if details and details != "None":
                    cand["duplicates"] = _parse_jira_details(details)
            elif line.startswith("JIRA_REGRESSION_DETAILS:"):
                details = line.split(":", 1)[1].strip()
                if details and details != "None":
                    cand["regressions"] = _parse_jira_details(details)

        if cand["error_signature"]:
            candidates.append(cand)

    return candidates


def _parse_jira_details(details_str):
    """Parse 'KEY|summary|status, KEY|summary|status' format."""
    result = []
    # Split at comma boundaries followed by a JIRA key pattern
    entries = re.split(r",\s+(?=[A-Z]+-\d+\|)", details_str)
    for entry in entries:
        entry = entry.strip()
        if not entry:
            continue
        parts = entry.split("|")
        if len(parts) >= 3:
            result.append({
                "key": parts[0].strip(),
                "summary": parts[1].strip(),
                "status": parts[2].strip(),
            })
        elif len(parts) == 2:
            result.append({
                "key": parts[0].strip(),
                "summary": parts[1].strip(),
                "status": "",
            })
    return result


# ---------------------------------------------------------------------------
# Fuzzy matching
# ---------------------------------------------------------------------------

def _tokenize(text):
    """Extract significant words for fuzzy matching."""
    words = re.findall(r"[a-z0-9][a-z0-9_.-]*[a-z0-9]|[a-z0-9]", text.lower())
    return {w for w in words if w not in STOP_WORDS and len(w) >= 2}


def match_issue_to_bugs(issue_title, bug_candidates):
    """Return the best-matching bug candidate for an issue title, or None.

    Uses simple token overlap: if >= 50% of a candidate's distinctive tokens
    appear in the issue title, it's considered a match.  Picks the candidate
    with the highest overlap score.
    """
    if not bug_candidates:
        return None

    issue_tokens = _tokenize(issue_title)
    if not issue_tokens:
        return None

    best = None
    best_score = 0.0

    for cand in bug_candidates:
        sig_tokens = _tokenize(cand["error_signature"])
        if not sig_tokens:
            continue
        overlap = len(issue_tokens & sig_tokens)
        score = overlap / len(sig_tokens)
        if score > best_score:
            best_score = score
            best = cand

    return best if best_score >= 0.5 else None


# ---------------------------------------------------------------------------
# HTML helpers
# ---------------------------------------------------------------------------

def _e(text):
    """HTML-escape text."""
    return html_mod.escape(str(text)) if text else ""


def _badge_class(total_failed, has_critical=False):
    """Determine badge CSS class."""
    if total_failed == 0:
        return "badge-ok"
    if total_failed >= 5 or has_critical:
        return "badge-critical"
    return "badge-issues"


def _render_bug_links(bug_match):
    """Render the bug-links div contents for a matched bug candidate."""
    if not bug_match:
        return '<span class="no-bugs">No tracked bugs</span>'

    has_dups = bool(bug_match["duplicates"])
    has_regs = bool(bug_match["regressions"])

    if not has_dups and not has_regs:
        return '<span class="no-bugs">No tracked bugs</span>'

    parts = []
    if has_dups:
        parts.append("<strong>Bugs:</strong> ")
        for d in bug_match["duplicates"]:
            parts.append(
                f'<a class="bug-tag bug-tag-open" '
                f'href="https://issues.redhat.com/browse/{_e(d["key"])}" '
                f'target="_blank">{_e(d["key"])}</a> '
                f'<span class="job-date">{_e(d["summary"])} ({_e(d["status"])})</span> '
            )

    if has_regs:
        if has_dups:
            parts.append("<br>")
        parts.append("<strong>Regressions:</strong> ")
        for r in bug_match["regressions"]:
            parts.append(
                f'<a class="bug-tag bug-tag-regression" '
                f'href="https://issues.redhat.com/browse/{_e(r["key"])}" '
                f'target="_blank">{_e(r["key"])} &#x27F2;</a> '
                f'<span class="job-date">{_e(r["summary"])} ({_e(r["status"])})</span> '
            )

    return "".join(parts)


# ---------------------------------------------------------------------------
# HTML generation — release section
# ---------------------------------------------------------------------------

def render_release_section(version, rdata, bug_candidates):
    """Render a single release section as HTML."""
    if rdata is None:
        return (
            f'        <div class="release-section" id="release-{_e(version)}">\n'
            f'            <div class="release-header">\n'
            f'                <h2>Release {_e(version)}</h2>\n'
            f'                <span class="badge badge-nodata">no data</span>\n'
            f'            </div>\n'
            f"            <p>Analysis failed or produced no output.</p>\n"
            f"        </div>"
        )

    total = rdata["total_failed"]
    has_critical = any(
        i["severity"].upper() == "CRITICAL" for i in rdata["issues"]
        if i["severity"]
    )
    badge = _badge_class(total, has_critical)
    b = rdata["breakdown"]

    lines = []
    lines.append(f'        <div class="release-section" id="release-{_e(version)}">')
    lines.append(f'            <div class="release-header">')
    lines.append(f"                <h2>Release {_e(version)}</h2>")
    label = "failure" if total == 1 else "failures"
    lines.append(f'                <span class="badge {badge}">{total} {label}</span>')
    lines.append(f"            </div>")
    lines.append(f'            <div class="breakdown">')
    lines.append(
        f'                <span class="breakdown-item">'
        f'<strong>{b["build"]}</strong> Build</span>'
    )
    lines.append(
        f'                <span class="breakdown-item">'
        f'<strong>{b["test"]}</strong> Test</span>'
    )
    lines.append(
        f'                <span class="breakdown-item">'
        f'<strong>{b["infrastructure"]}</strong> Infrastructure</span>'
    )
    lines.append(f"            </div>")

    for issue in rdata["issues"]:
        bug_match = match_issue_to_bugs(issue["title"], bug_candidates)
        jc = issue["job_count"]
        job_word = "job" if jc == 1 else "jobs"
        sev = issue["severity"].upper() if issue["severity"] else "UNKNOWN"

        lines.append("")
        lines.append(
            f'            <h4 class="collapsible">'
            f'{issue["number"]}. {_e(issue["title"])} '
            f'<span class="job-date">({jc} {job_word}, {sev})</span></h4>'
        )
        lines.append(f'            <div class="collapsible-content">')
        if issue["pattern"]:
            lines.append(f"                <p>{_e(issue['pattern'])}</p>")
        if issue["root_cause"]:
            lines.append(
                f'                <div class="root-cause">'
                f"<strong>Root Cause:</strong> {_e(issue['root_cause'])}</div>"
            )
        lines.append(
            f'                <div class="bug-links">'
            f"{_render_bug_links(bug_match)}</div>"
        )
        if issue["affected_jobs"]:
            lines.append(f"                <p><strong>Affected Jobs:</strong></p>")
            lines.append(f"                <ul>")
            for job in issue["affected_jobs"]:
                if job["url"]:
                    lines.append(
                        f'                    <li><a href="{_e(job["url"])}" '
                        f'target="_blank">{_e(job["name"])}</a> '
                        f'<span class="job-date">[{_e(job["date"])}]</span></li>'
                    )
                else:
                    lines.append(
                        f"                    <li>{_e(job['name'])} "
                        f'<span class="job-date">[{_e(job["date"])}]</span></li>'
                    )
            lines.append(f"                </ul>")
        if issue["next_steps"]:
            lines.append(
                f"                <p><em>Next Steps:</em> "
                f"{_e(issue['next_steps'])}</p>"
            )
        lines.append(f"            </div>")

    lines.append(f"        </div>")
    return "\n".join(lines)


# ---------------------------------------------------------------------------
# HTML generation — PR section
# ---------------------------------------------------------------------------

def render_pr_section(pr_data, all_pr_bugs):
    """Render the Pull Requests tab content as HTML."""
    if pr_data is None or not pr_data["has_content"]:
        return (
            '        <div class="release-section">\n'
            '            <div class="release-header">\n'
            "                <h2>Rebase Pull Requests</h2>\n"
            '                <span class="badge badge-ok">0 failures</span>\n'
            "            </div>\n"
            "            <p>No open rebase pull requests found.</p>\n"
            "        </div>"
        )

    lines = []
    for pr in pr_data["prs"]:
        total_failed = pr["failed"]
        badge = _badge_class(total_failed)
        label = "failure" if total_failed == 1 else "failures"

        lines.append(
            f'        <div class="release-section" id="pr-{pr["number"]}">'
        )
        lines.append(f'            <div class="release-header">')
        lines.append(
            f'                <h2>PR #{pr["number"]}: {_e(pr["title"])}</h2>'
        )
        lines.append(
            f'                <span class="badge {badge}">'
            f'{total_failed} {label}</span>'
        )
        lines.append(f"            </div>")
        if pr["url"]:
            lines.append(
                f'            <p><a href="{_e(pr["url"])}" target="_blank">'
                f'{_e(pr["url"])}</a></p>'
            )
        if pr["passed"] or pr["failed"]:
            lines.append(
                f"            <p>Jobs: {pr['passed']} passed, "
                f"{pr['failed']} failed</p>"
            )

        for job in pr["failed_jobs"]:
            bug_match = match_issue_to_bugs(
                job.get("root_cause", "") or job.get("name", ""),
                all_pr_bugs,
            )
            lines.append("")
            lines.append(
                f'            <h4 class="collapsible">'
                f'<span class="job-date">[{_e(job["date"])}]</span> '
                f'{job["number"]}. {_e(job["name"])}</h4>'
            )
            lines.append(f'            <div class="collapsible-content">')
            if job["url"]:
                lines.append(
                    f'                <p><strong>Job:</strong> '
                    f'<span class="job-date">[{_e(job["date"])}]</span> '
                    f'<a href="{_e(job["url"])}" target="_blank">'
                    f"{_e(job['name'])}</a></p>"
                )
            if job["root_cause"]:
                lines.append(
                    f'                <div class="root-cause">'
                    f"<strong>Root Cause:</strong> "
                    f"{_e(job['root_cause'])}</div>"
                )
            lines.append(
                f'                <div class="bug-links">'
                f"{_render_bug_links(bug_match)}</div>"
            )
            lines.append(f"            </div>")

        lines.append(f"        </div>")

    return "\n".join(lines)


# ---------------------------------------------------------------------------
# HTML generation — full document
# ---------------------------------------------------------------------------

def generate_html(releases_data, bug_data, pr_data, all_pr_bugs, timestamp):
    """Assemble the complete HTML document."""
    date_str = timestamp.strftime("%Y-%m-%d")
    time_str = timestamp.strftime("%Y-%m-%d %H:%M:%S")

    # Overview cards
    cards = []
    for version, rdata in releases_data.items():
        if rdata is not None:
            count = rdata["total_failed"]
            css = "status-fail" if count > 0 else "status-pass"
        else:
            count = "?"
            css = ""
        cards.append(
            f'        <div class="overview-card">\n'
            f'            <div class="number {css}">{count}</div>\n'
            f'            <div class="label">Release {_e(version)}</div>\n'
            f"        </div>"
        )
    pr_failed = pr_data["overview"]["total_failed"] if pr_data else 0
    pr_css = "status-fail" if pr_failed > 0 else "status-pass"
    cards.append(
        f'        <div class="overview-card">\n'
        f'            <div class="number {pr_css}">{pr_failed}</div>\n'
        f'            <div class="label">Rebase PRs</div>\n'
        f"        </div>"
    )

    # Table of contents
    toc = []
    for version, rdata in releases_data.items():
        if rdata is not None:
            b = rdata["breakdown"]
            toc.append(
                f'                <li><a href="#release-{_e(version)}">'
                f"Release {_e(version)}</a> &mdash; "
                f'{rdata["total_failed"]} failures '
                f'({b["build"]} build, {b["test"]} test, '
                f'{b["infrastructure"]} infra)</li>'
            )
        else:
            toc.append(
                f'                <li><a href="#release-{_e(version)}">'
                f"Release {_e(version)}</a> &mdash; no data</li>"
            )

    # Release sections
    sections = []
    for version, rdata in releases_data.items():
        bugs = bug_data.get(version, [])
        sections.append(render_release_section(version, rdata, bugs))

    # PR section
    pr_section = render_pr_section(pr_data, all_pr_bugs)

    return f"""\
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>MicroShift CI Doctor Report - {date_str}</title>
    <style>
{CSS}
    </style>
</head>
<body>
<div class="container">
    <h1>MicroShift CI Doctor Report</h1>
    <p class="timestamp">Generated: {time_str} UTC</p>

    <div class="overview-grid">
{chr(10).join(cards)}
    </div>

    <div class="tab-bar">
        <button class="tab-btn active" onclick="showTab(event, 'periodics')">Periodics</button>
        <button class="tab-btn" onclick="showTab(event, 'pull-requests')">Pull Requests</button>
    </div>

    <div id="tab-periodics" class="tab-content active">

        <div class="toc">
            <h3>Table of Contents</h3>
            <ul>
{chr(10).join(toc)}
            </ul>
        </div>

{chr(10).join(sections)}

    </div>

    <div id="tab-pull-requests" class="tab-content">

{pr_section}

    </div>

    <p>&nbsp;</p><p>&nbsp;</p><p>&nbsp;</p><p>&nbsp;</p>
</div>
<script>
{JS}
</script>
</body>
</html>
"""


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    workdir = None
    releases_arg = None

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
            print(
                "Usage: analyze-ci-create-report.py [--workdir DIR] "
                "<release1,release2,...>",
                file=sys.stderr,
            )
            sys.exit(1)
        else:
            releases_arg = args[i]
            i += 1

    if not releases_arg:
        print(
            "Usage: analyze-ci-create-report.py [--workdir DIR] "
            "<release1,release2,...>",
            file=sys.stderr,
        )
        sys.exit(1)

    releases = [v.strip() for v in releases_arg.split(",") if v.strip()]
    if not releases:
        print("Error: at least one release version is required", file=sys.stderr)
        sys.exit(1)

    if workdir is None:
        workdir = (
            f"/tmp/analyze-ci-claude-workdir."
            f"{datetime.now().strftime('%y%m%d')}"
        )

    if not os.path.isdir(workdir):
        print(f"Error: work directory does not exist: {workdir}", file=sys.stderr)
        sys.exit(1)

    # Discover files
    files = discover_files(workdir, releases)

    # Report discovery
    print("Files discovered:")
    found_any = False
    for version in releases:
        entry = files["releases"][version]
        parts = []
        if entry["summary"]:
            parts.append("summary found")
            found_any = True
        else:
            parts.append("summary MISSING")
        if entry["bugs"]:
            parts.append("bug mapping found")
        else:
            parts.append("no bug mapping")
        print(f"  Release {version}: {', '.join(parts)}")

    pr_entry = files["prs"]
    if pr_entry["summary"]:
        found_any = True
        bug_note = (
            f"bug mapping found ({len(pr_entry['bugs'])} files)"
            if pr_entry["bugs"]
            else "no bug mapping"
        )
        print(f"  PRs: summary found, {bug_note}")
    else:
        print("  PRs: no summary")

    if not found_any:
        print(
            f"\nError: no analysis files found in {workdir}",
            file=sys.stderr,
        )
        sys.exit(1)

    # Parse everything
    releases_data = {}
    bug_data = {}
    for version in releases:
        entry = files["releases"][version]
        releases_data[version] = parse_release_summary(entry["summary"])
        bug_data[version] = parse_bug_mapping(entry["bugs"])

    pr_data = parse_pr_summary(pr_entry["summary"])

    # Merge all PR bug candidates
    all_pr_bugs = []
    for path in pr_entry["bugs"].values():
        all_pr_bugs.extend(parse_bug_mapping(path))

    # Generate HTML
    timestamp = datetime.now(timezone.utc)
    html_content = generate_html(
        releases_data, bug_data, pr_data, all_pr_bugs, timestamp
    )

    # Write output
    output_path = os.path.join(workdir, "microshift-ci-doctor-report.html")
    with open(output_path, "w") as f:
        f.write(html_content)

    # Summary
    print("")
    print("Summary:")
    print("  Periodics:")
    for version in releases:
        rdata = releases_data[version]
        if rdata:
            print(
                f"    Release {version}: "
                f"{rdata['total_failed']} failed periodic jobs"
            )
        else:
            print(f"    Release {version}: no data")
    print("  Pull Requests:")
    if pr_data and pr_data["has_content"]:
        total_prs = len(pr_data["prs"])
        total_failed = sum(pr["failed"] for pr in pr_data["prs"])
        print(
            f"    {total_prs} rebase PRs with "
            f"{total_failed} total failed jobs"
        )
    elif pr_data:
        print("    No open rebase PRs or no failures found")
    else:
        print("    No PR data")
    print("")
    print(f"HTML report generated: {output_path}")


if __name__ == "__main__":
    main()
