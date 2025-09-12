# OCP Advisory Publication Report

## Description

Each week for every minor OCP release, the MicroShift team decides if a new MicroShift version should be publish.
The decision is based on if there are important changes/fixes to MicroShift or the OCP images it depends on.
This script will generate a report with advisories, CVEs and jira tickets relevant to making the decision.

## Important Notice

**ART has switched to releasing OCP releases via Konflux.** However, since RPMs are still being shipped via errata, when checking to see if a release is required for MicroShift, users need to:

1. **Switch to the 4.19 branch** to use the advisory publication script
2. Run: `sh advisory_publication_report.sh <ocp_release_version>` to check for rpm advisories.
3. User will need to export both `GITLAB_API_TOKEN` and `JIRA_API_TOKEN` environment variables.
4. A warning already appears when running `advisory_publication_report.sh` script from 4.20 and above branches to run `advisory_publication_report.sh` from 4.19 branches for rpm advisories

### Steps
1. Find shipment URL from https://github.com/openshift-eng/ocp-build-data/blob/openshift-<version>/releases.yml
2. Iterate through the list of extras, metadata and image advisories.
3. Check if there are any CVEs affecting MicroShift
4. Get the list of CVEs from a Red Hat advisory using [Red Hat errata API](https://errata.devel.redhat.com/documentation/developer-guide/api-http-api.html#get-cveshowerrata_id.json)
   - Only available behind the Red Hat VPN
   - Example for [`144556` advisory](https://errata.devel.redhat.com/cve/show/144556.json)
5. Query Jira to find if there are any MicroShift tickets to address a CVE fix
   - For example [`summary  ~ "CVE-2024-21626" AND component = MicroShift and (affectedVersion = 4.17 or affectedVersion = 4.17.z)`](https://issues.redhat.com/issues/?jql=summary%20%20~%20%22CVE-2024-21626%22%20AND%20component%20%3D%20MicroShift%20and%20(affectedVersion%20%3D%204.17%20or%20affectedVersion%20%3D%204.17.z)) query

## Requisites

### Jira API token

Visit [the Profile page on the Jira
Server](https://issues.redhat.com/secure/ViewProfile.jspa?selectedTab=com.atlassian.pats.pats-plugin:jira-user-personal-access-tokens) and create a token.

Set the `JIRA_API_TOKEN` in your env:

```
export JIRA_API_TOKEN="TOKEN_VALUE"
```

### GitLab API token

For releases 4.20 and above, you will also need a GitLab API token. Use your personal access token and set the `GITLAB_API_TOKEN` in your env:

```
export GITLAB_API_TOKEN="YOUR_PERSONAL_ACCESS_TOKEN"
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
