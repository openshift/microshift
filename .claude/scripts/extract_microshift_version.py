#!/usr/bin/env python3
"""
Extract MicroShift version from Prow CI journal logs.

This script fetches the journal log from a Prow CI job and extracts the
exact MicroShift version being tested from the systemd journal output.

Usage:
    python3 extract_microshift_version.py <prow_url> <scenario>

Arguments:
    prow_url: The Prow CI job URL
        e.g.: "https://prow.ci.openshift.org/view/gs/test-platform-results/pr-logs/pull/openshift_microshift/5703/pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-release/1990417342856695808"
    scenario: Scenario name
        e.g.: "el96-lrel@backups"

Output:
    JSON object with:
    {
        "version": "4.20.0-202511160032.p0.gb46fe41.assembly.stream.el9",
        "build_type": "nightly",
        "success": true,
        "error": null
    }
"""

import sys
import json
import re
import urllib.request
import urllib.error
import urllib.parse
import ssl
from html.parser import HTMLParser


class GCSWebLinkParser(HTMLParser):
    """Parse GCSWeb HTML to extract file links."""

    def __init__(self):
        super().__init__()
        self.links = []

    def handle_starttag(self, tag, attrs):
        if tag == 'a':
            for attr, value in attrs:
                if attr == 'href' and value and not value.startswith('?'):
                    self.links.append(value)


def construct_journal_log_dir_url(job_id, version, job_type, pr_number=None, scenario="el96-lrel@ipv6"):
    """Construct the URL to the journal log directory for a given job."""
    base_url = "https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/test-platform-results"

    if pr_number:
        # PR job URL pattern
        job_name = f"pull-ci-openshift-microshift-release-{version}-{job_type}"
        path = f"pr-logs/pull/openshift_microshift/{pr_number}/{job_name}/{job_id}"
    else:
        # Periodic job URL pattern
        job_name = f"periodic-ci-openshift-microshift-release-{version}-periodics-{job_type}"
        path = f"logs/{job_name}/{job_id}"

    artifact_path = f"artifacts/{job_type}/openshift-microshift-e2e-metal-tests/artifacts/scenario-info/{scenario}/vms/host1/sos"

    return f"{base_url}/{path}/{artifact_path}/"


def fetch_url(url):
    """Fetch content from the given URL."""
    try:
        # Validate URL scheme
        if not url.startswith('https://'):
            return None, "Invalid URL scheme. Only HTTPS URLs are supported."

        # Create SSL context that doesn't verify certificates
        ssl_context = ssl.create_default_context()
        ssl_context.check_hostname = False
        ssl_context.verify_mode = ssl.CERT_NONE

        with urllib.request.urlopen(url, timeout=30, context=ssl_context) as response:
            content = response.read().decode('utf-8')
            return content, None
    except urllib.error.URLError as e:
        return None, f"Failed to fetch URL: {e}"
    except (UnicodeDecodeError, OSError) as e:
        return None, f"Error reading URL: {e}"


def find_journal_log_file(dir_url):
    """Find the journal log file in the directory listing."""
    html_content, error = fetch_url(dir_url)
    if error:
        return None, error

    # Parse HTML to find links
    parser = GCSWebLinkParser()
    parser.feed(html_content)

    # Find journal_*.log files
    journal_files = []
    for link in parser.links:
        # Extract just the filename from the full path
        filename = link.split('/')[-1]
        # URL-decode the filename
        decoded_filename = urllib.parse.unquote(filename)
        if decoded_filename.startswith('journal_') and decoded_filename.endswith('.log'):
            # Keep the URL-encoded filename for URL construction
            journal_files.append(filename)

    if not journal_files:
        return None, "No journal log files found in directory"

    # Return the first journal log file (there should typically be only one)
    return journal_files[0], None


def extract_version_from_journal(log_content):
    """
    Extract MicroShift version from journal log content.

    Looks for pattern: "Version" microshift="4.20.0-202511160032.p0.gb46fe41.assembly.stream.el9"
    """
    # Pattern to match: "Version" microshift="<version>"
    pattern = r'"Version"\s+microshift="([^"]+)"'

    matches = re.findall(pattern, log_content)
    if matches:
        version_string = matches[-1].strip()
        return version_string, None

    return None, "Could not find MicroShift version in journal log"


def determine_build_type(version_string):
    """
    Determine the build type from the version string.

    Returns one of: "nightly", "ec", "rc", "zstream"
    """
    if "nightly" in version_string.lower():
        return "nightly"
    elif "-ec." in version_string:
        return "ec"
    elif "-rc." in version_string:
        return "rc"
    else:
        # Check if it's a date-based build (likely nightly/stream)
        if re.search(r'\d{12}\.p\d+\.g[a-f0-9]+', version_string):
            return "nightly"
        return "zstream"


def parse_prow_url(prow_url):
    """
    Parse a Prow CI URL to extract job information.

    Supported URL formats:
    - PR jobs: https://prow.ci.openshift.org/view/gs/test-platform-results/pr-logs/pull/openshift_microshift/{pr_number}/{job_name}/{job_id}
    - Periodic jobs: https://prow.ci.openshift.org/view/gs/test-platform-results/logs/{job_name}/{job_id}

    Returns:
        Tuple of (job_id, version, job_type, pr_number) or (None, None, None, None) on error
    """
    # Remove the prow.ci.openshift.org prefix and normalize
    url_parts = prow_url.replace("https://prow.ci.openshift.org/view/gs/test-platform-results/", "")

    # Split by '/'
    parts = url_parts.split('/')

    if len(parts) < 2:
        return None, None, None, None, "Invalid URL format"

    # Check if it's a PR job or periodic job
    if parts[0] == "pr-logs" and len(parts) >= 6:
        # PR job format: pr-logs/pull/openshift_microshift/{pr_number}/{job_name}/{job_id}
        pr_number = parts[3]
        job_name = parts[4]
        job_id = parts[5]

        # Extract version and job_type from job_name
        # Format: pull-ci-openshift-microshift-release-{version}-{job_type}
        # or: pull-ci-openshift-microshift-release-{version}-{job_type}-{arch}
        if job_name.startswith("pull-ci-openshift-microshift-release-"):
            name_parts = job_name.replace("pull-ci-openshift-microshift-release-", "").split("-", 1)
            if len(name_parts) >= 2:
                version = name_parts[0]
                job_type = name_parts[1]
                return job_id, version, job_type, pr_number, None

        return None, None, None, None, f"Could not parse PR job name: {job_name}"

    elif parts[0] == "logs" and len(parts) >= 3:
        # Periodic job format: logs/{job_name}/{job_id}
        job_name = parts[1]
        job_id = parts[2]

        # Extract version and job_type from job_name
        # Format: periodic-ci-openshift-microshift-release-{version}-periodics-{job_type}
        if job_name.startswith("periodic-ci-openshift-microshift-release-"):
            name_parts = job_name.replace("periodic-ci-openshift-microshift-release-", "")
            # Split on "-periodics-" to separate version from job_type
            if "-periodics-" in name_parts:
                version, job_type = name_parts.split("-periodics-", 1)
                return job_id, version, job_type, None, None

        return None, None, None, None, f"Could not parse periodic job name: {job_name}"

    return None, None, None, None, "Unsupported URL format"


def main():
    """Main entry point."""
    if len(sys.argv) != 3:
        print(json.dumps({
            "success": False,
            "error": "Usage: extract_microshift_version.py <prow_url> <scenario>"
        }))
        sys.exit(1)

    prow_url = sys.argv[1]
    scenario = sys.argv[2]

    # Validate inputs
    if not prow_url or not prow_url.strip():
        print(json.dumps({
            "success": False,
            "error": "prow_url cannot be empty"
        }))
        sys.exit(1)

    if not scenario or not scenario.strip():
        print(json.dumps({
            "success": False,
            "error": "scenario cannot be empty"
        }))
        sys.exit(1)

    # Parse the Prow URL
    job_id, version, job_type, pr_number, parse_error = parse_prow_url(prow_url)
    if parse_error:
        print(json.dumps({
            "success": False,
            "error": parse_error,
            "prow_url": prow_url
        }))
        sys.exit(1)

    # Construct journal log directory URL
    dir_url = construct_journal_log_dir_url(job_id, version, job_type, pr_number, scenario)

    # Find journal log file
    journal_file, error = find_journal_log_file(dir_url)
    if error:
        print(json.dumps({
            "success": False,
            "error": error,
            "url": dir_url
        }))
        sys.exit(1)

    # Construct full journal log URL
    log_url = dir_url + journal_file

    # Fetch journal log
    log_content, error = fetch_url(log_url)
    if error:
        print(json.dumps({
            "success": False,
            "error": error,
            "url": log_url
        }))
        sys.exit(1)

    # Extract version
    microshift_version, error = extract_version_from_journal(log_content)
    if error:
        print(json.dumps({
            "success": False,
            "error": error,
            "url": log_url
        }))
        sys.exit(1)

    # Determine build type
    build_type = determine_build_type(microshift_version)

    # Output result
    result = {
        "success": True,
        "version": microshift_version,
        "build_type": build_type,
        "url": log_url,
        "error": None
    }

    print(json.dumps(result, indent=2))
    sys.exit(0)


if __name__ == "__main__":
    main()
