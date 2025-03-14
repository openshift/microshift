FROM localhost/cos9-bootc-source:latest

# SOURCE_IMAGES contains a comma-separated list of container image references.
# Split the variable and pull each image in a separate layer.
#
# Note:
# - Gomplate blocks are commented out to avoid hadolint warnings.
# - Retries work around sporadic "cannot set user namespace" podman error.
#
# {{ range (.Env.SOURCE_IMAGES | strings.Split ",") }}
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
