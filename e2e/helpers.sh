#!/usr/bin/env bash

gscp() {
    local src="${1}"
    local dst="${2}"

    gcloud compute \
        --project "${GOOGLE_PROJECT_ID}" \
        --quiet \
        scp \
        --zone "${GOOGLE_COMPUTE_ZONE}" \
        --recurse \
        "${src}" "rhel8user@${INSTANCE_PREFIX}:${dst}"
}

gssh() {
    local cmd="${1}"

    # TODO: Consider flag --quiet
    gcloud compute \
        --project "${GOOGLE_PROJECT_ID}" \
        ssh \
        --zone "${GOOGLE_COMPUTE_ZONE}" \
        "rhel8user@${INSTANCE_PREFIX}" \
        --command \
        "${cmd}"
}
