---
name: Analyze SOS Report
argument-hint: <sos-report-path> [log.html-url]
description: Investigate MicroShift runtime problems from SOS report
allowed-tools: WebFetch, Bash, Read, Glob, Grep
---

## Name
analyze-sos-report

## Synopsis
```
/analyze-sos-report <sos-report-path> [log.html-url]
```

## Description
The `analyze-sos-report` command investigates MicroShift runtime problems by analyzing journal logs, Pod logs, YAML manifests, and configuration from a SOS report. Optionally, it can cross-reference findings with a Robot Framework test log.

This command focuses on:
- MicroShift and CRI-O journal logs for errors
- Pod status and container logs
- Kubernetes resource YAMLs (Deployments, Pods, Services, etc.)
- MicroShift configuration issues
- etcd errors and problems
- OVN networking issues
- **Robot Framework test results correlation** (when log.html URL is provided)

## Arguments
- `$1` (sos-report-path): Path or URL to the SOS report - **Required**
  - **Local directory**: Path to an already extracted sosreport (e.g., `/tmp/sosreport-hostname-2025-01-15`)
  - **Remote URL**: HTTP/HTTPS URL to a `.tar.xz` file that will be downloaded and extracted to `/tmp`
  - The extracted directory should contain `sos_commands/microshift` subdirectory.

- `$2` (log.html-url): URL to Robot Framework log.html file - **Optional**
  - When provided, test results will be analyzed and cross-referenced with SOS report findings
  - Helps correlate test failures with system-level issues
  - Example: `https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/.../log.html`

## Return Value
- **Format**: Markdown
- **Location**: Output directly to the conversation
- **Content**:
  - Identified runtime issues
  - Root cause analysis
  - Relevant log excerpts
  - Recommendations

## SOS Report Structure Reference

### Required directories (always present):
- `sos_commands/microshift/` - MicroShift specific data

### Key files in `sos_commands/microshift/`:
- `journalctl_--no-pager_--unit_microshift` - MicroShift service logs
- `journalctl_--no-pager_--unit_microshift-etcd.scope` - etcd logs
- `microshift_version` - MicroShift version
- `microshift_show-config_-m_effective` - Effective configuration
- `systemctl_status_microshift` - Service status
- `event-filter.html` - Kubernetes events viewer

### Namespace structure in `sos_commands/microshift/namespaces/<NAMESPACE>/`:
- `<namespace>.yaml` - Namespace definition
- `core/pods.yaml` - Pod definitions
- `core/events.yaml` - Namespace events
- `core/configmaps.yaml` - ConfigMaps
- `core/services.yaml` - Services
- `apps/deployments.yaml` - Deployments
- `apps/daemonsets.yaml` - DaemonSets
- `pods/<POD>/<POD>.yaml` - Individual Pod YAML
- `pods/<POD>/<CONTAINER>/<CONTAINER>/<CONTAINER>/logs/current.log` - Container logs
- `pods/<POD>/<CONTAINER>/<CONTAINER>/<CONTAINER>/logs/previous.log` - Previous container logs

### Cluster-scoped resources in `sos_commands/microshift/cluster-scoped-resources/`:
- `core/nodes.yaml` - Node information
- `storage.k8s.io/storageclasses.yaml` - Storage classes

### Optional directories (may not be present in minimal reports):
- `sos_commands/crio/` - CRI-O container runtime
- `sos_commands/logs/` - System journal logs
- `sos_commands/microshift_ovn/` - OVN networking
- `sos_commands/openvswitch/` - Open vSwitch
- `sos_commands/networking/` - Network configuration
- `etc/microshift/` - MicroShift configuration files

## Robot Framework Log Structure Reference

When a log.html URL is provided, the file contains Robot Framework test execution results:

### Key Information in log.html:
- **Test Suites**: Hierarchical organization of tests
- **Test Cases**: Individual test names, status (PASS/FAIL/SKIP), duration
- **Keywords**: Step-by-step execution details within each test
- **Timestamps**: Execution times for each test and keyword
- **Error Messages**: Failure reasons for failed tests
- **Log Messages**: DEBUG, INFO, WARN level messages during execution

### Timestamp Correlation:
- **Important**: log.html timestamps may be in a different timezone than SOS report journal logs
- Typically only the **hour component differs** (e.g., log.html shows 10:30:45, journal shows 05:30:45)
- When correlating events, match by **minutes and seconds** and allow for hour offset
- Look for events that occur within the same minute across both sources

### Test Sources:
- Robot Framework test files are located in `test/` directory of MicroShift repository
- Test file paths can be extracted from log.html suite names
- **Caveat**: The currently checked-out branch may not match the CI run that generated the log.html
- Use test sources as reference only; the actual executed code may differ

## Implementation Steps

### Step 1: Handle Input (URL or Local Path)

**Goal**: Determine if input is a URL or local path, and prepare the sosreport directory.

**Actions**:
1. Check if the input starts with `http://` or `https://`:
   - If YES: It's a remote URL, proceed to download and extract
   - If NO: It's a local path, proceed to validation

**For Remote URLs**:
1. Download the `.tar.xz` file to `/tmp`:
   ```bash
   curl -L -o /tmp/sosreport-download.tar.xz "<url>"
   ```
2. Extract the archive to `/tmp`:
   ```bash
   tar -xf /tmp/sosreport-download.tar.xz -C /tmp
   ```
3. Find the extracted directory:
   ```bash
   ls -dt /tmp/sosreport-*/ 2>/dev/null | head -1
   ```
4. Set the extracted directory as the working path
5. Clean up the downloaded archive:
   ```bash
   rm /tmp/sosreport-download.tar.xz
   ```

**For Local Paths**:
1. Check if the directory exists
2. Verify it contains `sos_commands/microshift/` subdirectory

### Step 2: Analyze MicroShift Journal Logs

**Goal**: Find errors and problems in MicroShift service logs.

**Actions**:
1. Read MicroShift journal logs:
   ```bash
   cat <sos-report-path>/sos_commands/microshift/journalctl_--no-pager_--unit_microshift
   ```
2. Search for errors, warnings, and failures - look for patterns like:
   - `error`, `fail`, `fatal`, `panic`
   - `timeout`, `refused`, `denied`
   - Startup failures
   - API server errors
   - Controller errors

### Step 3: Analyze etcd Logs

**Goal**: Check embedded etcd health.

**Actions**:
1. Read etcd journal logs:
   ```bash
   cat <sos-report-path>/sos_commands/microshift/journalctl_--no-pager_--unit_microshift-etcd.scope
   ```
2. Look for etcd-specific issues:
   - Database corruption
   - Slow disk warnings
   - Leader election problems
   - Timeout errors

### Step 4: Analyze CRI-O Logs (if available)

**Goal**: Check container runtime issues.

**Actions**:
1. Read CRI-O journal logs (if present):
   ```bash
   cat <sos-report-path>/sos_commands/crio/journalctl_--no-pager_--unit_crio
   ```
2. Check container status:
   ```bash
   cat <sos-report-path>/sos_commands/crio/crictl_ps_-a
   ```
3. Look for crashed or errored containers

### Step 5: Analyze Pod Status and Events

**Goal**: Check Pod health and Kubernetes events.

**Actions**:
1. Check Pod status in each namespace:
   ```bash
   cat <sos-report-path>/sos_commands/microshift/namespaces/*/core/pods.yaml
   ```
2. Look for Pods not in Running state (Pending, CrashLoopBackOff, Error, etc.)

3. Check events for problems:
   ```bash
   cat <sos-report-path>/sos_commands/microshift/namespaces/*/core/events.yaml
   ```
4. Look for warning events indicating issues

### Step 6: Analyze Container Logs

**Goal**: Check individual container logs for errors.

**Actions**:
1. Find container logs:
   ```bash
   find <sos-report-path>/sos_commands/microshift/namespaces -name "current.log" -o -name "previous.log"
   ```
2. Search for errors in container logs:
   ```bash
   grep -rE "error|fail|panic|fatal" <sos-report-path>/sos_commands/microshift/namespaces/*/pods/*/
   ```
3. Check previous.log files for containers that restarted

### Step 7: Analyze Resource YAMLs

**Goal**: Check resource configurations for issues.

**Actions**:
1. Check Deployments and DaemonSets:
   ```bash
   cat <sos-report-path>/sos_commands/microshift/namespaces/*/apps/deployments.yaml
   cat <sos-report-path>/sos_commands/microshift/namespaces/*/apps/daemonsets.yaml
   ```
2. Look for:
   - Replicas not matching desired count
   - Image pull errors
   - Resource constraint issues

### Step 8: Analyze MicroShift Configuration

**Goal**: Check for configuration issues.

**Actions**:
1. Read effective MicroShift config:
   ```bash
   cat <sos-report-path>/sos_commands/microshift/microshift_show-config_-m_effective
   ```
2. Check config files if present:
   ```bash
   cat <sos-report-path>/etc/microshift/config.yaml
   ```
3. Check for common misconfigurations

### Step 9: Analyze OVN/Networking (if available)

**Goal**: Check networking issues.

**Actions**:
1. Check OVN status if present:
   ```bash
   cat <sos-report-path>/sos_commands/microshift_ovn/*
   ```
2. Check OVN-related pods in openshift-ovn-kubernetes namespace
3. Look for network connectivity issues in logs

### Step 10: Analyze Robot Framework Log (if provided)

**Goal**: Extract test results from log.html and correlate with SOS report findings.

**Actions**:
1. Fetch and parse the log.html file:
   ```bash
   # Use WebFetch to retrieve and analyze the log.html
   ```

2. Extract key information:
   - Test suite name and overall status (PASS/FAIL)
   - Individual test cases with their status
   - Failed test names and error messages
   - Test execution timestamps

3. Cross-reference with SOS report:
   - **Timezone consideration**: Timestamps in log.html and SOS report may differ by timezone (typically only the hour component differs). When correlating events, match by minutes and seconds, allowing for hour offset.
   - Match test failure times with errors in MicroShift journal logs
   - Identify if pod restarts or errors occurred during specific test execution
   - Look for patterns: did a specific test cause system instability?

4. Test source reference:
   - Test sources are in the `test/` directory of the MicroShift repository
   - **Note**: The checked-out branch may not match the CI run, so test sources are for reference only
   - Extract test file paths from log.html to help locate relevant test code

### Step 11: Generate Investigation Report

**Goal**: Compile findings into a focused problem analysis.

**Report Structure**:
```markdown
# MicroShift Runtime Problem Analysis

## Summary
<Brief 1-2 sentence summary of the main issue found>

## Test Results (if log.html provided)
| Test Name | Status | Duration | Error |
|-----------|--------|----------|-------|
| ... | PASS/FAIL | ... | ... |

### Failed Tests Analysis
<For each failed test, provide details and correlation with system logs>

### Test-System Event Correlation
<Timeline showing test execution alongside system events, noting timezone differences>

## Identified Problems

### Problem 1: <Problem Title>
**Severity**: Critical/Warning/Info
**Component**: MicroShift/CRI-O/etcd/Pod/OVN/etc.

**Evidence**:
```
<Relevant log excerpts>
```

**Root Cause Analysis**:
<Explanation of what caused this issue>

**Recommendation**:
<How to fix or investigate further>

## Affected Pods/Containers
| Namespace | Pod | Status | Restarts | Issue |
|-----------|-----|--------|----------|-------|
| ... | ... | ... | ... | ... |

## Relevant Log Excerpts
```
<Key error messages from journals>
```

## Configuration Issues
<Any misconfigurations found>

## Next Steps
1. <Recommended action 1>
2. <Recommended action 2>
```

## Examples

### Example 1: Local Directory Analysis
```
/analyze-sos-report /tmp/sosreport-microshift-host-2025-01-15-abcdef
```

### Example 2: Remote URL Analysis
```
/analyze-sos-report https://example.com/sosreport-edge-device-01-2025-01-15.tar.xz
```

### Example 3: SOS Report with Robot Framework Log
```
/analyze-sos-report /tmp/sosreport-el96-host-2025-01-15 https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/test-platform-results/pr-logs/pull/openshift_microshift/5870/pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-arm/1997885403058671616/artifacts/e2e-aws-tests-bootc-arm/openshift-microshift-e2e-metal-tests/artifacts/scenario-info/el96-src@optional/log.html
```

## Notes
- Accepts either a local extracted directory or a remote `.tar.xz` URL
- Remote files are downloaded and extracted to `/tmp` automatically
- Focuses on runtime problem investigation, not system inventory
- Prioritizes actionable findings over comprehensive status reports
- Some directories (crio, logs, networking) are optional and may not be present in minimal reports
- For best results, ensure the sosreport was collected with MicroShift plugins enabled

## CI Environment Considerations

SOS reports from CI environments may contain logs from **multiple MicroShift restarts**. This is expected because:
- CI tests reuse VMs across multiple test scenarios
- Tests may change MicroShift configuration between runs
- The MicroShift service is restarted to apply configuration changes

When analyzing SOS reports:
- **Report all pod restarts** - even if restarts may occur during MicroShift service transitions, report them as potential concerns
- **Report all errors** - include connection refused errors, API server unavailability, and pod failures
- **Flag pods with any restart count > 0** - every restart is worth noting; excessive restarts may indicate pods that don't handle API server unavailability gracefully

### API Server Unavailability During Restarts

When MicroShift restarts, the API server becomes temporarily unavailable. Ideally, pods should:
- Retry API connections with backoff
- Not exit immediately on transient connection failures

Any pod restarts due to API server unavailability should be reported as a concern. Even if the pod eventually recovers, frequent restarts during MicroShift transitions may indicate:
- The pod lacks proper retry/backoff logic for API server connections
- A potential bug in the application's Kubernetes client configuration
- An opportunity to improve pod resilience
