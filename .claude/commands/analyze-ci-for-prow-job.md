---
name: Analyze CI for a Prow Job
argument-hint: <prow-job-url>
description: Download Prow job artifacts, identify root cause of failure, and produce a structured error report
allowed-tools: Skill, Bash, Read, Write, Glob, Grep, Agent
---

# analyze-ci-for-prow-job

## Synopsis
```bash
/analyze-ci-for-prow-job <prow-job-url>
```

## Description
Analyzes a single Prow CI test job by downloading artifacts, scanning for errors, and producing a structured failure report.

## Arguments
- `<prow-job-url>` (required): Full Prow job URL (e.g., `https://prow.ci.openshift.org/view/gs/test-platform-results/logs/periodic-ci-openshift-microshift-release-4.21-periodics-e2e-aws-ovn-ocp-conformance-serial/1984108354347208704`)

## Goal
Reduce noise for developers by processing large logs from a CI test pipeline and correctly classifying fatal errors with a false-positive rate of 0.01% and false-negative rate of 0.5%.

## Audience
Software Engineer

## Glossary

- **ci-config**: Top level configuration file specifying build inputs, versions, and test workflows to execute. Periodic tests are suffixed with `__periodic.yaml`.
- **test**: The set of configurations and commands that specify how to execute the test. Can be defined in-line in ci-config, or as individual "steps" (see below).
- **step-registry**: Root directory where all openshift-ci test step configs and commands are stored.
- **step**: Smallest component of the test infrastructure. A step yaml specifies the command or script to execute, environmental variables and default values, and step metadata. Also called "ref" or "step ref".
- **chain**: A yaml configuration specifying 1 or more steps or chains in an array. Steps and chains are exploded and executed serially by index. May override step environment variable values.
- **workflow**: A yaml configuration specifying 1 or more steps, chains, or workflows in an array. Steps, chains, and workflows are exploded and executed serially. May override chain or step environmental variable values. Typically referenced by a test in a ci-config.
- **scenario**: MicroShift integration tests are built on the robotframework test framework. A "scenario" represents the RF suite, the test's environment, the microshift deployment, and the virtual machine on which the entire testing process takes place. Scenarios also include the manner of deployment: rpm-ostree, rpm installation, or bootc container.

## Job Name and Job ID

The Job Name and Job ID are encoded in the URL. There are two URL formats depending on the job type:

**Periodic/postsubmit jobs:**
```
https://prow.ci.openshift.org/view/gs/test-platform-results/logs/{JOB_NAME}/{JOB_ID}
```
GCS path: `gs://test-platform-results/logs/{JOB_NAME}/{JOB_ID}/`

**Presubmit (PR) jobs:**
```
https://prow.ci.openshift.org/view/gs/test-platform-results/pr-logs/pull/openshift_microshift/{PR_NUMBER}/{JOB_NAME}/{JOB_ID}
```
GCS path: `gs://test-platform-results/pr-logs/pull/openshift_microshift/{PR_NUMBER}/{JOB_NAME}/{JOB_ID}/`

To determine the GCS path from any Prow URL, strip the `https://prow.ci.openshift.org/view/gs/` prefix — the remainder is the GCS bucket path (`gs://` + remainder).

## Important Files
> IMPORTANT! All files in this list will be downloaded after running the `gcloud storage cp -r` command in step 1 of the Workflow.
- `${TMP}/build-log.txt`: Log containing prow job output and most likely place to identify AWS infra related or hypervisor related errors.
- `${STEP}/build-log.txt`: Each step in the CI job is individually logged in a build-log.txt file.
- `${TMP}/artifacts/${TEST_NAME}/openshift-microshift-infra-sos-aws/artifacts/sosreport-*.tar.xz`: Compressed archive containing select portions of the test host's filesystem, relevant logs, and system configurations. `${TEST_NAME}` varies by job (e.g., `e2e-aws-tests`, `e2e-aws-ovn-ocp-conformance-arm64`).
- `${TMP}/artifacts/${TEST_NAME}/openshift-microshift-e2e-origin-conformance/build-log.txt`: Step-specific build log for origin conformance tests.

## Important Links

**Step Diagram URL** (found at the end of the main build-log):
```
https://steps.ci.openshift.org/job?org=openshift&repo=microshift&branch=release-4.19&test=e2e-aws-tests-bootc-nightly&variant=periodics
```
This link provides a diagram of the steps that make up the test. Think about reading this diagram when identifying step failures because not all fatal errors cause the current step to fail but may cause the next step to fail.

**SOS Report** (contains a cross-section of the test host's filesystem, including the microshift journal and container logs)

After downloading artifacts locally, find the SOS report at:
```
${TMP}/artifacts/${TEST_NAME}/openshift-microshift-infra-sos-aws/artifacts/sosreport-*.tar.xz
```
Where `${TEST_NAME}` is the test name directory (e.g., `e2e-aws-tests`, `e2e-aws-ovn-ocp-conformance-serial`). Use `find ${TMP}/artifacts -name 'sosreport-*.tar.xz'` to locate it.

## Work Directory

Set once at the start and reference throughout:
```bash
WORKDIR=/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)
mkdir -p ${WORKDIR}
```

## Common Commands

Create a temporary working directory to store artifacts for the current job:
```bash
mktemp -d ${WORKDIR}/openshift-ci-analysis-XXXX
```

Fetch the high level summary of the failed prow job:
```bash
curl https://prow.ci.openshift.org/view/gs/test-platform-results/logs/periodic-ci-openshift-microshift-release-4.21-periodics-e2e-aws-ovn-ocp-conformance-serial/1984108354347208704 -o ${TMP}/build-log.txt
```

Scan the build log for arbitrary text:
```bash
grep '${SOME_TEXT}' ${GREP_OPTS} ${TMP}/build-log.txt
```

Download all prow job artifacts (derive GCS path by stripping the web prefix from the job URL):
```bash
GCS_PATH=$(echo "${PROW_URL}" | sed 's|https://prow.ci.openshift.org/view/gs/|gs://|')
gcloud storage cp -r "${GCS_PATH}/" ${TMP}/
```

## Workflow

The user argument is: $ARGUMENTS

0. Create and use a temporary working directory. Use the mktemp -d command to create this directory.

1. **Download all artifacts**: Download all prow job artifacts using `gcloud storage cp -r` into the temporary working directory. Derive the GCS path by stripping the web prefix from the Prow URL:
   ```bash
   GCS_PATH=$(echo "${PROW_URL}" | sed 's|https://prow.ci.openshift.org/view/gs/|gs://|')
   gcloud storage cp -r "${GCS_PATH}/" ${TMP}/
   ```
   This works for both periodic (`logs/...`) and presubmit PR (`pr-logs/pull/...`) job URLs.
   This makes all build logs, step logs, and SOS reports available locally for analysis.

2. **Scan for errors**: Start by scanning the top level `build-log.txt` file for errors and determine the step where the error occurred. Record each error with the filepath and line number for later reference.

3. **Read context**: Iterate over each recorded error, locate the log file and line number, then read 50 lines before and 50 lines after the error. Use this information to characterize the error. Think about whether this error is transient and think about where in the stack the error occurs. Does it occur in the cloud infra, the openshift or prow ci-config, the hypervisor, or is it a legitimate test failure? If it is a legitimate test failure, determine what stage of the test failed: setup, testing, teardown.

4. **Analyze the error**: Based on the context of the error, think hard about whether this error caused the test to fail, is a transient error, or is a red herring.

    4.1 If it is a legitimate test error, analyze the test logs to determine the source of the error.
    4.2 If the source of the error appears to be due to microshift or a workload running on microshift, analyze the sos report's microshift journal and pod logs.

5. **Produce a report**: Create a concise report of the error. The report MUST specify:
   - Where in the pipeline the error occurred
   - The specific step the error occurred in
   - Whether the test failure was legitimate (i.e., a test failed) or due to an infrastructure failure (i.e., build image was not found, AWS infra failed due to quota, hypervisor failed to create test host VM, etc.)

## Prerequisites

- `gcloud` CLI must be installed and authenticated for GCS access
- Internet access to fetch job data from Prow/GCS
- Bash shell

## Tips

1. There are many setup and teardown stages so fatal errors may be buried by log output from the teardown phase. It is not common to find the fatal error at the end of the log.
2. You can quickly determine the failed step from the build-log.txt by reading the last `Running step e2e-aws-tests-bootc-nightly-openshift-microshift-e2e-metal-tests` line before the container logs appear.

## Output Template

Use this template for your error analysis reports:

```
Error Severity: {1-5}
Stack Layer: {AWS Infra, External Infrastructure, build phase, deploy phase, test setup phase, Test Configuration, test, teardown}
Step Name: {The specific step where the error occurred}
Error: {The exact error, including additional log context if it relates to the failure}
Suggested Remediation: {Based on where the error occurs, think hard about how to correct the error ONLY if it requires fixing. Infrastructure failures may not require code changes.}
```

After the human-readable report above, append a machine-readable block for downstream automation. This block MUST appear at the very end of the report, after all prose and analysis:

```text
--- STRUCTURED SUMMARY ---
SEVERITY: {1-5, same as Error Severity above}
STACK_LAYER: {AWS Infra, External Infrastructure, build phase, deploy phase, test setup phase, Test Configuration, test, teardown - same as Stack Layer above}
STEP_NAME: {same as Step Name above}
ERROR_SIGNATURE: {a concise, unique one-line description of the root cause - not the full error, just enough to identify and deduplicate this failure}
INFRASTRUCTURE_FAILURE: {true if Stack Layer is AWS Infra or the failure is due to CI infrastructure rather than product code, false otherwise}
JOB_URL: {the full prow job URL that was analyzed}
JOB_NAME: {the full job name extracted from the URL}
RELEASE: {the branch extracted from the URL, e.g. 4.22 from release-4.22, or main}
FINISHED: {the job finish date in YYYY-MM-DD format, extracted from finished.json or build log timestamps}
--- END STRUCTURED SUMMARY ---
```

The ERROR_SIGNATURE field is critical for deduplication. It should capture the essence of the failure in a way that two jobs failing for the same reason produce identical or near-identical signatures. Examples:
- `greenboot timeout waiting for MicroShift to start after bootc upgrade`
- `OCP conformance test NetworkPolicy timeout`
- `AWS EC2 quota exceeded in us-east-1`
- `rpm-ostree upgrade failed: package conflict microshift-selinux`
