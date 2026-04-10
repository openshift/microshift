#!/usr/bin/env python3
"""X/Y/Z release evaluation for MicroShift.

Evaluates whether MicroShift should participate in upcoming OCP X, Y, or Z
releases by checking lifecycle status, OCP availability, advisory CVEs,
code changes, and the 90-day rule.

Usage: precheck_xyz.py [version...] [--verbose]
"""

import argparse
import json
import logging
import os
import subprocess
import sys
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import datetime

from lib import art_jira, brew, git_ops, lifecycle, pyxis, release_controller

logging.basicConfig(
    level=logging.INFO,
    format="%(levelname)s: %(message)s",
    stream=sys.stderr,
)
logger = logging.getLogger(__name__)


def run_advisory_report(version, repo_root):
    """Call advisory_publication_report.sh as subprocess.

    Args:
        version: Full version, e.g., "4.21.8".
        repo_root: Path to the git repository root.

    Returns:
        dict: Parsed JSON report, or {"error": "...", "skipped": True} on failure.
    """
    # Check prerequisites
    missing = []
    for var in ["ATLASSIAN_API_TOKEN", "ATLASSIAN_EMAIL"]:
        if not os.environ.get(var, "").strip():
            missing.append(var)
    parts = version.split(".")
    if len(parts) >= 2 and parts[1].isdigit():
        minor_int = int(parts[1])
        if minor_int >= 20 and not os.environ.get("GITLAB_API_TOKEN", "").strip():
            missing.append("GITLAB_API_TOKEN")

    if missing:
        return {"error": f"Missing env vars: {', '.join(missing)}", "skipped": True}

    # Check VPN
    if not brew.check_vpn():
        return {"error": "VPN not connected", "skipped": True}

    script = os.path.join(
        repo_root, "scripts", "advisory_publication", "advisory_publication_report.sh"
    )
    if not os.path.exists(script):
        return {"error": f"Script not found: {script}", "skipped": True}

    logger.info("Running advisory publication report for %s...", version)
    try:
        result = subprocess.run(
            ["bash", script, version],
            capture_output=True, text=True, timeout=120,
        )
        if result.returncode == 0:
            stdout = result.stdout.strip()
            # The advisory script may print warnings before the JSON.
            # Find the outermost JSON object by matching the last '}'
            # back to its opening '{'.
            json_end = stdout.rfind("}")
            if json_end >= 0:
                json_start = stdout.find("{")
                if json_start >= 0 and json_start < json_end:
                    return json.loads(stdout[json_start:json_end + 1])
            return {"error": "No JSON found in advisory report output", "skipped": True}
        return {"error": result.stderr.strip(), "skipped": True}
    except subprocess.TimeoutExpired:
        return {"error": "Advisory report timed out (120s)", "skipped": True}
    except json.JSONDecodeError as e:
        return {"error": f"Invalid JSON from advisory report: {e}", "skipped": True}


def interpret_cves(advisory_report):
    """Interpret CVE results from the advisory report.

    Rules from the MicroShift release process:
    - Empty cves dict -> no CVEs -> no action
    - CVE with empty dict (no Jira ticket) -> does NOT affect MicroShift -> no action
    - CVE with resolution "Done-Errata" -> MUST release
    - CVE with resolution "Not a Bug" -> no action
    - CVE with status "In Progress" -> flag as NEEDS REVIEW

    Args:
        advisory_report: Parsed JSON from advisory_publication_report.sh.

    Returns:
        dict: {"impact": "none"|"must_release"|"needs_review", "details": [...]}
    """
    if not advisory_report or advisory_report.get("skipped"):
        return {"impact": "unknown", "details": ["Advisory report was skipped"]}

    must_release_cves = []
    needs_review_cves = []
    advisory_types_checked = []

    for advisory_name, advisory_data in advisory_report.items():
        advisory_type = advisory_data.get("type", "unknown")
        # Skip metadata advisories
        if advisory_type == "metadata":
            continue
        advisory_types_checked.append(advisory_type)

        cves = advisory_data.get("cves", {})
        for cve_id, cve_data in cves.items():
            jira_ticket = cve_data.get("jira_ticket")
            if not jira_ticket:
                # No MicroShift Jira ticket -> CVE does not affect MicroShift
                continue

            resolution = jira_ticket.get("resolution", "")
            status = jira_ticket.get("status", "")

            if resolution == "Done-Errata":
                must_release_cves.append({
                    "cve": cve_id,
                    "jira": jira_ticket.get("id", ""),
                    "reason": "Fix released via errata",
                })
            elif resolution == "Not a Bug":
                continue
            elif status == "In Progress":
                needs_review_cves.append({
                    "cve": cve_id,
                    "jira": jira_ticket.get("id", ""),
                    "reason": "Fix in progress",
                })

    if must_release_cves:
        return {
            "impact": "must_release",
            "details": must_release_cves,
            "advisory_types": advisory_types_checked,
        }
    if needs_review_cves:
        return {
            "impact": "needs_review",
            "details": needs_review_cves,
            "advisory_types": advisory_types_checked,
        }
    return {
        "impact": "none",
        "details": [],
        "advisory_types": advisory_types_checked,
    }


def compute_recommendation(evaluation):
    """Compute the final recommendation for a version.

    Decision rules:
    - RELEASE: critical fix, applicable CVE (Done-Errata), or 90-day rule
    - SKIP: no changes, no CVEs, within 90 days
    - NEEDS REVIEW: ambiguous cases

    Args:
        evaluation: Dict with version evaluation data.

    Returns:
        tuple[str, str]: (recommendation, reason).
    """
    cve_impact = evaluation.get("cve_impact", {}).get("impact", "unknown")
    commits = evaluation.get("commits", 0)
    days_since = evaluation.get("days_since")

    # Must release: CVE with Done-Errata
    if cve_impact == "must_release":
        cve_details = evaluation.get("cve_impact", {}).get("details", [])
        cve_list = ", ".join(d["cve"] for d in cve_details)
        return "ASK ART TO CREATE ARTIFACTS", f"CVE fix: {cve_list}"

    # 90-day rule
    if days_since is not None and days_since >= 90 and commits > 0:
        return "ASK ART TO CREATE ARTIFACTS", f"90-day rule ({days_since}d since last release, {commits} commits)"

    # Needs review: CVE in progress
    if cve_impact == "needs_review":
        return "NEEDS REVIEW", "CVE fix in progress"

    # Needs review: advisory report skipped
    if cve_impact == "unknown":
        if commits > 0:
            return "NEEDS REVIEW", f"{commits} commits, advisory report unavailable"
        return "SKIP", "No commits, advisory report unavailable"

    # Skip: no changes
    if commits == 0:
        days_str = f"{days_since}d since last release" if days_since is not None else "unknown last release"
        return "SKIP", f"No commits ({days_str})"

    # Has commits but no CVEs and within 90 days
    if days_since is not None:
        return "SKIP", f"{days_since}d since last release, {commits} commits, no CVEs"

    return "SKIP", f"{commits} commits, no CVEs"


def evaluate_version(version, lifecycle_data, repo_root):
    """Run full evaluation pipeline for one version.

    Args:
        version: Full version, e.g., "4.21.8".
        lifecycle_data: Output from lifecycle.fetch_lifecycle_data().
        repo_root: Path to the git repository root.

    Returns:
        dict: Evaluation result with recommendation.
    """
    minor = ".".join(version.split(".")[:2])
    result = {"version": version, "minor": minor}

    # Lifecycle check
    lc = lifecycle.get_lifecycle_status(minor, lifecycle_data)
    if lc:
        result["lifecycle_status"] = lc["phase"]
        result["lifecycle_end_date"] = lc.get("end_date", "")
        if not lc["active"]:
            result["recommendation"] = "SKIP"
            result["reason"] = f"{lc['phase']} (until {lc.get('end_date', 'N/A')})"
            return result
    else:
        result["lifecycle_status"] = "unknown"

    # Already released check (Pyxis)
    logger.info("Checking if %s is already released...", version)
    if pyxis.is_version_published(version):
        result["already_released"] = True
        result["recommendation"] = "ALREADY RELEASED"
        result["reason"] = "MicroShift errata published"
        return result
    result["already_released"] = False

    # OCP payload status
    logger.info("Checking OCP payload for %s...", version)
    result["ocp_status"] = release_controller.check_ocp_payload_accepted(version)

    # ART ticket lookup
    art_tickets = art_jira.query_art_releases_due(specific_version=version)
    if art_tickets:
        result["art_ticket"] = art_tickets[0]["key"]
        result["due_date"] = art_tickets[0].get("due_date", "")
    else:
        result["art_ticket"] = None
        result["due_date"] = ""

    # Z-stream evaluation
    # 4a: Code changes since last release
    branch = f"release-{minor}"
    logger.info("Fetching commits on %s...", branch)
    git_ops.fetch_branch(branch)

    last_pub = pyxis.find_latest_published_zstream(minor)
    if last_pub:
        result["last_released"] = last_pub["version"]
        since_version = last_pub["version"]
    else:
        result["last_released"] = f"{minor}.0"
        since_version = None

    commit_list = git_ops.commits_since(branch, since_version)
    result["commits"] = len(commit_list)
    result["commit_list"] = commit_list

    # 4b: Advisory publication report
    result["advisory_report"] = run_advisory_report(version, repo_root)

    # 4c: Interpret CVEs
    result["cve_impact"] = interpret_cves(result["advisory_report"])

    # 4d: 90-day rule — get date of last release from git tags
    if last_pub:
        release_date = git_ops.get_release_date(last_pub["version"])
        if release_date:
            try:
                build_date = datetime.strptime(release_date, "%Y-%m-%d")
                result["days_since"] = (datetime.now() - build_date).days
                result["last_release_date"] = release_date
            except (ValueError, TypeError):
                result["days_since"] = None
        else:
            result["days_since"] = None
    else:
        result["days_since"] = None

    # 4e: Recommendation
    result["recommendation"], result["reason"] = compute_recommendation(result)

    return result


def expand_versions(version_args, lifecycle_data):
    """Expand version arguments (X.Y -> query ART for specific z-stream).

    Args:
        version_args: List of version strings from CLI.
        lifecycle_data: Lifecycle data.

    Returns:
        list[str]: Expanded version strings.
    """
    versions = []
    for v in version_args:
        parts = v.split(".")
        if len(parts) == 2:
            # Minor version: query ART for specific releases
            art_tickets = art_jira.query_art_releases_due(minor_version=v)
            if art_tickets:
                for ticket in art_tickets:
                    versions.append(ticket["version"])
            else:
                # No ART tickets — can't determine specific z-stream
                logger.warning(
                    "No ART tickets found for %s, cannot determine specific z-stream", v
                )
        elif len(parts) == 3:
            versions.append(v)
        else:
            logger.warning("Invalid version format: %s", v)
    return versions


def _build_reason(e):
    """Build the reason string for a version evaluation."""
    parts = []

    # CVE / advisory impact
    cve_impact = e.get("cve_impact", {})
    impact = cve_impact.get("impact", "unknown")
    if impact == "must_release":
        details = cve_impact.get("details", [])
        cve_list = ", ".join(d.get("cve", "") for d in details)
        parts.append(f"CVE fix: {cve_list}")
    elif impact == "needs_review":
        parts.append("CVE in progress")
    elif impact == "none":
        parts.append("no CVEs")
    elif impact == "unknown":
        advisory = e.get("advisory_report", {})
        if advisory and advisory.get("skipped"):
            parts.append("advisory report unavailable")
        else:
            parts.append("advisory unknown")

    # Last released
    days = e.get("days_since")
    last = e.get("last_released", "")
    if days is not None and last:
        parts.append(f"last: {last} ({days}d ago)")
    elif last:
        parts.append(f"last: {last}")

    return " | ".join(parts) if parts else "no data"


def format_text_short(evaluations):
    """Format evaluations as one-line-per-version text.

    Format: ACTION x.y.z [OCP: available/NOT available] [reason]

    Args:
        evaluations: List of evaluation result dicts.

    Returns:
        str: Pre-formatted text output.
    """
    if not evaluations:
        return "No versions to evaluate."

    REC_WIDTH = 28  # len("ASK ART TO CREATE ARTIFACTS")
    lines = []

    for e in evaluations:
        version = e.get("version", "?")
        rec = e.get("recommendation", "UNKNOWN")

        if rec == "ALREADY RELEASED":
            lines.append(f"{rec:<{REC_WIDTH}} {version}")
            continue

        # OCP status
        ocp = e.get("ocp_status", "")
        if ocp == "available":
            ocp_str = "available"
        elif not ocp:
            ocp_str = "unknown"
        else:
            ocp_str = "NOT available"

        # Build reason — prefer canonical reason from compute_recommendation()
        lc = e.get("lifecycle_status", "")
        if lc in ("End of life", "Extended Support"):
            end = e.get("lifecycle_end_date", "N/A")
            reason = f"{lc}, until {end}"
        else:
            reason = e.get("reason") or _build_reason(e)

        lines.append(f"{rec:<{REC_WIDTH}} {version} [OCP: {ocp_str}] [{reason}]")

    return "\n".join(lines)


def format_text_full(output):
    """Format evaluations as detailed markdown report.

    Args:
        output: Full output dict with lifecycle and evaluations.

    Returns:
        str: Markdown-formatted report.
    """
    evaluations = output.get("evaluations", [])
    if not evaluations:
        return "No versions to evaluate."

    sections = []

    # Release Schedule table
    sections.append("## Release Schedule\n")
    sections.append("| Version | ART Ticket | Due Date | OCP Status | Lifecycle |")
    sections.append("|---------|-----------|----------|------------|-----------|")
    for e in evaluations:
        v = e.get("version", "?")
        art = e.get("art_ticket", "--")
        due = e.get("due_date", "--") or "--"
        ocp = e.get("ocp_status", "--")
        lc = e.get("lifecycle_status", "--")
        sections.append(f"| {v} | {art} | {due} | {ocp} | {lc} |")

    # Z-Stream Evaluation table
    sections.append("\n## Z-Stream Evaluation\n")
    sections.append("| Version | Last Released | Days Since | Commits | CVE Impact |")
    sections.append("|---------|--------------|------------|---------|------------|")
    for e in evaluations:
        if e.get("already_released") or e.get("recommendation") == "ALREADY RELEASED":
            continue
        v = e.get("version", "?")
        last = e.get("last_released", "--")
        days = str(e.get("days_since", "--")) if e.get("days_since") is not None else "--"
        commits = str(e.get("commits", 0))
        impact = e.get("cve_impact", {}).get("impact", "--")
        sections.append(f"| {v} | {last} | {days} | {commits} | {impact} |")

    # Advisory Report table
    has_advisories = any(
        e.get("advisory_report") and not e["advisory_report"].get("skipped")
        for e in evaluations
    )
    if has_advisories:
        sections.append("\n## Advisory Report\n")
        sections.append("| Version | Advisory | Type | CVEs | MicroShift Impact |")
        sections.append("|---------|----------|------|------|-------------------|")
        for e in evaluations:
            report = e.get("advisory_report", {})
            if not report or report.get("skipped"):
                continue
            v = e.get("version", "?")
            for adv_name, adv_data in report.items():
                # Skip non-advisory keys (e.g., "error", "skipped")
                if not isinstance(adv_data, dict) or "type" not in adv_data:
                    continue
                adv_type = adv_data.get("type", "?")
                cves = adv_data.get("cves", {})
                if not cves:
                    sections.append(f"| {v} | {adv_name} | {adv_type} | none | -- |")
                else:
                    for cve_id, cve_data in cves.items():
                        jira_ticket = cve_data.get("jira_ticket")
                        if jira_ticket:
                            impact = f"{jira_ticket.get('id', '?')} ({jira_ticket.get('resolution', '?')})"
                        else:
                            impact = "not affected"
                        sections.append(f"| {v} | {adv_name} | {adv_type} | {cve_id} | {impact} |")

    # Recommendations table
    sections.append("\n## Recommendations\n")
    sections.append("| Version | Recommendation | Reason |")
    sections.append("|---------|---------------|--------|")
    for e in evaluations:
        v = e.get("version", "?")
        rec = e.get("recommendation", "UNKNOWN")
        reason = e.get("reason", "").replace("|", "\\|").replace("\n", " ")
        sections.append(f"| {v} | {rec} | {reason} |")

    return "\n".join(sections)


def main():
    parser = argparse.ArgumentParser(description="MicroShift X/Y/Z release evaluation")
    parser.add_argument("versions", nargs="*", help="X.Y or X.Y.Z versions")
    parser.add_argument("--verbose", action="store_true",
                        help="Show detailed tables instead of one-line summary")
    parser.add_argument("--json", action="store_true", dest="json_output",
                        help="Output raw JSON instead of formatted text")
    args = parser.parse_args()

    # Step 1: Fetch lifecycle data
    logger.info("Fetching lifecycle data...")
    lifecycle_data = lifecycle.fetch_lifecycle_data()

    repo_root = git_ops.get_repo_root()

    # Step 2: Determine versions to evaluate
    if args.versions:
        versions = expand_versions(args.versions, lifecycle_data)
    else:
        # No versions specified: query ART Jira for releases due within 7 days
        logger.info("Querying ART Jira for releases due within 7 days...")
        art_tickets = art_jira.query_art_releases_due(days_ahead=7)
        versions = []
        for ticket in art_tickets:
            v = ticket["version"]
            minor = ".".join(v.split(".")[:2])
            if lifecycle.is_version_active(minor, lifecycle_data):
                versions.append(v)
        if not versions:
            logger.info("No ART releases due within 7 days")

    # Step 3: Evaluate each version (parallel when multiple)
    evaluations = []
    if len(versions) > 1:
        with ThreadPoolExecutor(max_workers=4) as executor:
            futures = {
                executor.submit(evaluate_version, v, lifecycle_data, repo_root): v
                for v in versions
            }
            for future in as_completed(futures):
                try:
                    evaluations.append(future.result())
                except Exception as e:
                    v = futures[future]
                    logger.warning("Evaluation failed for %s: %s", v, e)
                    evaluations.append({
                        "version": v,
                        "recommendation": "NEEDS REVIEW",
                        "reason": f"evaluation error: {e}",
                    })
        # Restore original version ordering
        version_order = {v: i for i, v in enumerate(versions)}
        evaluations.sort(key=lambda e: version_order.get(e["version"], 0))
    else:
        for version in versions:
            logger.info("Evaluating %s...", version)
            try:
                result = evaluate_version(version, lifecycle_data, repo_root)
            except Exception as e:
                logger.warning("Evaluation failed for %s: %s", version, e)
                result = {
                    "version": version,
                    "recommendation": "NEEDS REVIEW",
                    "reason": f"evaluation error: {e}",
                }
            evaluations.append(result)

    # Step 4: Output
    output = {
        "command": "precheck_xyz",
        "verbose": args.verbose,
        "timestamp": datetime.now().isoformat(),
        "lifecycle": lifecycle_data,
        "evaluations": evaluations,
    }

    if args.json_output:
        print(json.dumps(output, indent=2))
    elif args.verbose:
        print(format_text_full(output))
    else:
        print(format_text_short(evaluations))


if __name__ == "__main__":
    main()
