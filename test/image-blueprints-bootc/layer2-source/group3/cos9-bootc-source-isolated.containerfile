FROM localhost/cos9-bootc-source:latest

# Embed images based on contents of release-info RPMs.
#
# Note:
# - Retries work around sporadic "cannot set user namespace" podman error.
RUN --mount=type=secret,id=pullsecret,dst=/run/secrets/pull-secret.json \
    images="$(jq -r '.images[]' /usr/share/microshift/release/release-"$(uname -m)".json)" ; \
    images="${images} quay.io/microshift/busybox:1.36" ; \
    for img in ${images} ; do \
        for i in 1 2 3; do \
            GOMAXPROCS=8 podman pull \
                --authfile /run/secrets/pull-secret.json \
                --root /usr/lib/containers/storage \
                "docker://${img}" && break; \
            if [ $i -eq 3 ] ; then \
                echo "ERROR: Failed to pull ${img} image after 3 attempts"; \
                exit 1; \
            fi; \
            echo "Attempt $i failed. Retrying in 5 seconds..." && sleep 5; \
        done ; \
    done

# Edit the container storage configuration file to include the new path
RUN sed -i '/^additionalimagestores.*/a\   "/usr/lib/containers/storage",' /etc/containers/storage.conf
