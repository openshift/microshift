---
name: Analyze CI for a Release
argument-hint: <release-version>
description: Analyze CI for a MicroShift release using analyze-ci:prow-job command and produce a summary
allowed-tools: Skill, Bash, Read, Write, Glob, Grep, Agent
---

# analyze-ci:release

## Synopsis
```bash
/analyze-ci:release <release-version>
```

## Description
Analyzes all failed periodic jobs for a specific MicroShift release by leveraging existing tools and agents. This command orchestrates the analysis workflow by:

1. Fetching list of failed periodic jobs using `.claude/scripts/microshift-prow-jobs-for-release.sh`
2. Analyzing each job individually using the `/analyze-ci:prow-job` command
3. Aggregating results and presenting a concise summary with common failure patterns

This approach reuses existing analysis capabilities rather than duplicating logic.

## Arguments
- `<release-version>` (required): OpenShift release version (e.g., 4.22, 4.21, 4.20)

## Work Directory

Set once at the start and reference throughout:
```bash
WORKDIR=/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)
```

## Implementation Steps

### Step 1: Validate Arguments and Fetch Failed Jobs

**Goal**: Get the list of failed periodic jobs for the release.

**Actions**:
1. Validate release version argument is provided
2. Execute `.claude/scripts/microshift-prow-jobs-for-release.sh <release>` to get all failed jobs as a JSON array
3. Filter the JSON to only include periodic jobs (where `.type == "periodic"`)
4. Extract job URLs and build IDs from the JSON

**Example Command**:
```bash
bash .claude/scripts/microshift-prow-jobs-for-release.sh 4.22 | jq '[.[] | select(.type == "periodic")]'
```

**JSON output format** (array of objects):
```json
[
  {
    "job": "periodic-ci-openshift-microshift-release-4.22-periodics-e2e-aws-tests-bootc-nightly",
    "type": "periodic",
    "status": "failure",
    "finished": "2026-04-01T06:38:10Z",
    "duration": "4h37m20s",
    "url": "https://prow.ci.openshift.org/view/gs/...",
    "build_id": "2039161134610649088"
  }
]
```

**Expected Output**:
```text
Found 17 failed periodic jobs for release 4.22
```

**Error Handling**:
- If no release version provided, show usage and exit
- If no periodic jobs found (empty JSON array after filtering), report "No failed periodic jobs found for release X.XX" and exit successfully
- If microshift-prow-jobs-for-release.sh fails, report error and exit

### Step 2: Download All Job Artifacts

**Goal**: Download artifacts for all failed jobs in parallel before analysis.

**Actions**:
1. Run `WORKDIR=/tmp/analyze-ci-claude-workdir.$(date +%y%m%d) && mkdir -p ${WORKDIR}` using the `Bash` tool
2. Pipe the filtered JSON from Step 1 into the download script and save the enriched output to a prescribed file:
   ```bash
   bash .claude/scripts/microshift-prow-jobs-for-release.sh <release> | \
       jq '[.[] | select(.type == "periodic")]' | \
       WORKDIR=${WORKDIR} bash .claude/scripts/analyze-ci-download-jobs.sh \
       > ${WORKDIR}/analyze-ci-release-<release>-jobs.json
   ```
3. The script downloads all artifacts in parallel to `${WORKDIR}/artifacts/<build_id>/` and outputs enriched JSON with `artifacts_dir` fields added
4. The output is saved to `${WORKDIR}/analyze-ci-release-<release>-jobs.json` — do NOT use any other filename

**Error Handling**:
- If some downloads fail, note the failures but proceed with successfully downloaded jobs

### Step 3: Analyze Each Job Using /analyze-ci:prow-job

**Goal**: Get detailed root cause analysis for each failed job using pre-downloaded artifacts.

**IMPORTANT - Mandatory Per-Job Agent Requirements**:
Each job MUST be analyzed by launching a **separate Agent** (using the `Agent` tool, NOT the `Skill` tool) for each job. Do NOT analyze jobs inline or combine multiple jobs into a single agent. Each agent MUST:
1. Run `/analyze-ci:prow-job <artifacts_dir>` using the local artifacts directory (not by reimplementing the analysis)
2. After the analysis completes, write the analysis report to an **exact file path** (see File Storage below)
3. The file MUST contain the full report including the `--- STRUCTURED SUMMARY ---` block

**Actions**:
1. Read the enriched JSON from `${WORKDIR}/analyze-ci-release-<release>-jobs.json`
2. For each job in the JSON, launch a separate **Agent** with this exact prompt template, using the `.artifacts_dir` field as `<ARTIFACTS_DIR>` and the `.build_id` field as `<JOB_ID>`:
   ```
   Agent: subagent_type=general_purpose, prompt="Analyze this Prow job and save the report:
   1. Run /analyze-ci:prow-job <ARTIFACTS_DIR>
   2. After the analysis completes, save the FULL report output (including the --- STRUCTURED SUMMARY --- block) to:
      ${WORKDIR}/analyze-ci-release-<RELEASE>-job-<N>-<JOB_ID>.txt
      Use the Write tool to save the file. The file must contain the complete analysis report."
   ```
   Replace `<ARTIFACTS_DIR>`, `<RELEASE>`, `<N>` (1-based job index), and `<JOB_ID>` with actual values from the JSON.
2. Launch **ALL** job agents in parallel using `run_in_background: true`
3. Wait for all agents to complete before proceeding to Step 4

**Progress Reporting**:
```text
Analyzing N jobs in parallel...
Job 1/N: <job-name> (background)
Job 2/N: <job-name> (background)
...
```

**File Storage**:
All per-job report files MUST be saved at this exact path:
- `${WORKDIR}/analyze-ci-release-<release>-job-<N>-<job-id>.txt`

Where:
- `<release>` is the release version (e.g., `4.22`, `main`)
- `<N>` is the 1-based job index
- `<job-id>` is the numeric Prow job ID from the URL

**Verification**: After all agents complete, verify that per-job files exist:
```bash
ls ${WORKDIR}/analyze-ci-release-<release>-job-*.txt
```
If any files are missing, note the gap in the summary report but do NOT re-run the analysis.

### Step 4: Aggregate Results into Summary

**Goal**: Group per-job results by failure pattern and produce a JSON summary.

**Actions**:
1. Run the aggregation script:
   ```bash
   python3 .claude/scripts/analyze-ci-aggregate.py --release <release> --workdir ${WORKDIR}
   ```
2. The script deterministically:
   - Parses `STRUCTURED SUMMARY` blocks and prose (`Error:`, `Suggested Remediation:`) from each per-job file
   - Groups jobs by `ERROR_SIGNATURE` similarity (token overlap >= 50%)
   - Assigns severity: `CRITICAL` (max severity >= 4), `HIGH` (3+ jobs), `MEDIUM` (2 jobs), `LOW` (1 job)
   - Classifies breakdown by `STACK_LAYER` → build/test/infrastructure
   - Writes `${WORKDIR}/analyze-ci-release-<release>-summary.json`
3. Display the script output to the user
```

## Examples

### Example 1: Analyze All Failed Jobs

```bash
/analyze-ci:release 4.22
```

**Behavior**:
- Fetches all failed periodic jobs for 4.22
- Analyzes each job using /analyze-ci:prow-job command
- Presents aggregated summary

### Example 2: Different Release

```bash
/analyze-ci:release 4.21
```

**Behavior**:
- Analyzes 4.21 release jobs
- Same workflow as 4.22

## Performance Considerations

- **Execution Time**: Significantly reduced through parallel execution - typically 2-3 minutes for 15-20 jobs (depends on analyze-ci:prow-job execution time)
- **Network Usage**: Moderate to high - all jobs analyzed in parallel fetch logs from GCS simultaneously
- **Parallelization**: All jobs are analyzed in parallel for maximum efficiency
- **File Storage**: All intermediate and report files are stored in `${WORKDIR}` directory

## Prerequisites

- `.claude/scripts/microshift-prow-jobs-for-release.sh` script must exist and be executable
- `/analyze-ci:prow-job` command must be available
- `gcloud` CLI must be installed and authenticated for GCS access (used by analyze-ci:prow-job)
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
Usage: /analyze-ci:release <release-version>
Example: /analyze-ci:release 4.22
```

### microshift-prow-jobs-for-release.sh Not Found
```text
Error: Could not find .claude/scripts/microshift-prow-jobs-for-release.sh
Please ensure you're in the microshift project directory.
```

## Related Skills

- **analyze-ci:prow-job**: Detailed analysis of a single job (used internally)
- **analyze-ci:create-bugs**: Creates JIRA bugs from this command's output
- **analyze-ci:doctor**: Multi-release orchestrator that delegates to this command
- **analyze-ci:test-scenario**: Analyze specific test scenario results
- **analyze-sos-report**: Investigate runtime issues from SOS reports

## Use Cases

### Pre-Release Verification
```bash
/analyze-ci:release 4.22
```
Complete analysis before cutting a release

### Trend Analysis
Run periodically and compare summaries over time to identify regression patterns

## Notes

- This skill focuses on **periodic** jobs only (not presubmit/postsubmit)
- Analysis is read-only - no modifications to CI data
- Results are saved in files in ${WORKDIR} directory with a timestamp
- Provide links to the jobs in the summary
- Only present a concise analysis summary for each job
- Pattern detection improves with more jobs analyzed
