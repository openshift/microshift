# OCP Advisory Publication Report

## Description

Every week for every OCP release MicroShift team decide if a new MicroShift version should be publish.
The decision is done if there are important changes/fixes to the published.
This script will generate a report with advisory, CVEs and jira ticket relevant info for every OCP version to make the decision.

### Steps
1. Find advisory ids for a OCP version from [ocp-build-data repository](https://github.com/openshift-eng/ocp-build-data)
   - For example, for release `4.17` check [`releases.yml`](https://github.com/openshift-eng/ocp-build-data/blob/openshift-4.17/releases.yml)
2. Get list of CVEs from a Red Hat advisory using [Red Hat errata API](https://errata.devel.redhat.com/documentation/developer-guide/api-http-api.html#get-cveshowerrata_id.json)
   - Only available behind the Red Hat VPN
   - Example for [`144556` advisory](https://errata.devel.redhat.com/cve/show/144556.json)
3. Query Jira to find if there are any MicroShift ticket to address a CVE fix
   - Example for [`summary  ~ "CVE-2024-21626" AND component = MicroShift and (affectedVersion = 4.17 or affectedVersion = 4.17.z)`](https://issues.redhat.com/issues/?jql=summary%20%20~%20%22CVE-2024-21626%22%20AND%20component%20%3D%20MicroShift%20and%20(affectedVersion%20%3D%204.17%20or%20affectedVersion%20%3D%204.17.z)) query

## Requisites

### Jira API token

Visit [the Profile page on the Jira
Server](https://issues.redhat.com/secure/ViewProfile.jspa?selectedTab=com.atlassian.pats.pats-plugin:jira-user-personal-access-tokens) and create a token.

Set the `JIRA_API_TOKEN` in your env:

```
export JIRA_API_TOKEN="TOKEN_VALUE"
```

### Connect to Red Hat VPN

Red Hat VPN connection is mandatory to get info from Red Hat errata tool.

## Generate report

Run `./scripts/advisory_publication/advisory_publication_report.sh X.Y.Z` to generate json report.

`X.Y.Z` is the target OCP version, for example `4.17.12`

Output format:

```
{
  "RHSA-YYYY:NNNNN": {
    "type": "extras",
    "url": "https://errata.devel.redhat.com/advisory/XXXXX",
    "cves": {
      "CVE-YYYY-NNNN": {},
      "CVE-YYYY-NNNN": {}
    }
  },
  "RHSA-2025:0364": {
    "type": "image",
    "url": "https://errata.devel.redhat.com/advisory/YYYYY",
    "cves": {
      "CVE-YYYY-NNNN": {},
      "CVE-YYYY-NNNN": {
        "jira_ticket": {
          "id": "OCPBUGS-MMMMM",
          "summary": "CVE-YYYY-NNNN title of the jira ticket",
          "status": "Closed",
          "resolution": "Not a Bug"
        }
      }
    }
  },
  "RHSA-YYYY:NNNNN": {
    "type": "metadata",
    "url": "https://errata.devel.redhat.com/advisory/ZZZZZ",
    "cves": {}
  }
}

```
