"""ART Jira queries for release schedule (optional, graceful degradation)."""

import logging
import os
import re
from datetime import datetime, timedelta

ATLASSIAN_URL = "https://redhat.atlassian.net"

# JQL special characters that must be escaped in user-supplied values
_JQL_SPECIAL_RE = re.compile(r'[\\"\'\[\](){}+\-&|!~*?^/]')

logger = logging.getLogger(__name__)


def _sanitize_jql_value(value):
    """Escape JQL special characters in a value string.

    Args:
        value: Raw string to be embedded in a JQL query.

    Returns:
        str: Escaped string safe for JQL interpolation.
    """
    return _JQL_SPECIAL_RE.sub(r"\\\g<0>", value)


def create_jira_client():
    """Create a Jira client from environment variables.

    Requires ATLASSIAN_API_TOKEN and ATLASSIAN_EMAIL.

    Returns:
        jira.JIRA or None: Client instance, or None if env vars are missing.
    """
    token = os.environ.get("ATLASSIAN_API_TOKEN", "").strip()
    email = os.environ.get("ATLASSIAN_EMAIL", "").strip()

    if not token or not email:
        logger.info("Jira credentials not set, skipping Jira queries")
        return None

    try:
        import jira as jira_lib
    except ImportError:
        logger.warning("jira package not installed, skipping Jira queries")
        return None

    try:
        return jira_lib.JIRA(server=ATLASSIAN_URL, basic_auth=(email, token))
    except Exception as e:
        logger.warning("Failed to connect to Jira: %s", e)
        return None


def _extract_version_from_summary(summary):
    """Extract version from ART ticket summary.

    ART tickets use format: "Release X.Y.Z [YYYY-Mon-DD]"

    Args:
        summary: Ticket summary string.

    Returns:
        str or None: Extracted version, e.g., "4.21.8".
    """
    match = re.search(r"Release\s+(4\.\d+\.\d+)", summary)
    return match.group(1) if match else None


def _extract_date_from_summary(summary):
    """Extract date from ART ticket summary.

    ART tickets use format: "Release X.Y.Z [YYYY-Mon-DD]"

    Args:
        summary: Ticket summary string.

    Returns:
        str or None: Date in YYYY-MM-DD format.
    """
    match = re.search(r"\[(\d{4}-\w{3}-\d{2})\]", summary)
    if not match:
        return None
    try:
        dt = datetime.strptime(match.group(1), "%Y-%b-%d")
        return dt.strftime("%Y-%m-%d")
    except ValueError:
        return None


def query_art_releases_due(days_ahead=7, minor_version=None, specific_version=None):
    """Query ART Jira for in-progress release stories.

    Args:
        days_ahead: Look ahead N days when no version is specified.
        minor_version: Filter to a specific minor, e.g., "4.21".
        specific_version: Filter to exact version, e.g., "4.21.8".

    Returns:
        list[dict]: Each dict has keys: key, summary, status, due_date, version.
            Returns empty list if Jira is unavailable.
    """
    client = create_jira_client()
    if not client:
        return []

    if specific_version:
        safe_version = _sanitize_jql_value(specific_version)
        jql = (
            f'project = ART AND summary ~ "Release {safe_version}" '
            f'ORDER BY created DESC'
        )
    elif minor_version:
        safe_version = _sanitize_jql_value(minor_version)
        jql = (
            f'project = ART AND summary ~ "Release {safe_version}" '
            f'AND status = "In Progress" ORDER BY duedate ASC'
        )
    else:
        next_week = (datetime.now() + timedelta(days=days_ahead)).strftime("%Y-%m-%d")
        jql = (
            f'project = ART AND summary ~ "Release" '
            f'AND status = "In Progress" AND duedate <= "{next_week}" '
            f'ORDER BY duedate ASC'
        )

    try:
        issues = client.search_issues(jql, maxResults=20, fields="summary,status,duedate")
    except Exception as e:
        logger.warning("Jira query failed: %s", e)
        return []

    results = []
    for issue in issues:
        summary = issue.fields.summary
        version = _extract_version_from_summary(summary)
        if not version:
            continue
        # Only include 4.14+
        parts = version.split(".")
        if len(parts) >= 2 and int(parts[1]) < 14:
            continue

        due_date = str(issue.fields.duedate) if issue.fields.duedate else _extract_date_from_summary(summary)

        results.append({
            "key": issue.key,
            "summary": summary,
            "status": str(issue.fields.status),
            "due_date": due_date,
            "version": version,
        })

    return results


def query_art_ecrc(version_pattern):
    """Look up a single ART ticket for an EC/RC version.

    Args:
        version_pattern: e.g., "4.22.0-ec.5".

    Returns:
        dict or None: {"key": "ART-14768", "status": "In Progress"} or None.
    """
    client = create_jira_client()
    if not client:
        return None

    safe_pattern = _sanitize_jql_value(version_pattern)
    jql = (
        f'project = ART AND summary ~ "{safe_pattern}" '
        f'ORDER BY created DESC'
    )

    try:
        issues = client.search_issues(jql, maxResults=1, fields="summary,status")
    except Exception as e:
        logger.warning("Jira EC/RC query failed: %s", e)
        return None

    if not issues:
        return None

    issue = issues[0]
    return {
        "key": issue.key,
        "status": str(issue.fields.status),
    }
