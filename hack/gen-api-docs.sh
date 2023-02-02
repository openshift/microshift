#!/bin/bash -x

set -e
set -o pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPOROOT=$(git rev-parse --show-toplevel)
OUTDIR="$REPOROOT/_output/openshift-apidocs-gen"

##
## Build the image we need to run openshift-api-docs
##

podman build -f "$SCRIPTDIR/Dockerfile.apidocs" -t openshift-apidocs-gen

mkdir -p "$OUTDIR"
cd "$OUTDIR"

# Get the APIs present in the local microshift instance and clean up
# the JSON data produced.
oc get --raw /openapi/v2 | jq . > openapi-raw.json
python3 $SCRIPTDIR/cleanup_openapi.py openapi-raw.json openapi-clean.json
cat openapi-clean.json | jq . > openapi.json

cp "$SCRIPTDIR/openshift-apidocs-gen-config.yaml" api-config.yaml

# We do not want selinux relabeling done if we're on a Mac.
# https://github.com/containers/podman/issues/13631
systype=$(uname)
if [ "$systype" = "Darwin" ]; then
    selinux_suffix=""
else
    selinux_suffix=":Z"
fi

function run {
    podman run \
           --tty \
           --rm=true \
           --workdir /workdir \
           --env LC_ALL=C.UTF-8 \
           -v $(pwd):/workdir${selinux_suffix} \
           openshift-apidocs-gen \
           $@
}

##
## Produce the asciidoc files for the API docs
##

# This step to create the config file doesn't seem to work, so we
# created the config file by copying the one from openshift-docs and
# modifying it.
#
# run openshift-apidocs-gen create-resources openapi.json >> config.yaml

run openshift-apidocs-gen build openapi.json

##
## Generate the snippet of the topic map file based on the generated
## content.
##

function get_name {
    typeset f="$1"
    grep '^= ' $f | cut -c3- | sed -e "s|^|'|" -e "s|\$|'|"
}

function gen_topic_map {
    oc get configmap -n kube-public microshift-version -o yaml > $OUTDIR/version.yaml
    version=$(cat $OUTDIR/version.yaml | yq .data.version)

    # The first few entries are always the same and are manually ordered.
    cat - <<EOF
# Last updated for $version
# $(date)
Name: API reference
Dir: microshift_rest_api
Distros: microshift
Topics:
- Name: Understanding API tiers
  File: understanding-api-support-tiers
- Name: API compatibility guidelines
  File: understanding-compatibility-guidelines
- Name: API list
  File: index
- Name: Common object reference
  Dir: objects
  Topics:
  - Name: Index
    File: index
EOF

    # The entries for each API are ordered alphabetically.

    cd "$OUTDIR/microshift_rest_api"
    for d in $(ls -d *_apis); do
        pushd $d >/dev/null
        indexfile=$(ls *index.adoc)
        if [ -z "$indexfile" ]; then
            echo "No index file in $d" 1>&2
            return 1
        fi
        name=$(grep '^= ' $indexfile | cut -c3-)
        cat -<<EOF
- Name: $(get_name $indexfile)
  Dir: $d
  Topics:
EOF
        for f in $(ls *.adoc); do
            if [ "$f" = "$indexfile" ]; then
                continue
            fi
            cat - <<EOF
  - Name: $(get_name $f)
    File: $(basename $f .adoc)
EOF
        done
        popd >/dev/null
    done
}

gen_topic_map > $OUTDIR/_topic_map_segment.yml
