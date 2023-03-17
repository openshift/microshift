import jira
import os
from tabulate import tabulate
import argparse
from tqdm import tqdm

JQL_GLOBAL_QUERY = 'filter = "MicroShift - Bugs in Project" and resolution = Unresolved order by key asc'
JQL_ISSUE_QUERY = 'filter = "MicroShift - Bugs in Project" and resolution = Unresolved and key in ({}) order by key asc'
JIRA_SERVER = 'https://issues.redhat.com'
JIRA_URL_PREFIX = JIRA_SERVER+'/browse/'

def is_original_issue(issue):
    for link in issue.fields.issuelinks:
        if link.type.name != 'Cloners':
            continue
        if hasattr(link, 'outwardIssue'):
            return False
    return True

def get_issue_fix_versions(issue):
    l = []
    if hasattr(issue.fields, 'fixVersions') and issue.fields.fixVersions is not None:
        for x in issue.fields.fixVersions:
            if hasattr(x, 'name'):
                l.append(x.name)
    if len(l) == 0:
        raise RuntimeError(f"[{issue.key}] Fix versions empty")
    return sorted(l)[::-1]

def get_issue_target_versions(issue):
    l = []
    if hasattr(issue.fields, 'customfield_12319940') and issue.fields.customfield_12319940 is not None:
        for x in issue.fields.customfield_12319940:
            if hasattr(x, 'name'):
                l.append(x.name)
    if len(l) == 0:
        raise RuntimeError("Target version empty")
    if len(l) > 1:
        raise RuntimeError("Target version must have only 1 value")
    return l[0]

def get_clone_by_issues(issue, connection):
    l = []
    for link in issue.fields.issuelinks:
        if link.type.name != 'Cloners':
            continue
        if hasattr(link, 'inwardIssue'):
            l.append(connection.issue(link.inwardIssue.key))
    return l

def get_clones_issue(issue, connection):
    for link in issue.fields.issuelinks:
        if link.type.name != 'Cloners':
            continue
        if hasattr(link, 'outwardIssue'):
            return connection.issue(link.outwardIssue.key)
    return None

def set_fix_versions(issue, fixVersions, connection):
    issue.update(fields={
        'fixVersions': [{'name': x} for x in fixVersions]
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
        data_dict['customfield_12315948'] = {'value': issue.fields.customfield_12315948.value}
    data_dict['customfield_12320947'] = [{'value': x.value} for x in issue.fields.customfield_12320947]
    # Target version
    data_dict['customfield_12319940'] = [{'name': target}]

    new_issue = connection.create_issue(data_dict)
    connection.create_issue_link("Cloners", inwardIssue=new_issue.key, outwardIssue=issue.key)
    connection.create_issue_link("Blocks", inwardIssue=issue.key, outwardIssue=new_issue.key)
    return new_issue

def check_clones(issue, connection, dry_run):
    actions = []
    try:
        fix_versions = get_issue_fix_versions(issue)
    except RuntimeError as e:
        actions.append([JIRA_URL_PREFIX+issue.key, "", "Fix versions field empty. Please fill it."])
        return actions

    try:
        target_version = get_issue_target_versions(issue)
    except RuntimeError as e:
        actions.append([JIRA_URL_PREFIX+issue.key, "", f"Target field problems: {e}. Please check."])
        return actions

    if not is_original_issue(issue):
        return actions

    if target_version != fix_versions[0]:
        actions.append([JIRA_URL_PREFIX+issue.key, "", f"Target version does not target first Fix version. {target_version} != {fix_versions[0]}. Please check."])
        return actions

    fix_versions_missing = set(fix_versions[1:])
    clones = get_clone_by_issues(issue, connection)
    for clone in clones:
        clone_fix_versions = []
        try:
            clone_fix_versions = get_issue_fix_versions(clone)
        except RuntimeError as e:
            if dry_run:
                actions.append([JIRA_URL_PREFIX+clone.key, JIRA_URL_PREFIX+issue.key, f"Dry run. Skip set Fix versions."])
            else:
                try:
                    set_fix_versions(clone, fix_versions, connection)
                except RuntimeError as e:
                    actions.append([JIRA_URL_PREFIX+clone.key, JIRA_URL_PREFIX+issue.key, f"Unable to update Fix versions: {e}"])
                    continue
                finally:
                    actions.append([JIRA_URL_PREFIX+clone.key, JIRA_URL_PREFIX+issue.key, "Fix versions updated to match original"])

        if clone_fix_versions != fix_versions:
            if dry_run:
                actions.append([JIRA_URL_PREFIX+clone.key, JIRA_URL_PREFIX+issue.key, "Dry run. Skip set Fix versions of clone to original"])
            else:
                try:
                    set_fix_versions(clone, fix_versions, connection)
                except RuntimeError as e:
                    actions.append([JIRA_URL_PREFIX+clone.key, JIRA_URL_PREFIX+issue.key, f"Clone Fix versions does not match original issue. Unable to update Fix versions: {e}"])
                    continue
                finally:
                    actions.append([JIRA_URL_PREFIX+clone.key, JIRA_URL_PREFIX+issue.key, "Fix versions updated to match original"])

        try:
            target_version = get_issue_target_versions(clone)
        except RuntimeError as e:
            actions.append([JIRA_URL_PREFIX+clone.key, JIRA_URL_PREFIX+issue.key, f"Target version problems: {e}"])
            continue

        if target_version not in fix_versions:
            actions.append([JIRA_URL_PREFIX+clone.key, JIRA_URL_PREFIX+issue.key, f"Target version {target_version} not included in parent Fix versions"])
            continue

        if target_version in fix_versions_missing:
            fix_versions_missing.remove(target_version)

    for version in fix_versions_missing:
        if dry_run:
            actions.append([JIRA_URL_PREFIX+issue.key, "", f"Dry run. Skip cloning bug for version {version}"])
            continue
        try:
            new_issue = clone_issue(issue, version, connection)
        except RuntimeError as e:
            actions.append([JIRA_URL_PREFIX+issue.key, "", f"Unable to clone for version {version}: {e}"])
        finally:
            actions.append([JIRA_URL_PREFIX+new_issue.key, JIRA_URL_PREFIX+issue.key, f"New clone created targetting {version}"])

    return actions

def check_blocks(issue, connection, dry_run):
    actions = []
    original = get_clones_issue(issue, connection)
    if original is None:
        return actions

    for link in original.fields.issuelinks:
        if link.type.name != 'Blocks':
            continue
        if hasattr(link, 'outwardIssue'):
            if link.outwardIssue.key == issue.key:
                return actions
    if dry_run:
        actions.append([JIRA_URL_PREFIX+issue.key, JIRA_URL_PREFIX+original.key, "Dry run. Skip adding Blocks link"])
    else:
        connection.create_issue_link("Blocks", inwardIssue=original.key, outwardIssue=issue.key)
        actions.append([JIRA_URL_PREFIX+issue.key, JIRA_URL_PREFIX+original.key, "Blocks link added"])
    return actions

if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        prog='cloner',
        description='Automatic MicroShift bug JIRA cloner'
    )
    parser.add_argument('-t', '--token', help='JIRA Auth token. Defaults to JIRA_TOKEN env var', default=os.environ['JIRA_TOKEN'])
    parser.add_argument('-i', '--issue', help='Target a specific issue. Can be comma separated list', default="")
    parser.add_argument('-d', '--dry-run', help='Run everything but do not perform server updates', default=False, action='store_true')

    args = parser.parse_args()

    connection = jira.JIRA(
        server=JIRA_SERVER,
        token_auth=args.token)
    
    jql_query = JQL_GLOBAL_QUERY
    if len(args.issue) > 0:
        jql_query = JQL_ISSUE_QUERY.format(args.issue)

    if args.dry_run:
        print("Dry run activated. Will not perform any server changes")

    query = connection.search_issues(jql_str=jql_query, json_result=True, fields="key", maxResults=9999)
    print(f"Scanning {len(query['issues'])} issues")

    actions = []
    for i in tqdm(range(len(query['issues']))):
        issue = connection.issue(query['issues'][i]['key'])
        actions.extend(check_clones(issue, connection, args.dry_run))
        actions.extend(check_blocks(issue, connection, args.dry_run))
    
    print(tabulate(actions, headers=["Issue", "Clones", "Action"]))
