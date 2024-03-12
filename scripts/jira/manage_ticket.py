#!/usr/bin/env python3

"""Tool for managing Jira ticket status from the command line.

Start
=====

Bring a ticket into the current sprint and update its status.

Ticket ID
---------

The default ticket ID discovered by parsing the current branch name,
looking for the pattern

  <project-id>-<ticket-num>[-<description>]

For example

  USHIFT-1069-jira-manage-ticket

produces a ticket ID of

  USHIFT-1069

Set a ticket ID explicitly using the `--ticket-id` option.

Sprint
------

The ticket is added to the current active sprint by looking at the
sprints visible through the MicroShift Scrum Board
(https://issues.redhat.com/secure/RapidBoard.jspa?rapidView=14885) and
finding the active sprint beginning with "uShift". Disable this
behavior using the `--no-sprint` flag.

Target Version
--------------

The `--target-version` flag can be used to set a target version if one
is not set already.

Story Points
------------

The `--story-points` flag can be used to set the number of points the
work represents, if it was not pre-planned.

Status
------

Use the `--in-progress` or `--review` options to control the
status. The default is `--review`.

Close
=====

Find open tickets with closed PRs and update their status.

Authentication
==============

The script requires a Jira token passed via the `JIRA_API_TOKEN`
environment variable. See README.md for details of creating the token.

"""

import argparse
import os
import subprocess
import urllib.parse

import github

import jira

SERVER_URL = 'https://issues.redhat.com/'
SCRUM_BOARD = 14885
NO_QE_LABEL = 'no-qe-needed'


def custom_field_manager(server):
    """Return callables for working with custom fields.

    Custom fields are stored on the ticket in ticket.fields with names
    like "customfield_12319940". The information about those fields
    can be queried from the Jira API to map the user-visible names
    (like "Target Version") to the less helpful custom field attribute
    name. This function looks up the mappings and then creates two
    closures that can get and set values from a ticket using the
    understandable names.

    The getter takes a ticket and custom field name and returns the
    value of the field for that ticket.

    The setter takes a ticket, custom field name, and new value and
    updates the ticket to set that field to the new value.

    """
    field_info = server.fields()
    fields_by_name = {
        f['name']: f
        for f in field_info
    }

    def get_field_value(ticket, name):
        field_details = fields_by_name[name]
        return getattr(ticket.fields, field_details['id'])

    def set_field_value(ticket, name, value):
        field_details = fields_by_name[name]
        return ticket.update(fields={
            field_details['id']: value,
        })

    return get_field_value, set_field_value


def get_active_sprint(server, project_id):
    """Return the active sprint for the USHIFT project."""
    valid_sprints = server.sprints(SCRUM_BOARD, state='active')
    for s in valid_sprints:
        if not s.name.lower().startswith(project_id.lower()):
            continue
        return s
    return None


def get_project_id_from_ticket_id(ticket_id):
    """Parse a ticket ID and return the project portion.

    "USHIFT-662" -> "USHIFT"

    """
    return ticket_id.partition('-')[0]


def guess_ticket_id():
    """Try to determine the ticket ID from the git branch."""
    # git branch --show-current
    completed = subprocess.run(
        ['git', 'branch', '--show-current'],
        stdout=subprocess.PIPE,
        check=False,  # no exception when we cannot find the branch
    )
    if completed.returncode != 0:
        return None
    branch_name = completed.stdout.decode('UTF-8').strip()
    parts = branch_name.split('-')
    if len(parts) < 2:
        print(f'Unable to determine ticket ID from "{branch_name}"')
        return None
    return parts[0] + '-' + parts[1]


def command_start(args):
    """Implement 'start' command."""
    actual_project_id = get_project_id_from_ticket_id(args.ticket_id)
    sprint_project_id = actual_project_id
    if sprint_project_id == 'OCPBUGS':
        sprint_project_id = 'USHIFT'

    server = jira.JIRA(
        server=SERVER_URL,
        token_auth=os.environ.get('JIRA_API_TOKEN'),
    )
    getter, setter = custom_field_manager(server)

    print(f'finding ticket {args.ticket_id}')
    ticket = server.issue(args.ticket_id)
    print(f'found: "{ticket.fields.summary}"')

    jira_id = server.myself()['key']
    print(f'...updating assignment to "{jira_id}"')
    server.assign_issue(ticket, jira_id)

    if args.target_version:
        # Validate the version
        for v in server.project(actual_project_id).versions:
            if args.target_version == v.name:
                break
        else:
            raise ValueError('Unknown version')
        print(f'...setting the target version to "{args.target_version}"')
        setter(ticket, 'Target Version', [{'name': args.target_version}])

    if args.story_points:
        print(f'...setting the story points to "{args.story_points}"')
        setter(ticket, 'Story Points', args.story_points)
    else:
        points = getter(ticket, 'Story Points')
        if not points:
            print('...WARNING: story points unset')
        else:
            print(f'...story points set to "{points}"')

    if args.no_qe:
        labels = ticket.fields.labels
        if 'no-qe-needed' not in labels:
            labels.append('no-qe-needed')
            ticket.update(fields={'labels': labels})
            print('...added no-qe-needed label')
        else:
            print('...already have no-qe-needed')

    if args.sprint:
        active_sprint = get_active_sprint(server, sprint_project_id)
        if not active_sprint:
            raise ValueError('No active sprint found')
        print(f'...setting the sprint to "{active_sprint}"')
        server.add_issues_to_sprint(active_sprint.id, [ticket.key])

    if actual_project_id == 'OCPBUGS':
        print('...ticket status is managed automatically')
    else:
        print(f'...setting ticket status to "{args.status}"')
        server.transition_issue(
            issue=ticket,
            transition=args.status,
        )


def is_pr_link(url):
    """Returns boolean indicating whether the link points to a pull request."""
    if not url.startswith('https://github.com/'):
        return False
    if '/pull/' not in url:
        return False
    return True


def parse_pr_link(url):
    """Return triple containing org, repo, and PR number."""
    parsed = urllib.parse.urlparse(url)
    path_parts = parsed.path.lstrip('/').split('/')
    org = path_parts[0]
    repo = path_parts[1]
    prnum = path_parts[3]
    return (org, repo, prnum)


def command_close(args):
    server = jira.JIRA(
        server=SERVER_URL,
        token_auth=os.environ.get('JIRA_API_TOKEN'),
    )
    getter, setter = custom_field_manager(server)
    active_sprint = get_active_sprint(server, 'USHIFT')
    jira_id = server.myself()['name']
    gh_auth = github.Auth.Token(os.environ['GITHUB_TOKEN'])
    gh = github.Github(auth=gh_auth)

    print(f'finding open tickets assigned to {jira_id} in {active_sprint}')

    query = f'Sprint = {active_sprint.id} and status in ("Code Review", "In Progress") and assignee = "{jira_id}"'
    results = server.search_issues(query)
    for ticket in results:
        print()
        print(f'{ticket}: {ticket.fields.summary}')
        print(f'  URL: {SERVER_URL}browse/{ticket}')
        print('  Status:', ticket.fields.status)
        if ticket.fields.labels:
            print('  Labels:', ticket.fields.labels)
        points = getter(ticket, 'Story Points')
        print(f'  Story Points: {points}')

        num_merged = 0
        num_closed = 0
        links = server.remote_links(ticket.id)
        print(f'  PRs: {len(links)}')
        for link in links:
            url = link.object.url
            if not is_pr_link(url):
                continue
            org_name, repo_name, pr_num = parse_pr_link(url)
            repo = gh.get_repo(f'{org_name}/{repo_name}')
            pr = repo.get_pull(int(pr_num))
            state = pr.state
            if pr.merged:
                state = 'merged'
            print(f'  Link: {url} ({state})')
            if pr.merged:
                num_merged += 1
            elif pr.closed_at:
                num_closed += 1
        # We can close the ticket if we have at least 1 merged PR and
        # no open PRs.
        is_closable = num_merged and ((num_merged + num_closed) == len(links))

        actual_project_id = get_project_id_from_ticket_id(ticket.key)
        if actual_project_id == 'OCPBUGS' or not is_closable:
            print('  Transition: none')
            continue

        if NO_QE_LABEL in ticket.fields.labels:
            next_state = 'Closed'
        else:
            next_state = 'Review'
        print(f'  Transition: {next_state}')
        if not points:
            print('  SKIPPING: story points are not set')
            continue
        if args.dry_run:
            print('  DRY RUN')
        else:
            server.transition_issue(
                issue=ticket,
                transition=next_state,
            )


def main():
    """The main program."""
    parser = argparse.ArgumentParser(
        description=__doc__,
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )

    subparsers = parser.add_subparsers(dest='command', help='commands')

    start_parser = subparsers.add_parser(
        'start', help='Mark a ticket as in progress or under review',
    )
    start_parser.add_argument(
        '--ticket-id',
        default=guess_ticket_id(),
        help='the ticket id, defaults to the prefix of the branch name (%(default)s)',
    )
    start_parser.add_argument(
        '--target-version',
        default=os.environ.get('DEFAULT_TARGET_VERSION'),
        help='the target version',
    )
    start_parser.add_argument(
        '--story-points',
        help='the story points',
        default=None,
        type=int,
    )
    start_parser.add_argument(
        '--no-sprint',
        dest='sprint',
        default=True,
        action='store_false',
        help='set the sprint to the active sprint',
    )
    start_parser.add_argument(
        '--no-qe',
        dest='no_qe',
        default=False,
        action='store_true',
        help='add the no-qe-needed label',
    )
    start_parser.add_argument(
        '--review',
        dest='status',
        default='Code Review',
        action='store_const',
        const='Code Review',
        help='mark the ticket as ready for code review',
    )
    start_parser.add_argument(
        '--in-progress',
        dest='status',
        action='store_const',
        const='In Progress',
        help='mark the ticket as having been started',
    )

    close_parser = subparsers.add_parser(
        'close', help='Find tickets with closed PRs and move the ticket to the next state',
    )
    close_parser.add_argument(
        '--dry-run',
        dest='dry_run',
        default=False,
        action='store_true',
        help='report but make no changes',
    )

    args = parser.parse_args()

    if args.command == 'start':
        command_start(args)
    elif args.command == 'close':
        command_close(args)
    else:
        parser.print_help()


if __name__ == '__main__':
    main()
