#!/bin/bash
#
# Analyze a MicroShift scenario file (bootc or ostree) to identify all image
# dependencies (direct and transitive) and produce a sorted list of build
# commands needed to build all required images.
#
# Usage: scenario_builder.sh <scenario-file>
#
set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(cd "${SCRIPTDIR}/../.." && pwd)"

# Validate arguments
if [ $# -ne 1 ]; then
    echo "ERROR: Usage: $0 <scenario-file>" >&2
    exit 1
fi

scenario_file="${1}"

# ============================================================================
# Step 1: Validate Scenario File Path and Detect Scenario Type
# ============================================================================

# Detect scenario type based on path
if [[ "${scenario_file}" == *"test/scenarios-bootc/"* ]]; then
    scenario_type="bootc"
elif [[ "${scenario_file}" == *"test/scenarios/"* ]]; then
    scenario_type="ostree"
else
    echo "ERROR: Unknown scenario type: ${scenario_file}" >&2
    echo "ERROR: Scenario must be in test/scenarios-bootc/ or test/scenarios/ directory" >&2
    exit 1
fi

# Check if the scenario file exists
if [ ! -f "${scenario_file}" ]; then
    echo "ERROR: Scenario file not found: ${scenario_file}" >&2
    exit 1
fi

# ============================================================================
# Step 2: Parse Scenario File to Extract Image Names
# ============================================================================

# 2.1 Extract kickstart_image (mandatory)
kickstart_image="$(grep -E "prepare_kickstart.*kickstart.*" "${scenario_file}" | awk '{print $NF}' || true)"

if [ -z "${kickstart_image}" ]; then
    echo "ERROR: No kickstart image found in scenario file: ${scenario_file}" >&2
    exit 1
fi

# 2.2 Extract boot_image (optional with fallback)
boot_image="$(grep -E "launch_vm.*boot_blueprint.*" "${scenario_file}" | grep -oE '\-\-boot_blueprint[[:space:]]+[^[:space:]]+' | awk '{print $2}' || true)"

# If not found, use DEFAULT_BOOT_BLUEPRINT from scenario.sh
if [ -z "${boot_image}" ]; then
    boot_image="$(grep -E '^DEFAULT_BOOT_BLUEPRINT=' "${REPO_ROOT}/test/bin/scenario.sh" | cut -d'=' -f2 | tr -d '"' || true)"
fi

# 2.3 Extract upgrade images (optional)
target_ref_image="$(awk -F'TARGET_REF:' '{print $2}' "${scenario_file}" | tr -d '\\[:space:]"' || true)"
failing_ref_image="$(awk -F'FAILING_REF:' '{print $2}' "${scenario_file}" | tr -d '\\[:space:]"' || true)"

# 2.4 Evaluate shell variables by sourcing the scenario file
kickstart_image="$(bash -c "source \"${scenario_file}\"; echo ${kickstart_image}" 2>/dev/null || echo "${kickstart_image}")"
boot_image="$(bash -c "source \"${scenario_file}\"; echo ${boot_image}" 2>/dev/null || echo "${boot_image}")"

if [ -n "${target_ref_image}" ]; then
    target_ref_image="$(bash -c "source \"${scenario_file}\"; echo ${target_ref_image}" 2>/dev/null || echo "${target_ref_image}")"
fi

if [ -n "${failing_ref_image}" ]; then
    failing_ref_image="$(bash -c "source \"${scenario_file}\"; echo ${failing_ref_image}" 2>/dev/null || echo "${failing_ref_image}")"
fi

# Collect all images found
all_images=("${kickstart_image}")
[ -n "${boot_image}" ] && all_images+=("${boot_image}")
[ -n "${target_ref_image}" ] && all_images+=("${target_ref_image}")
[ -n "${failing_ref_image}" ] && all_images+=("${failing_ref_image}")

# Debug output (commented for production)
# echo "Found images: kickstart=${kickstart_image} boot=${boot_image} target_ref=${target_ref_image} failing_ref=${failing_ref_image}" >&2

# ============================================================================
# Step 3 & 4: Find Blueprint Files and Dependencies Recursively
# ============================================================================

# Track processed blueprints to avoid infinite loops
declare -A processed_blueprints
# Track ordered list of blueprint files
declare -a blueprint_order

# Recursive function to find dependencies
find_dependencies_bootc() {
    local image_name="${1}"

    # Find all blueprint files matching the image name (may have both .containerfile and .image-bootc)
    local blueprint_files=()
    while IFS= read -r -d '' file; do
        blueprint_files+=("${file}")
    done < <(find "${REPO_ROOT}/test/image-blueprints-bootc" -type f -name "${image_name}.*" \( -name "*.containerfile" -o -name "*.image-bootc" \) -print0 2>/dev/null | sort -z)

    if [ ${#blueprint_files[@]} -eq 0 ]; then
        echo "ERROR: No blueprint file found for image: ${image_name}" >&2
        exit 1
    fi

    # Process each blueprint file for this image
    for blueprint_file in "${blueprint_files[@]}"; do
        # Convert to absolute path
        local abs_blueprint_file
        abs_blueprint_file="$(realpath "${blueprint_file}")"

        # Skip if already processed
        if [ -n "${processed_blueprints[${abs_blueprint_file}]:-}" ]; then
            continue
        fi

        # Mark as processed to prevent infinite recursion
        processed_blueprints["${abs_blueprint_file}"]=1

        # Extract localhost dependencies from the blueprint file
        # Works for both .containerfile (FROM localhost/...) and .image-bootc (localhost/...) files
        local deps=""
        deps="$(grep -Eo 'localhost/[a-zA-Z0-9-]+:latest' "${abs_blueprint_file}" | \
               awk -F'localhost/' '{print $2}' | sed 's/:latest//' || true)"

        # Recursively process each dependency
        for dep in ${deps}; do
            find_dependencies_bootc "${dep}"
        done

        # Add this blueprint to the ordered list (dependencies first)
        blueprint_order+=("${abs_blueprint_file}")
    done
}

# Recursive function to find dependencies for ostree blueprints
find_dependencies_ostree() {
    echo "ERROR: find_dependencies_ostree is not implemented" >&2
    exit 1
}

# Process all found images using the appropriate function
for image_name in "${all_images[@]}"; do
    if [ -n "${image_name}" ]; then
        if [ "${scenario_type}" = "bootc" ]; then
            find_dependencies_bootc "${image_name}"
        else
            find_dependencies_ostree "${image_name}"
        fi
    fi
done

# ============================================================================
# Step 5: Generate Build Commands
# ============================================================================

if [ ${#blueprint_order[@]} -eq 0 ]; then
    echo "ERROR: No dependencies found for scenario file: ${scenario_file}" >&2
    exit 1
fi

# Output build commands in dependency order
for blueprint_file in "${blueprint_order[@]}"; do
    if [ "${scenario_type}" = "bootc" ]; then
        echo "./test/bin/build_bootc_images.sh --template ${blueprint_file}"
    else
        echo "./test/bin/build_images.sh -t ${blueprint_file}"
    fi
done
