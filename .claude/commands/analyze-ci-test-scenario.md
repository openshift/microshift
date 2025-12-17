---
name: Analyze CI Test Scenario
argument-hint: <job-url> <scenario-name>
description: Analyze MicroShift Test Scenario results
allowed-tools: WebFetch, Bash, Read, Write, Glob, Grep
---


## Name
analyze-ci-test-scenario

## Synopsis
```
/analyze-ci-test-scenario <job-url> <scenario-name>
```

## Description
The `analyze-ci-test-scenario` command retrieves comprehensive information about a specific test scenario executed within a MicroShift CI job. It returns detailed information containing:
- Scenario configuration (OS version, test type, architecture)
- Test execution results (pass/fail counts, test names)
- MicroShift version tested
- Execution timing
- Links to logs and artifacts
- Test failure details (if any)

This command is useful for detailed investigation of specific test scenarios and understanding test execution results.

## Implementation

This command works by:

1. **Parsing the job URL** to extract job metadata (ID, name, version, architecture, image type)
2. **Constructing artifact URLs** for the specified scenario in the GCS bucket
3. **Fetching JUnit XML** test results using curl/WebFetch from the scenario's artifact directory
4. **Parsing test results** to extract pass/fail counts, test case names, and failure details
5. **Extracting scenario metadata** from the scenario name (RHEL version, release type, test category)
6. **Compiling artifact links** for all logs and diagnostic files
7. **Generating formatted Markdown output** containing all collected information

If no scenario name is provided it will prompt to the user what scenario to use.

The command uses the `.claude/scripts/extract_microshift_version.py` helper script to determine the exact MicroShift version tested in the scenario.

## Arguments
- `$1` (job-url): URL to the Prow CI job - **Required**
  - Formats accepted:
    - Full Prow dashboard URL: `https://prow.ci.openshift.org/view/gs/test-platform-results/logs/<job-name>/<job-id>`
    - GCS web URL: `https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/test-platform-results/logs/<job-name>/<job-id>`
- `$2` (scenario-name): Name of the scenario to analyze - **Required**
  - Examples: `el96-lrel@standard1`, `el96-lrel@lvm`, `el96-lrel@dual-stack`
  - If not provided, the command will list all available scenarios

## Return Value
- **Format**: Markdown
- **Location**: Output directly to the conversation
- **Content**: Comprehensive scenario information including test results, configuration, and artifacts

## Implementation Steps

### Step 1: Parse and Validate Input

**Goal**: Extract job information and scenario name from the arguments.

**Actions**:
1. Parse the job URL to extract:
   - Job name
   - Job ID
   - Version (e.g., "4.20")
   - Job type (bootc/rpm-ostree, x86_64/aarch64)
2. Validate scenario name format (should match pattern: `el[0-9]+-[a-z0-9]+@.+`)
3. If no scenario name provided, set `list_scenarios = true` flag

**Example**:
```javascript
// Input
job_url = "https://prow.ci.openshift.org/view/gs/test-platform-results/logs/periodic-ci-openshift-microshift-release-4.20-periodics-e2e-aws-tests-bootc-release-periodic/1979744605507162112"
scenario_name = "el96-lrel@standard1"

// Parsed
job_id = "1979744605507162112"
job_name = "periodic-ci-openshift-microshift-release-4.20-periodics-e2e-aws-tests-bootc-release-periodic"
version = "4.20"
job_type = "e2e-aws-tests-bootc-release-periodic"
arch = "x86_64"
image_type = "bootc"
```

### Step 2: Construct Artifact URLs

**Goal**: Build URLs to the scenario's artifacts.

**Actions**:
1. Construct base artifact URL:
   ```
   https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/test-platform-results/logs/<job-name>/<job-id>/artifacts/<job-type>/openshift-microshift-e2e-metal-tests/artifacts/scenario-info/<scenario-name>/
   ```
2. Construct specific artifact URLs:
   - JUnit XML: `<base-url>/junit.xml`
   - Boot log: `<base-url>/boot_and_run.log`
   - Debug log: `<base-url>/rf-debug.log`
   - Phase logs: `<base-url>/phase_*/*.log`

### Step 3: List Available Scenarios (if no scenario specified)

**Goal**: If no scenario name was provided, list all available scenarios in the job.

**Actions**:
1. Fetch the scenario-info directory listing:
   ```
   https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/test-platform-results/logs/<job-name>/<job-id>/artifacts/<job-type>/openshift-microshift-e2e-metal-tests/artifacts/scenario-info/
   ```
2. Use WebFetch to parse the HTML directory listing
3. Extract all scenario directory names
4. Return formatted list of scenarios

**Output Format** (if listing scenarios):
```
# Available Test Scenarios

Job: periodic-ci-openshift-microshift-release-4.20-periodics-e2e-aws-tests-bootc-release-periodic
Job ID: 1979744605507162112

## Scenarios (11 total)
- el94-y2@el96-lrel@standard1
- el94-y2@el96-lrel@standard2
- el96-lrel@ai-model-serving-online
- el96-lrel@dual-stack
- el96-lrel@ginkgo-tests
- el96-lrel@ipv6
- el96-lrel@low-latency
- el96-lrel@lvm
- el96-lrel@multi-nic
- el96-lrel@standard1
- el96-lrel@standard2
```

### Step 4: Fetch and Parse JUnit XML

**Goal**: Get test execution results from JUnit XML.

**Actions**:
1. Fetch the junit.xml file using curl or WebFetch
2. Parse XML to extract:
   - Total test count
   - Passed tests count
   - Failed tests count
   - Skipped tests count
   - Error count
   - Test execution time
   - Individual test case names and statuses
   - Failure messages and stack traces (if any)

**Example Parsing**:
```python
import xml.etree.ElementTree as ET

root = ET.fromstring(xml_content)
testsuite = root.find('.//testsuite')

test_results = {
    'total': int(testsuite.get('tests', '0')),
    'passed': 0,
    'failures': int(testsuite.get('failures', '0')),
    'errors': int(testsuite.get('errors', '0')),
    'skipped': int(testsuite.get('skipped', '0')),
    'time': float(testsuite.get('time', '0')),
    'test_cases': []
}

for testcase in testsuite.findall('.//testcase'):
    name = testcase.get('name')
    status = 'passed'
    message = None

    if testcase.find('failure') is not None:
        status = 'failed'
        message = testcase.find('failure').get('message')
    elif testcase.find('error') is not None:
        status = 'error'
        message = testcase.find('error').get('message')
    elif testcase.find('skipped') is not None:
        status = 'skipped'

    test_results['test_cases'].append({
        'name': name,
        'status': status,
        'message': message
    })

test_results['passed'] = test_results['total'] - test_results['failures'] - test_results['errors'] - test_results['skipped']
```

### Step 5: Extract Scenario Metadata

**Goal**: Parse scenario name to extract configuration details.

**Actions**:
1. Parse scenario name to extract components:
   - RHEL version (e.g., "el96" → "RHEL 9.6")
   - Release type (e.g., "lrel" → "Latest Release")
   - Test type (e.g., "standard1", "lvm", "dual-stack")
   - Upgrade path (if format is `el94-y2@el96-lrel@...` → upgrade from 9.4 to 9.6)

2. Determine test category from test type:
   - `standard1`, `standard2` → "Standard Tests"
   - `lvm` → "LVM Storage Tests"
   - `dual-stack` → "Dual-Stack Networking Tests"
   - `ipv6` → "IPv6 Networking Tests"
   - `multi-nic` → "Multi-NIC Configuration Tests"
   - `low-latency` → "Low-Latency Tests"
   - `ginkgo-tests` → "Ginkgo Integration Tests"
   - `ai-model-serving-online` → "AI Model Serving Tests"

**Example**:
```javascript
// Scenario: el96-lrel@standard1
{
  "rhel_version": "9.6",
  "release_type": "latest",
  "test_category": "Standard Tests",
  "test_variant": "1",
  "is_upgrade": false
}

// Scenario: el94-y2@el96-lrel@standard1
{
  "source_rhel_version": "9.4",
  "target_rhel_version": "9.6",
  "release_type": "latest",
  "test_category": "Standard Tests",
  "test_variant": "1",
  "is_upgrade": true
}
```

### Step 6: Get Execution Timing

**Goal**: Extract when the scenario was executed and how long it took.

**Actions**:
1. Check boot_and_run.log for timestamps
2. Look for start and end markers in the log
3. Calculate duration if both timestamps available
4. Extract from junit.xml `time` attribute as fallback

### Step 7: Compile Artifact Links

**Goal**: Provide direct links to all relevant artifacts for the scenario.

**Actions**:
1. Build URLs for common artifacts:
   - JUnit XML report
   - Boot and run log
   - Debug log
   - Phase logs (if they exist)
   - Sosreport (if test failed)

2. Categorize artifacts:
   - Test results: junit.xml
   - Execution logs: boot_and_run.log, rf-debug.log
   - Phase logs: All logs under phase_* directories
   - Diagnostics: sosreports, system logs


### Error Handling

**Common Issues and Responses**:

1. **Scenario not found**:
```
# Error: Scenario Not Found

Scenario 'el96-lrel@invalid' does not exist in job 1979744605507162112

## Available Scenarios
- el96-lrel@standard1
- el96-lrel@lvm
- ...
```

2. **Job not found**:
```
# Error: Job Not Found

Could not fetch artifacts for job ID 1234567890

Please verify the job URL and ensure the job has completed.
```

3. **Missing artifacts**:
```
# Warning: Partial Data Available

Some artifacts were not available for this scenario.

## Missing Artifacts
- junit.xml

Displaying available information below...
```

## Examples

### Example 1: Get scenario information
```
/analyze-ci-test-scenario https://prow.ci.openshift.org/view/gs/test-platform-results/logs/periodic-ci-openshift-microshift-release-4.20-periodics-e2e-aws-tests-bootc-release-periodic/1979744605507162112 el96-lrel@standard1
```

Output:
```
# Test Scenario Analysis: el96-lrel@standard1

## Job Information
- **Job Name**: periodic-ci-openshift-microshift-release-4.20-periodics-e2e-aws-tests-bootc-release-periodic
- **Job ID**: 1979744605507162112
- **Version**: 4.20
- **Architecture**: x86_64
- **Image Type**: bootc

## Scenario Configuration
- **Name**: el96-lrel@standard1
- **Description**: RHEL 9.6 Latest Release - Standard Tests
- **RHEL Version**: 9.6
- **Release Type**: Latest
- **Test Category**: Standard Tests
- **Upgrade Test**: No

## Test Results
**Status**: PASSED

### Summary
- **Total Tests**: 65
- **Passed**: 65
- **Failed**: 0
- **Errors**: 0
- **Skipped**: 0

## Artifacts
- [JUnit XML](https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/.../junit.xml)
- [Boot and Run Log](https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/.../boot_and_run.log)
```

### Example 2: List all scenarios (no scenario name provided)
```
/analyze-ci-test-scenario https://prow.ci.openshift.org/view/gs/test-platform-results/logs/periodic-ci-openshift-microshift-release-4.20-periodics-e2e-aws-tests-bootc-release-periodic/1979744605507162112
```

Output:
```
# Available Test Scenarios

Job: periodic-ci-openshift-microshift-release-4.20-periodics-e2e-aws-tests-bootc-release-periodic
Job ID: 1979744605507162112

## Scenarios (5 total)
- el96-lrel@standard1
- el96-lrel@standard2
- el96-lrel@lvm
- el96-lrel@dual-stack
- el96-lrel@ipv6
```

### Example 3: Get information about a failed scenario
```
/analyze-ci-test-scenario https://prow.ci.openshift.org/view/gs/test-platform-results/logs/periodic-ci-openshift-microshift-release-4.20-periodics-e2e-aws-tests-release-periodic/1234567890 el96-lrel@lvm
```

Output would include failure details:
```
# Test Scenario Analysis: el96-lrel@lvm

## Job Information
- **Job Name**: periodic-ci-openshift-microshift-release-4.20-periodics-e2e-aws-tests-release-periodic
- **Job ID**: 1234567890
- **Version**: 4.20

## Scenario Configuration
- **Name**: el96-lrel@lvm
- **Description**: RHEL 9.6 Latest Release - LVM Storage Tests
- **Test Category**: LVM Storage Tests

## Test Results
**Status**: FAILED

### Summary
- **Total Tests**: 45
- **Passed**: 43
- **Failed**: 2
- **Errors**: 0
- **Skipped**: 0

### Failed Tests
1. **LVM volume creation**
   - **Error**: Failed to create LVM volume: insufficient space
   - **Log**: [create-lvm.log](https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/.../phase_create-and-run/create-lvm.log)

## Artifacts
- [JUnit XML](https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/.../junit.xml)
- [Boot and Run Log](https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/.../boot_and_run.log)
- [Debug Log](https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/.../rf-debug.log)
```

## Notes
- This command outputs detailed information in Markdown format for easy reading
- The command is read-only and does not modify any CI job data
- If a scenario doesn't have junit.xml, the command will attempt to infer results from logs
- The command caches scenario lists internally to avoid repeated fetches when listing scenarios
- Artifact links in the output are direct URLs to GCS storage for immediate access
