#!/usr/bin/env python3

import json
import os
import sys
from urllib.parse import quote

import requests
import urllib3
import yaml

import jira
import jira.client

JIRA_URL = 'https://issues.redhat.com/'
JIRA_API_TOKEN = os.environ.get('JIRA_API_TOKEN')
GITLAB_API_TOKEN = os.environ.get('GITLAB_API_TOKEN')
GITLAB_BASE_URL = 'https://gitlab.cee.redhat.com'
GITLAB_PROJECT_ID = 'hybrid-platforms/art/ocp-shipment-data'


def usage():
    """Print usage information."""
    print("""\
        usage: advisory_publication_report.py OCP_VERSION

        arguments:
            OCP_VERSION: The OCP versions to analyse if MicroShift version should be published. Format: "4.X.Z"

        environment variables:
            JIRA_API_TOKEN: API token for Jira access
            GITLAB_API_TOKEN: API token for GitLab access\
    """)


def get_shipment_merge_request_url(ocp_version: str) -> str:
    """
    Get merge request URL from GitHub releases.yml file.

    Parameters:
        ocp_version (str): OCP version with format: "X.Y.Z"

    Returns:
        str: GitLab merge request URL
    """
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

    try:
        microshift_xy_version = '.'.join(ocp_version.split('.')[:2])
        releases_url = (
            f'https://raw.githubusercontent.com/openshift-eng/ocp-build-data/'
            f'refs/heads/openshift-{microshift_xy_version}/releases.yml'
        )

        response = requests.get(releases_url, verify=False)
        response.raise_for_status()

        releases_dict = yaml.load(response.text, Loader=yaml.SafeLoader)
        return releases_dict['releases'][ocp_version]['assembly']['group']['shipment']['url']
    except requests.exceptions.HTTPError as err:
        raise SystemExit(f"Failed to fetch releases.yml: {err}")


def get_yaml_files_from_mr(mr_info: dict, headers: dict) -> dict:
    """
    Get YAML files from a merge request.

    Parameters:
        mr_info (dict): merge request information
        headers (dict): GitLab API headers

    Returns:
        dict: dictionary containing parsed YAML content
    """
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

    encoded_project_id = quote(GITLAB_PROJECT_ID, safe='')
    mr_iid = mr_info['iid']

    try:
        # Get changes in the merge request
        url = f'{GITLAB_BASE_URL}/api/v4/projects/{encoded_project_id}/merge_requests/{mr_iid}/changes'
        response = requests.get(url, headers=headers, verify=False)
        response.raise_for_status()

        changes = response.json()
        yaml_content = {}

        # Process each changed file
        for change in changes.get('changes', []):
            file_path = change.get('new_path', change.get('old_path', 'unknown'))

            # Only process YAML files
            if file_path.endswith(('.yml', '.yaml')):
                file_url = f'{GITLAB_BASE_URL}/api/v4/projects/{encoded_project_id}/repository/files/{quote(file_path, safe="")}/raw'

                # Try target branch first, then source branch
                file_response = requests.get(file_url, headers=headers, params={'ref': mr_info['target_branch']}, verify=False)

                if file_response.status_code != 200:
                    file_response = requests.get(file_url, headers=headers, params={'ref': mr_info['source_branch']}, verify=False)

                if file_response.status_code == 200:
                    try:
                        yaml_content[file_path] = yaml.load(file_response.text, Loader=yaml.SafeLoader)
                    except yaml.YAMLError as e:
                        print(f"Warning: Could not parse YAML file {file_path}: {e}")

        return yaml_content

    except requests.exceptions.HTTPError as err:
        raise SystemExit(f"GitLab API error while fetching MR changes: {err}")


def extract_cves_recursively(data):
    """Recursively search for CVE patterns in the YAML data"""
    cves_found = []

    if isinstance(data, dict):
        for key, value in data.items():
            if isinstance(key, str) and key.startswith('CVE-'):
                cves_found.append(key)
            if isinstance(value, str) and value.startswith('CVE-'):
                cves_found.append(value)
            cves_found.extend(extract_cves_recursively(value))
    elif isinstance(data, list):
        for item in data:
            if isinstance(item, str) and item.startswith('CVE-'):
                cves_found.append(item)
            if isinstance(item, dict) and 'key' in item and isinstance(item['key'], str) and item['key'].startswith('CVE-'):
                cves_found.append(item['key'])
            cves_found.extend(extract_cves_recursively(item))
    return cves_found


def get_advisories(ocp_version: str) -> dict[str, str]:
    """
    Get a list of advisory URLs for a OCP version from GitLab merge request YAML files.

    Parameters:
        ocp_version (str): OCP version with format: "X.Y.Z"

    Returns:
        dict: advisory dict with type and URL
    """
    # Get MR URL from GitHub releases.yml
    mr_url = get_shipment_merge_request_url(ocp_version)

    # Convert web URL to API URL
    mr_iid = mr_url.split('/')[-1]
    encoded_project_id = quote(GITLAB_PROJECT_ID, safe='')
    api_url = f'{GITLAB_BASE_URL}/api/v4/projects/{encoded_project_id}/merge_requests/{mr_iid}'

    headers = {'PRIVATE-TOKEN': GITLAB_API_TOKEN}

    try:
        response = requests.get(api_url, headers=headers, verify=False)
        response.raise_for_status()
        mr_info = response.json()
    except requests.exceptions.HTTPError as err:
        raise SystemExit(f"GitLab API error: {err}")

    # Get YAML files from the merge request
    yaml_files = get_yaml_files_from_mr(mr_info, headers)

    # Search through all YAML files to find the advisory information
    advisories_found = {}

    for file_path, yaml_content in yaml_files.items():
        # Skip the fbc file as requested
        if 'fbc-openshift' in file_path or not yaml_content:
            continue

        # Extract advisory URL using dict.get() for safer navigation
        public_url = (yaml_content.get('shipment', {})
                      .get('environments', {})
                      .get('stage', {})
                      .get('advisory', {})
                      .get('url', ''))

        if public_url:
            # Determine advisory type from filename and extract CVEs from YAML content
            for advisory_type in ['image', 'extras', 'metadata', 'rpm']:
                if advisory_type in file_path:
                    # Extract advisory name from public URL
                    advisory_name = public_url.split('/')[-1] if '/' in public_url else public_url
                    # Extract CVEs from the entire YAML content
                    cves = extract_cves_recursively(yaml_content)
                    advisories_found[advisory_type] = {
                        'name': advisory_name,
                        'cves': list(set(cves))  # Remove duplicates
                    }
                    break

    if not advisories_found:
        raise KeyError(f"{ocp_version} OCP version advisory data not found in any YAML files from the merge request")

    return advisories_found


def search_microshift_tickets(affects_version: str, cve_id: str) -> jira.client.ResultList:
    """
    Query Jira for MicroShift ticket with CVE id and MicroShift version.

    Parameters:
        affects_version (str): MicroShift affected version with format: "X.Y"
        cve_id (str): the CVE id with format: "CVE-YYYY-NNNNN"

    Returns:
        jira.client.ResultList: a list with all the Jira tickets matching the query
    """
    server = jira.JIRA(server=JIRA_URL, token_auth=JIRA_API_TOKEN)
    jira_tickets = server.search_issues(f'''
        summary  ~ "{cve_id}" and component = MicroShift and (affectedVersion = {affects_version} or affectedVersion = {affects_version}.z)
    ''')

    if not isinstance(jira_tickets, jira.client.ResultList):
        raise TypeError
    return jira_tickets


def get_report(ocp_version: str) -> dict[str, dict]:
    """
    Get a json object with all the advisories, CVEs and jira tickets linked.

    Parameters:
        ocp_version (str): OCP version with format: "X.Y.Z"

    Returns:
        dict: json object with all the advisories, CVEs and jira tickets linked
    """
    result_json = {}
    advisories = get_advisories(ocp_version)
    for advisory_type, advisory_data in advisories.items():
        advisory_name = advisory_data['name']
        cve_list = advisory_data['cves']
        advisory_dict = {
            'type': advisory_type,
            'cves': {}
        }

        for cve in cve_list:
            jira_tickets = search_microshift_tickets(".".join(ocp_version.split(".")[:2]), cve)
            advisory_dict['cves'][cve] = {}
            if jira_tickets:
                for ticket in jira_tickets:
                    jira_ticket_dict = {
                        'id': ticket.key,
                        'summary': ticket.fields.summary,
                        'status': ticket.fields.status.name,
                        'resolution': str(ticket.fields.resolution)
                    }
                    advisory_dict['cves'][cve]['jira_ticket'] = jira_ticket_dict
        result_json[advisory_name] = advisory_dict
    return result_json


def main():
    """Main function to run the advisory publication report."""
    if len(sys.argv) != 2:
        usage()
        raise ValueError('Invalid number of arguments')

    if JIRA_API_TOKEN is None or GITLAB_API_TOKEN is None:
        missing_tokens = []
        if JIRA_API_TOKEN is None:
            missing_tokens.append('JIRA_API_TOKEN')
        if GITLAB_API_TOKEN is None:
            missing_tokens.append('GITLAB_API_TOKEN')
        raise ValueError(f"Missing required environment variables: {', '.join(missing_tokens)}")

    ocp_version = str(sys.argv[1])
    result_json = get_report(ocp_version)
    print(json.dumps(result_json, indent=4))


if __name__ == '__main__':
    main()
