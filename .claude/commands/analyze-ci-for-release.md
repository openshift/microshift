---
name: Analyze CI for a Release
argument-hint: <release>
description: Analyze CI for a MicroShift release using openshift-ci-analysis agent and produce a summary
allowed-tools: Skill, Bash, Read, Write, Glob, Grep, Agent
---

# analyze-ci-for-release

## Name
analyze-ci-for-release - Analyze all failed periodic CI jobs for a MicroShift release

## Synopsis
```bash
/analyze-ci-for-release <release-version> [--limit N]
```

## Description
Analyzes all failed periodic jobs for a specific MicroShift release by leveraging existing tools and agents. This command orchestrates the analysis workflow by:

1. Fetching list of failed periodic jobs using `.claude/scripts/microshift-prow-jobs-for-release.sh`
2. Analyzing each job individually using the `openshift-ci-analysis` agent
3. Aggregating results and presenting a concise summary with common failure patterns

This approach reuses existing analysis capabilities rather than duplicating logic.

## Arguments
- `<release-version>` (required): OpenShift release version (e.g., 4.22, 4.21, 4.20)
- `--limit N` (optional): Limit analysis to first N jobs (useful for quick checks, default: all jobs)

## Implementation Steps

### Step 1: Validate Arguments and Fetch Failed Jobs

**Goal**: Get the list of failed periodic jobs for the release.

**Actions**:
1. Validate release version argument is provided
2. Execute `.claude/scripts/microshift-prow-jobs-for-release.sh <release>` to get all failed jobs
3. Filter output to only include periodic jobs (containing `-periodics-`)
4. Extract job URLs from the output
5. Apply limit if specified

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

### Step 2: Analyze Each Job Using openshift-ci-analysis agent

**Goal**: Get detailed root cause analysis for each failed job.

**Actions**:
1. For each job URL in the list:
   - Call the `openshift-ci-analysis` agent with the job URL **in parallel**
   - Capture the analysis result (failure reason, error summary)
   - Track common patterns across jobs
   - Store all intermediate analysis files in `/tmp`

2. Progress reporting:
   - Show "Analyzing job X/Y: <job-name>" for each job
   - Use the agent tool to invoke `openshift-ci-analysis` for each URL
   - **Run all job analyses in parallel** to maximize efficiency

**Example**:
```text
Analyzing 17 jobs in parallel...
Job 1/17: periodic-ci-openshift-microshift-release-4.22-periodics-e2e-aws-ovn-ocp-conformance
Job 2/17: periodic-ci-openshift-microshift-release-4.22-periodics-e2e-aws-tests-bootc-nightly
...
```

**Data Collection**:
For each job analysis, extract:
- Job name
- Job ID
- Failure type (build failure, test failure, infrastructure issue)
- Primary error/root cause
- Affected test scenarios (if applicable)

**File Storage**:
All intermediate analysis files are stored in `/tmp` with naming pattern:
- `/tmp/analyze-ci-release-<release>-job-<N>-<job-id>.txt`

### Step 3: Aggregate Results and Identify Patterns

**Goal**: Find common failure patterns across all jobs from parallel execution.

**Actions**:
1. Collect results from all parallel job analyses
   - Read individual job analysis files from `/tmp`
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
3. Generate summary report and save to `/tmp/analyze-ci-release-<release>-summary.<timestamp>.txt`
4. Display the summary to the user

**Report Structure**:

```text
═══════════════════════════════════════════════════════════════
MICROSHIFT 4.22 RELEASE - FAILED JOBS ANALYSIS
═══════════════════════════════════════════════════════════════

📊 OVERVIEW
  Total Failed Jobs: 17
  Analysis Date: 2026-03-14
  Report saved to: /tmp/analyze-ci-release-4.22-summary.txt

📋 FAILURE BREAKDOWN
  Build Failures:        0 jobs
  Test Failures:        15 jobs
  Infrastructure:        2 jobs

🔍 TOP ISSUES (by frequency)

1. OCP Conformance Test Failures (8 jobs)
   Severity: HIGH
   Pattern: Tests timeout or fail in conformance suite
   Affected Jobs:
   • periodic-ci-openshift-microshift-release-4.22-periodics-e2e-aws-ovn-ocp-conformance
   • periodic-ci-openshift-microshift-release-4.22-periodics-e2e-aws-ovn-ocp-conformance-serial
   • ... (6 more)

   Root Cause: [summarized from openshift-ci-analysis results]
   Next Steps: [recommended actions]

2. Bootc Image Test Failures (4 jobs)
   Severity: MEDIUM
   Pattern: Image build or deployment issues
   Affected Jobs:
   • periodic-ci-openshift-microshift-release-4.22-periodics-e2e-aws-tests-bootc-nightly
   • ... (3 more)

   Root Cause: [summarized]
   Next Steps: [recommended actions]

3. Infrastructure/Timeout Issues (2 jobs)
   Severity: LOW
   Pattern: Jobs timeout or fail to allocate resources
   Affected Jobs:
   • periodic-ci-openshift-microshift-release-4.22-periodics-rebase-on-nightlies
   • periodic-ci-openshift-microshift-release-4.22-periodics-update-versions-releases

   Root Cause: [summarized]

═══════════════════════════════════════════════════════════════

Individual job reports available in:
  /tmp/analyze-ci-release-4.22-job-*.txt
```

## Examples

### Example 1: Analyze All Failed Jobs

```bash
/analyze-ci-for-release 4.22
```

**Behavior**:
- Fetches all failed periodic jobs for 4.22
- Analyzes each job using openshift-ci-analysis agent
- Presents aggregated summary

### Example 2: Quick Analysis (First 5 Jobs)

```bash
/analyze-ci-for-release 4.22 --limit 5
```

**Behavior**:
- Analyzes only first 5 failed jobs
- Useful for quick health check
- Still provides pattern analysis

### Example 3: Different Release

```bash
/analyze-ci-for-release 4.21
```

**Behavior**:
- Analyzes 4.21 release jobs
- Same workflow as 4.22

## Performance Considerations

- **Execution Time**: Significantly reduced through parallel execution - typically 2-3 minutes for 15-20 jobs (depends on openshift-ci-analysis execution time)
- **Network Usage**: Moderate to high - all jobs analyzed in parallel fetch logs from GCS simultaneously
- **Parallelization**: All jobs are analyzed in parallel for maximum efficiency
- **Use --limit**: For quick checks, use --limit flag to analyze subset
- **File Storage**: All intermediate and report files are stored in `/tmp` directory

## Prerequisites

- `.claude/scripts/microshift-prow-jobs-for-release.sh` script must exist and be executable
- `openshift-ci-analysis` agent must be available
- Internet access to fetch job data from Prow/GCS
- Bash shell

## Error Handling

### No Failed Jobs
```text
No failed periodic jobs found for release 4.22
This is good news - all periodic jobs are passing! ✓
```

### Invalid Release Version
```text
Error: Invalid release version
Usage: /analyze-ci-for-release <release-version> [--limit N]
Example: /analyze-ci-for-release 4.22
```

### microshift-prow-jobs-for-release.sh Not Found
```text
Error: Could not find .claude/scripts/microshift-prow-jobs-for-release.sh
Please ensure you're in the microshift project directory.
```

## Related Skills

- **openshift-ci-analysis**: Detailed analysis of a single job (used internally)
- **analyze-ci-test-scenario**: Analyze specific test scenario results
- **analyze-microshift-start**: Analyze MicroShift startup performance
- **analyze-sos-report**: Investigate runtime issues from SOS reports

## Use Cases

### Daily CI Health Check
```bash
/analyze-ci-for-release 4.22 --limit 10
```
Quick morning check of CI status

### Pre-Release Verification
```bash
/analyze-ci-for-release 4.22
```
Complete analysis before cutting a release

### Root Cause Investigation
```bash
/analyze-ci-for-release 4.22
```
When multiple jobs fail, identify common issues

### Trend Analysis
Run periodically and compare summaries over time to identify regression patterns

## Notes

- This skill focuses on **periodic** jobs only (not presubmit/postsubmit)
- Analysis is read-only - no modifications to CI data
- Results are saved in files in /tmp directory with a timestamp
- Provide links to the jobs in the summary
- Only present a concise analysis summary for each job
- Pattern detection improves with more jobs analyzed (avoid limiting unless needed)
