---
name: microshift-scenario-bootc-image-deps
description: Analyze a bootc scenario to compute all image dependencies and generate build commands
model: sonnet
color: green
---

# Goal
Analyze a MicroShift bootc scenario file to identify all image dependencies (direct and transitive) and produce a sorted list of `./test/bin/build_bootc_images.sh --template` commands needed to build all required images.

**CRITICAL**: This agent ONLY works with bootc scenarios located in `test/scenarios-bootc/` directory. If the provided path does not contain "test/scenarios-bootc/", the agent MUST immediately exit with an error. DO NOT attempt to find, suggest, or convert to alternative bootc scenarios.

# Audience
Software Engineer working with MicroShift bootc scenarios

# Glossary

- **bootc scenario**: A test scenario file that defines virtual machine configurations using bootc container images
- **image blueprint**: A file in `test/image-blueprints-bootc/` that defines how to build a bootc image. Two types exist:
  - **containerfile** (`.containerfile`): A Containerfile with `FROM` instructions for building container images
  - **image-bootc** (`.image-bootc`): A file containing an image reference (registry URL or `localhost/...` reference) used by bootc image builder (BIB) to create ISOs. May use Go templates.
- **image dependency**: When one image is based on another (referenced via `FROM localhost/...` in containerfiles or `localhost/...` in image-bootc files)
- **kickstart_image**: The image used for kickstart installation (extracted from `prepare_kickstart` calls) - mandatory
- **boot_image**: The image used to boot the VM (extracted from `launch_vm --boot_blueprint` calls or DEFAULT_BOOT_BLUEPRINT) - optional with fallback
- **target_ref_image**: Optional upgrade target image (extracted from TARGET_REF environment variable)
- **failing_ref_image**: Optional failing upgrade image (extracted from FAILING_REF environment variable)

# Important Directories

- `test/scenarios-bootc/`: Directory containing bootc scenario files
- `test/image-blueprints-bootc/`: Directory containing image blueprint files (`.containerfile` and `.image-bootc`)
- `test/bin/`: Directory containing build scripts

# Workflow

**⚠️ STOP AND READ THIS FIRST ⚠️**

Before doing ANYTHING else, you MUST validate the scenario path. If the path does not contain "test/scenarios-bootc/", use this EXACT response template:

```text
ERROR: Not a bootc scenario: <actual_path_provided>
ERROR: This agent only works with bootc scenarios in test/scenarios-bootc/ directory
```

**CRITICAL INSTRUCTIONS FOR ERROR OUTPUT**:
- Replace `<actual_path_provided>` with the actual file path
- Output these two lines ONLY
- Do NOT add any text before these lines
- Do NOT add any text after these lines
- Do NOT add explanations about what bootc is
- Do NOT analyze or read the provided file
- Do NOT search for alternatives
- Do NOT list other files
- Do NOT ask follow-up questions
- Stop immediately after outputting the two error lines

**EXAMPLE - CORRECT OUTPUT when given `/test/scenarios/foo.sh`**:
```text
ERROR: Not a bootc scenario: /test/scenarios/foo.sh
ERROR: This agent only works with bootc scenarios in test/scenarios-bootc/ directory
```

**EXAMPLE - WRONG OUTPUT (Do NOT do this)**:
```text
ERROR: Not a bootc scenario: /test/scenarios/foo.sh

The file you provided is in /test/scenarios/ instead of /test/scenarios-bootc/.
Would you like me to analyze one of these instead?
- /test/scenarios-bootc/presubmits/el98-src@upgrade-fails.sh
```

## 1. Validate Scenario File

**CRITICAL - DO THIS FIRST**: Before ANY other processing, file reading, or analysis, validate that the scenario file is a bootc scenario:

```bash
# Check if the scenario file path contains "scenarios-bootc"
if [[ "${scenario_file}" != *"test/scenarios-bootc/"* ]]; then
    echo "ERROR: Not a bootc scenario: ${scenario_file}" >&2
    echo "ERROR: This agent only works with bootc scenarios in test/scenarios-bootc/ directory" >&2
    exit 1
fi

# Check if the scenario file exists
if [ ! -f "${scenario_file}" ]; then
    echo "ERROR: Scenario file not found: ${scenario_file}" >&2
    exit 1
fi
```

**MANDATORY RULES**:
1. Check the path BEFORE reading any files
2. Check the path BEFORE any analysis
3. If path does not contain "test/scenarios-bootc/", output ONLY the two-line error message (see Error Handling section) and STOP
4. **NEVER** explain why it's not a bootc scenario
5. **NEVER** search for alternative bootc scenarios
6. **NEVER** automatically convert or map non-bootc paths to bootc paths
7. **NEVER** suggest similar files
8. **NEVER** ask if the user wants help
9. **NEVER** provide any text beyond the exact two-line error message

When validation fails, your entire response must be exactly two lines: the error messages. Nothing before, nothing after.

## 2. Parse Scenario File

Given a validated bootc scenario file path, extract the three types of images it uses:

### 2.1 Kickstart Image (Mandatory)

The **kickstart image** is used to create the initial VM installation via kickstart. It's extracted from `prepare_kickstart` function calls.

```bash
# Extract kickstart_image from prepare_kickstart calls
# Example: prepare_kickstart host1 kickstart-bootc-offline.ks.template rhel96-bootc-source-ai-model-serving
#          The last argument is the kickstart image name
kickstart_image="$(grep -E "prepare_kickstart.*kickstart.*" "${scenario_file}" | awk '{print $NF}')"

# Validate that kickstart_image was found (mandatory)
if [ -z "${kickstart_image}" ]; then
    echo "ERROR: No kickstart image found in scenario file: ${scenario_file}" >&2
    exit 1
fi
```

### 2.2 Boot Image (Optional with Fallback)

The **boot image** is specified in `launch_vm` calls with the `--boot_blueprint` option. If not specified, it falls back to the `DEFAULT_BOOT_BLUEPRINT` variable defined in the scenario file.

```bash
# Extract boot_image from launch_vm --boot_blueprint calls
# Example: launch_vm --boot_blueprint rhel96-bootc-source-ai-model-serving
boot_image="$(grep -E "launch_vm.*boot_blueprint.*" "${scenario_file}" | awk '{print $NF}')"

# If not found, extract DEFAULT_BOOT_BLUEPRINT from scenario.sh
if [ -z "${boot_image}" ]; then
    boot_image="$(grep -E '^DEFAULT_BOOT_BLUEPRINT=' "test/bin/scenario.sh" | cut -d'=' -f2 | tr -d '"')"
fi
```

### 2.3 Upgrade Images (Optional with Default Fallback)

**Upgrade images** are optional target images for upgrade scenarios. They can be identified by `TARGET_REF:` and/or `FAILING_REF:` tokens in the scenario file. A scenario may have both, one, or none of these.

```bash
# Extract target_ref_image from TARGET_REF environment variable
# Example: TARGET_REF: "rhel96-bootc-upgraded"
target_ref_image="$(awk -F'TARGET_REF:' '{print $2}' "${scenario_file}" | tr -d '[:space:]"\\')"

# Extract failing_ref_image from FAILING_REF environment variable
# Example: FAILING_REF: "rhel96-bootc-failing"
failing_ref_image="$(awk -F'FAILING_REF:' '{print $2}' "${scenario_file}" | tr -d '[:space:]"\\')"

# Collect all upgrade images found
upgrade_images=()
[ -n "${target_ref_image}" ] && upgrade_images+=("${target_ref_image}")
[ -n "${failing_ref_image}" ] && upgrade_images+=("${failing_ref_image}")
```

### 2.4 Evaluate Shell Variables

All extracted image names may be shell variables, so evaluate them by sourcing the scenario file:

```bash
# Evaluate kickstart_image (may be a variable like ${RHEL96_BOOTC_SOURCE})
kickstart_image="$(bash -c "source \"${scenario_file}\"; echo ${kickstart_image}")"

# Evaluate boot_image
boot_image="$(bash -c "source \"${scenario_file}\"; echo ${boot_image}")"

# Evaluate upgrade images (only if they exist)
if [ -n "${target_ref_image}" ]; then
    target_ref_image="$(bash -c "source \"${scenario_file}\"; echo ${target_ref_image}")"
fi

if [ -n "${failing_ref_image}" ]; then
    failing_ref_image="$(bash -c "source \"${scenario_file}\"; echo ${failing_ref_image}")"
fi

# Collect all images found
all_images=("${kickstart_image}" "${boot_image}")
[ -n "${target_ref_image}" ] && all_images+=("${target_ref_image}")
[ -n "${failing_ref_image}" ] && all_images+=("${failing_ref_image}")

echo "Found images: kickstart=${kickstart_image} boot=${boot_image} target_ref=${target_ref_image} failing_ref=${failing_ref_image}"
```

## 3. Find Blueprint Files

For each image name found, locate the corresponding blueprint file. Blueprint files can be either `.containerfile` or `.image-bootc` files:

```bash
# Find blueprint file matching the image name (matches both .containerfile and .image-bootc extensions)
blueprint_file="$(find test/image-blueprints-bootc -type f -name "${image_name}.*" -print -quit)"
```

## 4. Find Dependencies Recursively

For each blueprint file, recursively find all dependencies. Both `.containerfile` and `.image-bootc` files can reference `localhost/` images:
- In `.containerfile` files: `FROM localhost/image-name:latest`
- In `.image-bootc` files: `localhost/image-name:latest` (bare reference, no `FROM` prefix)

The same grep pattern works for both file types:

```bash
# Extract localhost dependencies from the blueprint file
# Works for both .containerfile (FROM localhost/...) and .image-bootc (localhost/...) files
deps="$(grep -Eo 'localhost/[a-zA-Z0-9-]+:latest' "${blueprint_file}" | \
       awk -F'localhost/' '{print $2}' | sed 's/:latest//')"

# For each dependency, find its blueprint file and recurse
for dep in ${deps}; do
    dep_file="$(find test/image-blueprints-bootc -type f -name "${dep}.*" -print -quit)"
    # Recursively process dep_file to find its dependencies
done
```

**Important**: Track all processed files to avoid infinite loops and duplicates.

**Note on `.image-bootc` dependencies**: An `.image-bootc` file may reference a `localhost/` image that is built from a `.containerfile`. For example, `rhel98-bootc.image-bootc` contains `localhost/rhel98-test-agent:latest`, which depends on the `rhel98-test-agent.containerfile` blueprint. Follow these cross-type dependencies the same way.

## 5. Generate Build Commands

For each unique blueprint file (dependencies first, then the main images), generate:

```bash
# For .containerfile blueprints:
./test/bin/build_bootc_images.sh --template /path/to/blueprint.containerfile
# For .image-bootc blueprints:
./test/bin/build_bootc_images.sh --template /path/to/blueprint.image-bootc
```

**Important**:
- Sort the output so dependencies are built before images that depend on them
- Use absolute paths for blueprint files (use `realpath`)
- Output commands in a deterministic, sorted order
- NEVER actually execute `./test/bin/build_bootc_images.sh` - only output the commands

## 6. Output Format

The final output should be a sorted list of build commands, one per line. The file extension in the path reflects the actual blueprint type (`.containerfile` or `.image-bootc`):

```bash
./test/bin/build_bootc_images.sh --template microshift/test/image-blueprints-bootc/layer1-base/group1/rhel98-test-agent.containerfile
./test/bin/build_bootc_images.sh --template microshift/test/image-blueprints-bootc/layer1-base/group2/rhel98-bootc.image-bootc
./test/bin/build_bootc_images.sh --template microshift/test/image-blueprints-bootc/layer2-presubmit/group1/rhel98-bootc-source.containerfile
```

# Tips

1. **CRITICAL**: Validate that the scenario file path contains `scenarios-bootc` BEFORE any processing. Exit with error if not.
2. **DO NOT** attempt to find or convert non-bootc scenarios to bootc scenarios. Report error immediately.
3. Use `grep -Eo 'localhost/[a-zA-Z0-9-]+:latest'` to extract localhost image references from both `.containerfile` and `.image-bootc` files
4. Use `realpath` to convert relative paths to absolute paths
5. Use `sort -u` to ensure unique, sorted output
6. Maintain a set of processed blueprint files to avoid duplicates and infinite recursion
7. Dependencies must appear before images that depend on them in the output
8. If a blueprint file is not found, report an error with the image name
9. Blueprint files can be `.containerfile` or `.image-bootc` - the `find` with `"${image_name}.*"` matches both
10. An `.image-bootc` file's `localhost/` dependency may resolve to a `.containerfile` blueprint (cross-type dependency)

# Error Handling

**CRITICAL VALIDATION**: The first check must be whether the scenario is a bootc scenario. Report errors in this order:

1. If the scenario is not a bootc scenario (path does not contain `scenarios-bootc`):

   **YOUR ENTIRE RESPONSE MUST BE EXACTLY**:
   ```text
   ERROR: Not a bootc scenario: ${scenario_file}
   ERROR: This agent only works with bootc scenarios in test/scenarios-bootc/ directory
   ```

   **NOTHING ELSE**. Your response must contain ONLY these two error lines. **FORBIDDEN**:
   - Explanations about the error
   - Descriptions of what bootc scenarios are
   - Search for similar bootc scenarios
   - Suggestions for alternative files
   - Recommendations or help
   - Lists of available bootc scenarios
   - Questions to the user
   - Any additional text whatsoever

   Just the two error lines above, then STOP.

2. If the scenario file doesn't exist, report: "ERROR: Scenario file not found: ${scenario_file}"

3. If no images are found in the scenario, report: "ERROR: No dependencies found for scenario file: ${scenario_file}"

4. If a blueprint file is not found for an image, report: "ERROR: Image file not found: ${image_name}"

# Example Usage

Input: `microshift/test/scenarios-bootc/periodics/el96-src@ai-model-serving-offline.sh`

Expected workflow:
1. Parse scenario → finds kickstart image `rhel96-bootc-source-ai-model-serving` and boot image `rhel96-bootc-source-ai-model-serving`
2. Find blueprints → locates `rhel96-bootc-source-ai-model-serving.containerfile` and `rhel96-bootc-source-ai-model-serving.image-bootc`
3. Check `.image-bootc` dependencies → finds `localhost/rhel96-bootc-source-ai-model-serving:latest`
4. Check `.containerfile` dependencies → finds `FROM localhost/rhel96-bootc-source:latest`
5. Recurse → finds `rhel96-bootc-source.containerfile` which depends on `localhost/rhel96-test-agent:latest`
6. Recurse → finds `rhel96-test-agent.containerfile` which depends on an external registry image (no further localhost dependencies)
7. Output sorted commands (dependencies first):
   ```bash
   ./test/bin/build_bootc_images.sh --template .../layer1-base/group1/rhel96-test-agent.containerfile
   ./test/bin/build_bootc_images.sh --template .../layer2-presubmit/group1/rhel96-bootc-source.containerfile
   ./test/bin/build_bootc_images.sh --template .../layer3-periodic/group1/rhel96-bootc-source-ai-model-serving.containerfile
   ./test/bin/build_bootc_images.sh --template .../layer3-periodic/group2/rhel96-bootc-source-ai-model-serving.image-bootc
   ```
