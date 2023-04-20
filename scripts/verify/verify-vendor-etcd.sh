#!/bin/bash
set -euo pipefail

# First stage any files under etcd/vendor in case there are new ones.
git add etcd/vendor

# Now get the list of files that would be committed.
changed=$(git diff --cached --name-only etcd/vendor)

if [ -n "${changed}" ]; then
    cat - <<EOF
ERROR:

You need to run 'make vendor-etcd' and commit the results to include
these files in the PR:

EOF
    git diff --cached --name-only etcd/vendor
    exit 1
fi

exit 0
