#!/usr/bin/bash
set -euo pipefail

SCANNER_IMAGE="images.paas.redhat.com/exd-sp-guild-security/rh-crypto-scanner-image:latest"
REPOROOT="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")/../..")"
CBOM_OUTPUT="${REPOROOT}/cbom-microshift.json"
SPDX_OUTPUT="${REPOROOT}/sbom-microshift-crypto.spdx.json"

echo "Running crypto scanner against MicroShift source..."
echo "Scanner image: ${SCANNER_IMAGE}"
echo "Source directory: ${REPOROOT}"

if ! podman pull "${SCANNER_IMAGE}"; then
  echo "WARNING: failed to pull scanner image, skipping CBOM generation" >&2
  exit 0
fi

# The scanner skips directories named "vendor/" by default.
# Mount vendor directories under non-skipped names so the scanner
# processes them. The "deps/" directory is excluded because it is a
# patched copy of upstream Kubernetes vendor and would duplicate
# findings already covered by "vendor/".
# We stage a scan directory with only the sources we want scanned.
SCAN_DIR=$(mktemp -d)
trap 'rm -rf "${SCAN_DIR}"' EXIT

cp -a "${REPOROOT}/pkg" "${SCAN_DIR}/pkg"
cp -a "${REPOROOT}/cmd" "${SCAN_DIR}/cmd"
cp -a "${REPOROOT}/vendor" "${SCAN_DIR}/vendor-deps"
cp -a "${REPOROOT}/etcd/cmd" "${SCAN_DIR}/etcd-cmd"
cp -a "${REPOROOT}/etcd/vendor" "${SCAN_DIR}/etcd-deps"

if ! podman run --rm \
  -v "${SCAN_DIR}:/workspace:z" \
  "${SCANNER_IMAGE}" \
  /workspace > "${CBOM_OUTPUT}"; then
  echo "WARNING: crypto scan failed, skipping CBOM generation" >&2
  exit 0
fi

echo "CBOM generated: "
ls -lh "${CBOM_OUTPUT}"

# SPDX 2.3 has no native crypto asset support. As a workaround agreed
# with prod sec, cryptoProperties are serialized into the SPDX package
# "comment" field.
if ! jq '{
  spdxVersion: "SPDX-2.3",
  dataLicense: "CC0-1.0",
  SPDXID: "SPDXRef-DOCUMENT",
  name: "MicroShift-CBOM",
  documentNamespace: "https://microshift.io/spdx/cbom",
  packages: (
    [{
      SPDXID: "SPDXRef-Package-RootPackage",
      name: "microshift",
      downloadLocation: "NOASSERTION",
      filesAnalyzed: false
    }] +
    [.components[] | {
      SPDXID: ("SPDXRef-Package-" + (.name | gsub("[^a-zA-Z0-9.-]"; "-"))),
      name: .name,
      comment: (.cryptoProperties | tojson),
      description: "Converted CBOM component from CycloneDX to SPDX",
      downloadLocation: "NOASSERTION",
      filesAnalyzed: false
    }]
  ),
  relationships: (
    [{
      spdxElementId: "SPDXRef-DOCUMENT",
      relatedSpdxElement: "SPDXRef-Package-RootPackage",
      relationshipType: "DESCRIBES"
    }] +
    [.components[] | {
      spdxElementId: "SPDXRef-Package-RootPackage",
      relatedSpdxElement: ("SPDXRef-Package-" + (.name | gsub("[^a-zA-Z0-9.-]"; "-"))),
      relationshipType: "CONTAINS"
    }]
  )
}' "${CBOM_OUTPUT}" > "${SPDX_OUTPUT}"; then
  echo "WARNING: SPDX conversion failed" >&2
  exit 0
fi

echo "SPDX generated: "
ls -lh "${SPDX_OUTPUT}"