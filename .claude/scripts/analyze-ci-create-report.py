#!/usr/bin/env python3
"""
Generate an HTML report from analyze-ci JSON files.

Reads JSON summary files (from analyze-ci-aggregate.py) and JSON bug mapping
files (from analyze-ci:create-bugs) to produce a consolidated HTML report.

Usage:
    analyze-ci-create-report.py [--workdir DIR] <release1,release2,...>
"""

import json
import sys
import os
import re
import html as html_mod
import glob as glob_mod
from datetime import datetime, timezone


# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------

# Threshold for fuzzy matching issue titles to bug candidate signatures.
# Uses asymmetric formula: overlap / len(sig_tokens) — measures what fraction
# of the bug candidate's signature is covered by the issue title. This differs
# from the symmetric min-based formula in aggregate.py/search-bugs.py because
# issue titles are short summaries while signatures are detailed.
MATCH_THRESHOLD = 0.50

STOP_WORDS = frozenset({
    "the", "a", "an", "in", "on", "at", "to", "for", "of", "with", "by",
    "is", "was", "are", "were", "be", "been", "and", "or", "not", "no",
    "but", "from", "that", "this", "all", "has", "have", "had", "do",
    "does", "did", "will", "would", "could", "should", "may", "might",
})

CSS = """\
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; color: #333; }
        .container { max-width: 1200px; margin: 0 auto; }
        h1 { color: #1a1a2e; border-bottom: 3px solid #e94560; padding-bottom: 8px; font-size: 1.4em; margin: 10px 0; }
        h2 { font-size: 1.15em; margin: 0; }
        h3 { font-size: 1.05em; margin: 0 0 8px 0; }
        .release-section { background: white; border-radius: 8px; padding: 15px; margin: 15px 0; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .release-header { display: flex; justify-content: space-between; align-items: center; }
        .release-header h2 { color: #16213e; margin: 0; }
        .badge { padding: 4px 12px; border-radius: 12px; font-size: 0.85em; font-weight: 600; }
        .badge-ok { background: #d4edda; color: #155724; }
        .badge-issues { background: #fff3cd; color: #856404; }
        .badge-critical { background: #f8d7da; color: #721c24; }
        .badge-nodata { background: #e2e3e5; color: #383d41; }
        .root-cause { background: #fff8e1; border-left: 3px solid #ffc107; padding: 8px 12px; margin: 8px 0; font-size: 0.9em; }
        .status-pass { color: #28a745; }
        .status-fail { color: #dc3545; }
        .overview-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 10px; margin: 15px 0; }
        .overview-card { background: white; border-radius: 8px; padding: 12px; text-align: center; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .overview-card .number { font-size: 1.6em; font-weight: 700; }
        .overview-card .label { color: #6c757d; font-size: 0.9em; }
        .job-date { font-weight: 400; color: #6c757d; font-size: 0.85em; }
        .issues-table { width: 100%; border-collapse: collapse; margin: 15px 0; }
        .issues-table td { padding: 5px 6px; vertical-align: middle; }
        .issues-table .col-num { width: 30px; text-align: right; font-weight: 700; color: #495057; padding-right: 10px; }
        .issues-table .col-sev { width: 78px; }
        .issues-table .col-ftype { width: 58px; }
        .issues-table .col-title { cursor: pointer; user-select: none; }
        .issues-table .col-title::before { content: '\\25B6  '; font-size: 0.7em; color: #6c757d; }
        .issues-table .col-title.active::before { content: '\\25BC  '; }
        .issues-table .col-jobs { width: 70px; text-align: center; color: #6c757d; font-size: 0.85em; white-space: nowrap; }
        .issues-table .detail-row td { padding: 0 6px 12px 40px; }
        .issues-table .detail-row { display: none; }
        .issues-table .detail-row.show { display: table-row; }
        .issues-table tr.issue-row { border-top: 1px solid #eee; }
        .issues-table tr.issue-row:first-child { border-top: none; }
        .bug-links { margin: 8px 0; padding: 8px 12px; background: #f0f4ff; border-left: 3px solid #0366d6; font-size: 0.9em; }
        .bug-links .bug-tag { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 0.85em; font-weight: 600; margin: 2px 4px 2px 0; text-decoration: none; }
        .bug-tag-open { background: #fff3cd; color: #856404; border: 1px solid #ffc107; }
        .bug-tag-regression { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
        .no-bugs { color: #6c757d; font-style: italic; font-size: 0.85em; }
        .toc { background: white; border-radius: 8px; padding: 15px; margin: 15px 0; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .toc ul { list-style: none; padding-left: 0; }
        .toc li { padding: 5px 0; }
        .toc a { color: #0366d6; text-decoration: none; }
        .toc a:hover { text-decoration: underline; }
        .timestamp { color: #6c757d; font-size: 0.9em; }
        a { color: #0366d6; }
        .tab-bar { display: flex; gap: 0; margin: 20px 0 0 0; border-bottom: 2px solid #dee2e6; }
        .tab-btn { padding: 12px 24px; border: none; background: transparent; font-size: 1em; font-weight: 600;
            color: #6c757d; cursor: pointer; border-bottom: 3px solid transparent;
            margin-bottom: -2px; transition: color 0.2s, border-color 0.2s; }
        .tab-btn:hover { color: #333; }
        .tab-btn.active { color: #e94560; border-bottom-color: #e94560; }
        .tab-content { display: none; }
        .tab-content.active { display: block; }
        .breakdown { display: flex; gap: 15px; margin: 10px 0; flex-wrap: wrap; }
        .breakdown-item { font-size: 0.9em; color: #495057; }
        .breakdown-item strong { color: #333; }
        .severity-badge { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 0.75em; font-weight: 700; text-transform: uppercase; }
        .severity-high { background: #f8d7da; color: #721c24; }
        .severity-medium { background: #fff3cd; color: #856404; }
        .severity-low { background: #d4edda; color: #155724; }
        .severity-critical { background: #721c24; color: #fff; }
        .ftype-badge { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 0.75em; font-weight: 700; text-transform: uppercase; }
        .ftype-test { background: #cce5ff; color: #004085; }
        .ftype-build { background: #e2d5f1; color: #4a235a; }
        .ftype-infra { background: #fde2cc; color: #7d4e24; }"""

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
document.querySelectorAll('.col-title').forEach(function(el) {
    el.addEventListener('click', function() {
        this.classList.toggle('active');
        var row = this.closest('tr').nextElementSibling;
        if (row && row.classList.contains('detail-row')) {
            row.classList.toggle('show');
        }
    });
});"""


# ---------------------------------------------------------------------------
# File discovery
# ---------------------------------------------------------------------------

def discover_files(workdir, releases):
    result = {"releases": {}, "prs": {"summary": None, "status": None, "bugs": []}}

    for version in releases:
        entry = {"summary": None, "bugs": None}
        path = os.path.join(workdir, f"analyze-ci-release-{version}-summary.json")
        if os.path.exists(path):
            entry["summary"] = path
        path = os.path.join(workdir, f"analyze-ci-bugs-{version}.json")
        if os.path.exists(path):
            entry["bugs"] = path
        result["releases"][version] = entry

    path = os.path.join(workdir, "analyze-ci-prs-summary.json")
    if os.path.exists(path):
        result["prs"]["summary"] = path

    path = os.path.join(workdir, "analyze-ci-prs-status.json")
    if os.path.exists(path):
        result["prs"]["status"] = path

    for path in glob_mod.glob(os.path.join(workdir, "analyze-ci-bugs-rebase-release-*.json")):
        result["prs"]["bugs"].append(path)

    return result


# ---------------------------------------------------------------------------
# JSON loading (replaces all text parsers)
# ---------------------------------------------------------------------------

def load_json(filepath):
    if not filepath or not os.path.exists(filepath):
        return None
    try:
        with open(filepath, "r") as f:
            return json.load(f)
    except (json.JSONDecodeError, IOError) as exc:
        print(f"WARNING: failed to load {filepath}: {exc}", file=sys.stderr)
        return None


def load_bug_candidates(filepath):
    data = load_json(filepath)
    if not data:
        return []
    return data.get("candidates", [])


# ---------------------------------------------------------------------------
# Fuzzy matching
# ---------------------------------------------------------------------------

def _tokenize(text):
    words = re.findall(r"[a-z0-9][a-z0-9_.-]*[a-z0-9]|[a-z0-9]", text.lower())
    return {w for w in words if w not in STOP_WORDS and len(w) >= 2}


def match_issue_to_bugs(issue_title, bug_candidates):
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
        score = len(issue_tokens & sig_tokens) / len(sig_tokens)
        if score > best_score:
            best_score = score
            best = cand
    return best if best_score >= MATCH_THRESHOLD else None


def _extract_pr_numbers(candidate):
    """Extract PR numbers from a bug candidate's job names/URLs.

    Handles two patterns:
    - File-derived job names: "-pr123-" (from analyze-ci-prs-job-*-pr<N>-*.txt)
    - Prow URLs: ".../pull/openshift_microshift/123/..."
    """
    pr_nums = set()
    for job in candidate.get("jobs", []):
        url = job.get("job_url", "")
        m = re.search(r"/pull/[^/]+/(\d+)/", url)
        if m:
            pr_nums.add(int(m.group(1)))
        name = job.get("job_name", "")
        m = re.search(r"-pr(\d+)-", name)
        if m:
            pr_nums.add(int(m.group(1)))
    return pr_nums


def _index_pr_bugs(bug_paths):
    """Load PR bug candidates and index them by PR number.

    Returns a dict mapping PR number (int) to list of bug candidates.
    Candidates affecting multiple PRs appear under each PR.
    """
    by_pr = {}
    for path in bug_paths:
        for cand in load_bug_candidates(path):
            pr_nums = _extract_pr_numbers(cand)
            for num in pr_nums:
                by_pr.setdefault(num, []).append(cand)
    return by_pr


# ---------------------------------------------------------------------------
# HTML helpers
# ---------------------------------------------------------------------------

def _e(text):
    return html_mod.escape(str(text)) if text else ""


def _badge_class(total_failed, has_critical=False):
    if total_failed == 0:
        return "badge-ok"
    if total_failed >= 5 or has_critical:
        return "badge-critical"
    return "badge-issues"


def _render_bug_links(bug_match):
    if not bug_match:
        return '<span class="no-bugs">No tracked bugs</span>'
    has_dups = bool(bug_match.get("duplicates"))
    has_regs = bool(bug_match.get("regressions"))
    if not has_dups and not has_regs:
        return '<span class="no-bugs">No tracked bugs</span>'

    parts = []
    if has_dups:
        parts.append("<strong>Bugs:</strong><br>")
        for d in bug_match["duplicates"]:
            parts.append(
                f'<a class="bug-tag bug-tag-open" '
                f'href="https://issues.redhat.com/browse/{_e(d["key"])}" '
                f'target="_blank">{_e(d["key"])}</a> '
                f'<span class="job-date">{_e(d["summary"])} ({_e(d["status"])})</span><br>'
            )
    if has_regs:
        parts.append("<strong>Regressions:</strong><br>")
        for r in bug_match["regressions"]:
            parts.append(
                f'<a class="bug-tag bug-tag-regression" '
                f'href="https://issues.redhat.com/browse/{_e(r["key"])}" '
                f'target="_blank">{_e(r["key"])} &#x27F2;</a> '
                f'<span class="job-date">{_e(r["summary"])} ({_e(r["status"])})</span><br>'
            )
    return "".join(parts)


# ---------------------------------------------------------------------------
# HTML rendering
# ---------------------------------------------------------------------------

def render_release_section(version, rdata, bug_candidates):
    if rdata is None:
        return (
            f'        <div class="release-section" id="release-{_e(version)}">\n'
            '            <div class="release-header">\n'
            f'                <h2>Release {_e(version)}</h2>\n'
            '                <span class="badge badge-nodata">no data</span>\n'
            '            </div>\n'
            "            <p>Analysis failed or produced no output.</p>\n"
            "        </div>"
        )

    total = rdata["total_failed"]
    has_critical = any(i.get("severity", "").upper() == "CRITICAL" for i in rdata["issues"])
    badge = _badge_class(total, has_critical)
    b = rdata["breakdown"]

    lines = []
    lines.append(f'        <div class="release-section" id="release-{_e(version)}">')
    lines.append('            <div class="release-header">')
    lines.append(f"                <h2>Release {_e(version)}</h2>")
    label = "failure" if total == 1 else "failures"
    lines.append(f'                <span class="badge {badge}">{total} {label}</span>')
    lines.append("            </div>")
    lines.append('            <div class="breakdown">')
    lines.append(f'                <span class="breakdown-item"><strong>{b["build"]}</strong> Build</span>')
    lines.append(f'                <span class="breakdown-item"><strong>{b["test"]}</strong> Test</span>')
    lines.append(f'                <span class="breakdown-item"><strong>{b["infrastructure"]}</strong> Infrastructure</span>')
    lines.append("            </div>")

    lines.append('            <table class="issues-table">')
    for issue in rdata["issues"]:
        bug_match = match_issue_to_bugs(issue["title"], bug_candidates)
        jc = issue["job_count"]
        sev = issue.get("severity", "UNKNOWN").upper()
        sev_css = f"severity-{sev.lower()}" if sev in ("HIGH", "MEDIUM", "LOW", "CRITICAL") else ""
        ftype = issue.get("failure_type", "test")
        ftype_label = "INFRA" if ftype == "infrastructure" else ftype.upper()
        ftype_css = "ftype-infra" if ftype == "infrastructure" else f"ftype-{ftype}"
        jobs_label = f'{jc} {"job" if jc == 1 else "jobs"}'

        lines.append('            <tr class="issue-row">')
        lines.append(f'                <td class="col-num">{issue["number"]}.</td>')
        lines.append(f'                <td class="col-sev"><span class="severity-badge {sev_css}">{sev}</span></td>')
        lines.append(f'                <td class="col-ftype"><span class="ftype-badge {ftype_css}">{ftype_label}</span></td>')
        lines.append(f'                <td class="col-title">{_e(issue["title"])}</td>')
        lines.append(f'                <td class="col-jobs">{jobs_label}</td>')
        lines.append('            </tr>')
        lines.append('            <tr class="detail-row"><td colspan="5">')
        if issue.get("root_cause"):
            lines.append(f'                <div class="root-cause"><strong>Root Cause:</strong> {_e(issue["root_cause"])}</div>')
        lines.append(f'                <div class="bug-links">{_render_bug_links(bug_match)}</div>')
        if issue.get("affected_jobs"):
            lines.append("                <p><strong>Affected Jobs:</strong></p><ul>")
            for job in issue["affected_jobs"]:
                if job.get("url"):
                    lines.append(f'                    <li><span class="job-date">[{_e(job["date"])}]</span> <a href="{_e(job["url"])}" target="_blank">{_e(job["name"])}</a></li>')
                else:
                    lines.append(f'                    <li><span class="job-date">[{_e(job["date"])}]</span> {_e(job["name"])}</li>')
            lines.append("                </ul>")
        if issue.get("next_steps"):
            lines.append(f"                <p><em>Next Steps:</em> {_e(issue['next_steps'])}</p>")
        lines.append("            </td></tr>")
    lines.append('            </table>')

    lines.append("        </div>")
    return "\n".join(lines)


def render_pr_section(pr_data, all_pr_bugs, pr_status):
    """Render the Pull Requests tab.

    pr_data: analyzed PR summary (from aggregate), may be None.
    all_pr_bugs: dict mapping PR number (int) to list of bug candidates.
    pr_status: list of all PR status snapshots (from prepare), may be None.
    """
    # Build a lookup of analyzed PRs by number
    analyzed = {}
    if pr_data and pr_data.get("has_content"):
        for pr in pr_data["prs"]:
            analyzed[pr["number"]] = pr

    # Build the full PR list: all PRs from status, merged with analysis
    all_prs = []
    if pr_status:
        for s in pr_status:
            num = s["pr_number"]
            entry = {
                "number": num,
                "title": s.get("title", ""),
                "url": s.get("url", ""),
                "passed": s.get("passed", 0),
                "failed": s.get("failed", 0),
                "pending": s.get("pending", 0),
                "total": s.get("total", 0),
            }
            if num in analyzed:
                entry["analysis"] = analyzed[num]
            all_prs.append(entry)
    elif analyzed:
        # No status file — fall back to analyzed data only
        for pr in pr_data["prs"]:
            all_prs.append({
                "number": pr["number"],
                "title": pr.get("title", ""),
                "url": pr.get("url", ""),
                "passed": 0,
                "failed": pr.get("failed", 0),
                "pending": 0,
                "total": pr.get("failed", 0),
                "analysis": pr,
            })

    if not all_prs:
        return (
            '        <div class="release-section">\n'
            '            <div class="release-header">\n'
            "                <h2>Rebase Pull Requests</h2>\n"
            '                <span class="badge badge-ok">0 failures</span>\n'
            "            </div>\n"
            "            <p>No open rebase pull requests found.</p>\n"
            "        </div>"
        )

    # TOC
    toc_lines = []
    toc_lines.append('        <div class="toc">')
    toc_lines.append('            <h3>Table of Contents</h3>')
    toc_lines.append('            <ul>')
    for pr in all_prs:
        analysis = pr.get("analysis")
        if analysis:
            b = analysis.get("breakdown", {})
        else:
            b = {"build": 0, "test": 0, "infrastructure": 0}
        pending = pr.get("pending", 0)
        suffix = f' &mdash; {pending} running' if pending else ''
        toc_lines.append(
            f'                <li><a href="#pr-{pr["number"]}">PR# {pr["number"]}</a>'
            f' &mdash; {pr["failed"]} failures ({b.get("build", 0)} build, {b.get("test", 0)} test, {b.get("infrastructure", 0)} infra){suffix}</li>'
        )
    toc_lines.append('            </ul>')
    toc_lines.append('        </div>')

    # Sections
    lines = []
    for pr in all_prs:
        analysis = pr.get("analysis")
        total_failed = pr["failed"]
        badge = _badge_class(total_failed)

        lines.append(f'        <div class="release-section" id="pr-{pr["number"]}">')
        lines.append('            <div class="release-header">')
        pr_link = f'<a href="{_e(pr["url"])}" target="_blank" title="{_e(pr["title"])}">PR# {pr["number"]}</a>' if pr.get("url") else f'<span title="{_e(pr["title"])}">PR# {pr["number"]}</span>'
        pr_release_m = re.search(r"rebase-(release-\d+\.\d+|main)", pr.get("title", ""))
        pr_release_label = f' (rebase {pr_release_m.group(1)})' if pr_release_m else f': {_e(pr["title"])}' if pr.get("title") else ''
        lines.append(f'                <h2>{pr_link}{pr_release_label}</h2>')
        label = "failure" if total_failed == 1 else "failures"
        lines.append(f'                <span class="badge {badge}">{total_failed} {label}</span>')

        lines.append("            </div>")

        # Breakdown: same format as periodics (Build/Test/Infrastructure)
        # Plus job status (passed/running) when available
        pending = pr.get("pending", 0)
        if analysis and analysis.get("breakdown"):
            b = analysis["breakdown"]
        else:
            b = {"build": 0, "test": 0, "infrastructure": 0}
        lines.append('            <div class="breakdown">')
        lines.append(f'                <span class="breakdown-item"><strong>{b.get("build", 0)}</strong> Build</span>')
        lines.append(f'                <span class="breakdown-item"><strong>{b.get("test", 0)}</strong> Test</span>')
        lines.append(f'                <span class="breakdown-item"><strong>{b.get("infrastructure", 0)}</strong> Infrastructure</span>')
        if pr["passed"]:
            lines.append(f'                <span class="breakdown-item"><strong>{pr["passed"]}</strong> Passed</span>')
        if pending:
            lines.append(f'                <span class="breakdown-item"><strong>{pending}</strong> Running</span>')
        lines.append("            </div>")

        pr_bugs = all_pr_bugs.get(pr["number"], [])
        if analysis and analysis.get("issues"):

            lines.append('            <table class="issues-table">')
            for issue in analysis["issues"]:
                bug_match = match_issue_to_bugs(issue.get("title", ""), pr_bugs)
                jc = issue["job_count"]
                sev = issue.get("severity", "UNKNOWN").upper()
                sev_css = f"severity-{sev.lower()}" if sev in ("HIGH", "MEDIUM", "LOW", "CRITICAL") else ""
                ftype = issue.get("failure_type", "test")
                ftype_label = "INFRA" if ftype == "infrastructure" else ftype.upper()
                ftype_css = "ftype-infra" if ftype == "infrastructure" else f"ftype-{ftype}"
                jobs_label = f'{jc} {"job" if jc == 1 else "jobs"}'

                lines.append('            <tr class="issue-row">')
                lines.append(f'                <td class="col-num">{issue["number"]}.</td>')
                lines.append(f'                <td class="col-sev"><span class="severity-badge {sev_css}">{sev}</span></td>')
                lines.append(f'                <td class="col-ftype"><span class="ftype-badge {ftype_css}">{ftype_label}</span></td>')
                lines.append(f'                <td class="col-title">{_e(issue["title"])}</td>')
                lines.append(f'                <td class="col-jobs">{jobs_label}</td>')
                lines.append('            </tr>')
                lines.append('            <tr class="detail-row"><td colspan="5">')
                if issue.get("root_cause"):
                    lines.append(f'                <div class="root-cause"><strong>Root Cause:</strong> {_e(issue["root_cause"])}</div>')
                lines.append(f'                <div class="bug-links">{_render_bug_links(bug_match)}</div>')
                if issue.get("affected_jobs"):
                    lines.append("                <p><strong>Affected Jobs:</strong></p><ul>")
                    for job in issue["affected_jobs"]:
                        if job.get("url"):
                            lines.append(f'                    <li><span class="job-date">[{_e(job["date"])}]</span> <a href="{_e(job["url"])}" target="_blank">{_e(job["name"])}</a></li>')
                        else:
                            lines.append(f'                    <li><span class="job-date">[{_e(job["date"])}]</span> {_e(job["name"])}</li>')
                    lines.append("                </ul>")
                if issue.get("next_steps"):
                    lines.append(f"                <p><em>Next Steps:</em> {_e(issue['next_steps'])}</p>")
                lines.append("            </td></tr>")
            lines.append('            </table>')

        lines.append("        </div>")
    return "\n".join(toc_lines) + "\n\n" + "\n".join(lines)


def generate_html(releases_data, bug_data, pr_data, all_pr_bugs, pr_status, timestamp):
    date_str = timestamp.strftime("%Y-%m-%d")
    time_str = timestamp.strftime("%Y-%m-%d %H:%M:%S")

    cards = []
    for version, rdata in releases_data.items():
        count = rdata["total_failed"] if rdata else "?"
        css = "status-fail" if rdata and rdata["total_failed"] > 0 else ("status-pass" if rdata else "")
        cards.append(
            '        <div class="overview-card">\n'
            f'            <div class="number {css}">{count}</div>\n'
            f'            <div class="label">Release {_e(version)}</div>\n'
            "        </div>"
        )
    # PR overview: count failures from status (all PRs) or analysis
    if pr_status:
        pr_failed = sum(p.get("failed", 0) for p in pr_status)
    elif pr_data:
        pr_failed = pr_data.get("total_failed", 0)
    else:
        pr_failed = 0
    pr_css = "status-fail" if pr_failed > 0 else "status-pass"
    cards.append(
        '        <div class="overview-card">\n'
        f'            <div class="number {pr_css}">{pr_failed}</div>\n'
        f'            <div class="label">Rebase PRs</div>\n'
        "        </div>"
    )

    toc = []
    for version, rdata in releases_data.items():
        if rdata:
            b = rdata["breakdown"]
            toc.append(
                f'                <li><a href="#release-{_e(version)}">Release {_e(version)}</a> &mdash; '
                f'{rdata["total_failed"]} failures ({b["build"]} build, {b["test"]} test, {b["infrastructure"]} infra)</li>'
            )
        else:
            toc.append(f'                <li><a href="#release-{_e(version)}">Release {_e(version)}</a> &mdash; no data</li>')

    sections = []
    for version, rdata in releases_data.items():
        bugs = bug_data.get(version, [])
        sections.append(render_release_section(version, rdata, bugs))

    pr_section = render_pr_section(pr_data, all_pr_bugs, pr_status)

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
            sys.exit(1)
        else:
            releases_arg = args[i]
            i += 1

    if not releases_arg:
        print("Usage: analyze-ci-create-report.py [--workdir DIR] <release1,release2,...>", file=sys.stderr)
        sys.exit(1)

    releases = [v.strip() for v in releases_arg.split(",") if v.strip()]
    if not releases:
        print("Error: at least one release version is required", file=sys.stderr)
        sys.exit(1)

    if workdir is None:
        workdir = f"/tmp/analyze-ci-claude-workdir.{datetime.now().strftime('%y%m%d')}"

    if not os.path.isdir(workdir):
        print(f"Error: work directory does not exist: {workdir}", file=sys.stderr)
        sys.exit(1)

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
        parts.append("bug mapping found" if entry["bugs"] else "no bug mapping")
        print(f"  Release {version}: {', '.join(parts)}")

    pr_entry = files["prs"]
    if pr_entry["summary"] or pr_entry["status"]:
        found_any = True
        parts = []
        if pr_entry["summary"]:
            parts.append("summary found")
        if pr_entry["status"]:
            parts.append("status found")
        parts.append(f'{len(pr_entry["bugs"])} bug mapping files')
        print(f"  PRs: {', '.join(parts)}")
    else:
        print("  PRs: no data")

    if not found_any:
        print(f"\nError: no analysis files found in {workdir}", file=sys.stderr)
        sys.exit(1)

    # Load everything via json.load
    releases_data = {}
    bug_data = {}
    for version in releases:
        entry = files["releases"][version]
        releases_data[version] = load_json(entry["summary"])
        bug_data[version] = load_bug_candidates(entry["bugs"])

    pr_data = load_json(pr_entry["summary"])
    pr_status = load_json(pr_entry["status"])

    all_pr_bugs = _index_pr_bugs(pr_entry["bugs"])

    # Generate HTML
    timestamp = datetime.now(timezone.utc)
    html_content = generate_html(releases_data, bug_data, pr_data, all_pr_bugs, pr_status, timestamp)

    output_path = os.path.join(workdir, "microshift-ci-doctor-report.html")
    with open(output_path, "w") as f:
        f.write(html_content)

    # Summary
    print("\nSummary:")
    print("  Periodics:")
    for version in releases:
        rdata = releases_data[version]
        if rdata:
            print(f"    Release {version}: {rdata['total_failed']} failed periodic jobs")
        else:
            print(f"    Release {version}: no data")
    print("  Pull Requests:")
    if pr_status:
        pr_total_failed = sum(p.get("failed", 0) for p in pr_status)
        pr_total_pending = sum(p.get("pending", 0) for p in pr_status)
        parts = [f"{len(pr_status)} rebase PRs", f"{pr_total_failed} failed jobs"]
        if pr_total_pending:
            parts.append(f"{pr_total_pending} running")
        print(f"    {', '.join(parts)}")
    elif pr_data and pr_data.get("has_content"):
        print(f"    {len(pr_data['prs'])} rebase PRs with {pr_data['total_failed']} total failed jobs")
    else:
        print("    No PR data")
    print(f"\nHTML report generated: {output_path}")


if __name__ == "__main__":
    main()
