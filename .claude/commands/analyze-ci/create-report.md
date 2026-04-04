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
Reads analysis output files from `${WORKDIR}/` (produced by `analyze-ci:release` and `analyze-ci:pull-requests`) and optionally bug mapping files (produced by `analyze-ci:create-bugs` in both dry-run and create modes) to generate a single consolidated HTML report with tabbed navigation. When bug mapping files are present, the report shows linked JIRA bugs per issue. This command does NOT run any CI analysis — it only reads existing files and generates HTML.

## Arguments
- `$ARGUMENTS` (required): Comma-separated list of release versions that were analyzed (e.g., `4.19,4.20,4.21,4.22`)

## Work Directory

Set once at the start and reference throughout:
```bash
WORKDIR=/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)
```

## Implementation Steps

### Step 1: Run the Report Generator Script

**Actions**:
1. Set `WORKDIR=/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)`
2. Run the Python script:
   ```bash
   python3 .claude/scripts/analyze-ci-create-report.py --workdir ${WORKDIR} $ARGUMENTS
   ```
3. The script handles everything deterministically:
   - Discovers JSON summary files and bug mapping files in `${WORKDIR}`
   - Loads release summaries, PR summaries, and bug candidates from JSON
   - Fuzzy-matches issues to bug candidates using token overlap (>= 50% threshold)
   - Generates a self-contained HTML file with tabbed navigation, collapsible issues, and JIRA bug links
4. Report the script's stdout output to the user (it includes file discovery, summary, and the output path)

**Error Handling**:
- If `$ARGUMENTS` is empty, the script shows usage and exits
- If no analysis files are found, the script reports an error and exits
- Missing releases are shown as "no data" in the report

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
- **analyze-ci:create-bugs**: Creates JIRA bugs from analysis output; also produces bug mapping files (`analyze-ci-bugs-<source>.json`) consumed by this command to show JIRA links in the HTML report

## Notes
- This command is read-only — it only reads existing analysis files and generates HTML
- The HTML report is self-contained (no external CSS/JS dependencies)
- The HTML file can be opened in any browser for convenient examination
- If analysis files are missing for a release, it is noted in the report but does not block generation
- If no PR files exist, the Pull Requests tab shows "No open rebase pull requests found"
- This command can be re-run to regenerate the HTML without re-running the analyses
- If bug mapping files (`analyze-ci-bugs-*.json`) exist in `${WORKDIR}/`, JIRA bug links are shown per issue in the HTML; if not, the report still generates correctly without bug links
