---
name: Analyze CI for Pull Requests
argument-hint: [--rebase]
description: Analyze CI for open MicroShift pull requests and produce a summary of failures
allowed-tools: Skill, Bash, Read, Write, Glob, Grep, Agent
---

# analyze-ci-for-pull-requests

## Synopsis
```
/analyze-ci-for-pull-requests [--rebase]
```

## Description
Fetches all open MicroShift pull requests, identifies failed Prow CI jobs for each PR, analyzes each failure using the `/analyze-ci-for-prow-job` command, and produces a text summary report.

This command orchestrates the analysis workflow by:

1. Fetching the list of open PRs and their failed jobs using `.claude/scripts/microshift-prow-jobs-for-pull-requests.sh --mode detail`
2. Filtering to only PRs that have at least one failed job
3. Analyzing each failed job individually using the `/analyze-ci-for-prow-job` command
4. Aggregating results into a summary report saved to `/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)`

## Arguments
- `--rebase` (optional): Only analyze rebase PRs (titles containing `NO-ISSUE: rebase-release-`)

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

### Step 2: Analyze Each Failed Job Using /analyze-ci-for-prow-job

**Goal**: Get detailed root cause analysis for each failed job.

**IMPORTANT - Mandatory Per-Job Agent Requirements**:
Each job MUST be analyzed by launching a **separate Agent** (using the `Agent` tool, NOT the `Skill` tool) for each job URL. Do NOT analyze jobs inline or combine multiple jobs into a single agent. Each agent MUST:
1. Run `/analyze-ci-for-prow-job <job-url>` (not by reimplementing the analysis)
2. After the analysis completes, write the analysis report to an **exact file path** (see File Storage below)
3. The file MUST contain the full report including the `--- STRUCTURED SUMMARY ---` block

**Actions**:
1. Run `mkdir -p /tmp/analyze-ci-claude-workdir.$(date +%y%m%d)` using the `Bash` tool
2. For each failed job URL, launch a separate **Agent** with this exact prompt template:
   ```
   Agent: subagent_type=general_purpose, prompt="Analyze this Prow job and save the report:
   1. Run /analyze-ci-for-prow-job <JOB_URL>
   2. After the analysis completes, save the FULL report output (including the --- STRUCTURED SUMMARY --- block) to:
      /tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/analyze-ci-prs-job-<N>-pr<PR>-<JOB_NAME_SUFFIX>.txt
      Use the Write tool to save the file. The file must contain the complete analysis report."
   ```
   Replace `<JOB_URL>`, `<N>` (1-based job index), `<PR>` (PR number), and `<JOB_NAME_SUFFIX>` (last segment of job name) with actual values.
3. Launch **ALL** job agents in parallel using `run_in_background: true`
4. Wait for all agents to complete before proceeding to Step 3

**Progress Reporting**:
```text
Analyzing N jobs in parallel...
Job 1/N: <job-name> (PR #NNN) (background)
Job 2/N: <job-name> (PR #NNN) (background)
...
```

**Data Collection**:
For each job analysis, extract:
- Job name
- Job URL
- PR number and title
- Failure type (build failure, test failure, infrastructure issue)
- Primary error/root cause
- Affected test scenarios (if applicable)

**File Storage**:
All per-job report files MUST be saved at this exact path:
- `/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/analyze-ci-prs-job-<N>-pr<PR>-<job-name-suffix>.txt`

Where:
- `<N>` is the 1-based job index
- `<PR>` is the PR number
- `<job-name-suffix>` is the last segment of the job name

**Verification**: After all agents complete, verify that per-job files exist:
```bash
ls /tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/analyze-ci-prs-job-*.txt
```
If any files are missing, note the gap in the summary report but do NOT re-run the analysis.

### Step 3: Aggregate Results and Identify Patterns

**Goal**: Find common failure patterns across all PRs and jobs.

**Actions**:
1. Collect results from all parallel job analyses
   - Read each per-job file from `/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)`
   - Extract the `--- STRUCTURED SUMMARY ---` block from each file for pattern detection (including the `FINISHED` field for job date tracking)
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
3. Generate summary report and save to `/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/analyze-ci-prs-summary.<timestamp>.txt`
4. Display the summary to the user

**Important**: Each failed job MUST include the finish date in `[YYYY-MM-DD]` format (from the per-job `FINISHED` field) after the job name. This ensures the HTML report generator can extract dates without reading per-job files.

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
  Report: /tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/analyze-ci-prs-summary.20260315-143022.txt

PER-PR BREAKDOWN

PR #6313: USHIFT-6636: Change test-agent impl to align with greenboot-rs
  https://github.com/openshift/microshift/pull/6313
  Jobs: 8 passed, 7 failed

  Failed Jobs:
  1. pull-ci-openshift-microshift-main-e2e-aws-tests [2026-03-15]
     Status: FAILURE
     Root Cause: [summarized from analyze-ci-for-prow-job]
     URL: https://prow.ci.openshift.org/view/gs/...

  2. pull-ci-openshift-microshift-main-e2e-aws-tests-arm [2026-03-15]
     Status: FAILURE
     Root Cause: [summarized]
     URL: https://prow.ci.openshift.org/view/gs/...

  ... (more failed jobs)

PR #6116: USHIFT-6491: Improve gitops test
  https://github.com/openshift/microshift/pull/6116
  Jobs: 15 passed, 2 failed

  Failed Jobs:
  1. pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-periodic [2026-03-15]
     Status: FAILURE
     Root Cause: [summarized]
     URL: https://prow.ci.openshift.org/view/gs/...

COMMON PATTERNS (across PRs)
  If the same failure pattern appears in multiple PRs, list it here
  with the affected PRs and jobs.

═══════════════════════════════════════════════════════════════

Individual job reports: /tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/analyze-ci-prs-job-*.txt
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

## Performance Considerations

- **Execution Time**: Depends on number of failed jobs; parallel execution helps significantly
- **Network Usage**: Each job analysis fetches logs from GCS
- **Parallelization**: All job analyses run in parallel for maximum efficiency
- **File Storage**: All intermediate and report files are stored in `/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)` directory

## Prerequisites

- `.claude/scripts/microshift-prow-jobs-for-pull-requests.sh` script must exist and be executable
- `/analyze-ci-for-prow-job` command must be available
- `gh` CLI must be authenticated with access to openshift/microshift
- `gcloud` CLI must be installed and authenticated for GCS access (used by analyze-ci-for-prow-job)
- Internet access to fetch job data from GCS
- Bash shell

## Error Handling

### No Open PRs
```
No open pull requests found.
```

### No Failed Jobs
```
All open PR jobs are passing.
No failures to analyze.
```

### Script Not Found
```
Error: Could not find .claude/scripts/microshift-prow-jobs-for-pull-requests.sh
Please ensure you're in the microshift project directory.
```

## Related Skills

- **analyze-ci-for-prow-job**: Detailed analysis of a single job (used internally)
- **analyze-ci-for-release**: Similar analysis for periodic release jobs
- **analyze-ci-for-release-manager**: Multi-release analysis with HTML output
- **analyze-ci-test-scenario**: Analyze specific test scenario results

## Notes

- This skill focuses on **presubmit** PR jobs (not periodic/postsubmit)
- Analysis is read-only - no modifications to CI data or PRs
- Results are saved in files in /tmp/analyze-ci-claude-workdir.$(date +%y%m%d) directory with a timestamp
- Provide links to the jobs in the summary
- Only present a concise analysis summary for each job
- PRs with no Prow jobs (e.g., drafts without triggered tests) are skipped
- Pattern detection improves with more jobs analyzed
