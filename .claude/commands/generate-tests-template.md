# Test Generator Template

This is a **generic, customizable template** for creating test generation workflows. It's designed to be adapted for different projects, testing frameworks, and workflows.

## Why Use This Template?

The pre-configured `/generate-tests` command is optimized for MicroShift/OpenShift projects with specific assumptions about:
- Jira OCPSTRAT tickets
- Robot Framework test structure
- MicroShift repository layout
- Red Hat workflow conventions

**This template allows you to:**
- Adapt the workflow for your specific project
- Use different testing frameworks (Pytest, Jest, Ginkgo, etc.)
- Customize for different issue trackers (GitHub, Jira, Linear, etc.)
- Work with Claude Code, Cursor, or any AI assistant
- Match your team's conventions and preferences

## How to Use This Template

### Option 1: Direct Prompting (Recommended for Cursor users)
1. **Copy the relevant sections** from this template
2. **Customize them** for your project
3. **Paste into your AI assistant** as context with your request

**Example**:
```
I want to generate tests for my project. Here's my setup:
- Project: MyAPI
- Issue tracker: GitHub Issues (https://github.com/myorg/myapi/issues/)
- Test framework: Pytest
- Directory structure: tests/unit/, tests/integration/

[paste customized template sections]

Please generate tests for issue #123
```

### Option 2: Create a Custom Slash Command (For Claude Code)
1. **Copy this file** to create your own command (e.g., `.claude/commands/my-generate-tests.md`)
2. **Customize the sections** below to match your project
3. **Use the command**: `/my-generate-tests TICKET-ID`

### Option 3: Use as Reference
Simply reference sections of this template when asking your AI assistant to generate tests, adapting on the fly.

---

## Template Sections to Customize

### 🎯 Project Context (CUSTOMIZE THIS)
```
Replace with your project details:
- Project name: [MicroShift/OpenShift/Your Project]
- Issue tracker: [Jira/GitHub Issues/Linear/etc.]
- Repository structure: [mono-repo/multi-repo]
- Test framework: [Robot Framework/Pytest/Jest/etc.]
```

### 📋 Ticket Analysis (CUSTOMIZE THIS)
```
Define your ticket workflow:
- Ticket URL pattern: https://issues.redhat.com/browse/$1
- Required ticket fields: [Description, Acceptance Criteria, etc.]
- Linked tickets to check: [Implementation tickets, dependent tickets]
- Where to find PR links: [Ticket fields, comments, linked items]
```

### 🔍 Test Coverage Analysis (CUSTOMIZE THIS)
```
Define what to analyze:
- Test file patterns: *.robot, *_test.py, *.spec.js
- Test directories: test/, tests/, __tests__/
- Coverage sources: [PRs, existing test suites, test plans]
- What constitutes "covered": [Unit tests, integration tests, etc.]
```

### ✍️ Test Case Generation (CUSTOMIZE THIS)
```
Define your test case format:
- Test case document format: [Markdown, Confluence, etc.]
- Number of test cases: [Top 10, All scenarios, etc.]
- Test case fields:
  * Test ID format: USHIFT-XXX-TC-YYY
  * Priority levels: Critical/High/Medium/Low
  * Required sections: Description, Steps, Expected Results, etc.
- Output location: [Current directory, specific folder, etc.]
```

### 🧪 Test Categories to Cover (CUSTOMIZE THIS)
```
Define categories relevant to YOUR project:
1. Core Functionality
2. Configuration Edge Cases
3. Dynamic Behavior
4. Integration Points
5. Multi-tenant/Namespace
6. Error Handling
7. Upgrade/Compatibility
8. Performance
9. Security
10. Real Customer Scenarios

Add/remove categories as needed for your domain.
```

### 🤖 Test Implementation (CUSTOMIZE THIS)
```
Define your test structure:
- Test framework: Robot Framework / Pytest / Jest / etc.
- File naming convention: kebab-case / snake_case / PascalCase
- Directory structure: test/suites/<category>/ or tests/unit/
- Reusable components location: resources/, utils/, fixtures/
- Setup/Teardown patterns: Your project's patterns
```

### 📁 File Organization (CUSTOMIZE THIS)
```
Map feature types to test directories:
- Network features → test/suites/network/
- Storage features → test/suites/storage/
- API features → test/api/
- UI features → test/e2e/

Customize based on YOUR project structure.
```

### 🔧 Keyword/Helper Reuse (CUSTOMIZE THIS - Robot Framework specific)
```
If using Robot Framework, define:
- Common keyword locations: test/extended/util/*.robot
- Resource file patterns: resources/**/*.robot
- Naming conventions: Action_Object pattern
- When to create new vs. reuse: Your team's guidelines
```

### 🌿 Git Workflow (CUSTOMIZE THIS)
```
Define your branching strategy:
- Branch naming: test-OCPSTRAT-XXXX-$(date +%Y%m%d) or feature/tests-XXX
- Base branch: main / master / develop
- Commit message format: Your team's convention
- Auto-commit: Yes/No (get confirmation first)
- Repository location: Ask user / fixed path / auto-detect
```

### 📊 Output Format (CUSTOMIZE THIS)
```
Define what the final output should include:
- Test case document: Format and location
- Test implementation files: Locations and count
- Coverage report: What was tested vs. what's missing
- Summary report: Tickets analyzed, PRs reviewed, etc.
- Next steps: Instructions for running tests, creating PR, etc.
```

---

## Example Customized Prompt

Here's an example of how to customize this template for a different project:

### Example: Python/Pytest Project

```markdown
# Pytest Test Generator for MyAPI Project

Generate comprehensive Pytest coverage for MyAPI features based on GitHub Issues.

## Workflow:

### Step 1: Analyze GitHub Issue
- Fetch issue from: https://github.com/myorg/myapi/issues/$1
- Extract feature description and acceptance criteria
- Find linked PRs in issue body and comments

### Step 2: Analyze Existing Tests
- Check for test files: tests/**/*_test.py
- Look for pytest fixtures in: tests/conftest.py
- Identify what's already covered

### Step 3: Generate Test Cases
- Create markdown document: test_cases_issue_$1.md
- Generate 5-7 critical test scenarios
- Format: Given/When/Then style
- Include: unit tests, integration tests, edge cases

### Step 4: Implement Pytest Tests
- Create test file: tests/unit/test_<feature>.py
- Reuse fixtures from conftest.py
- Follow pytest best practices
- Use parametrize for multiple scenarios

### Step 5: Update conftest.py if needed
- Add new fixtures only if truly reusable
- Document fixture purpose and scope

### Step 6: Create PR (optional)
- Branch: feature/tests-issue-$1
- Run: pytest tests/ to verify
- Provide summary of coverage added
```

---

## Instructions for AI Assistants

When a user provides this template:

1. **Ask clarifying questions** about sections marked "CUSTOMIZE THIS"
2. **Don't assume** - let the user specify their preferences
3. **Adapt the workflow** to the user's specific tools and conventions
4. **Confirm before creating files/branches** - respect the user's workflow
5. **Be flexible** - this is a starting point, not a rigid spec

---

## Tips for Customization

### For Different Testing Frameworks
- **Pytest**: Focus on fixtures, parametrize, conftest.py
- **Jest**: Focus on mocks, setup/teardown, snapshot testing
- **Robot Framework**: Focus on keywords, resources, libraries
- **Selenium**: Focus on page objects, waits, selectors
- **Postman/REST**: Focus on request builders, assertions, environments

### For Different Issue Trackers
- **GitHub Issues**: Use GitHub API or web scraping
- **Jira**: Adjust ticket URL pattern, field names
- **Linear**: Different API structure
- **Azure DevObs**: Work items structure

### For Different Project Types
- **Microservices**: Multi-repo coordination, API contracts
- **Frontend**: Component testing, accessibility, visual regression
- **CLI Tools**: Command execution, output validation, error handling
- **Libraries**: API testing, documentation examples, edge cases

---

## License

This template is provided as-is for customization. Adapt freely to your needs.
