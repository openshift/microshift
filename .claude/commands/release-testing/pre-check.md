---
name: Release Pre-Check
argument-hint: [Z|X|Y|RC|EC|nightly] [version|time-range...] [--verbose]
description: Check OCP release schedule, verify availability, evaluate z-stream need, or check nightly build gaps
allowed-tools: Bash, mcp__productpages__search_entities, mcp__productpages__browse_schedule, mcp__productpages__list_entities
---

# release-testing:pre-check

## Synopsis
```bash
/release-testing:pre-check [release_type] [version|time-range...] [--verbose]
```

## Description
MicroShift ships as a layered product on top of OCP. Every time OCP releases a new version (z-stream, EC, RC, or nightly), the MicroShift team must evaluate whether to participate — checking for CVEs, verifying RPM builds exist in Brew, and deciding whether to ask ART to create artifacts.

This command automates that evaluation (Phase 0 of the release process). It checks lifecycle status, OCP payload availability, advisory CVEs, nightly build gaps, and EC/RC readiness — then outputs a clear action per version: OK, SKIP, ASK ART, NEEDS REVIEW, or ALREADY RELEASED.

When a time range is provided (e.g., "this week"), it queries Red Hat Product Pages for OCP versions scheduled in that period and evaluates each one.

## Prerequisites

| Requirement | Needed for | Mandatory? |
|---|---|---|
| VPN | Brew RPM checks (nightly, EC/RC), advisory report | Yes for nightly/ecrc — xyz degrades gracefully (skips advisory, 90-day rule) |
| `ATLASSIAN_API_TOKEN` | ART Jira queries, advisory report | No — scripts degrade gracefully, but advisory/CVE analysis is skipped |
| `ATLASSIAN_EMAIL` | ART Jira queries, advisory report | No — same as above |
| `GITLAB_API_TOKEN` | Advisory report for 4.20+ (shipment MR data) | No — advisory skipped for 4.20+ without it |
| Product Pages MCP | Time range lookups (e.g., "this week") | Only when using time ranges — not needed for explicit versions |

## Arguments
- `release_type` (optional): One or more of `Z`, `X`, `Y`, `RC`, `EC`, `nightly` (case-insensitive). If omitted, defaults to `Z`.
- `version` (optional): Specific version (e.g., `4.19.27`) or minor stream (e.g., `4.21`)
- `time-range` (optional): Natural language time range instead of explicit versions. Detected by keywords like:
  - `today`, `tomorrow`
  - `this week`, `next week`, `last week`
  - `next 3 days`, `next 7 days`
  - `April`, `this month`
- `--verbose` (optional): Show extra detail (tables for xyz, NVR/nightly names for nightly, next versions for EC/RC).

## Implementation

### Step 1: Parse Arguments

1. Identify `release_type`(s) — tokens matching `Z`, `X`, `Y`, `RC`, `EC`, `nightly` (case-insensitive)
2. Identify `version`(s) — tokens matching `X.Y` or `X.Y.Z` pattern
3. Identify `time range` — remaining tokens that are not release types, versions, or flags (e.g., "this week", "next 3 days", "tomorrow")
4. Identify `--verbose` flag
5. **Default**: If no release_type found, default to `Z` and treat version/time-range tokens accordingly

### Step 2: Resolve Versions from Product Pages (when time range is detected)

If a time range is present instead of explicit versions, query Product Pages to find OCP z-stream versions scheduled in that period:

1. **Convert the time range** to concrete dates (date_from, date_to) based on today's date
2. **Find active OCP z-release entities**: Use `mcp__productpages__list_entities` with `public_parent_id=146` (OCP product), `kind="release"`, `shortname="%z"`, `is_maintained=True`. Filter results to 4.14+ only (MicroShift GA'd at 4.14 — older z-release entities have no MicroShift images).
3. **For each z-release entity**: Call `mcp__productpages__browse_schedule` and find tasks where:
   - `flags` contains `"ga"` (these are the "X.Y.Z in Fast Channel" milestones)
   - `date_finish` falls within [date_from, date_to]
   - Extract the version from the task name (e.g., "4.21.10 in Fast Channel" → `4.21.10`)
4. **Collect all matching versions** across all streams and pass them to the xyz script

If no versions found, report "No OCP releases scheduled in <range>."

### Step 3: Run the Script

Map each release type to the corresponding `precheck.sh` subcommand and run via Bash:

| Release Type | Command |
|---|---|
| `Z`, `X`, `Y` (default) | `bash .claude/scripts/release_testing/precheck.sh xyz [versions...] [--verbose]` |
| `nightly` | `bash .claude/scripts/release_testing/precheck.sh nightly [version] [--verbose]` |
| `EC` | `bash .claude/scripts/release_testing/precheck.sh ecrc EC [version] [--verbose]` |
| `RC` | `bash .claude/scripts/release_testing/precheck.sh ecrc RC [version] [--verbose]` |

Stderr contains progress messages — only display it if the script exits non-zero.

**Multiple types** (e.g., `nightly EC RC`): Run each command as a separate Bash call in parallel.

### Step 4: Display Output

Display the script output **verbatim** — do not reformat, add tables, or change the layout. The scripts produce deterministic pre-formatted text. Do NOT add any commentary, explanation, or summary after the output.

### Step 5: Handle Errors

If the script exits non-zero, display stderr and suggest:
- VPN not connected → connect to VPN (Brew requires it)
- Missing env vars → set `ATLASSIAN_API_TOKEN`, `ATLASSIAN_EMAIL`, `GITLAB_API_TOKEN` (for 4.20+)

## Examples

```bash
/release-testing:pre-check this week                   # OCP versions releasing this week
/release-testing:pre-check next week                   # OCP versions releasing next week
/release-testing:pre-check today                       # OCP versions releasing today
/release-testing:pre-check next 3 days                 # OCP versions in next 3 days
/release-testing:pre-check 4.21.10                     # specific version
/release-testing:pre-check 4.20 4.21 4.22              # xyz eval for multiple streams
/release-testing:pre-check 4.19.27 --verbose           # specific version, detailed report
/release-testing:pre-check nightly                     # nightly gaps for all active branches
/release-testing:pre-check EC                          # latest EC status
/release-testing:pre-check RC                          # latest RC status
/release-testing:pre-check nightly EC RC               # combined report
```

## Product Pages Reference

- OCP product entity ID: **146** (from `search_entities("OpenShift Container Platform", kind="product")` — hardcoded to skip one API call)
- Z-release entities: `openshift-X.Y.z` (e.g., `openshift-4.21.z`)
- Schedule tasks with `"ga"` flag = version GA dates ("X.Y.Z in Fast Channel")
- `date_finish` on ga-flagged tasks = release date

## Notes
- Read-only — does NOT create tickets or modify external state
- Scripts support `--json` for raw JSON output when called directly (e.g., `bash .claude/scripts/release_testing/precheck.sh xyz 4.21.10 --json`)
- `--verbose` works for all types: detailed tables for xyz, NVR/nightly names for nightly, next versions for EC/RC
- Jira enrichment is optional (scripts handle gracefully without credentials)
- VPN required for Brew and errata access
