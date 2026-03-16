---
name: Analyze CI for Pull Requests
argument-hint: [--rebase] [--limit N]
description: Analyze CI for open MicroShift pull requests and produce a summary of failures
allowed-tools: Skill, Bash, Read, Write, Glob, Grep, Agent
---

# analyze-ci-for-pull-requests

## Synopsis
```
/analyze-ci-for-pull-requests [--rebase] [--limit N]
```

## Description
Fetches all open MicroShift pull requests, identifies failed Prow CI jobs for each PR, analyzes each failure using the `openshift-ci-analysis` agent, and produces a text summary report.

This command orchestrates the analysis workflow by:

1. Fetching the list of open PRs and their failed jobs using `.claude/scripts/microshift-prow-jobs-for-pull-requests.sh --mode detail`
2. Filtering to only PRs that have at least one failed job
3. Analyzing each failed job individually using the `openshift-ci-analysis` agent
4. Aggregating results into a summary report saved to `/tmp`

## Arguments
- `--rebase` (optional): Only analyze rebase PRs (titles containing `NO-ISSUE: rebase-release-`)
- `--limit N` (optional): Limit analysis to first N failed jobs total (useful for quick checks, default: all failed jobs)

## Implementation Steps

### Step 1: Fetch Open PRs and Their Job Results

**Goal**: Get the list of open PRs with failed Prow jobs.

**Actions**:
1. Execute `.claude/scripts/microshift-prow-jobs-for-pull-requests.sh --mode detail` to get all open PRs and their job statuses. If `--rebase` was specified, add `--filter "NO-ISSUE: rebase-release-"` to only include rebase PRs
2. Parse the output to identify PRs with failed jobs (lines containing `✗`)
3. For each failed job, extract:
   - PR number and title (from the `=== PR #NNN: ... ===` header lines)
   - PR URL (line following the header)
   - Job name (first column)
   - Job URL (last column, the Prow URL)
4. Apply `--limit N` if specified

**Example Commands**:
```bash
# All open PRs
bash .claude/scripts/microshift-prow-jobs-for-pull-requests.sh --mode detail 2>/dev/null

# Rebase PRs only
bash .claude/scripts/microshift-prow-jobs-for-pull-requests.sh --mode detail --filter "NO-ISSUE: rebase-release-" 2>/dev/null
```

**Expected Output Format**:
```
=== PR #6313: USHIFT-6636: Change test-agent impl to align with greenboot-rs ===
    https://github.com/openshift/microshift/pull/6313

    JOB                                                  STATUS  URL
    pull-ci-openshift-microshift-main-e2e-aws-tests      ✗       https://prow.ci.openshift.org/view/gs/...
    pull-ci-openshift-microshift-main-e2e-aws-tests-arm  ✗       https://prow.ci.openshift.org/view/gs/...
    ...
```

**Error Handling**:
- If no open PRs found, report "No open pull requests found" and exit successfully
- If no failed jobs across all PRs, report "All PR jobs are passing" and exit successfully
- If `microshift-prow-jobs-for-pull-requests.sh` fails, report error and exit

### Step 2: Analyze Each Failed Job Using openshift-ci-analysis Agent

**Goal**: Get detailed root cause analysis for each failed job.

**Actions**:
1. For each failed job URL from Step 1:
   - Call the `openshift-ci-analysis` agent with the job URL **in parallel**
   - Capture the analysis result (failure reason, error summary)
   - Store all intermediate analysis files in `/tmp`

2. Progress reporting:
   - Show "Analyzing job X/Y: <job-name> (PR #NNN)" for each job
   - Use the Agent tool to invoke `openshift-ci-analysis` for each URL
   - **Run all job analyses in parallel** to maximize efficiency

**Data Collection**:
For each job analysis, extract:
- Job name
- Job URL
- PR number and title
- Failure type (build failure, test failure, infrastructure issue)
- Primary error/root cause
- Affected test scenarios (if applicable)

**File Storage**:
All intermediate analysis files are stored in `/tmp` with naming pattern:
- `/tmp/analyze-ci-prs-job-<N>-pr<PR>-<job-name-suffix>.txt`

### Step 3: Aggregate Results and Identify Patterns

**Goal**: Find common failure patterns across all PRs and jobs.

**Actions**:
1. Collect results from all parallel job analyses
   - Read individual job analysis files from `/tmp`
   - Extract key findings from each analysis

2. Group failures by PR:
   - List each PR with its failed jobs and root causes

3. Identify common errors across PRs:
   - Count occurrences of similar error messages
   - Group jobs with identical root causes (e.g., same infrastructure issue affecting multiple PRs)

### Step 4: Generate Summary Report

**Goal**: Present actionable summary to the user.

**Actions**:
1. Aggregate all job analysis results from parallel execution
2. Identify common patterns and group by PR and failure type
3. Generate summary report and save to `/tmp/analyze-ci-prs-summary.<timestamp>.txt`
4. Display the summary to the user

**Report Structure**:

```
═══════════════════════════════════════════════════════════════
MICROSHIFT OPEN PULL REQUESTS - FAILED JOBS ANALYSIS
═══════════════════════════════════════════════════════════════

OVERVIEW
  Total Open PRs: 6
  PRs with Failures: 2
  Total Failed Jobs: 9
  Analysis Date: 2026-03-15
  Report: /tmp/analyze-ci-prs-summary.20260315-143022.txt

PER-PR BREAKDOWN

PR #6313: USHIFT-6636: Change test-agent impl to align with greenboot-rs
  https://github.com/openshift/microshift/pull/6313
  Jobs: 8 passed, 7 failed

  Failed Jobs:
  1. pull-ci-openshift-microshift-main-e2e-aws-tests
     Status: FAILURE
     Root Cause: [summarized from openshift-ci-analysis]
     URL: https://prow.ci.openshift.org/view/gs/...

  2. pull-ci-openshift-microshift-main-e2e-aws-tests-arm
     Status: FAILURE
     Root Cause: [summarized]
     URL: https://prow.ci.openshift.org/view/gs/...

  ... (more failed jobs)

PR #6116: USHIFT-6491: Improve gitops test
  https://github.com/openshift/microshift/pull/6116
  Jobs: 15 passed, 2 failed

  Failed Jobs:
  1. pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-periodic
     Status: FAILURE
     Root Cause: [summarized]
     URL: https://prow.ci.openshift.org/view/gs/...

COMMON PATTERNS (across PRs)
  If the same failure pattern appears in multiple PRs, list it here
  with the affected PRs and jobs.

═══════════════════════════════════════════════════════════════

Individual job reports: /tmp/analyze-ci-prs-job-*.txt
```

## Examples

### Example 1: Analyze All Failed PR Jobs

```
/analyze-ci-for-pull-requests
```

**Behavior**:
- Fetches all open PRs and their Prow job results
- Identifies PRs with failed jobs
- Analyzes each failed job
- Presents aggregated summary grouped by PR

### Example 2: Analyze Only Rebase PRs

```
/analyze-ci-for-pull-requests --rebase
```

**Behavior**:
- Only includes PRs with `NO-ISSUE: rebase-release-` in the title
- Useful for release manager workflow to check rebase CI status

### Example 3: Quick Analysis (First 5 Failed Jobs)

```
/analyze-ci-for-pull-requests --limit 5
```

**Behavior**:
- Analyzes only first 5 failed jobs across all PRs
- Useful for quick health check

## Performance Considerations

- **Execution Time**: Depends on number of failed jobs; parallel execution helps significantly
- **Network Usage**: Each job analysis fetches logs from GCS
- **Parallelization**: All job analyses run in parallel for maximum efficiency
- **Use --limit**: For quick checks, use --limit flag to analyze a subset
- **File Storage**: All intermediate and report files are stored in `/tmp` directory

## Prerequisites

- `.claude/scripts/microshift-prow-jobs-for-pull-requests.sh` script must exist and be executable
- `openshift-ci-analysis` agent must be available
- `gh` CLI must be authenticated with access to openshift/microshift
- Internet access to fetch job data from GCS
- Bash shell

## Error Handling

### No Open PRs
```
No open pull requests found.
```

### No Failed Jobs
```
All open PR jobs are passing! ✓
No failures to analyze.
```

### Script Not Found
```
Error: Could not find .claude/scripts/microshift-prow-jobs-for-pull-requests.sh
Please ensure you're in the microshift project directory.
```

## Related Skills

- **openshift-ci-analysis**: Detailed analysis of a single job (used internally)
- **analyze-ci-for-release**: Similar analysis for periodic release jobs
- **analyze-ci-for-release-manager**: Multi-release analysis with HTML output
- **analyze-ci-test-scenario**: Analyze specific test scenario results

## Notes

- This skill focuses on **presubmit** PR jobs (not periodic/postsubmit)
- Analysis is read-only - no modifications to CI data or PRs
- Results are saved in files in /tmp directory with a timestamp
- Provide links to the jobs in the summary
- Only present a concise analysis summary for each job
- PRs with no Prow jobs (e.g., drafts without triggered tests) are skipped
- Pattern detection improves with more jobs analyzed (avoid limiting unless needed)
