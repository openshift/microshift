# Setting up Developer Tools

## Jira API Token

Visit [the Profile page on the Jira
Server](https://issues.redhat.com/secure/ViewProfile.jspa?selectedTab=com.atlassian.pats.pats-plugin:jira-user-personal-access-tokens) and create a token.

Set the `JIRA_API_TOKEN` to that value by adding a line like this to
your shell login script (`~/.bashrc`, etc.).

```
export JIRA_API_TOKEN="TOKEN_VALUE"
```

## Github API Token

Create a token following [Github's
directions](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens)
and set the environment variable `GITHUB_TOKEN`.

```
export GITHUB_TOKEN="ghp_THETOKENVALUE"
```

## Start Ticket

Run `./scripts/jira/start_ticket.sh` to mark a ticket as started (In
Progress or Code Review).

The default ticket ID discovered by parsing the current branch name,
looking for the pattern

  <project-id>-<ticket-num>[-<description>]

For example

  USHIFT-1069-jira-manage-ticket

produces a ticket ID of

  USHIFT-1069

Several updates are made to the ticket:

1. The owner is assigned to the current jira user (based on the API
   token).
2. The ticket is added to the current active sprint.
3. The ticket is transitioned to either "In Progress" or "Code
   Review", depending on the command line options.

## Close Tickets

Run `./scripts/jira/manage_ticket.sh close` to find all open tickets with
closed PRs. When all of the PRs linked to a ticket are closed, it is
transitioned to the next state. Tickets with the `no-qe-needed` label
are moved to `Done`. Tickets without the label are moved to `Review`.

```
$ ./scripts/jira/manage_ticket.sh close
finding open tickets assigned to dhellman@redhat.com in uShift Sprint 238

USHIFT-1409: provide dev tool for logging into vm used for scenario test
  URL: https://issues.redhat.com/browse/USHIFT-1409
  Status: Code Review
  Labels: ['no-qe-needed']
  Link: https://github.com/openshift/microshift/pull/2009 (True)
  Transition: Closed

USHIFT-1402: error trap causes logic errors
  URL: https://issues.redhat.com/browse/USHIFT-1402
  Status: Code Review
  Labels: ['no-qe-needed']
  Link: https://github.com/openshift/microshift/pull/2005 (True)
  Transition: Closed

USHIFT-1401: rebase should only clean up files it creates
  URL: https://issues.redhat.com/browse/USHIFT-1401
  Status: Code Review
  Labels: ['no-qe-needed']
  Link: https://github.com/openshift/microshift/pull/2004 (True)
  Transition: Closed

USHIFT-1388: developer tool to make rebuilding scenario images easier
  URL: https://issues.redhat.com/browse/USHIFT-1388
  Status: Code Review
  Labels: ['no-qe-needed']
  Link: https://github.com/openshift/microshift/pull/1990 (True)
  Transition: Closed

USHIFT-1385: scenario test harness should launch VMs in parallel
  URL: https://issues.redhat.com/browse/USHIFT-1385
  Status: Code Review
  Link: https://github.com/openshift/microshift/pull/1987 (False)
  Transition: none

USHIFT-1098: CI: Periodically, run an upgrade e2e job starting from the most recent release on the previous y-stream and upgrading to HEAD of the branch
  URL: https://issues.redhat.com/browse/USHIFT-1098
  Status: In Progress
  Labels: ['ushift-updatability-ci']
  Link: https://github.com/openshift/microshift/pull/1989 (False)
  Transition: none
```
