# MicroShift Containerized Dev Environment

---

## Goal

Run the full MicroShift build and test pipeline (RPMs, bootc images, ISOs, VMs)
from any Linux host with podman — no dedicated RHEL VM required.

---

## Approach 1: devenv-toolbox (Toolbox Container)

---

### Technique

- Uses `toolbox create/enter/run` with a custom UBI9-based image
- Shares host's filesystem, network, `/dev/kvm`, SELinux kernel
- No `--privileged`, no nested systemd — uses host's init system
- `chcon` works because toolbox sees the real SELinux via `/sys/fs/selinux`

---

### What Works

- RPM builds (`make rpm`)
- Container image builds (`podman build`)
- `chcon` / SELinux xattrs (host kernel, no overlay xattr limitation)
- Host filesystem access (no bind-mount needed)
- subscription-manager registration
- configure-vm.sh (package installation)

---

### Limitations

- **No rootful podman** — toolbox maps host root to `nobody` via user namespaces.
  `sudo podman` fails with "permission denied" on `/dev/shm/libpod_lock`
  because the SHM lock file is owned by host root, inaccessible inside toolbox.
  - Mirror registry (`mirror_registry.sh`) uses `sudo podman` for quay/redis/postgres
  - Not a simple fix — containers use UID/GID mapping and privileged port binding
- **No systemd inside the container** — toolbox shares host PID 1
  - `osbuild-composer` requires systemd (socket activation, D-Bus)
  - Cannot run `systemctl enable/start` for container-local services
  - `manage_composer_config.sh` depends on osbuild-composer
- **ostree image builds** — current build system uses `osbuild-composer`
  (composer-cli). Porting to `image-builder-cli` would remove the systemd
  dependency but requires rewriting the ostree image build pipeline.
- **bootc ISO builds** — `image-builder-cli` works without systemd and
  supports `--bootc-ref` for building from container images directly.
  This is the viable path for ISO generation in a toolbox environment.

---

### Key Insight

Toolbox solves the SELinux problem but introduces two new blockers:
- **No rootful podman** — mirror registry and other `sudo podman` workflows fail
- **No systemd** — osbuild-composer and service management don't work

The build pipeline splits into:
- **Works now:** RPMs, container images, rootless podman builds
- **Blocked:** mirror registry, osbuild-composer, ISO builds via `sudo podman`
- **Needs porting:** ostree images (from `osbuild-composer` to `image-builder-cli`)

---

## Approach 2: devenv-bootc (Privileged Container)

---

### Technique

- Privileged RHEL bootc container with systemd (`podman run --privileged --systemd=true`)
- Source tree bind-mounted via git worktree
- RHSM registration inside the container for subscription-manager access
- configure-vm.sh installs build dependencies (Go, rpmbuild, osbuild, etc.)

---

### Problem: RHSM Container Mode

- On RHEL hosts, podman auto-mounts `/etc/rhsm` into `/run/secrets/rhsm`,
  causing `subscription-manager` to enter "container mode" (exit 78)
- **Fix:** Remove `/etc/pki/entitlement-host` symlink in Containerfile so
  subscription-manager registers a fresh subscription

---

### Problem: Nested Podman Network Conflict

- Host podman and container podman both use `10.88.0.0/16` default subnet
- Nested containers can't reach each other (e.g., quay ↔ redis)
- **Fix:** Set `default_subnet = "10.200.0.0/16"` in container's `containers.conf.d`

---

### Problem: SELinux in Nested Containers

- `chcon` fails with "Operation not supported" inside nested containers
- Overlay filesystem doesn't support SELinux xattrs in nested environments
- Affects: `bootc-image-builder`, `image-builder-cli`, `osbuild`
- **Attempted fixes:**
  - `--security-opt label=disable` — not enough, nested containers create own namespaces
  - `--tmpfs /sys/fs/selinux` — hides SELinux from container but not from nested builders
  - `SMDEV_CONTAINER_OFF=1` — not respected by newer subscription-manager
  - Mount `/dev/null` over `/usr/bin/chcon` — builder errors on missing binary
  - Replace `chcon` with no-op script — fragile
- **Status:** Unsolved. RPM builds work, but ISO/image builds fail.

---

### What Works

- RPM builds (`make rpm`)
- Container image builds (`podman build`)
- Mirror registry (quay/redis/postgres)
- configure-vm.sh, manage_composer_config.sh
- libvirt/KVM (with `/dev/kvm` access via `--privileged`)

### What Doesn't Work

- `bootc-image-builder` / `image-builder-cli` ISO builds (SELinux xattr issue)

---

## Approach 3: devenv-bootc-vm (Podman Machine)

---

### Technique

- Use `podman machine` (QEMU/KVM VM) instead of a privileged container
- Full kernel = real SELinux, no nested container limitations
- Build RHEL bootc OCI image, boot it as a VM

---

### Problem: podman machine init --image with OCI archive

- `podman machine init` converts OCI archive to qcow2 disk
- VM boots but never sends readiness signal on `org.fedoraproject.port.0`
- podman machine hangs on "Starting machine..."
- **Cause:** RHEL bootc image lacks Ignition (processes SSH keys, creates users,
  sends ready signal). Fedora CoreOS has this built in.

---

### Problem: Adding Ignition to RHEL bootc

- Installed `ignition`, `ignition-edge`, `afterburn`, `openssh-server`
- Rebuilt initramfs with `dracut --add ignition`
- Added `ignition.platform.id=qemu` to kernel args
- **Result:** Ignition dracut module present, but kernel args don't take effect —
  `podman machine init` converts OCI to raw disk, bypassing bootc's kargs mechanism.
  Ignition never runs.

---

### Problem: podman machine os apply (bootc switch)

- Boot default Fedora CoreOS VM first, then `os apply` to switch to RHEL
- `bootc switch` stages the RHEL deployment successfully
- **Problem 1:** `semodule` not found in bwrap sandbox during finalization
  - **Fix:** Symlink `/usr/sbin/semodule` → `/usr/bin/semodule`
- **Problem 2:** `/boot` partition too small (350MB) for two kernels
  - Fedora CoreOS ships with small `/boot`
  - RHEL kernel needs 257MB additional space, only 174MB available
  - Cannot delete running deployment's kernel files (ostree-protected)
- **Status:** Unsolved.

---

### Problem: image-builder-cli qcow2

- Build proper qcow2 disk using `image-builder-cli --bootc-ref`
- Pass qcow2 to `podman machine init --image`
- VM boots but SSH handshake fails
- **Cause:** qcow2 has no Ignition, no user/SSH key config for podman machine.
  Added blueprint with SSH key — still fails because podman machine's gvproxy
  networking requires Fedora CoreOS-specific setup.

---

### Key Insight

`podman machine` is fundamentally designed for Fedora CoreOS. It relies on:
- Ignition for user/SSH/network configuration
- `org.fedoraproject.port.0` virtio serial readiness signal
- Specific gvproxy networking setup
- rpm-ostree/bootc deployment model

None of these are present in standard RHEL bootc images.

---

## Next Steps

| Option | Pros | Cons |
|--------|------|------|
| Toolbox + `image-builder-cli` | SELinux works, no VM needed | Requires porting ostree builds from osbuild-composer |
| Direct `qemu-kvm` / `virt-install` | Full kernel, real SELinux, no nesting issues | More scripting, no podman machine UX |
| Fedora CoreOS + configure-vm.sh | Works with podman machine | Not RHEL, different package set |
