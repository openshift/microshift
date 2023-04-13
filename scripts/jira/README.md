# Setting up Developer Tools

## Jira API Token

Visit [the Profile page on the Jira
Server](https://issues.redhat.com/secure/ViewProfile.jspa?selectedTab=com.atlassian.pats.pats-plugin:jira-user-personal-access-tokens) and create a token.

Set the `JIRA_API_TOKEN` to that value by adding a line like this to
your shell login script (`~/.bashrc`, etc.).

```
export JIRA_API_TOKEN="TOKEN_VALUE"
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
