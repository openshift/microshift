---
name: Analyze CI Test Job
argument-hint: <job-url>
description: Analyze a MicroShift Prow CI Test Job execution
allowed-tools: WebFetch, Bash, Read, Write, Glob, Grep
---

## Name
analyze-ci-test-job

## Synopsis
```
/analyze-ci-test-job <job-url>
```

## Description
The `analyze-ci-test-job` command fetches comprehensive information from a Prow CI job execution and displays it in both JSON and Markdown formats.

This command provides:
- Job metadata (status, timing, architecture, image type)
- MicroShift version being tested
- Test scenarios executed and their results
- Build information
- Links to logs and artifacts

This command is useful for understanding what was tested in a specific job run, identifying failures, and accessing detailed logs and artifacts.

## Implementation

This command works by:

1. **Parsing the job URL** to extract job name, ID, and configuration (architecture, image type, version)
2. **Fetching job metadata** from `finished.json` and `started.json` to get status, timing, and result information
3. **Extracting MicroShift version** using the `extract_microshift_version.py` helper script from build logs
4. **Listing test scenarios** by fetching the scenario-info directory structure from GCS artifacts
5. **Analyzing test results** for each scenario using the `analyze-test-scenario` command to get comprehensive JSON data
6. **Compiling artifacts and logs** by constructing URLs to build logs, test execution logs, and failure diagnostics
7. **Generating a detailed Markdown report** with job overview, version info, scenario results, and artifact links

The command integrates with the `analyze-test-scenario` command to provide detailed per-scenario analysis and aggregates all information into a human-readable report with proper formatting (status icons, duration calculations, failure summaries).

## Arguments
- `$1` (job-url): URL to the Prow CI job - **Required**
  - Formats accepted:
    - Full Prow dashboard URL: `https://prow.ci.openshift.org/view/gs/test-platform-results/logs/<job-name>/<job-id>`
    - GCS web URL: `https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/test-platform-results/logs/<job-name>/<job-id>`
    - Job ID only (e.g., "1979744605507162112") - will attempt to infer job type from context

## Return Value
- **Format**: Markdown
- **Location**: Output directly to the conversation
- **Content**:
  - Job overview (status, timing, configuration)
  - MicroShift version details
  - Test scenario results
  - Build information
  - Links to logs and artifacts

## Implementation Steps

### Step 1: Parse Arguments and Validate Job URL

**Goal**: Extract the job name, job ID, and job configuration.

**Actions**:
1. Parse the job URL to extract:
   - Job name (e.g., "periodic-ci-openshift-microshift-release-4.20-periodics-e2e-aws-tests-bootc-release-periodic")
   - Job ID (e.g., "1979744605507162112")
2. Determine job configuration from job name:
   - Architecture: x86_64 or aarch64 (look for "arm" in job name)
   - Image type: bootc or rpm-ostree (look for "bootc" in job name)
   - Version: extract from job name (e.g., "4.20")
3. Validate URL format
4. If only job ID provided, ask user for job type or attempt to determine from recent jobs

**Example Parsing**:
```
URL: https://prow.ci.openshift.org/view/gs/test-platform-results/logs/periodic-ci-openshift-microshift-release-4.20-periodics-e2e-aws-tests-bootc-release-periodic/1979744605507162112

Extracted:
- job_name: "periodic-ci-openshift-microshift-release-4.20-periodics-e2e-aws-tests-bootc-release-periodic"
- job_id: "1979744605507162112"
- version: "4.20"
- arch: "x86_64"
- image_type: "bootc"
```

### Step 2: Fetch Job Metadata

**Goal**: Get job information (status, timing, result).

**Actions**:
1. Construct the GCS URL for the `finished.json` file:
   ```
   https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/test-platform-results/logs/<job-name>/<job-id>/finished.json
   ```
2. Fetch the `finished.json` file using curl or WebFetch
3. Parse the JSON to extract:
   - Job result (SUCCESS/FAILURE/ABORTED)
   - Timestamp (start/end times)
   - Duration
   - Passed status
   - Metadata (repo, revision, etc.)
4. Fetch `started.json` for additional metadata:
   ```bash
   curl -s "https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/test-platform-results/logs/<job-name>/<job-id>/started.json"
   ```

### Step 3: Extract MicroShift Version

**Goal**: Determine the exact MicroShift version being tested.

This command includes a Python script that automates version extraction from test logs.

**Script Location**: `.claude/scripts/extract_microshift_version.py`

**Usage**:
```bash
python3 .claude/scripts/extract_microshift_version.py <prow_url> <scenario>
```

**Arguments**:
- `prow_url`: The full Prow CI job URL (e.g., "https://prow.ci.openshift.org/view/gs/test-platform-results/logs/periodic-ci-openshift-microshift-release-4.20-periodics-e2e-aws-tests-bootc-release-periodic/1979744605507162112")
- `scenario`: The test scenario name (e.g., "el96-lrel@ipv6")

**Example**:
```bash
# Extract version for a specific job and scenario
python3 .claude/scripts/extract_microshift_version.py "https://prow.ci.openshift.org/view/gs/test-platform-results/logs/periodic-ci-openshift-microshift-release-4.20-periodics-e2e-aws-tests-bootc-release-periodic/1979744605507162112" "el96-lrel@ipv6"
```

**Output** (JSON):
```json
{
  "success": true,
  "version": "4.20.0-202510161342.p0.g17d1d9a.assembly.4.20.0.el9.x86_64",
  "build_type": "zstream",
  "url": "https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/test-platform-results/logs/...",
  "error": null
}
```

**Build Types Detected**:
- `"nightly"`: Nightly development builds
- `"ec"`: Engineering Candidate
- `"rc"`: Release Candidate
- `"zstream"`: Stable/zstream release

### Step 4: List Test Scenarios

**Goal**: Find all test scenarios executed in this job and list them.

**Output** (JSON):
```json
{
  "job_id": "1979744605507162112",
  "scenarios": [
    "el94-y2@el96-lrel@standard1",
    "el96-lrel@standard1",
    "el96-lrel@lvm",
    "el96-lrel@dual-stack",
    "el96-lrel@ipv6"
  ],
  "total_scenarios": 5
}
```

### Step 5: Analyze Test Results for Each Scenario

**Goal**: Get detailed test execution results for each scenario.

**Method**: Use the `analyze-test-scenario` command for each scenario to get comprehensive JSON data.

**Actions**:
For each scenario found in Step 4:

1. **Get scenario details** using the analyze-test-scenario command:
   ```bash
   /microshift-prow-job:analyze-test-scenario <job-url> <scenario-name>
   ```

2. **Parse the JSON response** which includes:
   - Test results summary (total, passed, failed, errors, skipped)
   - Individual test case details
   - Failure messages and details (if any)
   - Scenario configuration (RHEL version, test category)
   - Execution timing
   - Links to all artifacts

**Example JSON Response**:
```json
{
  "scenario": {
    "name": "el96-lrel@standard1",
    "description": "RHEL 9.6 Latest Release - Standard Tests",
    "configuration": {
      "rhel_version": "9.6",
      "release_type": "latest",
      "test_category": "Standard Tests"
    }
  },
  "test_results": {
    "status": "passed",
    "summary": {
      "total": 65,
      "passed": 65,
      "failed": 0,
      "errors": 0,
      "skipped": 0
    },
    "execution_time_seconds": 1234.56,
    "test_cases": [
      {
        "name": "MicroShift boots successfully",
        "status": "passed"
      }
    ],
    "failures": []
  },
  "artifacts": {
    "junit_xml": "https://...",
    "boot_log": "https://...",
    "debug_log": "https://..."
  }
}
```

3. **Extract key information** from each scenario:
   - Overall status (passed/failed)
   - Test counts
   - Failure details (for failed scenarios)
   - Execution time
   - Test category and configuration

**Alternative Manual Method** (if analyze-test-scenario command unavailable):
1. Fetch junit.xml directly from artifact URL
2. Parse XML to extract test counts
3. Check boot_and_run.log for execution details
4. Extract scenario metadata from directory structure

### Step 6: Compile Artifacts and Logs

**Goal**: Provide links to useful artifacts and logs.

**Actions**:
1. Compile key artifact URLs:
   - Build log: `artifacts/<job-type>/openshift-microshift-infra-iso-build/build-log.txt`
   - Test logs for each scenario
   - JUnit XML reports
   - Any failure logs or sosreports

2. Categorize artifacts by type:
   - Build artifacts
   - Test execution logs
   - Failure diagnostics
   - System information

### Step 7: Generate Detailed Report

**Goal**: Create a comprehensive, well-structured report.

**Report Structure**:
```markdown
# MicroShift CI Job Details

## Job Overview
- **Job ID**: <job-id>
- **Job Name**: <job-name>
- **Status**: ✓ SUCCESS / ✗ FAILURE / ⚠️ ABORTED
- **Architecture**: x86_64 / aarch64
- **Image Type**: bootc / rpm-ostree
- **Duration**: Xh Ym Zs
- **Started**: YYYY-MM-DD HH:MM:SS UTC
- **Finished**: YYYY-MM-DD HH:MM:SS UTC

## MicroShift Version
- **Full Version**: <full-version-string>
- **Build Type**: nightly / RC / EC / stable
- **Base Version**: X.Y.Z
- **Commit**: <commit-hash>
- **Build Timestamp**: YYYY-MM-DD-HHMMSS

## Test Scenarios

### Scenario: <scenario-name>
- **Description**: <RHEL version and test type>
- **Status**: ✓ PASS / ✗ FAIL
- **Tests**: X passed, Y failed, Z skipped
- **Duration**: Xm Ys

**Failures** (if any):
- Test: <test-name>
  - Error: <error-message>
  - Log: [View](<log-url>)

[Repeat for each scenario]

## Build Information
- **Build Status**: SUCCESS / FAILURE
- **Build Log**: [View](<build-log-url>)
- **Build Duration**: Xm Ys

## Artifacts & Logs
- [Build Log](<url>)
- [Test Execution Logs](<url>)
- [Scenario Details](<url>)
- [Full Artifacts](<artifacts-url>)

## Links
- [View on Prow CI](<prow-url>)
- [Browse All Artifacts](<gcsweb-url>)
```

### Step 8: Error Handling

**Goal**: Handle errors gracefully.

**Common Issues**:
1. **Job not found (404)**:
   - Verify job ID is correct
   - Check if job is still running (no finished.json yet)
   - Provide helpful error message to user
   - Handle network errors gracefully

2. **Artifacts not available**:
   - Some jobs may not have all artifacts
   - Gracefully handle missing files
   - Indicate which artifacts are unavailable in the report

3. **Invalid job URL**:
   - Validate URL format before making requests
   - Handle malformed URLs
   - Provide examples of valid formats
   - Suggest using job ID from create-report command

4. **Version extraction failures**:
   - Handle cases where version cannot be determined
   - Provide partial information if available
   - Include error message in report

## Examples

### Example 1: Successful Job Analysis
```
/microshift-prow-job:analyze-ci-test-job https://prow.ci.openshift.org/view/gs/test-platform-results/logs/periodic-ci-openshift-microshift-release-4.20-periodics-e2e-aws-tests-bootc-release-periodic/1979744605507162112
```

Output:
```markdown
# MicroShift CI Job Details

## Job Overview
- **Job ID**: 1979744605507162112
- **Job Name**: periodic-ci-openshift-microshift-release-4.20-periodics-e2e-aws-tests-bootc-release-periodic
- **Status**: ✓ SUCCESS
- **Architecture**: x86_64
- **Image Type**: bootc
- **Duration**: 1h 43m 40s
- **Started**: 2025-10-19 03:01:17 UTC
- **Finished**: 2025-10-19 04:44:57 UTC

## MicroShift Version
- **Full Version**: 4.20.0-0.nightly-2025-10-15-110252-20251017171355-4ad30ab2d
- **Build Type**: nightly
- **Base Version**: 4.20.0
- **Commit**: 4ad30ab2d
- **Build Date**: 2025-10-15

## Test Scenarios

### Scenario: el96-lrel@standard1
- **Description**: RHEL 9.6 Latest Release - Standard Tests
- **Status**: ✓ PASS
- **Tests**: 45 passed, 0 failed, 2 skipped

[Additional sections...]
```

### Example 2: Using GCS Web URL
```
/microshift-prow-job:analyze-ci-test-job https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/test-platform-results/logs/periodic-ci-openshift-microshift-release-4.20-periodics-e2e-aws-tests-release-arm-periodic/1979744608019550208
```

### Example 3: Failed Job Analysis
```
/microshift-prow-job:analyze-ci-test-job https://prow.ci.openshift.org/view/gs/test-platform-results/logs/some-failing-job/9876543210
```

Output would include failure details:
```markdown
## Job Overview
- **Status**: ✗ FAILURE
...

## Test Scenarios
...
**Failures**:
- Test: <test-name>
  - Error: <error-message>
  - Log: [View](<log-url>)
```

### Example 4: Job ID Only
```
/microshift-prow-job:analyze-ci-test-job 1979744605507162112
```
(May prompt for additional context or attempt to determine job type from recent jobs)

## Notes
- This command provides comprehensive analysis including job status, MicroShift version, test scenarios, and detailed results
- Works with MicroShift-specific Prow CI jobs
- Requires internet access to fetch job data from Prow CI
- All times are displayed in UTC
- Duration is calculated from the finished.json timestamp and start time from started.json
- The command is read-only and does not modify any CI job data
- Useful for debugging specific test failures or understanding what was tested
