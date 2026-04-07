---
name: Create JIRA Bugs from CI Analysis
argument-hint: <release|pr-NNNN> [--create]
description: Create JIRA bugs from analyze-ci failure reports (dry-run by default). Supports both release and PR job files.
allowed-tools: Bash, Read, Write, Glob, Grep, Agent, mcp__jira__jira_search, mcp__jira__jira_create_issue, mcp__jira__jira_get_issue, mcp__jira__jira_get_transitions, mcp__jira__jira_transition_issue, mcp__jira__jira_add_comment
---

# analyze-ci:create-bugs

## Synopsis
```bash
/analyze-ci:create-bugs <source> [--create]
```

## Description
Reads individual job analysis reports produced by `analyze-ci:doctor` and creates JIRA bugs in USHIFT for CI test failures. Operates in **dry-run mode by default** - it shows what bugs would be created without actually creating them. Use `--create` to perform actual issue creation.

This command does NOT re-analyze CI jobs. It consumes existing job analysis files from `${WORKDIR}/`.

## Arguments
- `$ARGUMENTS` (required): Source identifier, optionally followed by `--create`
  - `<source>` (required): One of the following:
    - **Release version** (e.g., `4.22`, `main`): Looks for files matching `analyze-ci-release-<release>-job-*.txt`
    - **PR number** (e.g., `pr-6396` or `pr6396`): Looks for files matching `analyze-ci-prs-job-*-pr<number>-*.txt`
    - **Rebase PR shorthand** (e.g., `rebase-release-4.22`): Resolves to the corresponding rebase PR by scanning existing `analyze-ci-prs-job-*` files for the matching release version in their content
  - `--create` (optional): Actually create JIRA issues. Without this flag, only a dry-run report is produced.

## Prerequisites

- Job analysis files must already exist in `${WORKDIR}/`:
  - For releases: `analyze-ci-release-<release>-job-*.txt` (produced by `/analyze-ci:doctor`)
  - For PRs: `analyze-ci-prs-job-*-pr<number>-*.txt` (produced by `/analyze-ci:doctor`)
- Each job file must contain a `--- STRUCTURED SUMMARY ---` block (see below)
- MCP Jira server must be configured and accessible
- User must have permissions to create issues in USHIFT

### STRUCTURED SUMMARY Block

Each job analysis file produced by `/analyze-ci:prow-job` must end with a machine-readable block:

```text
--- STRUCTURED SUMMARY ---
SEVERITY: <1-5>
STACK_LAYER: <AWS Infra|External Infrastructure|build phase|deploy phase|test setup phase|Test Configuration|test|teardown>
STEP_NAME: <the CI step where the error occurred>
ERROR_SIGNATURE: <concise, unique description of the root cause error>
RAW_ERROR: <verbatim primary error message from logs — used for deterministic grouping>
INFRASTRUCTURE_FAILURE: <true|false>
JOB_URL: <full prow job URL>
JOB_NAME: <full periodic job name>
RELEASE: <X.YY>
FINISHED: <job finish date in YYYY-MM-DD format>
--- END STRUCTURED SUMMARY ---
```

If a job file lacks this block, it is skipped with a warning.

## Work Directory

Set once at the start and reference throughout:
```bash
WORKDIR=/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)
```

## Implementation Steps

### Step 1: Prepare Bug Candidates (Deterministic Script)

**Actions**:
1. Parse `$ARGUMENTS` to extract `<source>` and detect `--create` flag
2. Determine mode: if `--create` is present, set `MODE=create`; otherwise `MODE=dry-run`
3. Run `WORKDIR=/tmp/analyze-ci-claude-workdir.$(date +%y%m%d) && mkdir -p ${WORKDIR}` using the `Bash` tool
4. Run the preparation script to parse job files, group by signature, and extract search keywords:
   ```bash
   python3 .claude/scripts/analyze-ci-search-bugs.py <source> --workdir ${WORKDIR}
   ```
5. The script writes `${WORKDIR}/analyze-ci-bug-candidates-<source>.json` containing:
   - Parsed and deduplicated bug candidates (grouped by ERROR_SIGNATURE similarity)
   - Pre-computed `keywords` (2-4 distinctive search terms per candidate)
   - Pre-computed `test_ids` (numeric IDs like `55394` for test case searches)
   - Full `analysis_text` for bug descriptions
   - Job lists with URLs and dates
6. Read the candidates JSON file for use in Step 2

**Error Handling**:
- No arguments: show usage and stop
- Script exits with error if no job files found — relay its error message to the user

### Step 2: Search Jira for Existing Bugs

For each bug candidate in the candidates JSON, run ALL of the following searches. The `keywords` and `test_ids` fields are pre-computed by the script — use them directly.

**Search A — Keyword search (multiple focused queries)**:
1. Use the pre-computed `keywords` array from the candidate (already filtered for stop words and ranked by specificity)
2. Run **2-3 separate searches in parallel**, each using 1-2 keywords from the array. Do NOT put all keywords into a single `text ~` query — Jira requires all terms to match, so queries with 3+ keywords are fragile and miss issues that use slightly different wording.
   ```python
   # Example: candidate.keywords = ["invalidclienttokenid", "cloudformation", "createstack", "aws-2"]
   # Search A1: most distinctive keyword
   mcp__jira__jira_search(jql='... AND issuetype = Bug AND text ~ "invalidclienttokenid" ...', limit=5)
   # Search A2: second keyword
   mcp__jira__jira_search(jql='... AND issuetype = Bug AND text ~ "cloudformation" ...', limit=5)
   ```
3. Merge and deduplicate results from all A-series queries before proceeding

**Search B — Test case ID search (MANDATORY when `test_ids` is non-empty)**:
Use the pre-computed `test_ids` array from the candidate. For EACH ID, run TWO separate searches:
```text
# Search B1: bare number
jql: ... AND issuetype = Bug AND text ~ "68256" AND status not in (Closed, Verified) ...

# Search B2: OCP-prefixed form (OpenShift Polarion convention)
jql: ... AND issuetype = Bug AND text ~ "OCP-68256" AND status not in (Closed, Verified) ...
```
**Why both forms are required**: Jira's text indexer treats `OCP-68256` as a single token, so `text ~ "68256"` will NOT match issues containing `OCP-68256`, and vice versa. Skipping either form WILL cause missed duplicates.

**After all searches**:
1. Merge and deduplicate results from all search queries (A, B1, B2)
2. If potential duplicates are found, fetch their details with `mcp__jira__jira_get_issue` to show summary and status

**Search C — Regression check (closed/verified issues)**:
After completing searches A and B, run an additional keyword search against closed/verified issues to detect potential regressions:
```python
mcp__jira__jira_search(
  jql='((project = OCPBUGS AND component = MicroShift) OR project = USHIFT) AND issuetype = Bug AND text ~ "<keywords>" AND status in (Closed, Verified) ORDER BY updated DESC',
  limit=5
)
```
If results are found, fetch their details with `mcp__jira__jira_get_issue` and flag them as **"Potential regression of closed bug"** — distinct from open duplicates. These should be shown to the user but do NOT block creation; they serve as a warning that a previously fixed issue may have resurfaced.

**Note**: Run searches in parallel where possible.

### Step 3: Present Bug Candidates to User

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
       Unique bug candidates: N
       Candidates with potential duplicates: N
       Candidates ready to file: N

     To create these bugs, run:
       /analyze-ci:create-bugs <source> --create
     ```
   - Do NOT prompt for any actions. Do NOT create any issues. Do NOT proceed to Steps 4/4a (create/reopen). Continue to Step 5 and Step 6.

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
   - **create**: Proceed to Step 4
   - **skip**: Skip this candidate, move to next
   - **link-to-existing**: Validate the key by calling `mcp__jira__jira_get_issue(issue_key=<JIRA-KEY>)`. If the issue exists, record the key and move to next. If the call fails or returns not-found, show an error (e.g., `"JIRA key <JIRA-KEY> not found — check for typos"`) and re-prompt with the same `Action?` choices.
   - **reopen**: Validate the provided JIRA key before proceeding. Call `mcp__jira__jira_get_issue(issue_key=<JIRA-KEY>)` to confirm the issue exists, then verify that the key matches one of the candidate's closed regressions found in Search C, that the issue status is Closed or Verified, and that the issue type is Bug. If validation fails (key not found, not in the candidate's closed regression list, not in Closed/Verified state, or not a Bug), show an error (e.g., `"JIRA key <JIRA-KEY> not eligible for reopen — must be a Bug closed regression"`) and re-prompt with the same `Action?` choices. If validation passes, proceed to Step 4a.

### Step 4: Create Bug via MCP (create mode only)

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

   **Source:** Auto-generated by /analyze-ci:create-bugs from CI analysis output.
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

### Step 4a: Reopen Closed Bug as Regression (create mode only)

**Precondition**: The JIRA issue must be a Bug in Closed or Verified state (validated in Step 3). If the issue type is not Bug, do not proceed — show an error and re-prompt.

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

   Reopened automatically by /analyze-ci:create-bugs.
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

### Step 5: Write Machine-Readable Bug Mapping File

**Actions**:
After processing all bug candidates (Steps 2-4a) and regardless of mode (dry-run or create), write a machine-readable bug mapping file that `analyze-ci-create-report.py` can consume to display JIRA bug links in the HTML report. The file content is based on the Jira search results from Step 2 — it is not affected by whether bugs were created or reopened in Steps 4/4a.

1. Save to `${WORKDIR}/analyze-ci-bugs-<source>.json` (overwrite if exists)
2. Use this JSON format:

```json
{
  "source": "<source>",
  "date": "YYYY-MM-DD",
  "candidates": [
    {
      "error_signature": "<error_signature>",
      "severity": <N>,
      "step_name": "<step_name>",
      "affected_jobs": <count>,
      "duplicates": [
        {"key": "<JIRA-KEY>", "summary": "<summary>", "status": "<status>"}
      ],
      "regressions": [
        {"key": "<JIRA-KEY>", "summary": "<summary>", "status": "<status>"}
      ]
    }
  ]
}
```

3. **IMPORTANT**: This file must be written in BOTH dry-run and create modes. The file enables `analyze-ci-create-report.py` to show linked bugs per issue in the HTML report.
4. Use empty arrays `[]` for `duplicates` and `regressions` when none are found.
5. Save using a Bash heredoc with `jq` or `python3 -c` to ensure valid JSON, or use the Write tool.

### Step 6: Generate Results Report

**Actions**:
1. Save report to `${WORKDIR}/analyze-ci-create-bugs-<source>.<timestamp>.txt`
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
  None (all failures included)

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
  /analyze-ci:create-bugs <source> --create

Report saved: ${WORKDIR}/analyze-ci-create-bugs-<source>.<timestamp>.txt
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

Report saved: ${WORKDIR}/analyze-ci-create-bugs-<source>.<timestamp>.txt
═══════════════════════════════════════════════════════════════
```

## Examples

### Example 1: Dry-Run for a Release (Default)
```bash
/analyze-ci:create-bugs 4.22
```
Shows what bugs would be created from release 4.22 analysis without creating anything.

### Example 2: Create Bugs for a Release
```bash
/analyze-ci:create-bugs 4.22 --create
```
Interactively creates bugs from release 4.22 analysis.

### Example 3: Dry-Run for a PR
```bash
/analyze-ci:create-bugs pr-6396
```
Shows what bugs would be created from PR #6396 analysis.

### Example 4: Create Bugs for a Rebase PR
```bash
/analyze-ci:create-bugs rebase-release-4.22 --create
```
Resolves the rebase PR for release 4.22, then interactively creates bugs.

### Example 5: No Job Files Found
```bash
/analyze-ci:create-bugs 4.19
```
```text
Error: No job analysis files found at ${WORKDIR}/analyze-ci-release-4.19-job-*.txt

Run the analysis first:
  /analyze-ci:doctor 4.19
```

### Example 6: No PR Job Files Found
```bash
/analyze-ci:create-bugs pr-9999
```
```text
Error: No job analysis files found at ${WORKDIR}/analyze-ci-prs-job-*-pr9999-*.txt

Run the analysis first:
  /analyze-ci:doctor <release>
```

## Notes

- This command does NOT run CI analysis — it only consumes existing analysis files from `${WORKDIR}`
- Supports two file naming patterns:
  - Release jobs: `analyze-ci-release-<release>-job-*.txt` (from `/analyze-ci:doctor`)
  - PR jobs: `analyze-ci-prs-job-*-pr<number>-*.txt` (from `/analyze-ci:doctor`)
- Dry-run is the default to prevent accidental bug creation
- The `--create` flag triggers interactive mode where each candidate requires user confirmation
- All failures are included without filtering — no entries are skipped based on severity, infrastructure status, or stack layer
- Bugs are created in USHIFT with component "MicroShift"; duplicate search covers both USHIFT and OCPBUGS
- All created bugs are labeled with `microshift-ci-ai-generated` for tracking
- Security level is set to "Red Hat Employee" on all created issues
- The STRUCTURED SUMMARY block in job files is required — this is a contract with `/analyze-ci:prow-job`
- In addition to the text report, a machine-readable bug mapping file (`analyze-ci-bugs-<source>.json`) is written in both dry-run and create modes — this file is consumed by `analyze-ci-create-report.py` to show JIRA bug links in the HTML report

## Related Skills

- **analyze-ci:doctor**: Produces job analysis files consumed by this command
- **analyze-ci:prow-job**: Command that produces individual job reports with STRUCTURED SUMMARY
- **jira:create-bug**: Interactive bug creation (not used here — we call MCP directly)
