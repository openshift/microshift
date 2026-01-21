---
name: microshift-add-enhancement
description: Create a new Microshift Enhancement Proposal
argument-hint: [area] <name> <description> <jira>
model: sonnet
color: blue
---

## Name
microshift:add-enhancement

## Synopsis
```
/microshift:add-enhancement [area] <name> <description> <jira>
```

## Description

The `microshift:add-enhancement` command creates a new microshift Enhancement
Proposal (EP) based on the official template from the openshift/enhancements
repository. It generates a comprehensive enhancement document with all required
sections, metadata, and guidance following OpenShift's enhancement process.

This command automates the creation of enhancement proposals by:
- Fetching the latest enhancement template from the openshift/enhancements
  repository
- Analyzing the provided description to extract what, why, and who information
- Generating user stories, goals, and non-goals based on the requirements
- Creating properly formatted enhancement files with all required sections
- Applying OpenShift-specific conventions for feature gates, API design, and
  testing requirements

## Arguments

- **area** (optional): Enhancement area (subdirectory under enhancements/). If not provided, the command will attempt to infer the best area from the enhancement description and ask for confirmation.
- **name**: One-line title describing the enhancement
- **description**: Detailed description (what, why, who)
- **jira**: JIRA ticket URL for tracking

## Implementation

Act as an experienced software architect to create a comprehensive enhancement proposal. Follow these steps:

**Important**: Reference the guidance in the OpenShift enhancements repository at https://github.com/openshift/enhancements/blob/master/dev-guide/feature-zero-to-hero.md, particularly the section "Writing an OpenShift Enhancement", when creating enhancement proposals. This guide provides essential context on the OpenShift Enhancement Proposal process, feature gates, API design conventions, testing requirements, and promotion criteria.

0. **Verify Repository**: Before proceeding, ensure you are operating in the correct repository:
   - Check if the current directory is a git repository: `git rev-parse --is-inside-work-tree`
   - Verify the repository is openshift/enhancements or a fork:
     - Get the remote URL: `git config --get remote.origin.url`
     - Check if the URL contains "openshift/enhancements" (e.g., `git@github.com:openshift/enhancements.git` or `https://github.com/openshift/enhancements.git`)
     - For forks, the URL might be different (e.g., `git@github.com:username/enhancements.git`) - in this case, check if it's a fork by examining the repository structure (presence of `enhancements/` directory and `guidelines/` directory with enhancement_template.md)
   - If verification fails, exit with a clear error message: "Error: This command must be run in the openshift/enhancements repository or a fork. Please clone the repository first or navigate to the correct directory."
   - If successful, continue to the next step

1. **Determine Enhancement Area**: Identify or validate the area where the enhancement should be created:
   - List available areas by examining subdirectories in `enhancements/microshift`: `ls -d enhancements/microshift/*/` (extract just the directory names)
   - **If the `<area>` argument WAS provided by the user:**
     - Check if `enhancements/microshift/<area>/` exists
     - If it exists: proceed with this area
     - If it does NOT exist:
       - Show the list of available areas
       - Use AskUserQuestion to ask: "The area '<area>' does not exist. Do you want to create a new area called '<area>'?" with options:
         - "Yes, create new area '<area>'" (then confirm they're sure about creating a new area)
         - "No, let me choose from existing areas" (then show available areas and ask them to select)
   - **If the `<area>` argument was NOT provided:**
     - Analyze the enhancement name and description to infer the most appropriate area
     - Look for keywords that match existing area names (e.g., "storage", "router", "apiserver", "coredns", "etcd", etc.)
     - Use AskUserQuestion to confirm: "Based on your enhancement description, I think the best area is '<inferred-area>'. Is this correct?" with options:
       - "Yes, use '<inferred-area>'"
       - "No, let me choose a different area" (then list available areas and ask them to select or specify a new one)
     - If the user chooses a different area that doesn't exist, ask if they want to create it (same flow as above)
   - **Creating a new area (if chosen):**
     - Create  markdown document in the new area : `touch enhancements/microshift/<new-area>.md`
   - Proceed with the validated/created area

2. **Fetch the Enhancement Template**: Before starting, fetch the latest template from the openshift/enhancements repository:
   - Download the template via HTTP using whatever web-fetch mechanism is available; otherwise ask the user to paste the template from: https://raw.githubusercontent.com/openshift/enhancements/master/guidelines/enhancement_template.md
   - Store this template for reference when creating the enhancement file
   - This ensures you always use the most current template structure and required sections

3. **Parse the Description**: Extract the following from the description:
   - **What**: What is this enhancement about
   - **Why**: Why this change is required (motivation)
   - **Who**: Which personas this applies to (use this to generate user stories)

4. **Ask Clarifying Questions** (if needed): Use the AskUserQuestion tool to gather:
   - Specific user stories or motivations if not clear from the description
   - Explicit Goals or Non-Goals the user wants to be included
   - Any specific technical constraints or requirements
   - Topology considerations should be MicroShift
   - Whether this proposal adds/changes CRDs, admission and conversion webhooks, ValidatingAdmissionPlugin, MutatingAdmissionPlugin, aggregated API servers, or finalizers (needed for API Extensions section)
   - Feature gate information: According to the OpenShift enhancement dev guide (https://github.com/openshift/enhancements/blob/master/dev-guide/feature-zero-to-hero.md), ALL new OpenShift features must start disabled by default using feature gates. Ask about the proposed feature gate name and initial feature set (DevPreviewNoUpgrade or TechPreviewNoUpgrade).
   - Ask clarifying questions about telemetry, security, upgrade and downgrade process, rollbacks, dependencies, in case it is not possible to assert these fields.

5. **Generate the Enhancement File**:
   - Create the file at `enhancements/microshift/<area>.md` where filename is the kebab-case version of the name argument
   - Fill in the template with:
     - **Title**: Use the provided name
     - **Summary**: One paragraph describing what this enhancement is about
     - **Motivation**: Explain why this change is required based on the description
     - **User Stories**: Generate 2-4 user stories based on the "who" information using the format:
       > "As a _role_, I want to _take some action_ so that I can _accomplish a goal_."
       Include a story on how the proposal will be operationalized: life-cycled, monitored and remediated at scale.
     - **Goals**: List specific, measurable goals (3-5 items). Goals should describe what users want from their perspective, not implementation details.
     - **Non-Goals**: List what is explicitly out of scope (2-3 items)
     - **Proposal**: High-level description of the proposed solution
     - **Workflow Description**: Detailed workflow with actors and steps
     - **Mermaid Diagram**: Add a sequence diagram when the workflow involves multiple actors or complex interactions between components (e.g., user -> API server -> controller -> operator). Simple single-actor workflows may not need a diagram.
     - **API Extensions**: Only fill this section if the user confirms the proposal adds/changes CRDs, admission and conversion webhooks, ValidatingAdmissionPlugin, MutatingAdmissionPlugin, aggregated API servers, or finalizers. Per the template, name the API extensions and describe if this enhancement modifies the behaviour of existing resources. Otherwise, add a TODO comment asking the user to complete this section if applicable.
     - **Topology Considerations**: Include subsections for Hypershift/Hosted Control Planes, Standalone Clusters, Single-node Deployments or MicroShift, and OKE (OpenShift Kubernetes Engine). Address how the proposal affects each topology.
     - **Implementation Details/Notes/Constraints**: Provide a high-level overview of the code changes required. Follow the guidance from the template: "While it is useful to go into the details of the code changes required, it is not necessary to show how the code will be rewritten in the enhancement." Keep it as an overview; the developer should fill in the specific implementation details. Include a reminder about creating a feature gate: Per the OpenShift dev guide (https://github.com/openshift/enhancements/blob/master/dev-guide/feature-zero-to-hero.md), all new features must be gated behind a feature gate in https://github.com/openshift/api/blob/master/features/features.go with the appropriate feature set (DevPreviewNoUpgrade or TechPreviewNoUpgrade initially).
     - **Test Plan**: Add a TODO comment with guidance on required test labels per the OpenShift dev guide (https://github.com/openshift/enhancements/blob/master/dev-guide/feature-zero-to-hero.md): Tests must include `[OCPFeatureGate:FeatureName]` label for the feature gate, `[Jira:"Component Name"]` for the component, and appropriate test type labels like `[Suite:...]`, `[Serial]`, `[Slow]`, or `[Disruptive]` as needed. Reference the test conventions guide (https://github.com/openshift/enhancements/blob/master/dev-guide/test-conventions.md) for details.
     - **Graduation Criteria**: Add a TODO comment referencing the specific promotion requirements from the OpenShift dev guide (https://github.com/openshift/enhancements/blob/master/dev-guide/feature-zero-to-hero.md): minimum 5 tests, 7 runs per week, 14 runs per supported platform, 95% pass rate, and tests running on all supported platforms (AWS, Azure, GCP, vSphere, Baremetal with various network stacks).
     - **Metadata**: Fill in creation-date with today's date, tracking-link with the provided JIRA ticket URL, set other fields to TBD. For api-approvers: use "None" if there are no API changes (no new/modified CRDs, webhooks, aggregated API servers, or finalizers); otherwise use "TBD" as a placeholder (the enhancement author will request an API reviewer from the #forum-api-review Slack channel later).

6. **Handle Unfilled Sections**: For sections that cannot be filled based on the input:
   - Add a clear comment like `<!-- TODO: This section needs to be filled in -->`
   - Provide guidance on what should be included

7. **Writing Guidelines**:
   - Write in a clear, concise, professional manner
   - Focus on the essential information
   - Use bullet points and structured formatting
   - Avoid unnecessary verbosity
   - **Line Length**: Keep lines in the generated enhancement at a maximum of 80 characters, but prioritize validity over line length limits. Only break lines at 80 characters if doing so will NOT create:
     - Invalid or broken URLs (URLs themselves should never be split, but the line CAN and SHOULD be broken before or after the URL)
     - Invalid Markdown syntax (e.g., breaking Markdown links, code blocks, or formatting)
     - Invalid code examples (e.g., breaking code in the middle of statements)
     If breaking at 80 characters would split a URL, code, or Markdown syntax, find the nearest valid break point such as: after a sentence, before a URL starts, after a URL ends, or at a natural paragraph break. For regular prose, it is acceptable to exceed 80 characters by 10-15 characters to avoid breaking words mid-word. Only allow lines >95 characters when the line contains a single unbreakable element (like a standalone URL with no surrounding text, or a single line of code).

8. **Validate**:
   - Create a valid filename from the name (lowercase, replace spaces with dashes)
   - Verify all required YAML metadata is present
   - Verify the JIRA ticket URL is included in the tracking-link metadata field
   - Ensure the enhancement file is created in the correct path: `enhancements/microshift/<area>.md`

## Output

After creating the enhancement file, provide:
- The full path to the created file
- A brief summary of what was included
- A list of sections that need further attention (marked with TODO comments)

Begin by analyzing the inputs and asking any necessary clarifying questions before generating the enhancement proposal.