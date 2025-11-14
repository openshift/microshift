---
name: Test Generator for MicroShift Features (Detailed test cases & Robot framework code)
argument-hint: JIRA-TICKET-ID
description: Generate Test Cases for a given Jira OCPSTRAT ticket & automates them using RobotFramework
allowed-tools: WebFetch, Bash, Read, Write, Glob, Grep
interaction: Ask user for confirmation before creating git branch
---

# Robot Framework Test Generator for MicroShift/OpenShift

**NOTE**: This is a **pre-configured template** for MicroShift/OpenShift projects. For other projects or to customize this workflow, see `.claude/commands/generate-tests-template.md` for a generic template you can adapt.

Generate comprehensive Robot Framework test coverage for MicroShift/OpenShift features based on Jira OCPSTRAT tickets.

## Prerequisites:

**Jira Access**: This command requires access to Red Hat Jira (https://issues.redhat.com).

- Ensure you're logged into Red Hat Jira in your browser before running this command
- Login at: https://issues.redhat.com

**Fallback**: If Jira access fails, you'll be prompted to manually provide:
- Ticket description and acceptance criteria
- Links to GitHub PRs
- Feature documentation

## Workflow:

### Step 1: Analyze OCPSTRAT Ticket
- Fetch the Jira ticket from https://issues.redhat.com/browse/$1
- Extract feature description, acceptance criteria, and technical details
- Identify all linked USHIFT tickets (look in "Issue Links" section)
- Find GitHub PR links from USHIFT tickets or OCPSTRAT ticket

**VALIDATION - Check for sufficient information**:
After analyzing the ticket, verify:
1. **Description Quality**: Does the ticket have detailed description, acceptance criteria, or technical implementation details?
2. **PR Links**: Are there any useful GitHub PR links in the OCPSTRAT ticket or linked USHIFT tickets?

**If BOTH conditions fail** (no detailed description AND no PR links):
- ⚠️ **Ask user for additional information**:
  "The OCPSTRAT ticket lacks detailed information for test generation. It's missing:
  - Detailed feature description or acceptance criteria
  - Linked USHIFT tickets with GitHub PRs

  Please provide ONE of the following to continue:
  1. A link to documentation (PR description, GitHub .md files, design docs)
  2. A brief description of the feature functionality and expected behavior
  3. Additional context about what this feature does

  Or type 'skip' to proceed with limited information (may result in generic tests)."

- **Wait for user response**:
  - If user provides documentation link: Fetch and analyze it
  - If user provides description: Use it as supplementary context
  - If user types 'skip': Proceed with warning that tests may be generic

**If at least ONE condition passes**, proceed to Step 2.

### Step 2: Analyze Existing Test Coverage
- For each GitHub PR found, check if it contains Robot Framework test files (*.robot)
- Note the test file names and their location in the PR
- Read and analyze existing test cases to understand what's already covered
- Identify functionality already covered by existing tests

### Step 3: Generate Top 10 Most Impactful Missing Tests

**Create test case document in current working directory**: `test_cases_OCPSTRAT-XXXX.md`

Focus ONLY on the **10 most impactful test scenarios NOT already covered**:

1. **Core Functionality** - Primary use cases not tested
2. **Configuration Edge Cases** - Invalid/boundary configurations not covered
3. **Dynamic Behavior** - Runtime changes, reloads not tested
4. **Integration Gaps** - Component interactions not validated
5. **Multi-tenant/Namespace** - Cross-namespace scenarios missing
6. **Error Handling** - Failure modes not covered
7. **Upgrade/Compatibility** - Version compatibility gaps
8. **Performance** - Load/scale testing if missing
9. **Security** - Permission/isolation tests not present
10. **Real Customer Scenarios** - Use cases from RFE not tested

For each test, provide:
- **Test ID**: USHIFT-XXX-TC-YYY (using the USHIFT ticket number)
- **Priority**: Critical/High/Medium
- **Coverage Gap**: What existing tests don't cover
- **Test Description**: Clear objective
- **Steps to Reproduce**: Concrete test steps with actual config/commands
- **Expected Results**: Expected outcome for each step

### Step 4: Create Robot Framework Test Files

**BEFORE writing any tests:**
1. Search for existing keyword files in the codebase
2. Identify reusable keywords for common operations
3. Check for existing resource files and utilities
4. Use Grep/Glob to find similar test patterns

Common keyword locations to check:
- `test/extended/util/*.robot` - Utility keywords
- `resources/**/*.robot` - Shared resource files
- Existing test files in same feature area

**When creating tests:**
- ✅ **REUSE** existing keywords whenever possible
- ❌ **DON'T** create duplicate keywords that already exist
- ✅ **EXTEND** existing keyword files if new keywords are truly needed
- ❌ **DON'T** reinvent common operations (oc commands, pod management, etc.)

**Robot Framework Test File Creation Process:**
- Generate test case documentation in current working directory as `test_cases_OCPSTRAT-XXXX.md`
- Robot Framework test files will be created in microshift repository AFTER git branch is created in Step 5
- Follow microshift repo naming convention: lowercase with hyphens (kebab-case)
  - Example files in microshift: `backup-restore-on-reboot.robot`, `multi-nic.robot`, `isolated-network.robot`
- If Robot Framework tests exist in PRs, create files with matching names
- If no existing Robot tests, create new file in appropriate `test/suites/<category>/` subdirectory:
  - For DNS/network features: `test/suites/network/<feature-name>.robot` (e.g., `coredns-hosts-file.robot`)
  - Use feature-based naming: `<feature-name>.robot` (kebab-case, lowercase)
- Implement the top 3-5 most critical missing tests in Robot Framework
- Follow existing patterns from similar test files in microshift codebase
- Include proper setup, teardown, and error handling

**CRITICAL Robot Framework Teardown Rules** (from `.cursor/rules/robot-framework-teardown.md` in the microshift repository):
- ✅ **DO**: Write clean teardowns without `Run Keyword And Ignore Error`
- ❌ **DON'T**: Use `Run Keyword And Ignore Error` in ANY teardown section
- Robot Framework teardowns automatically continue execution even if keywords fail
- This applies to: `[Teardown]`, `Suite Teardown`, `Test Teardown`, and keyword teardowns
- Only use `Run Keyword And Ignore Error` in test cases/setup when you need conditional logic

**Correct Teardown Example**:
```robot
Suite Teardown    Cleanup Test Resources

Cleanup Test Resources
    [Documentation]    Clean up all test resources
    Oc Delete    pod --all -n ${NAMESPACE}
    Remove Test Files
    Restart MicroShift
```

**Incorrect Teardown (DON'T DO THIS)**:
```robot
Suite Teardown    Run Keyword And Ignore Error    Cleanup Test Resources

Cleanup Test Resources
    Run Keyword And Ignore Error    Oc Delete    pod --all -n ${NAMESPACE}
    Run Keyword And Ignore Error    Remove Test Files
```

**Keyword Reuse Example**:
```robot
# ✅ GOOD - Reusing existing keywords
*** Test Cases ***
Test CoreDNS Hosts File
    [Documentation]    Test hosts file resolution
    Setup MicroShift Config    dns.hostsFile=/etc/custom-hosts
    Restart MicroShift
    Wait For MicroShift Ready
    Deploy Test Pod    dns-test    ${NAMESPACE}
    ${ip}=    Resolve Hostname In Pod    dns-test    ${NAMESPACE}    test.local
    Should Be Equal    ${ip}    192.168.1.100

# ❌ BAD - Creating duplicate keywords
*** Keywords ***
My Custom MicroShift Restart
    [Documentation]    Restart MicroShift (DON'T DO THIS - use existing keyword!)
    Run Process    systemctl    restart    microshift
    Sleep    10s
```

### Step 5: Create Git Branch and Robot Framework Tests in MicroShift Repository

**IMPORTANT**: Before creating git branch, **WAIT for user confirmation**.

Ask user for confirmation:

**Prompt user**: "Would you like me to create a new git branch in the microshift repository? The branch will be named `test-OCPSTRAT-XXXX-$(date +%Y%m%d)`. (yes/no)"

**If user confirms (yes)**:

#### 5.1: Create Git Branch
1. Ask user for the microshift repository path on their system
2. Navigate to the microshift repository
3. Check current git status
4. Create new branch: `test-OCPSTRAT-XXXX-$(date +%Y%m%d)`
5. Confirm branch creation and provide branch name

**Prompt user**: "Please provide the full path to your microshift repository (e.g., /home/username/microshift):"

**Commands** (use the path provided by user):
```bash
cd <user-provided-microshift-repo-path>
git status
git checkout -b test-OCPSTRAT-XXXX-$(date +%Y%m%d)
git branch --show-current
```

#### 5.2: Create Robot Framework Test Files in the New Branch

After branch is created, create Robot Framework test file directly in the microshift repository:

1. Determine appropriate test suite directory based on feature type
   - Network/DNS features: `test/suites/network/`
   - Storage features: `test/suites/storage/`
   - Backup features: `test/suites/backup/`
   - etc.

2. Create the .robot file with kebab-case naming convention
   - Example: `test/suites/network/coredns-hosts-file.robot`

3. Implement the top 3-5 most critical tests from Step 3

4. Add file to git staging

**Commands** (use the user-provided microshift repository path):
```bash
# File will be created at appropriate path, e.g.:
# <microshift-repo-path>/test/suites/network/<feature-name>.robot

# After file is created, stage it
cd <user-provided-microshift-repo-path>
git add test/suites/<category>/<feature-name>.robot
git status
```

**Output to user**:
- Branch created: test-OCPSTRAT-XXXX-YYYYMMDD
- Repository: <user-provided-microshift-repo-path>
- Current branch: [branch name]
- Robot Framework test file created: test/suites/<category>/<feature-name>.robot
- File staged for commit: Ready to commit

### Step 6: Summary Report
Provide:
- List of USHIFT tickets and PRs analyzed
- Existing test files found (with coverage summary)
- Test case document (10 test cases) - filename and path
- Robot Framework test files created (with test count) - filename and path
- Git branch created (branch name and repository path) - if created
- Coverage gaps addressed
- Remaining gaps (if any)
