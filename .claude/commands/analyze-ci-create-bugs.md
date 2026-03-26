---
name: Create JIRA Bugs from CI Analysis
argument-hint: <release|pr-NNNN> [--create]
description: Create JIRA bugs from analyze-ci failure reports (dry-run by default). Supports both release and PR job files.
allowed-tools: Bash, Read, Write, Glob, Grep, Agent, mcp__jira__jira_search, mcp__jira__jira_create_issue, mcp__jira__jira_get_issue, mcp__jira__jira_get_transitions, mcp__jira__jira_transition_issue, mcp__jira__jira_add_comment
---

# analyze-ci-create-bugs

## Synopsis
```bash
/analyze-ci-create-bugs <source> [--create]
```

## Description
Reads individual job analysis reports produced by `analyze-ci-for-release` or `analyze-ci-for-pull-requests` and creates JIRA bugs in USHIFT for legitimate test failures. Operates in **dry-run mode by default** - it shows what bugs would be created without actually creating them. Use `--create` to perform actual issue creation.

This command does NOT re-analyze CI jobs. It consumes existing job analysis files from `/tmp/analyze-ci-claude-workdir/`.

## Arguments
- `$ARGUMENTS` (required): Source identifier, optionally followed by `--create`
  - `<source>` (required): One of the following:
    - **Release version** (e.g., `4.22`, `main`): Looks for files matching `analyze-ci-release-<release>-job-*.txt`
    - **PR number** (e.g., `pr-6396` or `pr6396`): Looks for files matching `analyze-ci-prs-job-*-pr<number>-*.txt`
    - **Rebase PR shorthand** (e.g., `rebase-release-4.22`): Resolves to the corresponding rebase PR by scanning existing `analyze-ci-prs-job-*` files for the matching release version in their content
  - `--create` (optional): Actually create JIRA issues. Without this flag, only a dry-run report is produced.

## Prerequisites

- Job analysis files must already exist in `/tmp/analyze-ci-claude-workdir/`:
  - For releases: `analyze-ci-release-<release>-job-*.txt` (produced by `/analyze-ci-for-release`)
  - For PRs: `analyze-ci-prs-job-*-pr<number>-*.txt` (produced by `/analyze-ci-for-pull-requests`)
- Each job file must contain a `--- STRUCTURED SUMMARY ---` block (see below)
- MCP Jira server must be configured and accessible
- User must have permissions to create issues in USHIFT

### STRUCTURED SUMMARY Block

Each job analysis file produced by `/analyze-ci-for-prow-job` must end with a machine-readable block:

```text
--- STRUCTURED SUMMARY ---
SEVERITY: <1-5>
STACK_LAYER: <AWS Infra|External Infrastructure|build phase|deploy phase|test setup phase|Test Configuration|test|teardown>
STEP_NAME: <the CI step where the error occurred>
ERROR_SIGNATURE: <concise, unique description of the root cause error>
INFRASTRUCTURE_FAILURE: <true|false>
JOB_URL: <full prow job URL>
JOB_NAME: <full periodic job name>
RELEASE: <X.YY>
FINISHED: <job finish date in YYYY-MM-DD format>
--- END STRUCTURED SUMMARY ---
```

If a job file lacks this block, it is skipped with a warning.

## Implementation Steps

### Step 1: Parse Arguments and Locate Job Files

**Actions**:
1. Parse `$ARGUMENTS` to extract `<source>` and detect `--create` flag
2. Determine mode: if `--create` is present, set `MODE=create`; otherwise `MODE=dry-run`
3. Run `mkdir -p /tmp/analyze-ci-claude-workdir` using the `Bash` tool
4. Determine source type and locate job files:

   **a) Release version** (e.g., `4.22`, `main`, `4.19`):
   - Glob for: `/tmp/analyze-ci-claude-workdir/analyze-ci-release-<release>-job-*.txt`
   - Set `SOURCE_TYPE=release`, `SOURCE_LABEL="release <release>"`

   **b) PR number** (e.g., `pr-6396`, `pr6396`):
   - Extract the numeric PR number (strip `pr-` or `pr` prefix)
   - Glob for: `/tmp/analyze-ci-claude-workdir/analyze-ci-prs-job-*-pr<number>-*.txt`
   - Set `SOURCE_TYPE=pr`, `SOURCE_LABEL="PR #<number>"`

   **c) Rebase PR shorthand** (e.g., `rebase-release-4.22`):
   - Extract the release version from the shorthand (e.g., `4.22`)
   - Glob for all PR job files: `/tmp/analyze-ci-claude-workdir/analyze-ci-prs-job-*.txt`
   - Read each file's STRUCTURED SUMMARY and find files where JOB_NAME contains the release version (e.g., `release-4.22` or `main`) OR where the JOB_URL contains the release branch
   - Alternatively, check the PR summary file (`analyze-ci-prs-summary.*.txt`) to find the PR number for the given rebase release
   - Set `SOURCE_TYPE=pr`, `SOURCE_LABEL="rebase PR for <release> (PR #<number>)"`

5. If no files found, report error and stop

**Error Handling**:
- No arguments: show usage and stop
- No job files found: suggest running the appropriate analysis command first:
  - For releases: `/analyze-ci-for-release <release>`
  - For PRs: `/analyze-ci-for-pull-requests`

### Step 2: Parse STRUCTURED SUMMARY from Each Job File

**Actions**:
1. For each job file, extract the `--- STRUCTURED SUMMARY ---` block
2. Parse key-value pairs: SEVERITY, STACK_LAYER, STEP_NAME, ERROR_SIGNATURE, INFRASTRUCTURE_FAILURE, JOB_URL, JOB_NAME, RELEASE, FINISHED
3. Also capture the full file content for use in the bug description (the error context and analysis above the structured block)
4. If a file lacks the structured block, log a warning and skip it

**Parsing approach**: Use grep/sed in Bash to extract the block between `--- STRUCTURED SUMMARY ---` and `--- END STRUCTURED SUMMARY ---`, then parse each `KEY: value` line.

**Data structure per job**:
```text
{
  severity: number,
  stack_layer: string,
  step_name: string,
  error_signature: string,
  infrastructure_failure: boolean,
  job_url: string,
  job_name: string,
  release: string,
  finished: string,         # job finish date in YYYY-MM-DD format
  analysis_text: string,    # full file content for bug description
  source_file: string       # path to the job file
}
```

### Step 3: Filter Out Non-Bug-Worthy Failures

**Actions**:
1. Remove entries where `SEVERITY <= 2` (minor/flaky issues)
2. Remove entries where `INFRASTRUCTURE_FAILURE=true` **AND** the failure is transient/external (not a CI configuration bug). Use `STACK_LAYER` to distinguish:
   - **Filter out** (transient infrastructure — out of our control):
     - `STACK_LAYER` contains `AWS Infra` (AWS quota, VM creation, networking)
     - `STACK_LAYER` contains `External Infrastructure` (container registry outages, third-party services)
     - `STACK_LAYER` is `build phase` (release image import timeouts, registry 404s)
   - **Keep as bug-worthy** (CI configuration issues — need code fixes):
     - `STACK_LAYER` is `test setup phase`, `Test Configuration`, or similar (missing test files, wrong directory mappings, broken scenario selection logic)
     - Any `INFRASTRUCTURE_FAILURE=true` entry whose analysis text indicates the fix requires a code change in the repository (e.g., adding missing files, updating directory mappings, fixing CI scripts)
3. Log each filtered-out entry with reason. For kept CI configuration failures, log them as `"CI CONFIG (kept)"` to distinguish from product test failures.

**Rationale**: Not all "infrastructure failures" are transient. A missing test scenario directory or broken CI script mapping is a configuration bug that requires a code fix — these should result in JIRA bugs, not be silently filtered out. Only truly external/transient failures (AWS outages, registry issues, image import timeouts) should be excluded.

**Output**: Filtered list of bug-worthy failures with count of filtered entries reported to user.

### Step 4: Deduplicate by ERROR_SIGNATURE

**Actions**:
1. Group remaining entries by `ERROR_SIGNATURE` similarity
   - Exact matches are grouped together
   - Near-matches (same error but slightly different wording) should also be grouped — use your judgment to identify when two signatures describe the same root cause
2. For each group, create a "bug candidate":
   - **Representative signature**: the ERROR_SIGNATURE that best describes the group
   - **Affected jobs**: list of all JOB_NAME + JOB_URL in the group
   - **Max severity**: highest SEVERITY in the group
   - **Step names**: unique STEP_NAME values in the group
   - **Analysis text**: from the highest-severity job in the group

**Output**: List of deduplicated bug candidates.

### Step 5: Search Jira for Existing Bugs

For each bug candidate, run ALL of the following searches. Each search is MANDATORY — do not skip any.

**Search A — Keyword search**:
1. Extract 2-4 distinctive keywords from the error signature (avoid generic words like "error", "failed", "test")
2. Run:
   ```python
   mcp__jira__jira_search(
     jql='((project = OCPBUGS AND component = MicroShift) OR project = USHIFT) AND text ~ "<keywords>" AND status not in (Closed, Verified) ORDER BY created DESC',
     limit=5
   )
   ```

**Search B — Test case ID search (MANDATORY when IDs are present)**:
Extract ALL numeric IDs from the error signature that could be test case references (typically 4-6 digit numbers like `68256`). For EACH numeric ID found, run TWO separate searches:
```text
# Search B1: bare number
jql: ... AND text ~ "68256" AND status not in (Closed, Verified) ...

# Search B2: OCP-prefixed form (OpenShift Polarion convention)
jql: ... AND text ~ "OCP-68256" AND status not in (Closed, Verified) ...
```
**Why both forms are required**: Jira's text indexer treats `OCP-68256` as a single token, so `text ~ "68256"` will NOT match issues containing `OCP-68256`, and vice versa. Skipping either form WILL cause missed duplicates.

**After all searches**:
1. Merge and deduplicate results from all search queries (A, B1, B2)
2. If potential duplicates are found, fetch their details with `mcp__jira__jira_get_issue` to show summary and status

**Search C — Regression check (closed/verified issues)**:
After completing searches A and B, run an additional keyword search against closed/verified issues to detect potential regressions:
```python
mcp__jira__jira_search(
  jql='((project = OCPBUGS AND component = MicroShift) OR project = USHIFT) AND text ~ "<keywords>" AND status in (Closed, Verified) ORDER BY updated DESC',
  limit=5
)
```
If results are found, fetch their details with `mcp__jira__jira_get_issue` and flag them as **"Potential regression of closed bug"** — distinct from open duplicates. These should be shown to the user but do NOT block creation; they serve as a warning that a previously fixed issue may have resurfaced.

**Note**: Run searches in parallel where possible.

### Step 6: Present Bug Candidates to User

**Actions**:
1. Display a numbered list of all bug candidates with:
   - Summary (derived from error signature)
   - Severity and affected job count
   - Step name(s) where failure occurred
   - List of affected job URLs
   - Potential duplicate JIRAs found (if any), with key, summary, and status
   - Mode indicator: `[DRY-RUN]` or `[WILL CREATE]`

2. **In dry-run mode** (`--create` NOT specified):
   - Display all candidates with `[DRY-RUN]` prefix
   - After listing all candidates, show a summary:
     ```text
     DRY-RUN SUMMARY
       Source: <SOURCE_LABEL>
       Total job files parsed: N
       Filtered out (infra/low severity): N
       Unique bug candidates: N
       Candidates with potential duplicates: N
       Candidates ready to file: N

     To create these bugs, run:
       /analyze-ci-create-bugs <source> --create
     ```
   - Do NOT prompt for any actions. Do NOT create any issues. Stop here.

3. **In create mode** (`--create` specified):
   - For each candidate, prompt the user:
     ```text
     Bug Candidate N/M:
       Summary: "<derived summary>"
       Severity: X (affects Y jobs)
       Step: <step_name>
       Jobs:
         - <job_url_1>
         - <job_url_2>
       Potential Duplicates (open):
         - USHIFT-XXXXX: "<summary>" [Status] (or OCPBUGS-YYYYY)
         (or "None found")
       Potential Regressions (closed):
         - USHIFT-YYYYY (or OCPBUGS-YYYYY): "<summary>" [Status] potential regression
         (or "None found")

     # ACTION_PROMPT_WITH_REOPEN (use when closed regressions exist):
     Action? [c]reate / [s]kip / [l]ink-to-existing <JIRA-KEY> / [r]eopen <JIRA-KEY>:

     # ACTION_PROMPT_NO_REOPEN (use when no closed regressions exist):
     Action? [c]reate / [s]kip / [l]ink-to-existing <JIRA-KEY>:
     ```
   - Select the prompt template based on whether closed regressions were found for the candidate: use `ACTION_PROMPT_WITH_REOPEN` when the candidate has closed regressions from Search C, and `ACTION_PROMPT_NO_REOPEN` otherwise.
   - **create**: Proceed to Step 7
   - **skip**: Skip this candidate, move to next
   - **link-to-existing**: Validate the key by calling `mcp__jira__jira_get_issue(issue_key=<JIRA-KEY>)`. If the issue exists, record the key and move to next. If the call fails or returns not-found, show an error (e.g., `"JIRA key <JIRA-KEY> not found — check for typos"`) and re-prompt with the same `Action?` choices.
   - **reopen**: Validate the provided JIRA key before proceeding. Call `mcp__jira__jira_get_issue(issue_key=<JIRA-KEY>)` to confirm the issue exists, then verify that the key matches one of the candidate's closed regressions found in Search C, that the issue status is Closed or Verified, and that the issue type is Bug. If validation fails (key not found, not in the candidate's closed regression list, not in Closed/Verified state, or not a Bug), show an error (e.g., `"JIRA key <JIRA-KEY> not eligible for reopen — must be a Bug closed regression"`) and re-prompt with the same `Action?` choices. If validation passes, proceed to Step 7a.

### Step 7: Create Bug via MCP (create mode only)

**Actions**:
For each candidate where user chose "create":

1. **Construct the bug summary**:
   - Format: `"MicroShift CI: <error_signature>"` (truncate to 100 chars if needed)

2. **Construct the bug description** using **Markdown** format (the MCP Jira tool accepts Markdown and automatically converts it to Jira wiki markup — do NOT write Jira wiki markup directly):
   ```text
   ## Description of problem

   CI job failures detected for MicroShift <SOURCE_LABEL>.

   <concise description derived from the error signature and analysis text>

   ## Version-Release number of selected component (if applicable)

   <release from STRUCTURED SUMMARY>

   ## How reproducible

   Always (fails consistently in CI)

   ## Steps to Reproduce

   1. Run the CI job(s) listed below
   2. Observe failure in step: <step_name>

   ## Actual results

   ````
   <error details extracted from the analysis text — the specific error message and relevant log context>
   ````

   ## Expected results

   CI job should pass successfully.

   ## Additional info

   **Stack Layer:** <stack_layer>
   **CI Step:** <step_name>
   **Error Severity:** <severity>/5
   **Number of affected jobs:** <count>
   **Last observed:** <finished date from STRUCTURED SUMMARY, YYYY-MM-DD>

   **Affected Jobs:**
   <for each job in the group>
   - [<job_name>](<job_url>)
   </for each>

   **Source:** Auto-generated by /analyze-ci-create-bugs from CI analysis output.
   ```

3. **Create the issue**:
   ```python
   mcp__jira__jira_create_issue(
       project_key="USHIFT",
       summary="MicroShift CI: <error_signature>",
       issue_type="Bug",
       description="<constructed description>",
       components="MicroShift",
       additional_fields={
           "labels": ["microshift-ci-ai-generated"],
           "security": {"name": "Red Hat Employee"}
       }
   )
   ```

4. **Record the result**: Store the created issue key for the final report.

**Error Handling**:
- If MCP call fails, report error, ask user if they want to retry or skip
- Do NOT retry automatically

### Step 7a: Reopen Closed Bug as Regression (create mode only)

**Precondition**: The JIRA issue must be a Bug in Closed or Verified state (validated in Step 6). If the issue type is not Bug, do not proceed — show an error and re-prompt.

**Actions**:
For each candidate where user chose "reopen":

1. **Get available transitions** for the closed issue:
   ```python
   mcp__jira__jira_get_transitions(issue_key="<JIRA-KEY>")
   ```

2. **Find the reopen transition**: Look for a transition whose name is exactly "To Do", "New", or "Backlog" (case-insensitive). If no suitable transition is found, report the error and ask the user whether to create a new bug instead or skip.

3. **Construct a regression comment** describing the new occurrences:
   ```text
   ## Regression: issue has resurfaced

   This issue was previously closed but the same failure has been detected again in CI.

   **Error Signature:** <error_signature>
   **Error Severity:** <severity>/5
   **Number of affected jobs:** <count>
   **Last observed:** <finished date>

   **Affected Jobs:**
   - [<job_name>](<job_url>)
   ...

   Reopened automatically by /analyze-ci-create-bugs.
   ```

4. **Transition the issue** to reopen it:
   ```python
   mcp__jira__jira_transition_issue(
       issue_key="<JIRA-KEY>",
       transition_id="<reopen_transition_id>",
       comment="<regression comment>"
   )
   ```

5. If the transition call does not support inline comments, add the comment separately:
   ```python
   mcp__jira__jira_add_comment(
       issue_key="<JIRA-KEY>",
       body="<regression comment>"
   )
   ```

6. **Record the result**: Store the reopened issue key for the final report.

**Error Handling**:
- If no reopen-like transition is available, report available transitions to user and ask whether to create a new bug or skip
- If the transition fails, report error and ask user if they want to retry, create a new bug instead, or skip
- Do NOT retry automatically

### Step 8: Generate Results Report

**Actions**:
1. Save report to `/tmp/analyze-ci-claude-workdir/analyze-ci-create-bugs-<source>.<timestamp>.txt`
2. Display summary to user:

**Dry-run report format**:
```text
═══════════════════════════════════════════════════════════════
ANALYZE-CI CREATE BUGS - DRY-RUN REPORT
Source: <SOURCE_LABEL>
Date: YYYY-MM-DD
═══════════════════════════════════════════════════════════════

PARSING
  Job files found: N
  Successfully parsed: N
  Skipped (no structured summary): N

FILTERING
  Infrastructure failures removed: N
  Low severity (<=2) removed: N
  Bug-worthy failures: N

DEDUPLICATION
  Unique bug candidates: N

CANDIDATES

  1. MicroShift CI: <error_signature>
     Severity: X | Jobs: Y | Step: <step_name>
     Potential Duplicates: USHIFT-XXXXX, OCPBUGS-YYYYY (or "None")
     Potential Regressions: USHIFT-YYYYY (or OCPBUGS-YYYYY) [Closed] (or "None")

  2. MicroShift CI: <error_signature>
     ...

To create these bugs, run:
  /analyze-ci-create-bugs <source> --create

Report saved: /tmp/analyze-ci-claude-workdir/analyze-ci-create-bugs-<source>.<timestamp>.txt
═══════════════════════════════════════════════════════════════
```

**Create mode report format**:
```text
═══════════════════════════════════════════════════════════════
ANALYZE-CI CREATE BUGS - CREATION REPORT
Source: <SOURCE_LABEL>
Date: YYYY-MM-DD
═══════════════════════════════════════════════════════════════

RESULTS

  1. USHIFT-12345 (CREATED)
     MicroShift CI: <error_signature>
     URL: https://redhat.atlassian.net/browse/USHIFT-12345

  2. SKIPPED
     MicroShift CI: <error_signature>
     Reason: User skipped

  3. USHIFT-99999 (LINKED TO EXISTING)
     MicroShift CI: <error_signature>
     Reason: Duplicate of existing issue

  4. USHIFT-88888 (REOPENED)
     MicroShift CI: <error_signature>
     URL: https://redhat.atlassian.net/browse/USHIFT-88888
     Reason: Regression of previously closed bug

SUMMARY
  Created: N
  Skipped: N
  Linked to existing: N
  Reopened: N
  Failed: N

Report saved: /tmp/analyze-ci-claude-workdir/analyze-ci-create-bugs-<source>.<timestamp>.txt
═══════════════════════════════════════════════════════════════
```

## Examples

### Example 1: Dry-Run for a Release (Default)
```bash
/analyze-ci-create-bugs 4.22
```
Shows what bugs would be created from release 4.22 analysis without creating anything.

### Example 2: Create Bugs for a Release
```bash
/analyze-ci-create-bugs 4.22 --create
```
Interactively creates bugs from release 4.22 analysis.

### Example 3: Dry-Run for a PR
```bash
/analyze-ci-create-bugs pr-6396
```
Shows what bugs would be created from PR #6396 analysis.

### Example 4: Create Bugs for a Rebase PR
```bash
/analyze-ci-create-bugs rebase-release-4.22 --create
```
Resolves the rebase PR for release 4.22, then interactively creates bugs.

### Example 5: No Job Files Found
```bash
/analyze-ci-create-bugs 4.19
```
```text
Error: No job analysis files found at /tmp/analyze-ci-claude-workdir/analyze-ci-release-4.19-job-*.txt

Run the analysis first:
  /analyze-ci-for-release 4.19
```

### Example 6: No PR Job Files Found
```bash
/analyze-ci-create-bugs pr-9999
```
```text
Error: No job analysis files found at /tmp/analyze-ci-claude-workdir/analyze-ci-prs-job-*-pr9999-*.txt

Run the analysis first:
  /analyze-ci-for-pull-requests
```

## Notes

- This command does NOT run CI analysis — it only consumes existing analysis files from `/tmp/analyze-ci-claude-workdir`
- Supports two file naming patterns:
  - Release jobs: `analyze-ci-release-<release>-job-*.txt` (from `/analyze-ci-for-release`)
  - PR jobs: `analyze-ci-prs-job-*-pr<number>-*.txt` (from `/analyze-ci-for-pull-requests`)
- Dry-run is the default to prevent accidental bug creation
- The `--create` flag triggers interactive mode where each candidate requires user confirmation
- Transient infrastructure failures (AWS, VM creation, quota, registry outages) are automatically filtered out
- CI configuration failures (missing test files, broken directory mappings) are kept as bug-worthy even if marked as INFRASTRUCTURE_FAILURE=true
- Bugs are created in USHIFT with component "MicroShift"; duplicate search covers both USHIFT and OCPBUGS
- All created bugs are labeled with `microshift-ci-ai-generated` for tracking
- Security level is set to "Red Hat Employee" on all created issues
- The STRUCTURED SUMMARY block in job files is required — this is a contract with `/analyze-ci-for-prow-job`

## Related Skills

- **analyze-ci-for-release**: Produces release job analysis files consumed by this command
- **analyze-ci-for-pull-requests**: Produces PR job analysis files consumed by this command
- **analyze-ci-for-release-manager**: Multi-release orchestrator
- **analyze-ci-for-prow-job**: Command that produces individual job reports with STRUCTURED SUMMARY
- **jira:create-bug**: Interactive bug creation (not used here — we call MCP directly)
