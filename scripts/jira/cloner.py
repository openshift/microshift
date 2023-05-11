#!/usr/bin/env python3
"""
This script provides various functions for Automatic MicroShift bug JIRA cloner.

Args:
    To use this script, provide the following arguments:

    -t, --token, JIRA Auth token. Defaults to JIRA_TOKEN env var
    -i, --issue, Target a specific issue. Can be comma separated list
    -y, --auto-accept, Do not prompt for action execution confirmation
    -u, --user, Issues owned by this user

File : cloner.py
"""

import argparse
import os
import sys

from tabulate import tabulate
from tqdm import tqdm

import jira

JQL_FILTER_QUERY = 'filter = "MicroShift - Bugs in Project" and status in (New, Assigned, Post, "To Do", "In Progress", "Code Review")'
JQL_FILTER_ISSUE = 'key in ({})'
JQL_FILTER_USER = 'assignee in ({})'
JQL_ORDER_BY = 'order by key asc'
JIRA_SERVER = 'https://issues.redhat.com'
JIRA_URL_PREFIX = JIRA_SERVER+'/browse/'


def is_original_issue(issue):
    """Returns True if the given issue is an original issue (not a clone), False otherwise."""
    for link in issue.fields.issuelinks:
        if link.type.name != 'Cloners':
            continue
        if hasattr(link, 'outwardIssue'):
            return False
    return True


def get_clone_by_issues(issue, connection):
    """Returns a list of clone issues for the given issue."""
    l = []
    for link in issue.fields.issuelinks:
        if link.type.name != 'Cloners':
            continue
        if hasattr(link, 'inwardIssue'):
            l.append(connection.issue(link.inwardIssue.key))
    return l


def get_assignee(issue):
    """Returns the email address of the assignee of the given issue."""
    if not hasattr(issue.fields, 'assignee') or issue.fields.assignee is None:
        return None
    return issue.fields.assignee.emailAddress


def get_fix_versions(issue):
    """Returns a list of fix versions for the given issue."""
    l = []
    if hasattr(issue.fields, 'fixVersions') and issue.fields.fixVersions is not None:
        for fver in issue.fields.fixVersions:
            if hasattr(fver, 'name'):
                l.append(fver.name)
    return sorted(l)[::-1]


def get_target_versions(issue):
    """Returns a list of target versions for the given issue."""
    l = []
    if hasattr(issue.fields, 'customfield_12319940') and issue.fields.customfield_12319940 is not None:
        for tver in issue.fields.customfield_12319940:
            if hasattr(tver, 'name'):
                l.append(tver.name)
    return l


def get_parent_issue(issue, connection):
    """Returns the parent issue for the given issue."""
    for link in issue.fields.issuelinks:
        if link.type.name != 'Cloners':
            continue
        if hasattr(link, 'outwardIssue'):
            return connection.issue(link.outwardIssue.key)
    return None


def get_sprint(issue):
    """Returns the sprint that the given issue is currently assigned to, or None if not assigned."""
    if not hasattr(issue.fields, 'customfield_12310940'):
        return None
    if issue.fields.customfield_12310940 is None or len(issue.fields.customfield_12310940) == 0:
        return None
    last_sprint = issue.fields.customfield_12310940[-1]
    first_index = last_sprint.find('id=')+3
    last_index = last_sprint.find(',', first_index)
    return last_sprint[first_index:last_index]


def is_issue_a_cve(issue):
    """Returns True if the given issue has a label that starts with "CVE-", False otherwise."""
    for label in issue.fields.labels:
        if label.startswith('CVE-'):
            return True
    return False


def has_fix_versions_label(issue):
    """Returns True if the given issue has a label "needs-fix-version", False otherwise."""
    for label in issue.fields.labels:
        if label == 'needs-fix-version':
            return True
    return False


def set_fix_version(issue, version):
    """Sets the fix version for the given issue to the specified version."""
    issue.update(fields={
        'fixVersions': [{'name': version}]
    })


def set_needs_fix_version_label(issue):
    """Adds the label "needs-fix-version" to the given issue's labels."""
    labels = issue.fields.labels + ['needs-fix-version']
    issue.update(fields={
        'labels': list(labels)
    })


def remove_needs_fix_version_label(issue):
    """Removes the label "needs-fix-version" from the issue's labels."""
    labels = []
    for label in issue.fields.labels:
        if label == 'needs-fix-version':
            continue
        labels.append({'name': label})
    issue.update(fields={
        'labels': labels
    })


def set_target_version(issue, version):
    """Sets the target version for the given issue to the specified version."""
    issue.update(fields={
        'customfield_12319940': [{'name': version}]
    })


def set_assignee(issue, assignee):
    """Sets the assignee of the issue to the specified user."""
    issue.update(fields={
        'assignee': {
            'name': assignee
        }
    })


def set_fix_versions(issue, fix_versions):
    """Sets the fix versions of the issue to the specified list of fix versions."""
    issue.update(fields={
        'fixVersions': [{'name': x} for x in fix_versions]
    })


def set_qa_contact(issue, contact):
    """Sets the QA contact for the issue to the specified user."""
    issue.update(fields={
        'customfield_12315948': {'value': contact}
    })


def set_sprint(issue, sprint, connection):
    """Sets the issue to the specified sprint."""
    connection.add_issues_to_sprint(sprint, [issue.key])


def clone_issue(issue, target, connection):
    """Clones the specified issue."""
    data_dict = {}
    data_dict['priority'] = {'id': issue.fields.priority.id}
    data_dict['labels'] = issue.fields.labels + ['backport']
    data_dict['issuetype'] = {'id': issue.fields.issuetype.id}
    data_dict['project'] = {'id': issue.fields.project.id}
    data_dict['summary'] = issue.fields.summary
    data_dict['description'] = issue.fields.description
    data_dict['components'] = [{'id': x.id} for x in issue.fields.components]
    data_dict['versions'] = [{'name': x.name} for x in issue.fields.versions]
    # Target version
    data_dict['customfield_12319940'] = [{'name': target}]

    new_issue = connection.create_issue(data_dict)
    set_assignee(new_issue, issue.fields.assignee.name)
    set_fix_versions(new_issue, [x.name for x in issue.fields.fixVersions])
    if hasattr(issue.fields, 'customfield_12315948'):
        set_qa_contact(new_issue, issue.fields.customfield_12315948.name)

    connection.create_issue_link("Cloners", inwardIssue=new_issue.key, outwardIssue=issue.key)
    connection.create_issue_link("Blocks", inwardIssue=issue.key, outwardIssue=new_issue.key)

    sprint = get_sprint(issue)
    if sprint is not None:
        connection.add_issues_to_sprint(sprint, [new_issue.key])


def add_blocks_link(issue, parent, connection):
    """Adds a 'Blocks' link between an issue and parent issue."""
    connection.create_issue_link("Blocks", inwardIssue=parent.key, outwardIssue=issue.key)


class Action:
    """This class defines actions that can be taken on an issue, along with associated metadata."""
    def __init__(self, issuue, comm, func, **kwargs):
        def _internal_fn():
            return func(**kwargs)
        self.issue = issuue
        self.comment = comm
        self.action = None
        if func is not None:
            self.action = _internal_fn


def scan_original_issue(issue, connection):
    """Scans the spefied original issue, and returns a list of actions to take."""
    actions = []
    fix_versions = get_fix_versions(issue)
    target_versions = get_target_versions(issue)
    if len(target_versions) > 1:
        actions.append(Action(issue.key, "Too many Target versions. Set to latest in Fix versions.", set_target_version, issue=issue, version=fix_versions[0]))
    elif len(target_versions) == 0:
        actions.append(Action(issue.key, "Empty Target versions. Set to latest in Fix versions.", set_target_version, issue=issue, version=fix_versions[0]))
    elif target_versions[0] != fix_versions[0]:
        actions.append(Action(issue.key, "Target versions do not match Fix versions latest. Set to latest in Fix versions.", set_target_version, issue=issue, version=fix_versions[0]))

    if get_sprint(issue) is None:
        actions.append(Action(issue.key, "Issue needs to be included in a sprint", None))

    fix_versions_missing = set(fix_versions[1:])
    cloning_issues = get_clone_by_issues(issue, connection)
    for clone in cloning_issues:
        clone_target_versions = get_target_versions(clone)
        if len(target_versions) > 1:
            actions.append(Action(clone.key, "Target versions has more than one field. Fix manually.", None))
        elif len(target_versions) == 0:
            actions.append(Action(clone.key, "Target versions is empty. Fix manually", None))
        else:
            if clone_target_versions[0] in fix_versions_missing:
                fix_versions_missing.remove(clone_target_versions[0])
    for version in fix_versions_missing:
        actions.append(Action(issue.key, f"Clone for Target version {version}", clone_issue, issue=issue, target=version, connection=connection))
    return actions


def scan_cloned_issue(issue, connection):
    """Scans the spefied cloned issue, and returns a list of actions to take."""
    actions = []
    parent = get_parent_issue(issue, connection)
    parent_fix_versions = get_fix_versions(parent)
    fix_versions = get_fix_versions(issue)
    if parent_fix_versions != fix_versions:
        actions.append(Action(issue.key, "Fix versions update to match parent", set_fix_versions, issue=issue, fix_versions=parent_fix_versions))

    target_version = get_target_versions(issue)
    if len(target_version) > 1:
        actions.append(Action(issue.key, "Target versions has more than one value. Fix manually.", None))
    elif len(target_version) == 0:
        actions.append(Action(issue.key, "Target versions is empty. Fix manually.", None))
    else:
        if target_version[0] not in parent_fix_versions:
            actions.append(Action(issue.key, "Target version not in parent Fix versions. Fix manually.", None))

    if get_sprint(issue) is None:
        parent_sprint = get_sprint(parent)
        if parent_sprint is not None:
            actions.append(Action(issue.key, "Sprint empty, copy from parent", set_sprint, issue=issue, sprint=parent_sprint, connection=connection))
        else:
            actions.append(Action(issue.key, "Issue needs to be included in a sprint", None))

    for link in issue.fields.issuelinks:
        if link.type.name != 'Blocks':
            continue
        if hasattr(link, 'inwardIssue'):
            if link.inwardIssue.key == parent.key:
                return actions
    actions.append(Action(issue.key, "Add Blocks link with parent issue", add_blocks_link, issue=issue, parent=parent, connection=connection))

    return actions


def scan_issue(issue, connection):
    """Scans the specified JIRA issue and returns a list of potential actions."""
    actions = []
    if is_issue_a_cve(issue):
        return actions

    assignee = get_assignee(issue)
    if assignee is None:
        actions.append(Action(issue.key, "Issue needs to be assigned. Fix manually", None))
        return actions

    fix_versions = get_fix_versions(issue)
    if len(fix_versions) == 0:
        if not has_fix_versions_label(issue):
            actions.append(Action(issue.key, "Fix versions empty. Add label needs-fix-versions", set_needs_fix_version_label, issue=issue))
        else:
            actions.append(Action(issue.key, "Fix versions empty. Label needs-fix-versions present", None))
        return actions

    if has_fix_versions_label(issue):
        actions.append(Action(issue.key, "Remove Fix versions label", remove_needs_fix_version_label, issue=issue))

    if is_original_issue(issue):
        actions.extend(scan_original_issue(issue, connection))
    else:
        actions.extend(scan_cloned_issue(issue, connection))
    return actions


def query_build(issue, user):
    """Builds a JIRA JQL query string based on the specified issue and/or user names."""
    query_str = JQL_FILTER_QUERY
    if issue:
        query_str += f' and {JQL_FILTER_ISSUE.format(issue)}'
    if user:
        users = [f'"{x}"' for x in user.split(',')]
        query_str += f' and {JQL_FILTER_USER.format(",".join(users))}'
    query_str += f' {JQL_ORDER_BY}'
    return query_str


if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        prog='cloner',
        description='Automatic MicroShift bug JIRA cloner'
    )
    parser.add_argument('-t', '--token', help='JIRA Auth token. Defaults to JIRA_TOKEN env var', default=os.environ['JIRA_TOKEN'])
    parser.add_argument('-i', '--issue', help='Target a specific issue. Can be comma separated list', default='')
    parser.add_argument('-y', '--auto-accept', help='Do not prompt for action execution confirmation', action='store_true', default=False)
    parser.add_argument('-u', '--user', help='Issues owned by this user', default='')

    args = parser.parse_args()

    conn = jira.JIRA(
        server=JIRA_SERVER,
        token_auth=args.token)

    jql_query = query_build(args.issue, args.user)
    print(f"JQL Query: '{jql_query}'")

    try:
        query = conn.search_issues(jql_str=jql_query, maxResults=9999)
        print(f"Scanning {len(query)} issues")
    except Exception as e:
        print(f"Unable to retrieve issues: {e}")
        sys.exit(1)

    actions_list = []
    for i in tqdm(range(len(query))):
        try:
            issuee = query[i]
        except Exception as e:
            print(f"Unable to retrieve issue {query[i]}: {e}")
            continue
        actions_list.extend(scan_issue(issuee, conn))

    print(tabulate([[x.issue, x.comment, 'Y' if x.action is None else 'N'] for x in actions_list], headers=['Issue', 'Action', 'Manual']))
    print()

    actions_list = list(filter(lambda x: x.action is not None, actions_list))
    if len(actions_list) == 0:
        print("No automatic actions to perform.")
        sys.exit(0)

    answer = ''
    if args.auto_accept:
        answer = 'y'
    while answer not in ['y', 'n']:
        answer = input(f'Perform {len(actions_list)} non manual actions? [Y/N]').lower()
    if answer == 'n':
        sys.exit(0)

    for i in tqdm(range(len(actions_list))):
        if actions_list[i].action is not None:
            try:
                actions_list[i].action()
            except Exception as e:
                print(f'Error executing action "{actions_list[i].comment}" on issue {actions_list[i].issue}:\n\t{e}')
