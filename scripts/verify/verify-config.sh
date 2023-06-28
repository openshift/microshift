#!/bin/bash
set -euo pipefail

config_spec_path=cmd/generate-config/config/config-openapi-spec.json
config_file_path=packaging/microshift/config.yaml
config_doc_path=docs/user/howto_config.md

config_changed=$(git diff HEAD --name-only ${config_file_path})
doc_changed=$(git diff HEAD --name-only ${config_doc_path})
openapi_spec_changed=$(git diff HEAD --name-only ${config_spec_path})

if [ -n "${config_changed}" ] || [ -n "${doc_changed}" ] || [ -n "${openapi_spec_changed}" ]; then
    cat - <<EOF
ERROR:

You need to run 'make generate-config' and commit the results to include
these files in the PR:

EOF
    exit 1
fi
exit 0
