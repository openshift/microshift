---
name: Analyze CI for Release Manager
argument-hint: <release1,release2,...>
description: Analyze CI for multiple MicroShift releases and produce an HTML summary
allowed-tools: Bash, Read, Write, Glob, Grep, Agent
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
   Agent: subagent_type=general-purpose, prompt="Run /analyze-ci-for-release <version>"
   ```
2. Launch all releases **in parallel** as separate agents — do NOT wait for one to finish before starting the next

**Progress Reporting**:
```text
Analyzing release X/Y: <version>
```

### Step 3: Analyze Rebase Pull Requests

**Actions**:
1. Launch the `analyze-ci-for-pull-requests` command as an **Agent** (using the `Agent` tool, NOT the `Skill` tool) with `--rebase` argument:
   ```text
   Agent: subagent_type=general-purpose, prompt="Run /analyze-ci-for-pull-requests --rebase"
   ```
2. Launch this agent **in parallel** with the release agents in Step 2 — do NOT wait for Step 2 agents to finish first

### Step 4: Collect All Results

**Actions**:
1. **IMPORTANT**: Wait until ALL agents from Steps 2 and 3 are confirmed complete
2. Track which releases succeeded and which failed
3. If no rebase PRs were found, note "No open rebase PRs" for the report
4. Gather per-job files:
   - Per-job files: `/tmp/analyze-ci-claude-workdir/analyze-ci-release-<version>-job-*.txt` for each version
   - PR per-job files: `/tmp/analyze-ci-claude-workdir/analyze-ci-prs-job-*.txt`
5. If no per-job files exist for a release, note it as "Analysis failed or produced no output"
6. If no PR job files exist, note "No open rebase PRs or no failures found"
7. Do NOT read the files manually — the Python script in Step 5 will read and parse them directly

### Step 5: Generate HTML Summary Report

**Goal**: Create a single HTML file at `/tmp/analyze-ci-claude-workdir/microshift-ci-release-manager-<timestamp>.html` that consolidates all analyses with tabbed navigation. The report must be **concise at the top level** with expandable details per job.

**Actions**:
1. Write a Python script to `/tmp/analyze-ci-claude-workdir/gen_html.py` and execute it. Using Python (not a heredoc) is required because it handles HTML escaping and URL-to-link conversion properly.
2. Save output to `/tmp/analyze-ci-claude-workdir/microshift-ci-release-manager-<timestamp>.html` where `<timestamp>` is `YYYYMMDD-HHMMSS`
3. Display the file path to the user in the end, AFTER the summary

**HTML Structure**:

The HTML file must be a self-contained, single-file document with embedded CSS and JS. The layout is:

1. **Header**: Title + timestamp
2. **Tab bar**: "Periodics" and "Pull Requests" tabs
3. **Table of Contents**: Horizontal list of release links with failure counts
4. **Per-release sections**: Each contains a **concise table** of failed jobs (NOT large text blocks)
5. **Expandable detail rows**: Clicking a table row reveals the full per-job analysis text

**Per-Release Table Structure**:

Each release section contains a table with these columns:
| Column | Width | Content |
|--------|-------|---------|
| **Job** | 30% | Shortened job name (linked to Prow URL). Strip `periodic-ci-openshift-microshift-release-X.YY-periodics-` prefix. |
| **Date** | 8% | Job finish date from the `FINISHED:` line in the per-job file |
| **Sev** | 5% | Severity badge (LOW/MED/HIGH/CRIT) from `SEVERITY:` field in per-job structured summary |
| **Type** | 6% | "Infra" or "Test" badge from `INFRASTRUCTURE_FAILURE:` field in per-job structured summary |
| **Summary** | 51% | One-line summary from `ERROR_SIGNATURE:` field in per-job structured summary |

**Expandable Detail Rows**:

Each table row is clickable. Clicking it toggles a hidden detail row below containing:
- The full per-job analysis text (everything BEFORE `--- STRUCTURED SUMMARY ---`)
- Rendered as a monospace `pre-wrap` block
- All URLs converted to clickable `<a>` links
- HTML special characters properly escaped

**Python Script Requirements**:

The Python script (`gen_html.py`) must:
1. Use `html.escape()` to escape all text content inserted into HTML
2. Use `re.sub()` to convert bare `https://...` URLs to clickable `<a href="...">` links
3. Parse each per-job file to extract structured fields (`SEVERITY`, `INFRASTRUCTURE_FAILURE`, `ERROR_SIGNATURE`, `JOB_NAME`, `JOB_URL`, `FINISHED`)
4. Split per-job content at `--- STRUCTURED SUMMARY ---` to separate the human-readable analysis from the structured fields
5. Shorten job names by stripping the `periodic-ci-openshift-microshift-release-*-periodics-` prefix
6. Generate severity badges with appropriate colors:
   - SEVERITY 2 → LOW (yellow)
   - SEVERITY 3 → MED (yellow)
   - SEVERITY 4 → HIGH (red)
   - SEVERITY 5 → CRIT (red)
7. Generate type badges: `INFRASTRUCTURE_FAILURE: true` → "Infra" (blue), otherwise → "Test" (gray)

**Release Badge Colors**:
- `badge-ok`: 0 failed jobs
- `badge-issues`: 1-4 failed jobs
- `badge-critical`: 5+ failed jobs

**Pull Requests Tab**:

The Pull Requests tab uses the same table layout. For each PR with failures:
- Section header: "PR #NNN: title" with a badge
- Table of failed jobs with the same columns
- Expandable detail rows

If all PR jobs pass, show a simple table row with PR link, job count, and "All passing" status.
If no rebase PRs are open, show "No open rebase pull requests found."

**Reference CSS** (embed in the generated HTML):

```css
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;margin:0;padding:20px;background:#f5f5f5;color:#333}
.container{max-width:1400px;margin:0 auto}
h1{color:#1a1a2e;border-bottom:3px solid #e94560;padding-bottom:10px}
.section{background:#fff;border-radius:8px;padding:20px;margin:20px 0;box-shadow:0 2px 4px rgba(0,0,0,.1)}
.section-header{display:flex;justify-content:space-between;align-items:center;margin-bottom:12px}
.section-header h2{color:#16213e;margin:0}
.badge{padding:4px 12px;border-radius:12px;font-size:.85em;font-weight:600;white-space:nowrap}
.badge-ok{background:#d4edda;color:#155724}
.badge-issues{background:#fff3cd;color:#856404}
.badge-critical{background:#f8d7da;color:#721c24}
table{width:100%;border-collapse:collapse}
th{text-align:left;padding:8px 10px;background:#f8f9fa;border-bottom:2px solid #dee2e6;font-size:.85em;color:#495057}
td{padding:8px 10px;border-bottom:1px solid #eee;font-size:.85em;vertical-align:top}
tr.job-row{cursor:pointer}
tr.job-row:hover{background:#f0f4ff}
tr.job-row td:first-child::before{content:'\\25B6 ';font-size:.7em;color:#999}
tr.job-row.active td:first-child::before{content:'\\25BC ';color:#333}
.detail-row{display:none}
.detail-row.show{display:table-row}
.detail-cell{padding:0 10px 12px 28px}
.detail-block{white-space:pre-wrap;font-family:Consolas,'Liberation Mono',Menlo,monospace;font-size:.82em;background:#f8f9fa;padding:14px;border-radius:4px;border:1px solid #e9ecef;overflow-x:auto}
a{color:#0366d6}
.timestamp{color:#6c757d;font-size:.9em}
.tab-bar{display:flex;gap:0;margin:20px 0 0;border-bottom:2px solid #dee2e6}
.tab-btn{padding:12px 24px;border:none;background:0 0;font-size:1em;font-weight:600;color:#6c757d;cursor:pointer;border-bottom:3px solid transparent;margin-bottom:-2px}
.tab-btn:hover{color:#333}
.tab-btn.active{color:#e94560;border-bottom-color:#e94560}
.tab-content{display:none}
.tab-content.active{display:block}
.toc{background:#fff;border-radius:8px;padding:16px 20px;margin:20px 0;box-shadow:0 2px 4px rgba(0,0,0,.1)}
.toc ul{list-style:none;padding:0;margin:0;display:flex;flex-wrap:wrap;gap:8px 24px}
.toc a{color:#0366d6;text-decoration:none;font-weight:500}
.fail{color:#dc3545}
.pass{color:#28a745}
```

**Reference JS** (embed in the generated HTML):

```javascript
function showTab(e, name) {
    document.querySelectorAll('.tab-content').forEach(el => el.classList.remove('active'));
    document.querySelectorAll('.tab-btn').forEach(el => el.classList.remove('active'));
    document.getElementById('tab-' + name).classList.add('active');
    e.target.classList.add('active');
}
function toggleDetail(id) {
    var row = document.getElementById(id);
    row.classList.toggle('show');
    row.previousElementSibling.classList.toggle('active');
}
```

### Step 6: Report Completion

**Actions**:
1. Provide a brief text summary listing each release and its failed job count, plus rebase PR status
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

HTML report generated: /tmp/analyze-ci-claude-workdir/microshift-ci-release-manager-20260315-143022.html
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

## Notes
- Each release analysis launches `analyze-ci-for-release` as an **Agent** (not a Skill) - this command does NOT duplicate that logic
- Rebase PR analysis launches `analyze-ci-for-pull-requests --rebase` as an **Agent** (not a Skill)
- All agents (releases + PR analysis) are launched in parallel for maximum efficiency
- The HTML report is self-contained (no external CSS/JS dependencies)
- All intermediate files from `analyze-ci-for-release` and `analyze-ci-for-pull-requests` remain available in `/tmp/analyze-ci-claude-workdir`
- The HTML file can be opened in any browser for convenient examination
- If a release analysis fails, it is noted in the report but does not block other releases
- If no rebase PRs are open, the Pull Requests tab shows "No open rebase pull requests found"
