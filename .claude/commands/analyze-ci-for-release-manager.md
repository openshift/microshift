---
name: Analyze CI for Release Manager
argument-hint: <release1,release2,...>
description: Analyze CI for multiple MicroShift releases and produce an HTML summary
allowed-tools: Skill, Bash, Read, Write, Glob, Grep, Agent
---

# analyze-ci-for-release-manager

## Synopsis
```bash
/analyze-ci-for-release-manager <release1,release2,...>
```

## Description
Accepts a comma-separated list of MicroShift release versions, runs the `analyze-ci-for-release` command for each release and the `analyze-ci-for-pull-requests --rebase` command for open rebase PRs, and produces a single HTML summary file consolidating all results. The HTML report uses tabs to separate Periodics (per-release) and Pull Requests sections.

## Arguments
- `$ARGUMENTS` (required): Comma-separated list of release versions (e.g., `4.19,4.20,4.21,4.22`)

## Implementation Steps

### Step 1: Parse and Validate Arguments

**Actions**:
1. Split `$ARGUMENTS` by comma to get a list of release versions
2. Trim whitespace from each version
3. Validate that at least one release version is provided
4. If no arguments provided, show usage and stop

**Error Handling**:
- If `$ARGUMENTS` is empty, display: "Usage: /analyze-ci-for-release-manager <release1,release2,...>" and stop

### Step 2: Analyze Each Release (Periodics)

**Actions**:
1. For each release version from the parsed list, launch the `analyze-ci-for-release` command as an **Agent** (using the `Agent` tool, NOT the `Skill` tool):
   ```text
   Agent: subagent_type=general_purpose, prompt="Run /analyze-ci-for-release <version>"
   ```
2. Launch all releases **in parallel** as separate agents — do NOT wait for one to finish before starting the next
3. After each agent completes, note the summary report file path it produced (typically `/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/analyze-ci-release-<version>-summary.*.txt`)
4. Wait until all the parallel agents are complete
5. Track which releases succeeded and which failed

**Progress Reporting**:
```text
Analyzing release X/Y: <version>
```

### Step 3: Analyze Rebase Pull Requests

**Actions**:
1. Launch the `analyze-ci-for-pull-requests` command as an **Agent** (using the `Agent` tool, NOT the `Skill` tool) with `--rebase` argument:
   ```text
   Agent: subagent_type=general_purpose, prompt="Run /analyze-ci-for-pull-requests --rebase"
   ```
2. This agent can be launched in parallel with the release agents in Step 2
3. After the agent completes, note the summary report file path (typically `/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/analyze-ci-prs-summary.*.txt`)
4. If no rebase PRs are found, note "No open rebase PRs" for the report

**Progress Reporting**:
1. Keep updating the background task list and completion status

### Step 4: Generate HTML Report via Dedicated Agent

**Goal**: Delegate HTML generation to a sub-agent with a fresh context.

**Why**: By this point the main context has accumulated agent launch/completion messages for all releases and PRs. The `analyze-ci-generate-html-report` command runs in a fresh agent context, reads only the summary files (not per-job files), and generates the HTML report.

**Actions**:
1. **IMPORTANT**: Wait until ALL analysis agents (releases + PRs) are confirmed complete
2. Launch the `analyze-ci-generate-html-report` command as an **Agent** (using the `Agent` tool, NOT the `Skill` tool):
   ```text
   Agent: subagent_type=general_purpose, prompt="Run /analyze-ci-generate-html-report <comma-separated-release-versions>"
   ```
3. Wait for the agent to complete

### Step 5: Report Completion

**Actions**:
1. After the HTML generation agent completes, relay its summary to the user
2. Display the path to the generated HTML file

**Example Output**:
```text
Summary:
  Periodics:
    Release 4.19: 3 failed periodic jobs
    Release 4.20: 7 failed periodic jobs
    Release 4.21: 0 failed periodic jobs
    Release 4.22: 12 failed periodic jobs
  Pull Requests:
    2 rebase PRs with 5 total failed jobs

HTML report generated: /tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/microshift-ci-release-manager-20260315-143022.html
```

## Examples

### Example 1: Analyze Multiple Releases
```bash
/analyze-ci-for-release-manager 4.19,4.20,4.21,4.22
```

### Example 2: Analyze Two Releases
```bash
/analyze-ci-for-release-manager 4.21,4.22
```

### Example 3: Single Release (still produces HTML)
```bash
/analyze-ci-for-release-manager 4.22
```

## Prerequisites

- `/analyze-ci-for-release` command must be available
- `/analyze-ci-for-pull-requests` command must be available
- `gcloud` CLI must be installed and authenticated for GCS access (used by analyze-ci-for-prow-job)
- `gh` CLI must be authenticated with access to openshift/microshift (used by analyze-ci-for-pull-requests)
- Internet access to fetch job data from Prow/GCS
- Bash shell

## Related Skills

- **analyze-ci-for-release**: Per-release periodic job analysis (used internally)
- **analyze-ci-for-pull-requests**: PR job analysis (used internally)
- **analyze-ci-for-prow-job**: Single job analysis (used by the above)
- **analyze-ci-generate-html-report**: HTML report generation from analysis files (used internally in Step 4)
- **analyze-ci-create-bugs**: Creates JIRA bugs from analysis output (supports both release and PR job files — run separately after this command)

## Notes
- Each release analysis launches `analyze-ci-for-release` as an **Agent** (not a Skill) - this command does NOT duplicate that logic
- Rebase PR analysis launches `analyze-ci-for-pull-requests --rebase` as an **Agent** (not a Skill)
- All agents (releases + PR analysis) are launched in parallel for maximum efficiency
- HTML generation is delegated to `analyze-ci-generate-html-report` running as a separate agent — it only reads summary files (not per-job files), keeping context usage minimal
- The HTML report is self-contained (no external CSS/JS dependencies)
- All intermediate files from `analyze-ci-for-release` and `analyze-ci-for-pull-requests` remain available in `/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)`
- The HTML file can be opened in any browser for convenient examination
- If a release analysis fails, it is noted in the report but does not block other releases
- If no rebase PRs are open, the Pull Requests tab shows "No open rebase pull requests found"
