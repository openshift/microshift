---
name: Analyze CI for a Release
argument-hint: <release-version>
description: Analyze CI for a MicroShift release using analyze-ci-for-prow-job command and produce a summary
allowed-tools: Skill, Bash, Read, Write, Glob, Grep, Agent
---

# analyze-ci-for-release

## Synopsis
```bash
/analyze-ci-for-release <release-version>
```

## Description
Analyzes all failed periodic jobs for a specific MicroShift release by leveraging existing tools and agents. This command orchestrates the analysis workflow by:

1. Fetching list of failed periodic jobs using `.claude/scripts/microshift-prow-jobs-for-release.sh`
2. Analyzing each job individually using the `/analyze-ci-for-prow-job` command
3. Aggregating results and presenting a concise summary with common failure patterns

This approach reuses existing analysis capabilities rather than duplicating logic.

## Arguments
- `<release-version>` (required): OpenShift release version (e.g., 4.22, 4.21, 4.20)

## Implementation Steps

### Step 1: Validate Arguments and Fetch Failed Jobs

**Goal**: Get the list of failed periodic jobs for the release.

**Actions**:
1. Validate release version argument is provided
2. Execute `.claude/scripts/microshift-prow-jobs-for-release.sh <release>` to get all failed jobs
3. Filter output to only include periodic jobs (containing `-periodics-`)
4. Extract job URLs from the output

**Example Command**:
```bash
bash .claude/scripts/microshift-prow-jobs-for-release.sh 4.22 | grep -E "periodics-" | awk '{print $NF}'
```

**Expected Output**:
```text
Found 17 failed periodic jobs for release 4.22
```

**Error Handling**:
- If no release version provided, show usage and exit
- If no periodic jobs found, report "No failed periodic jobs found for release X.XX" and exit successfully
- If microshift-prow-jobs-for-release.sh fails, report error and exit

### Step 2: Analyze Each Job Using /analyze-ci-for-prow-job

**Goal**: Get detailed root cause analysis for each failed job.

**IMPORTANT - Mandatory Per-Job Agent Requirements**:
Each job MUST be analyzed by launching a **separate Agent** (using the `Agent` tool, NOT the `Skill` tool) for each job URL. Do NOT analyze jobs inline or combine multiple jobs into a single agent. Each agent MUST:
1. Run `/analyze-ci-for-prow-job <job-url>` (not by reimplementing the analysis)
2. After the analysis completes, write the analysis report to an **exact file path** (see File Storage below)
3. The file MUST contain the full report including the `--- STRUCTURED SUMMARY ---` block

**Actions**:
1. Run `mkdir -p /tmp/analyze-ci-claude-workdir.$(date +%y%m%d)` using the `Bash` tool
2. For each job URL, launch a separate **Agent** with this exact prompt template:
   ```
   Agent: subagent_type=general_purpose, prompt="Analyze this Prow job and save the report:
   1. Run /analyze-ci-for-prow-job <JOB_URL>
   2. After the analysis completes, save the FULL report output (including the --- STRUCTURED SUMMARY --- block) to:
      /tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/analyze-ci-release-<RELEASE>-job-<N>-<JOB_ID>.txt
      Use the Write tool to save the file. The file must contain the complete analysis report."
   ```
   Replace `<JOB_URL>`, `<RELEASE>`, `<N>` (1-based job index), and `<JOB_ID>` with actual values.
3. Launch **ALL** job agents in parallel using `run_in_background: true`
4. Wait for all agents to complete before proceeding to Step 3

**Progress Reporting**:
```text
Analyzing N jobs in parallel...
Job 1/N: <job-name> (background)
Job 2/N: <job-name> (background)
...
```

**File Storage**:
All per-job report files MUST be saved at this exact path:
- `/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/analyze-ci-release-<release>-job-<N>-<job-id>.txt`

Where:
- `<release>` is the release version (e.g., `4.22`, `main`)
- `<N>` is the 1-based job index
- `<job-id>` is the numeric Prow job ID from the URL

**Verification**: After all agents complete, verify that per-job files exist:
```bash
ls /tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/analyze-ci-release-<release>-job-*.txt
```
If any files are missing, note the gap in the summary report but do NOT re-run the analysis.

### Step 3: Aggregate Results and Identify Patterns

**Goal**: Find common failure patterns across all jobs from parallel execution.

**Actions**:
1. Collect results from all parallel job analyses
   - Read each per-job file: `/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/analyze-ci-release-<release>-job-<N>-<job-id>.txt`
   - Extract the `--- STRUCTURED SUMMARY ---` block from each file for pattern detection (including the `FINISHED` field for job date tracking)
   - Extract key findings from each analysis

2. Group jobs by failure type:
   - Build/infrastructure failures
   - Test execution failures
   - Configuration/setup issues

3. Identify most common errors:
   - Count occurrences of similar error messages
   - Group jobs with identical root causes

4. Categorize by severity:
   - CRITICAL: Affects multiple jobs, blocks release
   - HIGH: Affects several jobs
   - MEDIUM: Isolated failures
   - LOW: Flaky/intermittent issues

### Step 4: Generate Concise Summary Report

**Goal**: Present actionable summary to the user.

**Actions**:
1. Aggregate all job analysis results from parallel execution
2. Identify common patterns and group by failure type
3. Generate summary report and save to `/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/analyze-ci-release-<release>-summary.<timestamp>.txt`
4. Display the summary to the user

**Important**: Each job listed under "Affected Jobs" MUST include:
- The job name followed by the finish date in `[YYYY-MM-DD]` format (from the per-job `FINISHED` field)
- The Prow URL on the next line
This ensures the HTML report generator can extract dates without reading per-job files.

**Report Structure**:

```text
═══════════════════════════════════════════════════════════════
MICROSHIFT 4.22 RELEASE - FAILED JOBS ANALYSIS
═══════════════════════════════════════════════════════════════

OVERVIEW
  Total Failed Jobs: 17
  Analysis Date: 2026-03-14
  Report saved to: /tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/analyze-ci-release-4.22-summary.<timestamp>.txt

FAILURE BREAKDOWN
  Build Failures:        0 jobs
  Test Failures:        15 jobs
  Infrastructure:        2 jobs

TOP ISSUES (by frequency)

1. OCP Conformance Test Failures (8 jobs)
   Severity: HIGH
   Pattern: Tests timeout or fail in conformance suite
   Affected Jobs:
   * periodic-ci-openshift-microshift-release-4.22-periodics-e2e-aws-ovn-ocp-conformance [2026-03-14]
     https://prow.ci.openshift.org/view/gs/test-platform-results/logs/.../1234567890
   * periodic-ci-openshift-microshift-release-4.22-periodics-e2e-aws-ovn-ocp-conformance-serial [2026-03-14]
     https://prow.ci.openshift.org/view/gs/test-platform-results/logs/.../1234567891
   * ... (6 more)

   Root Cause: [summarized from analyze-ci-for-prow-job results]
   Next Steps: [recommended actions]

2. Bootc Image Test Failures (4 jobs)
   Severity: MEDIUM
   Pattern: Image build or deployment issues
   Affected Jobs:
   * periodic-ci-openshift-microshift-release-4.22-periodics-e2e-aws-tests-bootc-nightly [2026-03-14]
     https://prow.ci.openshift.org/view/gs/test-platform-results/logs/.../1234567892
   * ... (3 more)

   Root Cause: [summarized]
   Next Steps: [recommended actions]

3. Infrastructure/Timeout Issues (2 jobs)
   Severity: LOW
   Pattern: Jobs timeout or fail to allocate resources
   Affected Jobs:
   * periodic-ci-openshift-microshift-release-4.22-periodics-rebase-on-nightlies [2026-03-14]
     https://prow.ci.openshift.org/view/gs/test-platform-results/logs/.../1234567893
   * periodic-ci-openshift-microshift-release-4.22-periodics-update-versions-releases [2026-03-14]
     https://prow.ci.openshift.org/view/gs/test-platform-results/logs/.../1234567894

   Root Cause: [summarized]

═══════════════════════════════════════════════════════════════

Individual job reports available in:
  /tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/analyze-ci-release-4.22-job-*.txt
```

## Examples

### Example 1: Analyze All Failed Jobs

```bash
/analyze-ci-for-release 4.22
```

**Behavior**:
- Fetches all failed periodic jobs for 4.22
- Analyzes each job using /analyze-ci-for-prow-job command
- Presents aggregated summary

### Example 2: Different Release

```bash
/analyze-ci-for-release 4.21
```

**Behavior**:
- Analyzes 4.21 release jobs
- Same workflow as 4.22

## Performance Considerations

- **Execution Time**: Significantly reduced through parallel execution - typically 2-3 minutes for 15-20 jobs (depends on analyze-ci-for-prow-job execution time)
- **Network Usage**: Moderate to high - all jobs analyzed in parallel fetch logs from GCS simultaneously
- **Parallelization**: All jobs are analyzed in parallel for maximum efficiency
- **File Storage**: All intermediate and report files are stored in `/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)` directory

## Prerequisites

- `.claude/scripts/microshift-prow-jobs-for-release.sh` script must exist and be executable
- `/analyze-ci-for-prow-job` command must be available
- `gcloud` CLI must be installed and authenticated for GCS access (used by analyze-ci-for-prow-job)
- Internet access to fetch job data from Prow/GCS
- Bash shell

## Error Handling

### No Failed Jobs
```text
No failed periodic jobs found for release 4.22
All periodic jobs are passing.
```

### Invalid Release Version
```text
Error: Invalid release version
Usage: /analyze-ci-for-release <release-version>
Example: /analyze-ci-for-release 4.22
```

### microshift-prow-jobs-for-release.sh Not Found
```text
Error: Could not find .claude/scripts/microshift-prow-jobs-for-release.sh
Please ensure you're in the microshift project directory.
```

## Related Skills

- **analyze-ci-for-prow-job**: Detailed analysis of a single job (used internally)
- **analyze-ci-create-bugs**: Creates JIRA bugs from this command's output
- **analyze-ci-for-release-manager**: Multi-release orchestrator that delegates to this command
- **analyze-ci-test-scenario**: Analyze specific test scenario results
- **analyze-sos-report**: Investigate runtime issues from SOS reports

## Use Cases

### Pre-Release Verification
```bash
/analyze-ci-for-release 4.22
```
Complete analysis before cutting a release

### Trend Analysis
Run periodically and compare summaries over time to identify regression patterns

## Notes

- This skill focuses on **periodic** jobs only (not presubmit/postsubmit)
- Analysis is read-only - no modifications to CI data
- Results are saved in files in /tmp/analyze-ci-claude-workdir.$(date +%y%m%d) directory with a timestamp
- Provide links to the jobs in the summary
- Only present a concise analysis summary for each job
- Pattern detection improves with more jobs analyzed
