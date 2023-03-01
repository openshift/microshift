# Developer Review Dashboard

This directory contains tools for building a dashboard for GitHub pull
requests relevant to developers of MicroShift. It uses
[dinghy](https://github.com/nedbat/dinghy) to scan open pull requests
for activity and build a daily summary.

## GitHub API Token

Dinghy requires a GitHub [personal access
token](https://github.com/settings/tokens) associated with your
account. It only needs read permissions based on your user and the
`openshift` organization.

## Setup

The `setup.sh` script will create a directory `~/MicroShiftReviews`
containing everything needed to run dinghy and save the output. To put
the output in a different location, pass the name of a directory as
the first argument to the script.

`setup.sh` looks for your GitHub token using the `GITHUB_TOKEN`
environment variable and saves it to the script it generates to make
it easy to configure a cron job. If you do not want the token written
to disk, you can set the variable to a dummy value then modify the
generated script to extract the value from a keychain or other
source. Those changes are left as an exercise to the reader.

## Building the Dashboard

Run `$HOME/MicroShiftReviews/update.sh` to generate the dashboard as a
static HTML file written to `$HOME/MicroShiftReviews/microshift.html`.

## Cron Configuration

The default configuration file produced by `setup.sh` includes query
parameters to limit the data to changes within the last day. Setting
up a daily cron job ensures the dashboard is fresh each day.

```
@daily $HOME/MicroShiftReviews/update.sh
```

## The Dashboard Content

The generated dashboard includes several sections based on different
queries.

### Enhancements

This section includes all PRs in `openshift/enhancements` that mention
`microshift`.

### Code Changes

This section includes all PRs for all branches in
`openshift/microshift`.

### CI Changes

This section includes PRs that mention `microshift` in the
`openshift/origin` or `openshift/release` repositories.

### Documentation

This section includes PRs that mention `microshift` in the
`openshift/openshift-docs` repository.
