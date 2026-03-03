#!/usr/bin/env bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

# Build codegen-crds when it's not present and not overridden for a specific file.
if [ -z "${CODEGEN:-}" ];then
  ${TOOLS_MAKE} codegen
  CODEGEN="${TOOLS_OUTPUT}/codegen"
fi

# This runs the codegen utility against the entire set of API types.
# It has three args:
#
# GENERATOR is the name of the generator to run, which could be one of compatibility,
# deepcopy, schemapatch or swaggerdocs.
#
# API_GROUP_VERSIONS is an optional comma-separated list of fully qualified api
# <group>/<version> that codegen will be restricted to. Group names should be fully
# qualified (e.g., "config.openshift.io/v1,route.openshift.io/v1"). If not set,
# codegen will process all discovered API groups.
#
# EXTRA_ARGS are additional arguments to pass to the generator, usually this would be
# --verify so that the generator verifies the output rather than writing it.

codegen_args=()

if [ -n "${GENERATOR:-}" ]; then
  codegen_args+=("${GENERATOR}")
fi

codegen_args+=(--base-dir "${SCRIPT_ROOT}" -v 1)

if [ -n "${API_GROUP_VERSIONS:-}" ]; then
  codegen_args+=(--api-group-versions "${API_GROUP_VERSIONS}")
fi

"${CODEGEN}" "${codegen_args[@]}" ${EXTRA_ARGS:-}
