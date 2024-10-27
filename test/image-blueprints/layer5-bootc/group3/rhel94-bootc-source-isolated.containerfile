FROM localhost/rhel94-bootc-source:latest

# SOURCE_IMAGES contains a comma-separated list of container image references.
# Split the variable and pull each image in a separate layer.
#
# Note: Gomplate blocks are commented out to avoid hadolint warnings.
# {{ range (.Env.SOURCE_IMAGES | strings.Split ",") }}
RUN --mount=type=secret,id=pullsecret,dst=/run/secrets/pull-secret.json \
    GOMAXPROCS=8 skopeo copy \
        --retry-times 3 \
        --authfile /run/secrets/pull-secret.json \
        "docker://{{ . }}" \
        dir:/var/lib/containers/storage-preloaded
# {{ end }}

# Edit the container storage configuration file to include the new path
RUN sed -i '/^additionalimagestores.*/a\   "/var/lib/containers/storage-preloaded",' /etc/containers/storage.conf

# Apply a workaround to set the SELinux context on the new storage directory and
# also restore 'NET_BIND_SERVICE' capability that is currently lost when including
# images in the container.
#
# Note: This requires setting the additional image stores path to a read-write
# location on the file system. The images will still be treated as read-only by
# the container subsystem.
# See https://github.com/ostreedev/ostree-rs-ext/issues/654
COPY --chmod=755 ./bootc-images/microshift-imagestore-config.sh /usr/bin/microshift-imagestore-config
RUN printf '[Unit]\n\
Description=Configure the image store directory for MicroShift\n\
Before=microshift.service\n\
[Service]\n\
Type=oneshot\n\
ExecStart=/usr/bin/microshift-imagestore-config /var/lib/containers/storage /var/lib/containers/storage-preloaded\n\
[Install]\n\
WantedBy=multi-user.target\n' > /etc/systemd/system/microshift-imagestore-config.service && \
    systemctl enable microshift-imagestore-config.service
