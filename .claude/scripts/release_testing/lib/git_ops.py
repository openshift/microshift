"""Git operations via subprocess."""

import logging
import re
import subprocess

logger = logging.getLogger(__name__)


def get_repo_root():
    """Return the git repository root directory.

    Returns:
        str: Absolute path to the repo root.

    Raises:
        RuntimeError: If not inside a git repository.
    """
    result = subprocess.run(
        ["git", "rev-parse", "--show-toplevel"],
        capture_output=True, text=True, check=True,
    )
    return result.stdout.strip()


def list_release_branches():
    """List active release-X.Y branches from the remote.

    Returns:
        list[str]: Minor versions sorted descending, e.g., ["5.1", "5.0", "4.22", ...].
    """
    result = subprocess.run(
        ["git", "branch", "-r", "--list", "origin/release-[45].*"],
        capture_output=True, text=True,
    )
    if result.returncode != 0:
        logger.warning("git branch -r failed: %s", result.stderr.strip())
        return []

    # Match only X.Y branches (not feature branches like 4.21-common-versions-update).
    # Filter to 4.14+ — MicroShift GA'd at 4.14, older branches have no MicroShift builds.
    pattern = re.compile(r"origin/release-([45]\.\d+)$")
    versions = []
    for line in result.stdout.strip().split("\n"):
        match = pattern.search(line.strip())
        if match:
            v = match.group(1)
            parts = v.split(".")
            if parts[0] == "4" and int(parts[1]) < 14:
                continue
            versions.append(v)

    versions.sort(key=lambda v: [int(x) for x in v.split(".")], reverse=True)
    return versions


def get_release_date(version):
    """Get the release date for a version from git tags.

    Tags follow the format: X.Y.Z-YYYYMMDDHHMMSS.p0

    Args:
        version: Version string, e.g., "4.21.7".

    Returns:
        str or None: Date in YYYY-MM-DD format, or None if not found.
    """
    result = subprocess.run(
        ["git", "tag", "-l", f"{version}-*"],
        capture_output=True, text=True,
    )
    if result.returncode != 0 or not result.stdout.strip():
        return None

    # Take the latest tag for this version
    tags = result.stdout.strip().split("\n")
    tags.sort(reverse=True)
    tag = tags[0]

    # Extract date from tag: 4.21.7-202603230928.p0 → 2026-03-23
    match = re.search(rf"{re.escape(version)}-(\d{{4}})(\d{{2}})(\d{{2}})", tag)
    if match:
        return f"{match.group(1)}-{match.group(2)}-{match.group(3)}"
    return None


def fetch_branch(branch):
    """Run 'git fetch origin <branch>'.

    Args:
        branch: Branch name, e.g., "release-4.21".

    Returns:
        bool: True on success, False on failure.
    """
    result = subprocess.run(
        ["git", "fetch", "origin", branch],
        capture_output=True, text=True,
    )
    if result.returncode != 0:
        logger.warning("git fetch origin %s failed: %s", branch, result.stderr.strip())
        return False
    return True


def find_version_tag(version):
    """Find the git tag for a given version.

    Tags follow the format: X.Y.Z-YYYYMMDDHHMMSS.p0

    Args:
        version: Version string, e.g., "4.18.36".

    Returns:
        str or None: The latest tag for the version, or None if not found.
    """
    result = subprocess.run(
        ["git", "tag", "-l", f"{version}-*"],
        capture_output=True, text=True,
    )
    if result.returncode != 0 or not result.stdout.strip():
        return None

    tags = result.stdout.strip().split("\n")
    tags.sort(reverse=True)
    return tags[0]


def commits_since(branch, since_version):
    """Get commits on origin/<branch> since a version tag.

    Uses a tag-based range (tag..origin/branch) to count only commits
    after the specified version, rather than a time-based --since filter.

    Args:
        branch: Branch name, e.g., "release-4.21".
        since_version: Version string, e.g., "4.18.36", or None to get
            all commits on the branch.

    Returns:
        list[dict]: Each dict has keys: sha, subject, date.
    """
    tag = find_version_tag(since_version) if since_version else None
    if tag:
        revision = f"{tag}..origin/{branch}"
    else:
        revision = f"origin/{branch}"

    result = subprocess.run(
        [
            "git", "log", revision,
            "--format=%H%x00%s%x00%ai",
        ],
        capture_output=True, text=True,
    )
    if result.returncode != 0:
        logger.warning("git log origin/%s failed: %s", branch, result.stderr.strip())
        return []

    commits = []
    for line in result.stdout.strip().split("\n"):
        if not line:
            continue
        parts = line.split("\x00", 2)
        if len(parts) == 3:
            commits.append({
                "sha": parts[0][:12],
                "subject": parts[1],
                "date": parts[2].split(" ")[0],
            })

    return commits
