---
name: Analyze CI for Pull Requests
argument-hint: [--rebase]
description: Analyze CI for open MicroShift pull requests and produce a summary of failures
allowed-tools: Skill, Bash, Read, Write, Glob, Grep, Agent
---

# analyze-ci:pull-requests

## Synopsis
```
/analyze-ci:pull-requests [--rebase]
```

## Description
Fetches all open MicroShift pull requests, identifies failed Prow CI jobs for each PR, analyzes each failure using the `/analyze-ci:prow-job` command, and produces a text summary report.

This command orchestrates the analysis workflow by:

1. Fetching the list of open PRs and their failed jobs using `.claude/scripts/microshift-prow-jobs-for-pull-requests.sh --mode detail`
2. Filtering to only PRs that have at least one failed job
3. Analyzing each failed job individually using the `/analyze-ci:prow-job` command
4. Aggregating results into a summary report saved to `${WORKDIR}`

## Arguments
- `--rebase` (optional): Only analyze rebase PRs (authored by `microshift-rebase-script[bot]`)

## Work Directory

Set once at the start and reference throughout:
```bash
WORKDIR=/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)
```

## Implementation Steps

### Step 1: Fetch Open PRs and Their Job Results

**Goal**: Get the list of open PRs with failed Prow jobs.

**Actions**:
1. Execute `.claude/scripts/microshift-prow-jobs-for-pull-requests.sh --mode detail` to get all open PRs and their job statuses as a JSON array. If `--rebase` was specified, add `--author "microshift-rebase-script[bot]"` to only include rebase PRs
2. Parse the JSON to identify PRs with failed jobs (jobs where `.status == "FAILURE"`)
3. For each failed job, extract from the JSON:
   - PR number (`.pr_number`), title (`.title`), and URL (`.url`)
   - Job name (`.jobs[].job`), status (`.jobs[].status`), Prow URL (`.jobs[].url`), and build ID (`.jobs[].build_id`)

**Example Commands**:
```bash
# All open PRs
bash .claude/scripts/microshift-prow-jobs-for-pull-requests.sh --mode detail 2>/dev/null

# Rebase PRs only
bash .claude/scripts/microshift-prow-jobs-for-pull-requests.sh --mode detail --author "microshift-rebase-script[bot]" 2>/dev/null
```

**JSON output format** (array of PR objects with nested jobs):
```json
[
  {
    "pr_number": 6313,
    "title": "USHIFT-6636: Change test-agent impl to align with greenboot-rs",
    "url": "https://github.com/openshift/microshift/pull/6313",
    "jobs": [
      {
        "job": "pull-ci-openshift-microshift-main-e2e-aws-tests",
        "status": "FAILURE",
        "url": "https://prow.ci.openshift.org/view/gs/...",
        "build_id": "2039755516510474240",
        "finished": "2026-04-01T06:38:10Z"
      }
    ]
  }
]
```

**Error Handling**:
- If no open PRs found (empty JSON array), report "No open pull requests found" and exit successfully
- If no failed jobs across all PRs, report "All PR jobs are passing" and exit successfully
- If `microshift-prow-jobs-for-pull-requests.sh` fails, report error and exit

### Step 2: Download All Job Artifacts

**Goal**: Download artifacts for all failed jobs in parallel before analysis.

**Actions**:
1. Run `WORKDIR=/tmp/analyze-ci-claude-workdir.$(date +%y%m%d) && mkdir -p ${WORKDIR}` using the `Bash` tool
2. Filter the JSON from Step 1 to only include PRs with failed jobs, then pipe into the download script:
   ```bash
   echo '<json_from_step1>' | \
       jq '[.[] | select(.jobs | map(select(.status == "FAILURE")) | length > 0)]' | \
       WORKDIR=${WORKDIR} bash .claude/scripts/analyze-ci-download-jobs.sh
   ```
3. The script auto-detects the nested PR format, flattens the jobs, downloads artifacts in parallel to `${WORKDIR}/artifacts/<build_id>/`, and outputs enriched JSON with `artifacts_dir` fields added
4. Save the output JSON for use in Step 3

**Error Handling**:
- If some downloads fail, note the failures but proceed with successfully downloaded jobs

### Step 3: Analyze Each Failed Job Using /analyze-ci:prow-job

**Goal**: Get detailed root cause analysis for each failed job using pre-downloaded artifacts.

**IMPORTANT - Mandatory Per-Job Agent Requirements**:
Each job MUST be analyzed by launching a **separate Agent** (using the `Agent` tool, NOT the `Skill` tool) for each job. Do NOT analyze jobs inline or combine multiple jobs into a single agent. Each agent MUST:
1. Run `/analyze-ci:prow-job <artifacts_dir>` using the local artifacts directory (not by reimplementing the analysis)
2. After the analysis completes, write the analysis report to an **exact file path** (see File Storage below)
3. The file MUST contain the full report including the `--- STRUCTURED SUMMARY ---` block

**Actions**:
1. For each failed job in the enriched JSON from Step 2, launch a separate **Agent** with this exact prompt template, using the `.artifacts_dir` field as `<ARTIFACTS_DIR>`, `.pr_number` as `<PR>`, and the last segment of `.job` as `<JOB_NAME_SUFFIX>`:
   ```
   Agent: subagent_type=general_purpose, prompt="Analyze this Prow job and save the report:
   1. Run /analyze-ci:prow-job <ARTIFACTS_DIR>
   2. After the analysis completes, save the FULL report output (including the --- STRUCTURED SUMMARY --- block) to:
      ${WORKDIR}/analyze-ci-prs-job-<N>-pr<PR>-<JOB_NAME_SUFFIX>.txt
      Use the Write tool to save the file. The file must contain the complete analysis report."
   ```
   Replace `<ARTIFACTS_DIR>`, `<N>` (1-based job index), `<PR>` (PR number), and `<JOB_NAME_SUFFIX>` with actual values from the JSON.
2. Launch **ALL** job agents in parallel using `run_in_background: true`
3. Wait for all agents to complete before proceeding to Step 4

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
- `${WORKDIR}/analyze-ci-prs-job-<N>-pr<PR>-<job-name-suffix>.txt`

Where:
- `<N>` is the 1-based job index
- `<PR>` is the PR number
- `<job-name-suffix>` is the last segment of the job name

**Verification**: After all agents complete, verify that per-job files exist:
```bash
ls ${WORKDIR}/analyze-ci-prs-job-*.txt
```
If any files are missing, note the gap in the summary report but do NOT re-run the analysis.

### Step 4: Aggregate Results and Identify Patterns

**Goal**: Find common failure patterns across all PRs and jobs.

**Actions**:
1. Collect results from all parallel job analyses
   - Read each per-job file from `${WORKDIR}`
   - Extract the `--- STRUCTURED SUMMARY ---` block from each file for pattern detection (including the `FINISHED` field for job date tracking)
   - Extract key findings from each analysis

2. Group failures by PR:
   - List each PR with its failed jobs and root causes

3. Identify common errors across PRs:
   - Count occurrences of similar error messages
   - Group jobs with identical root causes (e.g., same infrastructure issue affecting multiple PRs)

### Step 5: Generate Summary Report

**Goal**: Present actionable summary to the user.

**Actions**:
1. Aggregate all job analysis results from parallel execution
2. Identify common patterns and group by PR and failure type
3. Generate summary report and save to `${WORKDIR}/analyze-ci-prs-summary.<timestamp>.txt`
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
  Report: ${WORKDIR}/analyze-ci-prs-summary.20260315-143022.txt

PER-PR BREAKDOWN

PR #6313: USHIFT-6636: Change test-agent impl to align with greenboot-rs
  https://github.com/openshift/microshift/pull/6313
  Jobs: 8 passed, 7 failed

  Failed Jobs:
  1. pull-ci-openshift-microshift-main-e2e-aws-tests [2026-03-15]
     Status: FAILURE
     Root Cause: [summarized from analyze-ci:prow-job]
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

Individual job reports: ${WORKDIR}/analyze-ci-prs-job-*.txt
```

## Examples

### Example 1: Analyze All Failed PR Jobs

```
/analyze-ci:pull-requests
```

**Behavior**:
- Fetches all open PRs and their Prow job results
- Identifies PRs with failed jobs
- Analyzes each failed job
- Presents aggregated summary grouped by PR

### Example 2: Analyze Only Rebase PRs

```
/analyze-ci:pull-requests --rebase
```

**Behavior**:
- Only includes PRs authored by `microshift-rebase-script[bot]`
- Useful for release manager workflow to check rebase CI status

## Performance Considerations

- **Execution Time**: Depends on number of failed jobs; parallel execution helps significantly
- **Network Usage**: Each job analysis fetches logs from GCS
- **Parallelization**: All job analyses run in parallel for maximum efficiency
- **File Storage**: All intermediate and report files are stored in `${WORKDIR}` directory

## Prerequisites

- `.claude/scripts/microshift-prow-jobs-for-pull-requests.sh` script must exist and be executable
- `/analyze-ci:prow-job` command must be available
- `gh` CLI must be authenticated with access to openshift/microshift
- `gcloud` CLI must be installed and authenticated for GCS access (used by analyze-ci:prow-job)
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

- **analyze-ci:prow-job**: Detailed analysis of a single job (used internally)
- **analyze-ci:release**: Similar analysis for periodic release jobs
- **analyze-ci:doctor**: Multi-release analysis with HTML output
- **analyze-ci:test-scenario**: Analyze specific test scenario results

## Notes

- This skill focuses on **presubmit** PR jobs (not periodic/postsubmit)
- Analysis is read-only - no modifications to CI data or PRs
- Results are saved in files in ${WORKDIR} directory with a timestamp
- Provide links to the jobs in the summary
- Only present a concise analysis summary for each job
- PRs with no Prow jobs (e.g., drafts without triggered tests) are skipped
- Pattern detection improves with more jobs analyzed
