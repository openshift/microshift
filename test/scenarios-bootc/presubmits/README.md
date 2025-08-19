# MicroShift BootC Presubmit Test Scenarios

This directory contains presubmit test scenarios for MicroShift BootC (Image Mode) functionality. These tests are executed as part of the continuous integration pipeline to validate changes before they are merged.

## Test Scenarios Overview

The following table lists all available test scenarios in this directory:

| Test Scenario | Base OS | Target OS | Test Type | Description |
|---------------|---------|-----------|-----------|-------------|
| `cos9-src@standard-suite1.sh` | CentOS Stream 9 | Source | Standard | Standard test suite 1 for CentOS Stream 9 |
| `cos9-src@standard-suite2.sh` | CentOS Stream 9 | Source | Standard | Standard test suite 2 for CentOS Stream 9 |
| `el94-yminus2@el96-src@upgrade-ok.sh` | RHEL 9.4 (Y-2) | RHEL 9.6 Source | Upgrade | Successful upgrade from RHEL 9.4 to RHEL 9.6 source |
| `el94-yminus2@prel@src@delta-upgrade-ok.sh` | RHEL 9.4 (Y-2) | Prerelease → Source | Delta Upgrade | Successful delta upgrade from Y-2 through prerelease to source |
| `el96-base@el96-src@upgrade-ok.sh` | RHEL 9.6 Base | RHEL 9.6 Source | Upgrade | Successful upgrade from RHEL 9.6 base to source |
| `el96-base@src@opt@delta-upgrade-ok.sh` | RHEL 9.6 Base | Source → Optional | Delta Upgrade | Successful delta upgrade from base to source with optional components |
| `el96-crel@el96-src@upgrade-ok.sh` | RHEL 9.6 Candidate | RHEL 9.6 Source | Upgrade | Successful upgrade from candidate release to source |
| `el96-prel@el96-src@upgrade-ok.sh` | RHEL 9.6 Prerelease | RHEL 9.6 Source | Upgrade | Successful upgrade from prerelease to source |
| `el96-src@ai-model-serving-online.sh` | RHEL 9.6 Source | - | Feature | AI model serving functionality with online connectivity |
| `el96-src@auto-recovery.sh` | RHEL 9.6 Source | - | Recovery | Automatic recovery functionality testing |
| `el96-src@backup-and-restore-on-reboot.sh` | RHEL 9.6 Source | - | Backup/Restore | Backup and restore functionality during system reboot |
| `el96-src@backups.sh` | RHEL 9.6 Source | - | Backup | General backup functionality testing |
| `el96-src@downgrade-block.sh` | RHEL 9.6 Source | - | Downgrade | Validation that downgrades are properly blocked |
| `el96-src@dual-stack.sh` | RHEL 9.6 Source | - | Networking | Dual-stack (IPv4/IPv6) networking configuration |
| `el96-src@ipv6.sh` | RHEL 9.6 Source | - | Networking | IPv6-only networking configuration |
| `el96-src@log-scan.sh` | RHEL 9.6 Source | - | Logging | Log scanning and analysis functionality |
| `el96-src@multi-nic.sh` | RHEL 9.6 Source | - | Networking | Multiple network interface configuration |
| `el96-src@optional.sh` | RHEL 9.6 Source | - | Components | Optional component installation and functionality |
| `el96-src@router.sh` | RHEL 9.6 Source | - | Networking | Router functionality and configuration |
| `el96-src@standard-suite1.sh` | RHEL 9.6 Source | - | Standard | Standard test suite 1 for RHEL 9.6 source |
| `el96-src@standard-suite2.sh` | RHEL 9.6 Source | - | Standard | Standard test suite 2 for RHEL 9.6 source |
| `el96-src@storage.sh` | RHEL 9.6 Source | - | Storage | Storage functionality and configuration |
| `el96-src@upgrade-fails-then-recovers.sh` | RHEL 9.6 Source | - | Recovery | Failed upgrade followed by successful recovery |
| `el96-src@upgrade-fails.sh` | RHEL 9.6 Source | - | Upgrade | Failed upgrade scenario validation |

## Test Categories

### Standard Test Suites
These tests validate core MicroShift functionality across different operating system versions:
- Standard suite 1 & 2 for both CentOS Stream 9 and RHEL 9.6

### Upgrade Tests
Tests that validate upgrade scenarios between different MicroShift versions and OS releases:
- Cross-version upgrades (Y-2 to current)
- Base to source upgrades
- Candidate/prerelease to source upgrades
- Delta upgrades with optional components

### Networking Tests
Tests focused on network configuration and functionality:
- IPv6-only networking
- Dual-stack (IPv4/IPv6) configuration
- Multi-NIC setups
- Router functionality

### Feature-Specific Tests
Tests for specific MicroShift features:
- AI model serving (online mode)
- Optional component management
- Storage functionality
- Logging and monitoring

### Recovery and Resilience Tests
Tests that validate system recovery and failure handling:
- Automatic recovery mechanisms
- Backup and restore functionality
- Failed upgrade recovery
- Downgrade prevention

## Naming Convention

Test scenario files follow the naming pattern:
```
<base-os>@<target-os>@<test-type>.sh
```

Where:
- `base-os`: Starting operating system (e.g., `el96-src`, `cos9-src`)
- `target-os`: Target system for upgrades (if applicable)
- `test-type`: Type of test being performed

Common OS abbreviations:
- `cos9`: CentOS Stream 9
- `el94`: RHEL 9.4
- `el96`: RHEL 9.6
- `src`: Source build
- `base`: Base image
- `crel`: Candidate release
- `prel`: Prerelease
- `yminus2`: Y-2 release (two versions back)

## Running Tests

These test scenarios are designed to be executed within the MicroShift CI/CD pipeline. Each test script contains the necessary configuration and steps to validate specific functionality or upgrade paths.

For more information about the testing framework and how to run these tests locally, refer to the main test documentation in the parent directories.

