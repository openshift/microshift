---
name: Generate HTML Report from CI Analysis
argument-hint: <release1,release2,...>
description: Generate an HTML report from analyze-ci analysis files in the WORKDIR
allowed-tools: Bash, Read, Write, Glob, Grep
---

# analyze-ci:create-report

## Synopsis
```bash
/analyze-ci:create-report <release1,release2,...>
```

## Description
Reads analysis output files from `${WORKDIR}/` (produced by `analyze-ci:release` and `analyze-ci:pull-requests`) and optionally bug mapping files (produced by `analyze-ci:create-bugs` dry-run) to generate a single consolidated HTML report with tabbed navigation. When bug mapping files are present, the report shows linked JIRA bugs per issue. This command does NOT run any CI analysis — it only reads existing files and generates HTML.

## Arguments
- `$ARGUMENTS` (required): Comma-separated list of release versions that were analyzed (e.g., `4.19,4.20,4.21,4.22`)

## Work Directory

Set once at the start and reference throughout:
```bash
WORKDIR=/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)
```

## Implementation Steps

### Step 1: Parse Arguments and Discover Files

**Actions**:
1. Split `$ARGUMENTS` by comma to get the list of release versions
2. Trim whitespace from each version
3. If no arguments provided, show usage and stop
4. Run `WORKDIR=/tmp/analyze-ci-claude-workdir.$(date +%y%m%d) && mkdir -p ${WORKDIR}`
5. For each release version, find the summary file:
   - `${WORKDIR}/analyze-ci-release-<version>-summary.*.txt`
6. Find the PR summary file:
   - `${WORKDIR}/analyze-ci-prs-summary.*.txt`
7. Find bug mapping files (produced by `analyze-ci:create-bugs` dry-run):
   - `${WORKDIR}/analyze-ci-bugs-*.txt`
   - These are optional — the report works without them, but when present, JIRA bug links are shown per issue
8. Report what was found:
   ```text
   Files discovered:
     Release 4.19: summary found, bug mapping found
     Release 4.20: summary found, bug mapping found
     PRs: summary found, bug mapping found (or "no bug mapping")
   ```

**Error Handling**:
- If `$ARGUMENTS` is empty, display: "Usage: /analyze-ci:create-report <release1,release2,...>" and stop
- If no files found at all, display error and stop

### Step 2: Read Summary Files

**Actions**:
1. Read each release summary file to extract the grouped analysis content (issue patterns, failure breakdown, affected jobs with URLs and dates, severity, root causes)
2. Read the PR summary file (if present) for grouped PR analysis
3. If a summary file is missing for a release, note "Analysis failed or produced no output"
4. If no PR summary file exists, note "No open rebase PRs or no failures found"

**Important**: Only summary files and bug mapping files are read — per-job files are NOT read. The summary files contain all information needed for the HTML report: severity levels, job names, Prow URLs, finish dates (in `[YYYY-MM-DD]` format after each job name), root cause descriptions, and affected scenarios. This keeps context usage minimal.

5. If bug mapping files were found, read and parse them:
   - Each file contains `--- BUG CANDIDATE ---` blocks with `ERROR_SIGNATURE`, `JIRA_DUPLICATES`, `JIRA_DUPLICATE_DETAILS`, `JIRA_REGRESSIONS`, and `JIRA_REGRESSION_DETAILS` fields
   - Build a lookup structure mapping error signatures to their JIRA bug information
   - The bug mapping file for a release is `analyze-ci-bugs-<version>.txt`; for rebase PRs it may be `analyze-ci-bugs-rebase-release-<version>.txt`
   - This data will be used in Step 3 to show JIRA links per issue in the HTML

### Step 3: Generate HTML Report

**Goal**: Create a single HTML file at `${WORKDIR}/microshift-ci-doctor-report.html` that consolidates all analyses with tabbed navigation.

**Actions**:
1. Determine `<timestamp>` as `YYYYMMDD-HHMMSS`
2. Generate the HTML report with the structure described below
3. **IMPORTANT**: Save using `cat <<'HTMLEOF' > ${WORKDIR}/microshift-ci-doctor-report.html` (heredoc via Bash tool), NOT the `Write` tool. This avoids permission prompts for the `/tmp` path.

**HTML Structure**:

The HTML file must be a self-contained, single-file document with embedded CSS and JS. Use the following structure:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>MicroShift CI Doctor Report - YYYY-MM-DD</title>
    <style>
        /* Clean, professional styling */
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; color: #333; }
        .container { max-width: 1200px; margin: 0 auto; }
        h1 { color: #1a1a2e; border-bottom: 3px solid #e94560; padding-bottom: 10px; }
        .release-section { background: white; border-radius: 8px; padding: 20px; margin: 20px 0; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .release-header { display: flex; justify-content: space-between; align-items: center; }
        .release-header h2 { color: #16213e; margin: 0; }
        .badge { padding: 4px 12px; border-radius: 12px; font-size: 0.85em; font-weight: 600; }
        .badge-ok { background: #d4edda; color: #155724; }
        .badge-issues { background: #fff3cd; color: #856404; }
        .badge-critical { background: #f8d7da; color: #721c24; }
        .badge-nodata { background: #e2e3e5; color: #383d41; }
        .summary-table { width: 100%; border-collapse: collapse; margin: 15px 0; }
        .summary-table th, .summary-table td { padding: 10px 14px; text-align: left; border-bottom: 1px solid #eee; }
        .summary-table th { background: #f8f9fa; font-weight: 600; color: #495057; }
        .summary-table tr:hover { background: #f8f9fa; }
        .severity-1, .severity-2 { color: #28a745; }
        .severity-3 { color: #856404; }
        .severity-4 { color: #dc3545; }
        .severity-5 { color: #721c24; font-weight: 700; }
        .infra-tag { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 0.8em; font-weight: 600; background: #cce5ff; color: #004085; }
        .root-cause { background: #fff8e1; border-left: 3px solid #ffc107; padding: 8px 12px; margin: 8px 0; font-size: 0.9em; }
        .status-pass { color: #28a745; }
        .status-fail { color: #dc3545; }
        .overview-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; margin: 20px 0; }
        .overview-card { background: white; border-radius: 8px; padding: 20px; text-align: center; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .overview-card .number { font-size: 2em; font-weight: 700; }
        .overview-card .label { color: #6c757d; font-size: 0.9em; }
        .content-block { white-space: pre-wrap; font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace; font-size: 0.85em; background: #f8f9fa; padding: 15px; border-radius: 4px; border: 1px solid #e9ecef; overflow-x: auto; margin: 10px 0; }
        .collapsible { cursor: pointer; user-select: none; }
        .collapsible::before { content: '\25B6  '; font-size: 0.8em; }
        .collapsible.active::before { content: '\25BC  '; }
        .collapsible .job-date { font-weight: 400; color: #6c757d; font-size: 0.85em; }
        .job-date { font-weight: 400; color: #6c757d; font-size: 0.85em; }
        .collapsible-content { display: none; }
        .collapsible-content.show { display: block; }
        .bug-links { margin: 8px 0; padding: 8px 12px; background: #f0f4ff; border-left: 3px solid #0366d6; font-size: 0.9em; }
        .bug-links .bug-tag { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 0.85em; font-weight: 600; margin: 2px 4px 2px 0; text-decoration: none; }
        .bug-tag-open { background: #fff3cd; color: #856404; border: 1px solid #ffc107; }
        .bug-tag-regression { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
        .bug-links-label { font-weight: 600; color: #495057; margin-right: 6px; }
        .no-bugs { color: #6c757d; font-style: italic; font-size: 0.85em; }
        .toc { background: white; border-radius: 8px; padding: 20px; margin: 20px 0; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .toc ul { list-style: none; padding-left: 0; }
        .toc li { padding: 5px 0; }
        .toc a { color: #0366d6; text-decoration: none; }
        .toc a:hover { text-decoration: underline; }
        .timestamp { color: #6c757d; font-size: 0.9em; }
        a { color: #0366d6; }

        /* Tab styling */
        .tab-bar { display: flex; gap: 0; margin: 20px 0 0 0; border-bottom: 2px solid #dee2e6; }
        .tab-btn { padding: 12px 24px; border: none; background: transparent; font-size: 1em; font-weight: 600; color: #6c757d; cursor: pointer; border-bottom: 3px solid transparent; margin-bottom: -2px; transition: color 0.2s, border-color 0.2s; }
        .tab-btn:hover { color: #333; }
        .tab-btn.active { color: #e94560; border-bottom-color: #e94560; }
        .tab-content { display: none; }
        .tab-content.active { display: block; }
    </style>
</head>
<body>
<div class="container">
    <h1>MicroShift CI Doctor Report</h1>
    <p class="timestamp">Generated: YYYY-MM-DD HH:MM:SS UTC</p>

    <!-- Overview cards: one per release + one for rebase PRs -->
    <div class="overview-grid">
        <!-- One card per release -->
        <div class="overview-card">
            <div class="number status-fail">N</div>
            <div class="label">Release X.YY Failed Jobs</div>
        </div>
        <!-- Card for rebase PRs -->
        <div class="overview-card">
            <div class="number status-fail">N</div>
            <div class="label">Rebase PRs Failed Jobs</div>
        </div>
    </div>

    <!-- Tab navigation -->
    <div class="tab-bar">
        <button class="tab-btn active" onclick="showTab(event, 'periodics')">Periodics</button>
        <button class="tab-btn" onclick="showTab(event, 'pull-requests')">Pull Requests</button>
    </div>

    <!-- Periodics tab content -->
    <div id="tab-periodics" class="tab-content active">

        <!-- Table of Contents -->
        <div class="toc">
            <h3>Releases Analyzed</h3>
            <ul>
                <li><a href="#release-X.YY">Release X.YY</a></li>
            </ul>
        </div>

        <!-- Per-release sections -->
        <div class="release-section" id="release-X.YY">
            <div class="release-header">
                <h2>Release X.YY</h2>
                <span class="badge badge-issues">N failed jobs</span>
            </div>

            <!-- One collapsible block per issue from the summary's TOP ISSUES section -->
            <div class="collapsible">1. SEVERITY: Issue Title (N jobs)</div>
            <div class="collapsible-content">
                <div class="root-cause"><strong>Root Cause:</strong> Root cause description from summary</div>
                <!-- Bug links from bug mapping file (if available) -->
                <!-- Match by comparing the issue title/error signature against ERROR_SIGNATURE values in the bug mapping -->
                <!-- Use fuzzy matching: if significant keywords from the issue title appear in a bug candidate's ERROR_SIGNATURE, consider it a match -->
                <div class="bug-links">
                    <span class="bug-links-label">JIRA Bugs:</span>
                    <!-- For each matching JIRA duplicate (open bugs): -->
                    <a class="bug-tag bug-tag-open" href="https://issues.redhat.com/browse/USHIFT-XXXXX" title="Bug summary text [Status]">USHIFT-XXXXX</a>
                    <!-- For each matching JIRA regression (closed bugs): -->
                    <a class="bug-tag bug-tag-regression" href="https://issues.redhat.com/browse/USHIFT-YYYYY" title="Bug summary text [Closed] regression">USHIFT-YYYYY ⟲</a>
                    <!-- If no matching bugs found in bug mapping: -->
                    <span class="no-bugs">No tracked bugs</span>
                </div>
                <p><strong>Affected Jobs:</strong></p>
                <ul>
                    <!-- Each job link includes finish date from the grep extraction -->
                    <li><span class="job-date">[YYYY-MM-DD]</span> <a href="JOB_URL">job-name</a></li>
                </ul>
                <p><strong>Next Steps:</strong> Next steps text from the summary</p>
            </div>
            <!-- Repeat for each issue -->
        </div>

        <!-- Repeat for each release -->
    </div>

    <!-- Pull Requests tab content -->
    <div id="tab-pull-requests" class="tab-content">

        <!-- Per-PR sections from analyze-ci:pull-requests --rebase -->
        <div class="release-section" id="pr-NNN">
            <div class="release-header">
                <h2>PR #NNN: title</h2>
                <span class="badge badge-issues">N failed jobs</span>
            </div>

            <!-- One entry per failed job, extracted from the PR summary -->
            <div class="collapsible"><span class="job-date">[YYYY-MM-DD]</span> 1. job-name - Root cause summary</div>
            <div class="collapsible-content">
                <p><strong>Job:</strong> <span class="job-date">[YYYY-MM-DD]</span> <a href="JOB_URL">job-name</a></p>
                <div class="root-cause"><strong>Root Cause:</strong> Root cause from PR summary</div>
                <!-- Bug links from bug mapping file (if available for this rebase PR) -->
                <!-- Match by comparing the job's root cause/error description against ERROR_SIGNATURE values in the bug mapping -->
                <div class="bug-links">
                    <span class="bug-links-label">JIRA Bugs:</span>
                    <a class="bug-tag bug-tag-open" href="https://issues.redhat.com/browse/USHIFT-XXXXX" title="Bug summary text [Status]">USHIFT-XXXXX</a>
                    <a class="bug-tag bug-tag-regression" href="https://issues.redhat.com/browse/USHIFT-YYYYY" title="Bug summary text [Closed] regression">USHIFT-YYYYY ⟲</a>
                    <span class="no-bugs">No tracked bugs</span>
                </div>
            </div>
        </div>

        <!-- If no rebase PRs found -->
        <div class="release-section">
            <p>No open rebase pull requests found.</p>
        </div>

    </div>

    <!-- Extra spacing to ensure tab content is fully visible in Prow Spyglass iframe -->
    <p>&nbsp;</p>
    <p>&nbsp;</p>
    <p>&nbsp;</p>
    <p>&nbsp;</p>
</div>

<script>
// Tab switching
function showTab(e, name) {
    document.querySelectorAll('.tab-content').forEach(function(el) {
        el.classList.remove('active');
    });
    document.querySelectorAll('.tab-btn').forEach(function(el) {
        el.classList.remove('active');
    });
    document.getElementById('tab-' + name).classList.add('active');
    e.target.classList.add('active');
}

// Collapsible sections
document.querySelectorAll('.collapsible').forEach(function(el) {
    el.addEventListener('click', function() {
        this.classList.toggle('active');
        var content = this.nextElementSibling;
        content.classList.toggle('show');
    });
});
</script>
</body>
</html>
```

**Content Guidelines**:
- Do NOT re-analyze or reinterpret the data — use summary file content as-is
- Convert the plain text summary reports into HTML-formatted content, preserving all information
- Ensure all Prow job URLs from the summaries remain clickable links in the HTML
- **Bug Correlation**: For each issue in the TOP ISSUES section (Periodics tab) and each failed job entry (Pull Requests tab), attempt to match it against the bug candidates from the corresponding bug mapping file (`analyze-ci-bugs-<release>.txt` for releases, `analyze-ci-bugs-rebase-release-<version>.txt` for rebase PRs). Match by comparing the issue title/description or job root cause against the `ERROR_SIGNATURE` in each `--- BUG CANDIDATE ---` block — use fuzzy keyword matching (shared distinctive terms like tool names, test IDs, error codes). When a match is found:
  - Show `JIRA_DUPLICATES` as clickable links with `bug-tag-open` styling (linking to `https://issues.redhat.com/browse/<KEY>`) with the summary from `JIRA_DUPLICATE_DETAILS` as the title attribute
  - Show `JIRA_REGRESSIONS` as clickable links with `bug-tag-regression` styling (with ⟲ suffix) with the summary from `JIRA_REGRESSION_DETAILS` as the title attribute
  - If no bug mapping file exists for a release, or no candidates match an issue, show `<span class="no-bugs">No tracked bugs</span>`
  - If the bug mapping file exists but a candidate has `JIRA_DUPLICATES: None` and `JIRA_REGRESSIONS: None`, show `<span class="no-bugs">No tracked bugs</span>`
- Use appropriate badge colors:
  - `badge-ok`: 0 failed jobs
  - `badge-issues`: 1+ failed jobs
  - `badge-critical`: 5+ failed jobs or CRITICAL severity issues present
  - `badge-nodata`: analysis failed or no data
- Each release section contains collapsible issue details: each issue from the summary's TOP ISSUES section gets a collapsible block showing root cause, affected jobs (linked to Prow), and next steps
- Each PR section shows the PR title, pass/fail counts, and collapsible failed job details with root causes — all extracted from the PR summary
- Severity labels from summaries (CRITICAL/HIGH/MEDIUM/LOW) should be color-coded using CSS classes
- The overview cards show the number of failed jobs per release and for rebase PRs at a glance
- The **Periodics** tab contains the per-release periodic job analyses
- The **Pull Requests** tab contains the rebase PR analyses grouped by PR
- The FAILURE BREAKDOWN section from each summary (Build/Test/Infrastructure counts) should be shown in the release header area

### Step 4: Report Completion

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

HTML report generated: ${WORKDIR}/microshift-ci-doctor-report.html
```

## Examples

### Example 1: Generate report after multi-release analysis
```bash
/analyze-ci:create-report 4.19,4.20,4.21,4.22
```

### Example 2: Regenerate report for a subset
```bash
/analyze-ci:create-report 4.21,4.22
```

## Prerequisites

- Analysis files must already exist in `${WORKDIR}/` (produced by `analyze-ci:release` and/or `analyze-ci:pull-requests`)
- Bash shell

## Related Skills

- **analyze-ci:doctor**: Orchestrator that runs analyses and then invokes this command
- **analyze-ci:release**: Per-release periodic job analysis (produces the input files)
- **analyze-ci:pull-requests**: PR job analysis (produces the input files)
- **analyze-ci:create-bugs**: Creates JIRA bugs from analysis output; also produces bug mapping files (`analyze-ci-bugs-<source>.txt`) consumed by this command to show JIRA links in the HTML report

## Notes
- This command is read-only — it only reads existing analysis files and generates HTML
- The HTML report is self-contained (no external CSS/JS dependencies)
- The HTML file can be opened in any browser for convenient examination
- If analysis files are missing for a release, it is noted in the report but does not block generation
- If no PR files exist, the Pull Requests tab shows "No open rebase pull requests found"
- This command can be re-run to regenerate the HTML without re-running the analyses
- If bug mapping files (`analyze-ci-bugs-*.txt`) exist in `${WORKDIR}/`, JIRA bug links are shown per issue in the HTML; if not, the report still generates correctly without bug links
