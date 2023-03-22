import jira
import os
from tabulate import tabulate
import argparse
from tqdm import tqdm

JQL_FILTER_QUERY = 'filter = "MicroShift - Bugs in Project" and status in (New, Assigned, Post, "To Do", "In Progress", "Code Review")'
JQL_FILTER_ISSUE = 'key in ({})'
JQL_FILTER_USER = 'assignee in ({})'
JQL_ORDER_BY = 'order by key asc'
JIRA_SERVER = 'https://issues.redhat.com'
JIRA_URL_PREFIX = JIRA_SERVER+'/browse/'

def is_original_issue(issue):
    for link in issue.fields.issuelinks:
        if link.type.name != 'Cloners':
            continue
        if hasattr(link, 'outwardIssue'):
            return False
    return True

def get_clone_by_issues(issue, connection):
    l = []
    for link in issue.fields.issuelinks:
        if link.type.name != 'Cloners':
            continue
        if hasattr(link, 'inwardIssue'):
            l.append(connection.issue(link.inwardIssue.key))
    return l

def get_assignee(issue):
    if not hasattr(issue.fields, 'assignee') or issue.fields.assignee is None:
        return None
    return issue.fields.assignee.emailAddress

def get_fix_versions(issue):
    l = []
    if hasattr(issue.fields, 'fixVersions') and issue.fields.fixVersions is not None:
        for x in issue.fields.fixVersions:
            if hasattr(x, 'name'):
                l.append(x.name)
    return sorted(l)[::-1]

def get_target_versions(issue):
    l = []
    if hasattr(issue.fields, 'customfield_12319940') and issue.fields.customfield_12319940 is not None:
        for x in issue.fields.customfield_12319940:
            if hasattr(x, 'name'):
                l.append(x.name)
    return l

def get_parent_issue(issue, connection):
    for link in issue.fields.issuelinks:
        if link.type.name != 'Cloners':
            continue
        if hasattr(link, 'outwardIssue'):
            return connection.issue(link.outwardIssue.key)
    return None

def get_sprint(issue):
    if not hasattr(issue.fields, 'customfield_12310940'):
        return None
    if issue.fields.customfield_12310940 is None or len(issue.fields.customfield_12310940) == 0:
        return None
    last_sprint = issue.fields.customfield_12310940[-1]
    first_index = last_sprint.find('id=')+3
    last_index = last_sprint.find(',', first_index)
    return last_sprint[first_index:last_index]

def is_issue_a_cve(issue):
    for label in issue.fields.labels:
        if label.startswith('CVE-'):
            return True
    return False

def has_fix_versions_label(issue):
    for label in issue.fields.labels:
        if label == 'needs-fix-version':
            return True
    return False

def set_target_version(issue, version):
    issue.update(fields={
        'fixVersions': [{'name': version}]
    })

def set_needs_fix_version_label(issue):
    labels = issue.fields.labels + ['needs-fix-version']
    issue.update(fields={
        'labels': [{'name': x} for x in labels]
    })

def remove_needs_fix_version_label(issue):
    labels = []
    for l in issue.fields.labels:
        if l == 'needs-fix-version':
            continue
        labels.append({'name': l})
    issue.update(fields={
        'labels': labels
    })

def set_target_version(issue, version):
    issue.update(fields={
        'customfield_12319940': [{'name': version}]
    })

def clone_issue(issue, target, connection):
    data_dict = {}
    data_dict['priority'] = {'id': issue.fields.priority.id}
    data_dict['labels'] = issue.fields.labels + ['backport']
    data_dict['assignee'] = {'name': issue.fields.assignee.name}
    data_dict['reporter'] = {'name': issue.fields.reporter.name}
    data_dict['issuetype'] = {'name': issue.fields.issuetype.name}
    data_dict['project'] = {'id': issue.fields.project.id}
    data_dict['summary'] = issue.fields.summary
    data_dict['description'] = issue.fields.description
    data_dict['components'] = [{'id': x.name} for x in issue.fields.components]
    data_dict['versions'] = [{'name': x.name} for x in issue.fields.versions]
    data_dict['fixVersions'] = [{'name': x.name} for x in issue.fields.fixVersions]
    # QA contact
    if hasattr(issue.fields, 'customfield_12315948'):
        data_dict['customfield_12315948'] = {'value': issue.fields.customfield_12315948.name}
    data_dict['customfield_12320947'] = [{'value': x.value} for x in issue.fields.customfield_12320947]
    # Target version
    data_dict['customfield_12319940'] = [{'name': target}]

    new_issue = connection.create_issue(data_dict)
    connection.create_issue_link("Cloners", inwardIssue=new_issue.key, outwardIssue=issue.key)
    connection.create_issue_link("Blocks", inwardIssue=issue.key, outwardIssue=new_issue.key)

    sprint = get_sprint(issue)
    if sprint is not None:
        connection.add_issues_to_sprint(sprint, [new_issue])

    return new_issue

def add_blocks_link(issue, parent, connection):
    connection.create_issue_link("Blocks", inwardIssue=parent.key, outwardIssue=issue.key)

class Action:
    def __init__(self, i, c, f, **kwargs):
        def _internal_fn():
            return f(**kwargs)
        self.issue = i
        self.comment = c
        self.action = None
        if f is not None:
            self.action = _internal_fn

def scan_original_issue(issue, connection):
    actions = []
    fix_versions = get_fix_versions(issue)
    target_versions = get_target_versions(issue)
    if len(target_versions) > 1:
        actions.append(Action(issue.key, "Too many Target versions. Set to latest in Fix versions.", set_target_version, issue=issue, version=fix_versions[0]))
    elif len(target_versions) == 0:
        actions.append(Action(issue.key, "Empty Target versions. Set to latest in Fix versions.", set_target_version, issue=issue, version=fix_versions[0]))
    elif target_versions[0] != fix_versions[0]:
        actions.append(Action(issue.key, "Target versions do not match Fix versions latest. Set to latest in Fix versions.", set_target_version, issue=issue, version=fix_versions[0]))

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
    actions = []
    parent = get_parent_issue(issue, connection)
    parent_fix_versions = get_fix_versions(parent)
    fix_versions = get_fix_versions(issue)
    if parent_fix_versions != fix_versions:
        actions.append(Action(issue.key, "Fix versions update to match parent's", None))

    target_version = get_target_versions(issue)
    if len(target_version) > 1:
        actions.append(Action(issue.key, "Target versions has more than one value. Fix manually.", None))
    elif len(target_version) == 0:
        actions.append(Action(issue.key, "Target versions is empty. Fix manually.", None))
    else:
        if target_version[0] not in parent_fix_versions:
            actions.append(Action(issue.key, "Target version not in parent Fix versions. Fix manually.", None))

    for link in issue.fields.issuelinks:
        if link.type.name != 'Blocks':
            continue
        if hasattr(link, 'inwardIssue'):
            if link.inwardIssue.key == parent.key:
                return actions
    actions.append(Action(issue.key, "Add Blocks link with parent issue", add_blocks_link, issue=issue, parent=parent, connection=connection))

    return actions


def scan_issue(issue, connection):
    actions = []
    if is_issue_a_cve(issue):
        return actions

    assignee = get_assignee(issue)
    if assignee is None:
        actions.append(Action(issue.key, "Issue needs to be assigned. Fix manually", None))
        return actions

    fix_versions = get_fix_versions(issue)
    if len(fix_versions) == 0:
        actions.append(Action(issue.key, "Fix versions empty. Add label needs-fix-versions", set_needs_fix_version_label, issue=issue))
        return actions
    else:
        if has_fix_versions_label(issue):
            actions.append(Action(issue.key, "Remove Fix versions label", remove_needs_fix_version_label, issue=issue))

    if get_sprint(issue) is None:
        actions.append(Action(issue.key, "Issue needs to be included in a sprint", None))

    if is_original_issue(issue):
        actions.extend(scan_original_issue(issue, connection))
    else:
        actions.extend(scan_cloned_issue(issue, connection))
    return actions

def query_build(issue, user):
    query = JQL_FILTER_QUERY
    if issue:
        query += f' and {JQL_FILTER_ISSUE.format(issue)}'
    if user:
        users = [f'"{x}"' for x in user.split(',')]
        query += f' and {JQL_FILTER_USER.format(",".join(users))}'
    query += f' {JQL_ORDER_BY}'
    return query

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

    connection = jira.JIRA(
        server=JIRA_SERVER,
        token_auth=args.token)

    jql_query = query_build(args.issue, args.user)
    print(f"JQL Query: '{jql_query}'")

    try:
        query = connection.search_issues(jql_str=jql_query, maxResults=9999)
        print(f"Scanning {len(query)} issues")
    except Exception as e:
        print(f"Unable to retrieve issues: {e}")
        exit(1)

    actions = []
    for i in tqdm(range(len(query))):
        try:
            issue = query[i]
        except Exception as e:
            print(f"Unable to retrieve issue {query[i]}: {e}")
            continue
        actions.extend(scan_issue(issue, connection))

    print(tabulate([[x.issue, x.comment, 'Y' if x.action is None else 'N'] for x in actions], headers=['Issue', 'Action', 'Manual']))

    answer = ''
    if args.auto_accept:
        answer = 'y'
    while answer not in ['y', 'n']:
        answer = input('Perform non manual actions? [Y/N]').lower()
    if answer == 'n':
        exit(0)
    
    for i in tqdm(range(len(actions))):
        if actions[i].action is not None:
            actions[i].action()
