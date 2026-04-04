---
name: Analyze CI Doctor
argument-hint: <release1,release2,...>
description: Analyze CI for multiple MicroShift releases and produce an HTML summary
allowed-tools: Skill, Bash, Read, Write, Glob, Grep, Agent
---

# analyze-ci:doctor

## Synopsis
```bash
/analyze-ci:doctor <release1,release2,...>
```

## Description
Accepts a comma-separated list of MicroShift release versions, runs the `analyze-ci:release` command for each release and the `analyze-ci:pull-requests --rebase` command for open rebase PRs, and produces a single HTML summary file consolidating all results. The HTML report uses tabs to separate Periodics (per-release) and Pull Requests sections.

## Arguments
- `$ARGUMENTS` (required): Comma-separated list of release versions (e.g., `4.19,4.20,4.21,4.22`)

## Work Directory

Set once at the start and reference throughout:
```bash
WORKDIR=/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)
```

## Implementation Steps

### Step 1: Parse and Validate Arguments

**Actions**:
1. Run `WORKDIR=/tmp/analyze-ci-claude-workdir.$(date +%y%m%d) && mkdir -p ${WORKDIR}` using the `Bash` tool
2. Split `$ARGUMENTS` by comma to get a list of release versions
3. Trim whitespace from each version
4. Validate that at least one release version is provided
5. If no arguments provided, show usage and stop

**Error Handling**:
- If `$ARGUMENTS` is empty, display: "Usage: /analyze-ci:doctor <release1,release2,...>" and stop

### Step 2: Analyze Each Release (Periodics)

**Actions**:
1. For each release version from the parsed list, launch the `analyze-ci:release` command as an **Agent** (using the `Agent` tool, NOT the `Skill` tool):
   ```text
   Agent: subagent_type=general_purpose, prompt="Run /analyze-ci:release <version>"
   ```
2. Launch all releases **in parallel** as separate agents — do NOT wait for one to finish before starting the next
3. After each agent completes, note the summary report file path it produced (typically `${WORKDIR}/analyze-ci-release-<version>-summary.json`)
4. Wait until all the parallel agents are complete
5. Track which releases succeeded and which failed

**Progress Reporting**:
```text
Analyzing release X/Y: <version>
```

### Step 3: Analyze Rebase Pull Requests

**Actions**:
1. Launch the `analyze-ci:pull-requests` command as an **Agent** (using the `Agent` tool, NOT the `Skill` tool) with `--rebase` argument:
   ```text
   Agent: subagent_type=general_purpose, prompt="Run /analyze-ci:pull-requests --rebase"
   ```
2. This agent can be launched in parallel with the release agents in Step 2
3. After the agent completes, note the summary report file path (typically `${WORKDIR}/analyze-ci-prs-summary.json`)
4. If no rebase PRs are found, note "No open rebase PRs" for the report

**Progress Reporting**:
1. Keep updating the background task list and completion status

### Step 4: Run Bug Correlation (Dry-Run)

**Goal**: For each release and for rebase PRs, run `analyze-ci:create-bugs` in dry-run mode to identify existing JIRA bugs that correlate with detected failures. This produces machine-readable bug mapping files that `analyze-ci:create-report` will use to show linked bugs in the HTML report.

**Why**: The `analyze-ci:create-bugs` command searches Jira for existing bugs matching each failure's error signature. Running it in dry-run mode before HTML generation allows the report to display known JIRA bugs next to each problem, helping users quickly see which issues are already tracked.

**Actions**:
1. **IMPORTANT**: Wait until ALL analysis agents (releases + PRs) are confirmed complete before starting this step
2. For each release version, launch `analyze-ci:create-bugs` in dry-run mode as an **Agent** (using the `Agent` tool, NOT the `Skill` tool):
   ```text
   Agent: subagent_type=general_purpose, prompt="Run /analyze-ci:create-bugs <version>"
   ```
3. If rebase PR analysis produced job files, also launch `analyze-ci:create-bugs` for rebase PRs. Check the PR summary file to identify rebase PR source identifiers (e.g., `rebase-release-4.22`) and launch an agent for each:
   ```text
   Agent: subagent_type=general_purpose, prompt="Run /analyze-ci:create-bugs rebase-release-<version>"
   ```
4. Launch all create-bugs agents **in parallel** — do NOT wait for one to finish before starting the next
5. Wait until all create-bugs agents complete
6. Each agent produces a bug mapping file at `${WORKDIR}/analyze-ci-bugs-<source>.json` that the create-report command will consume

**Progress Reporting**:
```text
Running bug correlation for release X.YY...
Running bug correlation for rebase PRs...
```

**Error Handling**:
- If create-bugs fails for a release (e.g., no job files found), note the failure but do not block other releases or the HTML report generation
- The HTML report will simply omit bug links for releases where no bug mapping file exists

### Step 5: Generate HTML Report via Dedicated Agent

**Goal**: Delegate HTML generation to a sub-agent with a fresh context.

**Why**: By this point the main context has accumulated agent launch/completion messages for all releases and PRs. The `analyze-ci:create-report` command runs in a fresh agent context, reads only the summary files and bug mapping files (not per-job files), and generates the HTML report.

**Actions**:
1. **IMPORTANT**: Wait until ALL create-bugs agents from Step 4 are confirmed complete
2. Launch the `analyze-ci:create-report` command as an **Agent** (using the `Agent` tool, NOT the `Skill` tool):
   ```text
   Agent: subagent_type=general_purpose, prompt="Run /analyze-ci:create-report <comma-separated-release-versions>"
   ```
3. Wait for the agent to complete

### Step 6: Report Completion

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

HTML report generated: ${WORKDIR}/microshift-ci-doctor-report.html
```

## Examples

### Example 1: Analyze Multiple Releases
```bash
/analyze-ci:doctor 4.19,4.20,4.21,4.22
```

### Example 2: Analyze Two Releases
```bash
/analyze-ci:doctor 4.21,4.22
```

### Example 3: Single Release (still produces HTML)
```bash
/analyze-ci:doctor 4.22
```

## Prerequisites

- `/analyze-ci:release` command must be available
- `/analyze-ci:pull-requests` command must be available
- `gcloud` CLI must be installed and authenticated for GCS access (used by analyze-ci:prow-job)
- `gh` CLI must be authenticated with access to openshift/microshift (used by analyze-ci:pull-requests)
- Internet access to fetch job data from Prow/GCS
- Bash shell

## Related Skills

- **analyze-ci:release**: Per-release periodic job analysis (used internally)
- **analyze-ci:pull-requests**: PR job analysis (used internally)
- **analyze-ci:prow-job**: Single job analysis (used by the above)
- **analyze-ci:create-report**: HTML report generation from analysis files (used internally in Step 5)
- **analyze-ci:create-bugs**: Creates JIRA bugs from analysis output (run automatically in dry-run mode during Step 4 for bug correlation; can also be run separately with `--create` after this command)

## Notes
- Each release analysis launches `analyze-ci:release` as an **Agent** (not a Skill) - this command does NOT duplicate that logic
- Rebase PR analysis launches `analyze-ci:pull-requests --rebase` as an **Agent** (not a Skill)
- All agents (releases + PR analysis) are launched in parallel for maximum efficiency
- Bug correlation runs `analyze-ci:create-bugs` in dry-run mode for each release to produce bug mapping files — these are consumed by `analyze-ci:create-report` to show JIRA links in the HTML
- HTML generation is delegated to `analyze-ci:create-report` running as a separate agent — it reads summary files and bug mapping files (not per-job files), keeping context usage minimal
- The HTML report is self-contained (no external CSS/JS dependencies)
- All intermediate files from `analyze-ci:release` and `analyze-ci:pull-requests` remain available in `${WORKDIR}`
- The HTML file can be opened in any browser for convenient examination
- If a release analysis fails, it is noted in the report but does not block other releases
- If no rebase PRs are open, the Pull Requests tab shows "No open rebase pull requests found"
