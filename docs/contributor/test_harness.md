# MicroShift Test Harness

The CI test harness provisions VMs, installs MicroShift, and runs Robot Framework tests against it. This document describes the test infrastructure. For CI job configuration and Prow details, see [openshift_ci.md](./openshift_ci.md).

## Deployment Modes

The harness supports two deployment models, each with its own scenarios and image blueprints:

| Mode | Scenarios | Image Blueprints | Description |
|------|-----------|-----------------|-------------|
| **ostree (RPM-based)** | `test/scenarios/` | `test/image-blueprints/` | MicroShift installed as RPMs on an rpm-ostree system |
| **bootc (container image-based)** | `test/scenarios-bootc/` | `test/image-blueprints-bootc/` | MicroShift embedded in a bootable container image |

## Scenarios

Each test scenario is a shell script that defines two functions:

- `scenario_create_vms()` — provisions VMs using kickstart templates and image blueprints
- `scenario_run_tests()` — runs Robot Framework test suites against the provisioned VMs

Scenarios are executed in parallel by `test/bin/ci_phase_boot_and_test.sh`, each provisioning its own independent VMs on a hypervisor.

### Naming Convention

```text
<os>-<branch>@<test-name>.sh
```

**OS prefixes:**

- `el98` — RHEL 9.8
- `cos10` — CentOS 10

**Branch prefixes:**

- `src` — current source code
- `crel` — current release
- `lrel` — latest release
- `prel` — previous release
- `yminus1` / `yminus2` — previous minor versions

**Examples:**

- `el98-src@standard-suite1.sh` — RHEL 9.8, current source, standard test suite 1
- `el98-prel@el98-src@upgrade-ok.sh` — upgrade from previous release to current source

### Scenario Types

| Type | Directory | When it runs |
|------|-----------|-------------|
| **Presubmit** | `presubmits/` | Every PR |
| **Periodic** | `periodics/` | Nightly/weekly |
| **Release** | `releases/` | Before release cuts (Brew RPM sourced) |
| **Upstream** | `upstream/` (bootc only) | CentOS-only testing |
| **C2CC** | `c2cc/` (bootc only) | Cluster-to-cluster connectivity testing |

### Scenario Definition Example

```bash
scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template rhel-9.8-microshift-source
    launch_vm rhel-9.8
}

scenario_run_tests() {
    run_tests host1 --variable "EXPECTED_OS_VERSION:9.8" suites/standard1/
}
```

## Image Blueprint Layers

Images are composed in layers with time-balanced caching. Each layer builds on the previous one:

### ostree Blueprints (`test/image-blueprints/`)

1. **Layer 1 — Base** (cached, ~30min build): OS-only base and historical MicroShift versions (y-2, y-1)
2. **Layer 2 — Presubmit** (not cached, ~10min): current source artifacts
3. **Layer 3 — Periodic** (not cached, ~15min): extended testing artifacts
4. **Layer 4 — Release** (cached behind VPN, ~30min): Brew RPM artifacts and upgrade paths

### bootc Blueprints (`test/image-blueprints-bootc/`)

1. **Layer 1 — Base** (cached, ~10min): basic prerequisites
2. **Layer 2 — Presubmit** (not cached, ~5min): current source on RHEL
3. **Layer 3 — Periodic** (not cached, ~15min): extended testing on RHEL
4. **Layer 4 — Upstream** (not cached, ~15min): CentOS-based testing
5. **Layer 5 — Release** (cached, ~15min): EC/RC/GA/z-stream releases

Within each layer, blueprints are organized in groups that can build in parallel. Groups within a layer are built sequentially when there are dependencies (e.g., base OS → y-2 → y-1).

## VM Lifecycle

The core VM orchestration lives in `test/bin/scenario.sh`:

1. **Kickstart preparation**: `prepare_kickstart` renders a kickstart template with scenario-specific variables (blueprint, registry URLs, pull secret, hostname, network config, FIPS mode, LVM size)
2. **VM provisioning**: `launch_vm` uses `virt-install` with configurable CPU/memory/disk. Retries up to 2 times with backoff.
3. **IP assignment**: waits for DHCP with a 20-minute timeout, validates via ping
4. **SSH access**: waits for SSH availability
5. **Greenboot health check**: waits for `greenboot-healthcheck` service to complete (up to 30 minutes, skippable per scenario)
6. **Test execution**: runs Robot Framework suites with scenario-specific variables
7. **Diagnostics collection**: on failure, collects SOS reports and PCP metrics from the VM
8. **Cleanup**: graceful shutdown with fallback to force destroy, removes libvirt domain and storage

### Kickstart Templates

Templates live in `test/kickstart-templates/`:

- `kickstart.ks.template` — standard ostree installations
- `kickstart-bootc.ks.template` — bootc container-based installations
- `kickstart-bootc-offline.ks.template` — bootc without network access
- `kickstart-bootc-isolated.ks.template` — bootc on isolated networks
- `kickstart-bootc-container.ks.template` — bootc container variant
- `kickstart-liveimg.ks.template` — live image booting
- `kickstart-offline.ks.template` — offline ostree installations
- `kickstart-isolated.ks.template` — ostree on isolated networks
- `kickstart-centos.ks.template` — CentOS variants

Templates use `REPLACE_*` placeholder variables substituted at runtime (e.g., `REPLACE_BOOT_COMMIT_REF`, `REPLACE_PULL_SECRET`, `REPLACE_HOST_NAME`).

### Network Configurations

VMs can be provisioned with different network setups:

- **default** — standard libvirt bridge
- **isolated** — dedicated network bridge for registry access
- **multus** — additional networks for Multus CNI testing
- **ipv6** — IPv6-only networking
- **dual-stack** — IPv4 + IPv6

## Robot Framework Tests

Test suites live in `test/suites/`, organized by feature area:

- `standard1/`, `standard2/` — core functionality
- `backup/` — backup/restore operations
- `greenboot/` — health check validation
- `upgrade/` — version upgrade scenarios
- `network/`, `ipv6/` — networking features
- `storage/` — storage operations
- `router/`, `gateway-api/` — ingress testing
- `configuration1/`, `configuration2/` — configuration scenarios
- `core-api/` — API validation
- `fips/` — FIPS mode validation
- `c2cc/`, `c2cc-ipsec/` — cluster-to-cluster connectivity
- `telemetry/` — telemetry testing
- `osconfig/` — OS configuration (cluster ID, etc.)
- `tuned/` — tuned profile testing
- `rpm/` — RPM package testing
- `ai-model-serving/` — AI model serving features
- `gitops/` — GitOps workflow testing
- `fault-tests/` — fault injection testing
- `optional/` — optional feature testing
- `otp-workloads/` — OTP-specific workloads

Each `.robot` file uses Resource files (`.resource`) for shared keywords and Python helper libraries for system interaction.

### Running Tests

Tests run against a remote MicroShift host via SSH. Copy the example variables file and configure it:

```bash
cp test/variables.yaml.example test/variables.yaml
```

Edit `test/variables.yaml` with your target host:

```yaml
USHIFT_HOST: microshift-dev
USHIFT_USER: microshift
SSH_PRIV_KEY: ~/.ssh/id_rsa
SSH_PORT: 22
```

The `variables.yaml` file is gitignored — each developer maintains their own copy.

Then run:

```bash
test/run.sh [suite paths...]
```

Without suite arguments, it runs a default set (standard1, standard2, and selected osconfig/storage tests). The script automatically sets up a Python virtual environment with Robot Framework.

## CI Orchestration

The main CI entry point is `test/bin/ci_phase_boot_and_test.sh`, which:

1. Sets up hypervisor configuration and monitoring (Prometheus, Loki)
2. Configures container registry mirroring via `test/bin/mirror_registry.sh`
3. Copies scenario definitions from `SCENARIO_SOURCES` (set by the Prow job)
4. Executes all scenarios in parallel using GNU `parallel`
5. Aggregates JUnit results across all scenarios

Supporting scripts in `test/bin/`:

| Script | Purpose |
|--------|---------|
| `scenario.sh` | Core VM lifecycle orchestration |
| `build_images.sh` | ostree image composition |
| `build_bootc_images.sh` | bootc image composition |
| `mirror_registry.sh` | Local Quay registry for image caching |
| `manage_hypervisor_config.sh` | libvirt networks, web server, monitoring setup |
| `common.sh` | Shared variables and utility functions |
