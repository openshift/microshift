FROM localhost/rhel96-bootc-source:latest

# The qemu-guest-agent is used in offline tests
RUN dnf install -y qemu-guest-agent && \
    systemctl enable qemu-guest-agent && \
    dnf clean all

# Override the default qemu-ga service configuration on the guest to allow all RPCs.
# BLOCK_RPCS is a deny-list of qemu-guest-agent RPCs to block, such as file read/write, process execution, etc. By
# default, all RPCs are blocked.
# FREEZE_HOOK_PATHNAME is the dir-path containing hook scripts. This override does not change the default value.
# These scripts are executed before or after guest image snapshot ops ("freezing" and "thawing" in libvirt speak).
# No hooks are installed by default.
RUN printf '\
\# This is a systemd environment file, not a shell script.\n\
\# It provides settings for "/lib/systemd/system/qemu-guest-agent.service".\n\
BLOCK_RPCS=\n\
FSFREEZE_HOOK_PATHNAME=/etc/qemu-ga/fsfreeze-hook\n' > /etc/sysconfig/qemu-ga

# SOURCE_IMAGES environment contains a comma-separated list of container image
# references to be set by the calling procedure. The list is prepended one more
# image used in the offline tests.
# Split the variable and pull each image in a separate layer.
#
# Note:
# - Gomplate blocks are commented out to avoid hadolint warnings.
# - Retries work around sporadic "cannot set user namespace" podman error.
#
# {{ $SOURCE_IMAGES := (printf "%s,%s" "quay.io/microshift/busybox:1.36" (.Env.SOURCE_IMAGES)) }}
# {{ range ($SOURCE_IMAGES  | strings.Split ",") }}
RUN --mount=type=secret,id=pullsecret,dst=/run/secrets/pull-secret.json \
    for i in 1 2 3; do \
        GOMAXPROCS=8 podman pull \
            --authfile /run/secrets/pull-secret.json \
            --root /usr/lib/containers/storage \
            "docker://{{ . }}" && break; \
        if [ $i -eq 3 ] ; then \
            echo "ERROR: Failed to pull {{ . }} image after 3 attempts"; \
            exit 1; \
        fi; \
        echo "Attempt $i failed. Retrying in 5 seconds..." && sleep 5; \
    done
# {{ end }}

# Edit the container storage configuration file to include the new path
RUN sed -i '/^additionalimagestores.*/a\   "/usr/lib/containers/storage",' /etc/containers/storage.conf
