#!/usr/bin/env bash

if [ ! -f ./_output/tools/bin/yq ]; then
    mkdir -p ./_output/tools/bin
    curl -s -f -L https://github.com/mikefarah/yq/releases/download/2.4.0/yq_$(go env GOHOSTOS)_$(go env GOHOSTARCH) -o ./_output/tools/bin/yq
    chmod +x ./_output/tools/bin/yq
fi

FAILS=false

for f in $(find . -name "*.yaml" -type f); do
    grep -qre "kind:\(.*\)CustomResourceDefinition" $f || continue
    grep -qre "name:\(.*\).openshift.io" $f || continue

    # skip the files that are merged to produce the final outcome
    if [[ "$f" == *"zz_generated.featuregated-crd-manifests"* ]]; then
      continue
    fi
    if [[ "$f" == *"manual-override-crd-manifests"* ]]; then
      continue
    fi
    if [[ "$f" == *"testdata"* ]]; then
      continue
    fi

    if [[ $(./_output/tools/bin/yq r $f apiVersion) == "apiextensions.k8s.io/v1beta1" ]]; then
        if [[ $(./_output/tools/bin/yq r $f spec.validation.openAPIV3Schema.properties.metadata.description) != "null" ]]; then
            echo "Error: cannot have a metadata description in $f"
            FAILS=true
        fi

        if [[ $(./_output/tools/bin/yq r $f spec.preserveUnknownFields) != "false" ]]; then
            echo "Error: pruning not enabled (spec.preserveUnknownFields != false) in $f"
            FAILS=true
        fi
    fi

    if [[ $(./_output/tools/bin/yq r $f metadata.annotations[api-approved.openshift.io]) == "null" ]]; then
        echo "Error: missing 'api-approved.openshift.io' annotation pointing to openshift/api pull request in $f"
        FAILS=true
    fi
done

if [ "$FAILS" = true ] ; then
    exit 1
fi

