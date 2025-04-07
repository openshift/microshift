#!/usr/bin/env python3

import os
import sys
import jira.client
import requests
import urllib3
import json
import jira
import yaml

SERVER_URL = 'https://issues.redhat.com/'
JIRA_API_TOKEN = os.environ.get('JIRA_API_TOKEN')


def usage():
    print("""\
        usage: advisory_publication_report.py OCP_VERSION

        arguments:
            OCP_VERSION: The OCP versions to analyse if MicroShift version should be published. Format: "4.X.Z"\
    """)


def get_advisories(ocp_version: str) -> dict[str, int]:
    """
    Get a list of advisory ids for a OCP version from github.com/openshift-eng/ocp-build-data repository
        Parameters:
            ocp_version (str): OCP version with format: "X.Y.Z"
        Returns:
            (dict): advisory dict with type and id
    """
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

    try:
        microshift_xy_version = '.'.join(ocp_version.split('.')[:2])
        request = requests.get(f'https://raw.githubusercontent.com/openshift-eng/ocp-build-data/refs/heads/openshift-{microshift_xy_version}/releases.yml', verify=False)
        request.raise_for_status()
    except requests.exceptions.HTTPError as err:
        raise SystemExit(err)
    releases_dict = yaml.load(str(request.text), Loader=yaml.SafeLoader)

    if ocp_version in releases_dict['releases']:
        return releases_dict['releases'][ocp_version]['assembly']['group']['advisories']
    else:
        raise KeyError(f"{ocp_version} OCP version does NOT exist")


def get_advisory_info(advisory_id: int) -> dict[str, str]:
    """
    Get a list of strings with the CVEs ids for an advisory
        Parameters:
            advisory_id (int): advisory id
        Returns:
            (list): list of strings with CVE ids
    """
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

    try:
        request = requests.get(f'https://errata.devel.redhat.com/cve/show/{advisory_id}.json', verify=False)
        request.raise_for_status()
    except requests.exceptions.HTTPError as err:
        raise SystemExit(err)
    advisory_info = json.loads(request.text)

    if advisory_info is None:
        raise ValueError
    if not isinstance(advisory_info, dict):
        raise TypeError
    return advisory_info


def search_microshift_tickets(affects_version: str, cve_id: str) -> jira.client.ResultList:
    """
    Query Jira for MicroShift ticket with CVE id and MicroShift version
        Parameters:
            affects_version (str): MicroShift affected version with format: "X.Y"
            cve_id (str): the CVE id with format: "CVE-YYYY-NNNNN"
        Returns:
            (jira.client.ResultList): a list with all the Jira tickets matching the query
    """
    server = jira.JIRA(server=SERVER_URL, token_auth=JIRA_API_TOKEN)
    jira_tickets = server.search_issues(f'''
        summary  ~ "{cve_id}" and component = MicroShift and (affectedVersion = {affects_version} or affectedVersion = {affects_version}.z)
    ''')

    if not isinstance(jira_tickets, jira.client.ResultList):
        raise TypeError
    return jira_tickets


def get_report(ocp_version: str) -> dict[str, dict]:
    """
    Get a json object with all the advisories, CVEs and jira tickets linked
        Parameters:
            ocp_version (str): OCP version with format: "X.Y.Z"
        Returns:
            (dict): json object with all the advisories, CVEs and jira tickets linked
    """
    result_json = dict()
    advisories = get_advisories(ocp_version)
    for advisory_type, advisory_id in advisories.items():
        advisory_info = get_advisory_info(advisory_id)
        cve_list = advisory_info['cve']
        advisory_dict = dict()
        advisory_dict['type'] = advisory_type
        advisory_dict['url'] = f'https://errata.devel.redhat.com/advisory/{advisory_id}'
        advisory_dict['cves'] = dict()
        for cve in cve_list:
            jira_tickets = search_microshift_tickets(".".join(ocp_version.split(".")[:2]), cve)
            advisory_dict['cves'][cve] = dict()
            for ticket in jira_tickets:
                jira_ticket_dict = dict()
                jira_ticket_dict['id'] = ticket.key
                jira_ticket_dict['summary'] = ticket.fields.summary
                jira_ticket_dict['status'] = ticket.fields.status.name
                jira_ticket_dict['resolution'] = str(ticket.fields.resolution)
                advisory_dict['cves'][cve]['jira_ticket'] = jira_ticket_dict
        result_json[advisory_info['advisory']] = advisory_dict
    return result_json


def main():
    if len(sys.argv) != 2:
        usage()
        raise ValueError('Invalid number of arguments')

    if JIRA_API_TOKEN is None:
        raise ValueError('JIRA_API_TOKEN var not found in the env')

    ocp_version = str(sys.argv[1])
    result_json = get_report(ocp_version)
    print(f"{json.dumps(result_json, indent=4)}")


if __name__ == '__main__':
    main()
