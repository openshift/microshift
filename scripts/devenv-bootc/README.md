# MicroShift Bootc Builder Container

A containerized build environment for MicroShift. Runs a privileged RHEL bootc
container with systemd, providing rpmbuild, osbuild-composer, and podman for
building RPMs, ostree images, and bootc images from any Linux host with podman.

## Prerequisites

- Linux host with `podman`, `git`, and `jq` installed
- Red Hat pull secret at `~/.pull-secret.json` (or set `PULL_SECRET`)
- RHSM activation key (`RHSM_ORG` and `RHSM_ACTIVATION_KEY` env vars)

## Quick Start

```bash
export RHSM_ORG="your-org-id"
export RHSM_ACTIVATION_KEY="your-activation-key"

# Build the container image
./scripts/devenv-bootc/devenv.sh setup

# Start and configure the container (registers subscription, installs build deps)
./scripts/devenv-bootc/devenv.sh start

# Open a shell inside the container
./scripts/devenv-bootc/devenv.sh shell

# Build RPMs (inside the container)
make rpm
```

## Commands

| Command | Description |
|---------|-------------|
| `setup` | Build the builder container image |
| `start` | Start the container and run configure-vm.sh + configure-composer.sh |
| `stop` | Stop the container (preserves state) |
| `delete` | Remove a stopped container |
| `shell` | Open an interactive shell as the builder user |
| `exec` | Run a command inside the container |
| `status` | Show container and subscription status |

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `RHSM_ORG` | setup, start | — | Red Hat subscription org ID |
| `RHSM_ACTIVATION_KEY` | setup, start | — | Red Hat subscription activation key |
| `PULL_SECRET` | setup, start | `~/.pull-secret.json` | Path to OpenShift pull secret |
| `DEVENV_BRANCH` | — | current branch | Target branch (creates a worktree) |

## Per-Release Builds

Each branch maps to a RHEL version in `rhel-versions.json`. Use `DEVENV_BRANCH`
to build for a different release:

```bash
export DEVENV_BRANCH=release-4.21
./scripts/devenv-bootc/devenv.sh setup
./scripts/devenv-bootc/devenv.sh start
./scripts/devenv-bootc/devenv.sh shell
```

Each branch gets its own container (`microshift-builder-release-4.21`) and image
(`microshift-builder:release-4.21`), so multiple releases can run side by side.

The source tree is checked out into a git worktree at `.worktrees/<branch>` and
bind-mounted into the container at `/opt/microshift`.

## How It Works

1. **`setup`** builds a minimal RHEL bootc image (podman, git, sudo, firewalld).
   Subscription is registered to install packages, then unregistered so
   credentials are not baked into the image.

2. **`start`** creates a privileged systemd container, adjusts the builder
   user's UID/GID to match the host user, sets up a symlink so git worktree
   references resolve inside the container, registers subscription, then runs
   the release branch's own `configure-vm.sh --no-build` and
   `configure-composer.sh` from the bind-mounted source tree. The repo's
   `.git` directory is also mounted so that git operations work correctly.

3. **`stop`/`start`** preserves state — a stopped container is restarted
   without re-running configuration. Use `delete` + `start` for a clean slate.
   A running container must be stopped before it can be deleted.

## Editing Code

Every branch (including the current one) uses a git worktree at
`.worktrees/<branch>`, which is bind-mounted into the container at
`/opt/microshift`. Edits made on the host are immediately visible inside the
container — edit in your IDE, build inside the container.

To work on the code, open the worktree directory in your IDE:

```bash
# For the default branch (e.g., main)
code .worktrees/main

# For a release branch
code .worktrees/release-4.21
```

The worktree is created with a detached HEAD. You can create and switch to
a private branch inside it:

```bash
cd .worktrees/main
git checkout -b my-feature
# ... make changes, commit, push
```

To return to the detached HEAD:

```bash
git checkout --detach main
```

## Files

| File | Description |
|------|-------------|
| `devenv.sh` | Container lifecycle entry point |
| `Containerfile.bootc-builder` | Minimal RHEL bootc builder image |
| `rhel-versions.json` | Branch name to RHEL version mapping |
