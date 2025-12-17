---
name: MicroShift Release Versions
argument-hint: <version>
description: Check if a given MicroShift version is available and where to find it
allowed-tools: WebFetch, Bash
---

## Name
microshift-release-versions

## Synopsis
```
/microshift-release-versions <version>
```

## Description
The `microshift-release-versions` command checks if a specific MicroShift version is available and provides information about where to find RPMs and bootc images.

This command provides:
- Version availability status (available/not yet available)
- Version type classification (nightly, EC, RC, GA, z-stream)
- RPM package URLs
- Bootc image URLs (for versions 4.18+)
- Brew build status (internal Red Hat, requires VPN)
- Links to browse packages and images

This command is useful for:
- Checking if a new EC/RC version has been published
- Finding the correct URLs for RPM repositories
- Locating bootc image pull specs
- Understanding what version type a given version is
- Finding Brew builds for internal testing (VPN required)

## Arguments
- `$1` (version): The MicroShift version to check - **Required**
  - Formats accepted:
    - Full version: `4.20.0`, `4.18.26`
    - EC version: `4.21.0-ec.3` or `4.21.0~ec.3`
    - RC version: `4.20.0-rc.3` or `4.20.0~rc.3`
    - Short version: `4.21` (will check latest available)

## Return Value
- **Format**: Markdown
- **Location**: Output directly to the conversation
- **Content**:
  - Version availability status
  - Version type (EC, RC, GA, z-stream)
  - RPM URLs for both architectures (x86_64, aarch64)
  - Bootc image pull specs (if available)
  - Links to package catalogs

## Version Types Reference

| Type | Version Format | Release Cadence | RPM Source | Image Source |
|------|----------------|-----------------|------------|--------------|
| Nightly | `X.Y.0~0.nightly_YYYY_MM_DD_HHMMSS` | Continuous | Brew | Not available |
| EC | `X.Y.0~ec.N` | Every sprint (3 weeks) | Brew, mirror repo | Mirror repo |
| RC | `X.Y.0~rc.N` | After branch cutoff until GA | Brew, mirror repo | Mirror repo |
| GA | `X.Y.0` | Every 4 months | Brew, rhocp repos | registry.redhat.io |
| Z-stream | `X.Y.Z` | On request | Brew, rhocp repos | registry.redhat.io |

## Implementation Steps

### Step 1: Parse Version and Determine Type

**Goal**: Parse the input version and classify its type.

**Actions**:
1. Normalize version format (replace `~` with `-` for URL construction)
2. Extract major.minor version (e.g., `4.20` from `4.20.0-rc.3`)
3. Classify version type:
   - Contains `nightly` → Nightly build
   - Contains `ec` → Engineering Candidate
   - Contains `rc` → Release Candidate
   - Format `X.Y.0` → GA release
   - Format `X.Y.Z` (Z > 0) → Z-stream release

**Example**:
```
Input: 4.21.0-ec.3
Parsed:
- normalized: 4.21.0-ec.3
- major_minor: 4.21
- type: EC
- base_version: 4.21.0
```

### Step 2: Check OpenShift Release Status

**Goal**: Check if OCP payload is Accepted for this version.

**Actions**:
1. Web Fetch this and look for the version:
   ```
   https://openshift-release.apps.ci.l2s4.p1.openshiftapps.com/
   ```
2. Check if the payload for the version is Accepted.

**Outcome**:
 - If the payload is Accepted continue with next step.
 - If the payload is Not Accepted do not run next steps. MicroShift version won't be created if the OCP is not Accepted.

**Note**: MicroShift packages become available when the corresponding OCP version is marked as "Accepted".

### Step 3: Check RPM Availability

**Goal**: Determine if RPM packages are available and provide URLs.

**Actions**:

**For EC versions**:
1. Check mirror.openshift.com for dev-preview packages:
   ```
   https://mirror.openshift.com/pub/openshift-v4/x86_64/microshift/ocp-dev-preview/<version>/el9/os/Packages/
   https://mirror.openshift.com/pub/openshift-v4/aarch64/microshift/ocp-dev-preview/<version>/el9/os/Packages/
   ```
2. Use WebFetch to verify the directory exists and list packages

**For RC versions**:
1. Check mirror.openshift.com for ocp packages:
   ```
   https://mirror.openshift.com/pub/openshift-v4/x86_64/microshift/ocp/<version>/el9/os/Packages/
   https://mirror.openshift.com/pub/openshift-v4/aarch64/microshift/ocp/<version>/el9/os/Packages/
   ```

**For GA/Z-stream versions**:
1. Check on brew:
   ```
   https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=82827
   ```
   Use this command to check for microshift releases:
   ```
   curl -sk "https://brewweb.engineering.redhat.com/brew/search?match=glob&type=build&terms={version}" 2>/dev/null
   ```

### Step 4: Check Bootc Image Availability

**Goal**: Find bootc image pull specs for versions 4.18+.

**Actions**:

**For EC versions**:
1. Fetch bootc-pullspec.txt:
   ```
   https://mirror.openshift.com/pub/openshift-v4/x86_64/microshift/ocp-dev-preview/<version>/el9/bootc-pullspec.txt
   https://mirror.openshift.com/pub/openshift-v4/aarch64/microshift/ocp-dev-preview/<version>/el9/bootc-pullspec.txt
   ```
2. Extract the quay.io pull spec from the file

**For RC versions**:
1. Fetch bootc-pullspec.txt:
   ```
   https://mirror.openshift.com/pub/openshift-v4/x86_64/microshift/ocp/<version>/el9/bootc-pullspec.txt
   https://mirror.openshift.com/pub/openshift-v4/aarch64/microshift/ocp/<version>/el9/bootc-pullspec.txt
   ```

**For GA/Z-stream versions**:
1. For pre-GA z-stream, check registry.stage.redhat.io:
   - Production: `https://catalog.redhat.com/en/software/containers/openshift4/microshift-bootc-rhel9/`
      - Use this command to check the catalog: `curl -sk "https://catalog.redhat.com/api/containers/v1/repositories/registry/registry.access.redhat.com/repository/openshift4/microshift-bootc-rhel9/images?page_size=500&page=0" 2>/dev/null`
         - if not found try few more pages
   - Stage: `https://catalog.stage.redhat.com/en/software/containers/openshift4/microshift-bootc-rhel9/`
      - Use this command to check the catalog: `curl -sk "https://catalog.stage.redhat.com/api/containers/v1/repositories/registry/registry.access.redhat.com/repository/openshift4/microshift-bootc-rhel9/images?page_size=500&page=0" 2>/dev/null`
         - if not found try few more pages

### Step 5: List Available Versions (if short version provided)

**Goal**: If user provides a short version like `4.21`, list all available sub-versions.

**Actions**:
1. Fetch directory listing from mirror.openshift.com:
   ```
   https://mirror.openshift.com/pub/openshift-v4/x86_64/microshift/ocp-dev-preview/
   https://mirror.openshift.com/pub/openshift-v4/x86_64/microshift/ocp/
   ```
2. Filter for versions matching the major.minor pattern
3. List all available versions with their types

### Step 6: Generate Report

**Goal**: Create a comprehensive availability report.

**Report Structure**:
```markdown
# MicroShift Version Availability: <version>

## Status
- **Version**: <full-version>
- **Type**: EC / RC / GA / Z-stream
- **Status**: Available / Not Yet Available
- **OCP Release Status**: Accepted / Pending

## Brew Builds (VPN Required)
- **Web**: https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=82827
- **Build**: <brew-build-nvr> (if found)
- **Status**: Available / Not Found / VPN Required

## RPM Packages

### x86_64
- **URL**: <mirror-url>
- **Status**: Available / Not Found

### aarch64
- **URL**: <mirror-url>
- **Status**: Available / Not Found

## Bootc Images (4.18+)

### x86_64
- **Pull Spec**: <quay.io/...@sha256:...>

### aarch64
- **Pull Spec**: <quay.io/...@sha256:...>

## Catalog Links
- [Red Hat Catalog](<catalog-url>) (GA/Z-stream only)
- [Stage Catalog](<stage-catalog-url>) (pre-GA Z-stream only)

## Notes
- <any relevant notes about availability>
```

## Examples

### Example 1: Check EC Version
```
/microshift-release-versions 4.21.0-ec.3
```

Output:
```markdown
# MicroShift Version Availability: 4.21.0-ec.3

## Status
- **Version**: 4.21.0-ec.3
- **Type**: Engineering Candidate (EC)
- **Status**: Available
- **OCP Release Status**: Accepted

## Brew Builds (VPN Required)
- **Web**: https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=82827
- **Build**: microshift-4.21.0-0.ec.3.el9
- **Status**: Available

## RPM Packages

### x86_64
- **URL**: https://mirror.openshift.com/pub/openshift-v4/x86_64/microshift/ocp-dev-preview/4.21.0-ec.3/el9/os/Packages/
- **Status**: Available

### aarch64
- **URL**: https://mirror.openshift.com/pub/openshift-v4/aarch64/microshift/ocp-dev-preview/4.21.0-ec.3/el9/os/Packages/
- **Status**: Available

## Bootc Images

### x86_64
- **Pull Spec**: quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:5ec944...

### aarch64
- **Pull Spec**: quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:a255da...
```

### Example 2: Check RC Version (Not Yet Available)
```
/microshift-release-versions 4.21.0-rc.1
```

Output:
```markdown
# MicroShift Version Availability: 4.21.0-rc.1

## Status
- **Version**: 4.21.0-rc.1
- **Type**: Release Candidate (RC)
- **Status**: Not Yet Available

## Notes
- 4.21 is currently in EC phase (latest: 4.21.0-ec.3)
- RC versions are created after branch cutoff until GA
- Check https://openshift-release.apps.ci.l2s4.p1.openshiftapps.com/ for release status
```

### Example 3: Check GA/Z-stream Version
```
/microshift-release-versions 4.18.26
```

Output:
```markdown
# MicroShift Version Availability: 4.18.26

## Status
- **Version**: 4.18.26
- **Type**: Z-stream
- **Status**: Available

## RPM Packages
- **Repo**: rhocp-4.18-for-rhel-9-x86_64-rpms
- **Catalog**: https://access.redhat.com/downloads/content/microshift/4.18.26/

## Bootc Images
- **Pull Spec**: registry.redhat.io/openshift4/microshift-bootc-rhel9:4.18.26
- **Catalog**: https://catalog.redhat.com/en/software/containers/openshift4/microshift-bootc-rhel9/
```

### Example 4: List Available Versions for a Release
```
/microshift-release-versions 4.21
```

Output:
```markdown
# Available MicroShift 4.21 Versions

## EC Versions (Dev Preview)
- 4.21.0-ec.3 (Accepted)
- 4.21.0-ec.2 (Accepted)
- 4.21.0-ec.1 (Accepted)
- 4.21.0-ec.0 (Accepted)

## RC Versions
- None available yet (4.21 is still in EC phase)

## Notes
- RC versions will be available after branch cutoff
- GA expected in approximately X months
```

## Reference Information

### Brew (Internal Red Hat)

Brew is Red Hat's internal build system. VPN is required to access it, assume VPN is enabled.

**Web Interface**:
- Browse all MicroShift packages: https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=82827

**Build NVR Format**:
- GA/Z-stream: `microshift-4.18.26-<release>.el9`
- EC: `microshift-4.21.0-0.ec.3.<release>.el9`
- RC: `microshift-4.21.0-0.rc.1.<release>.el9`
- Nightly: `microshift-4.21.0-0.nightly_YYYY_MM_DD_HHMMSS.<release>.el9`

### URL Patterns

**Mirror repos (EC/RC)**:
- EC RPMs: `https://mirror.openshift.com/pub/openshift-v4/{ARCH}/microshift/ocp-dev-preview/{VERSION}/el9/os/Packages/`
- RC RPMs: `https://mirror.openshift.com/pub/openshift-v4/{ARCH}/microshift/ocp/{VERSION}/el9/os/Packages/`
- EC Bootc: `https://mirror.openshift.com/pub/openshift-v4/{ARCH}/microshift/ocp-dev-preview/{VERSION}/el9/bootc-pullspec.txt`
- RC Bootc: `https://mirror.openshift.com/pub/openshift-v4/{ARCH}/microshift/ocp/{VERSION}/el9/bootc-pullspec.txt`

**Registry repos (GA/Z-stream)**:
- Production: `registry.redhat.io/openshift4/microshift-bootc-rhel9`
- Stage: `registry.stage.redhat.io/openshift4/microshift-bootc-rhel9`

**Catalogs**:
- Production: `https://catalog.redhat.com/en/software/containers/openshift4/microshift-bootc-rhel9/`
- Stage: `https://catalog.stage.redhat.com/en/software/containers/openshift4/microshift-bootc-rhel9/`

### Release Timing

- **EC versions**: Usually created from a commit on the last Wednesday of every sprint; available on the first Wednesday of the next sprint
- **RC versions**: Created after branch cutoff until GA; several may be created
- **GA versions**: Released every 4 months
- **Z-stream versions**: Created on request from MicroShift team to ART team

### CI Testing

Release versions are tested in CI:
- Scenarios: `test/scenarios/releases/` and `test/scenarios-bootc/releases/`
- Latest release version must be **manually updated** in `test/bin/common_versions.sh` when new RC or z-stream is available

## Notes
- VPN access is enabled by default
- Bootc images are only available for MicroShift 4.18 and later
- MicroShift packages become available when the OCP version is marked as "Accepted"
- For nightly builds, only RPMs are available in Brew (VPN required)
- The command is read-only and does not modify any data
- Use this command to verify availability before updating CI configurations or deployment scripts
