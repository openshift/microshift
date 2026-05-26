#!/bin/bash
set -euo pipefail

changed=$(git diff HEAD --name-only \
    assets/crd/ \
    pkg/apis/microshift/v1alpha1/ \
    pkg/generated/)

if [ -n "${changed}" ]; then
    cat - <<EOF
ERROR:

You need to run 'make generate-crds' and commit the results to include
these files in the PR:

${changed}

EOF
    exit 1
fi
exit 0
