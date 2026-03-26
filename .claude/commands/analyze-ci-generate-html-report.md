---
name: Generate HTML Report from CI Analysis
argument-hint: <release1,release2,...>
description: Generate an HTML report from analyze-ci analysis files in /tmp/analyze-ci-claude-workdir.$(date +%y%m%d)
allowed-tools: Bash, Read, Write, Glob, Grep
---

# analyze-ci-generate-html-report

## Synopsis
```bash
/analyze-ci-generate-html-report <release1,release2,...>
```

## Description
Reads analysis output files from `/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/` (produced by `analyze-ci-for-release` and `analyze-ci-for-pull-requests`) and generates a single consolidated HTML report with tabbed navigation. This command does NOT run any CI analysis — it only reads existing files and generates HTML.

## Arguments
- `$ARGUMENTS` (required): Comma-separated list of release versions that were analyzed (e.g., `4.19,4.20,4.21,4.22`)

## Implementation Steps

### Step 1: Parse Arguments and Discover Files

**Actions**:
1. Split `$ARGUMENTS` by comma to get the list of release versions
2. Trim whitespace from each version
3. If no arguments provided, show usage and stop
4. Run `mkdir -p /tmp/analyze-ci-claude-workdir.$(date +%y%m%d)`
5. For each release version, find the summary file:
   - `/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/analyze-ci-release-<version>-summary.*.txt`
6. Find the PR summary file:
   - `/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/analyze-ci-prs-summary.*.txt`
7. Report what was found:
   ```text
   Files discovered:
     Release 4.19: summary found
     Release 4.20: summary found
     PRs: summary found
   ```

**Error Handling**:
- If `$ARGUMENTS` is empty, display: "Usage: /analyze-ci-generate-html-report <release1,release2,...>" and stop
- If no files found at all, display error and stop

### Step 2: Read Summary Files

**Actions**:
1. Read each release summary file to extract the grouped analysis content (issue patterns, failure breakdown, affected jobs with URLs and dates, severity, root causes)
2. Read the PR summary file (if present) for grouped PR analysis
3. If a summary file is missing for a release, note "Analysis failed or produced no output"
4. If no PR summary file exists, note "No open rebase PRs or no failures found"

**Important**: Only summary files are read — per-job files are NOT read. The summary files contain all information needed for the HTML report: severity levels, job names, Prow URLs, finish dates (in `[YYYY-MM-DD]` format after each job name), root cause descriptions, and affected scenarios. This keeps context usage minimal.

### Step 3: Generate HTML Report

**Goal**: Create a single HTML file at `/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/microshift-ci-release-manager-<timestamp>.html` that consolidates all analyses with tabbed navigation.

**Actions**:
1. Determine `<timestamp>` as `YYYYMMDD-HHMMSS`
2. Generate the HTML report with the structure described below
3. **IMPORTANT**: Save using `cat <<'HTMLEOF' > /tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/microshift-ci-release-manager-<timestamp>.html` (heredoc via Bash tool), NOT the `Write` tool. This avoids permission prompts for the `/tmp` path.

**HTML Structure**:

The HTML file must be a self-contained, single-file document with embedded CSS and JS. Use the following structure:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>MicroShift CI Release Manager Report - YYYY-MM-DD</title>
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
    <h1>MicroShift CI Release Manager Report</h1>
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

        <!-- Per-PR sections from analyze-ci-for-pull-requests --rebase -->
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
            </div>
        </div>

        <!-- If no rebase PRs found -->
        <div class="release-section">
            <p>No open rebase pull requests found.</p>
        </div>
    </div>
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

HTML report generated: /tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/microshift-ci-release-manager-20260315-143022.html
```

## Examples

### Example 1: Generate report after multi-release analysis
```bash
/analyze-ci-generate-html-report 4.19,4.20,4.21,4.22
```

### Example 2: Regenerate report for a subset
```bash
/analyze-ci-generate-html-report 4.21,4.22
```

## Prerequisites

- Analysis files must already exist in `/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)/` (produced by `analyze-ci-for-release` and/or `analyze-ci-for-pull-requests`)
- Bash shell

## Related Skills

- **analyze-ci-for-release-manager**: Orchestrator that runs analyses and then invokes this command
- **analyze-ci-for-release**: Per-release periodic job analysis (produces the input files)
- **analyze-ci-for-pull-requests**: PR job analysis (produces the input files)
- **analyze-ci-create-bugs**: Creates JIRA bugs from analysis output

## Notes
- This command is read-only — it only reads existing analysis files and generates HTML
- The HTML report is self-contained (no external CSS/JS dependencies)
- The HTML file can be opened in any browser for convenient examination
- If analysis files are missing for a release, it is noted in the report but does not block generation
- If no PR files exist, the Pull Requests tab shows "No open rebase pull requests found"
- This command can be re-run to regenerate the HTML without re-running the analyses
