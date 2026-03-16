---
name: Analyze CI for Release Manager
argument-hint: <release1,release2,...>
description: Analyze CI for multiple MicroShift releases and produce an HTML summary
allowed-tools: Skill, Bash, Read, Write, Glob, Grep, Agent
---

# analyze-ci-for-release-manager

## Synopsis
```
/analyze-ci-for-release-manager <release1,release2,...>
```

## Description
Accepts a comma-separated list of MicroShift release versions, runs the `analyze-ci-for-release` skill for each release and the `analyze-ci-for-pull-requests --rebase` skill for open rebase PRs, and produces a single HTML summary file consolidating all results. The HTML report uses tabs to separate Periodics (per-release) and Pull Requests sections.

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
1. For each release version from the parsed list, invoke the `analyze-ci-for-release` skill:
   ```
   Skill: analyze-ci-for-release, args: "<version>"
   ```
2. Run releases **sequentially** (each skill invocation is a full analysis)
3. After each skill completes, note the summary report file path it produced (typically `/tmp/analyze-ci-release-<version>-summary.*.txt`)
4. Track which releases succeeded and which failed

**Progress Reporting**:
```
Analyzing release X/Y: <version>
```

### Step 3: Analyze Rebase Pull Requests

**Actions**:
1. Invoke the `analyze-ci-for-pull-requests` skill with `--rebase` argument:
   ```
   Skill: analyze-ci-for-pull-requests, args: "--rebase"
   ```
2. After the skill completes, note the summary report file path (typically `/tmp/analyze-ci-prs-summary.*.txt`)
3. If no rebase PRs are found, note "No open rebase PRs" for the report

**Progress Reporting**:
```
Analyzing rebase pull requests...
```

### Step 4: Collect All Results

**Actions**:
1. After all analyses complete, gather all summary files:
   - Periodics: `/tmp/analyze-ci-release-<version>-summary.*.txt` for each version
   - Pull Requests: `/tmp/analyze-ci-prs-summary.*.txt`
   - Per-job files: `/tmp/analyze-ci-release-<version>-job-*.txt` and `/tmp/analyze-ci-prs-job-*.txt`
2. Read each summary file to extract the analysis content
3. If a summary file is missing for a release, note it as "Analysis failed or produced no output"
4. If no PR summary file exists, note "No open rebase PRs or no failures found"

### Step 5: Generate HTML Summary Report

**Goal**: Create a single HTML file at `/tmp/microshift-ci-release-manager-<timestamp>.html` that consolidates all analyses with tabbed navigation.

**Actions**:
1. Generate the HTML report with the structure described below
2. Save to `/tmp/microshift-ci-release-manager-<timestamp>.html` where `<timestamp>` is `YYYYMMDD-HHMMSS`
3. **IMPORTANT**: Use the `Bash` tool with `cat <<'HTMLEOF' > /tmp/microshift-ci-release-manager-<timestamp>.html` (heredoc) to write the file, NOT the `Write` tool. This ensures the absolute `/tmp` path is used and avoids permission prompts.
4. Display the file path to the user in the end.

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
        <button class="tab-btn active" onclick="showTab('periodics')">Periodics</button>
        <button class="tab-btn" onclick="showTab('pull-requests')">Pull Requests</button>
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

            <!-- Embed the full analysis content from analyze-ci-for-release -->
            <div class="content-block">
                ... periodics analysis content ...
            </div>
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

            <!-- Embed analysis content from analyze-ci-for-pull-requests -->
            <div class="content-block">
                ... PR analysis content ...
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
function showTab(name) {
    document.querySelectorAll('.tab-content').forEach(function(el) {
        el.classList.remove('active');
    });
    document.querySelectorAll('.tab-btn').forEach(function(el) {
        el.classList.remove('active');
    });
    document.getElementById('tab-' + name).classList.add('active');
    event.target.classList.add('active');
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
- Do NOT re-analyze or reinterpret the data from `analyze-ci-for-release` or `analyze-ci-for-pull-requests` - use their output as-is
- Convert the plain text analysis reports into HTML-formatted content, preserving all information
- Ensure all Prow job URLs from the original analyses remain clickable links in the HTML
- Use appropriate badge colors:
  - `badge-ok`: 0 failed jobs
  - `badge-issues`: 1+ failed jobs
  - `badge-critical`: 5+ failed jobs or CRITICAL severity issues present
  - `badge-nodata`: analysis failed or no data
- Make per-job details collapsible to keep the page manageable
- The overview cards should show the number of failed jobs per release and for rebase PRs at a glance
- The **Periodics** tab contains the per-release periodic job analyses (same as before)
- The **Pull Requests** tab contains the rebase PR analyses grouped by PR

### Step 6: Report Completion

**Actions**:
1. Display the path to the generated HTML file
2. Provide a brief text summary listing each release and its failed job count, plus rebase PR status

**Example Output**:
```
HTML report generated: /tmp/microshift-ci-release-manager-20260315-143022.html

Summary:
  Periodics:
    Release 4.19: 3 failed periodic jobs
    Release 4.20: 7 failed periodic jobs
    Release 4.21: 0 failed periodic jobs
    Release 4.22: 12 failed periodic jobs
  Pull Requests:
    2 rebase PRs with 5 total failed jobs
```

## Examples

### Example 1: Analyze Multiple Releases
```
/analyze-ci-for-release-manager 4.19,4.20,4.21,4.22
```

### Example 2: Analyze Two Releases
```
/analyze-ci-for-release-manager 4.21,4.22
```

### Example 3: Single Release (still produces HTML)
```
/analyze-ci-for-release-manager 4.22
```

## Notes
- Each release analysis uses the `analyze-ci-for-release` skill - this command does NOT duplicate that logic
- Rebase PR analysis uses the `analyze-ci-for-pull-requests --rebase` skill
- The HTML report is self-contained (no external CSS/JS dependencies)
- All intermediate files from `analyze-ci-for-release` and `analyze-ci-for-pull-requests` remain available in `/tmp`
- Releases are analyzed sequentially since each invocation is resource-intensive
- The rebase PR analysis runs after all releases are analyzed
- The HTML file can be opened in any browser for convenient examination
- If a release analysis fails, it is noted in the report but does not block other releases
- If no rebase PRs are open, the Pull Requests tab shows "No open rebase pull requests found"
