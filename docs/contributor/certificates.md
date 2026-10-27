# Certificate Renewal Transactions

The renewal CLI stages new certificates without replacing active files.
`pkg/admin/certificates/` manages staging, publication, and recoverable replacement;
`pkg/cmd/certs_pending.go` validates and activates the pending generation during
startup, after pre-run data management and before components read certificates.
See the [user guide](../user/howto_certificates.md) for command examples and
recovery precautions.

## On-Disk State

The active data directory is `/var/lib/microshift/`. Renewal uses these paths
inside it; entries exist only in the phases that need them:

| Path relative to the data directory | Purpose |
|------------------------------------|---------|
| `certs/` | Active certificates, private keys, and CA bundles. |
| `resources/` | Active generated kubeconfigs and other component resources. |
| `.cert-renewal/stage-<id>/` | Prepared copies of `certs/` and `resources/`. |
| `.cert-renewal/pending.json` | Atomically published record selecting a complete staged generation. |
| `.cert-renewal/committing` | Durable marker indicating activation may require recovery. |
| `.cert-renewal/previous/` | Original active directories retained until activation is validated. |
| `.cert-renewal-rejected/generation-<id>/renewal/` | Quarantined transaction whose pending metadata was rejected. |

`pending.json` is private transaction state, not a supported configuration file or
public API. Its `directory` field names the selected `stage-<id>/`; its `metadata`
contains the record version, a `CertificateRenewalResult`, and three hashes:

- `configHash`: certificate-affecting MicroShift configuration inputs, such as
  host names, IP addresses, and service networks. It is not a copy of
  `/etc/microshift/config.yaml` or its drop-ins.
- `originalHash`: names and contents of the active files under `certs/` at preparation.
- `renewedHash`: names and contents of the staged files under `certs/` after renewal.

The JSON record contains neither raw configuration nor private key bytes. The
staged and quarantined directories do contain private keys and remain root-only.
The public `certs status` field `pendingRenewal` exposes the result, not this
internal transaction record.

Publication is the CLI's success boundary. Obsolete stages from a successful
replacement are cleaned by the next preparation or activation, so cleanup cannot
turn a published renewal into a reported failure.

## Activation and Recovery

At activation, MicroShift validates the hashes and certificate material. A pending
generation that has reached its automatic rotation deadline is discarded before
replacement, allowing normal certificate initialization in the same startup.
The warning explicitly notes that normal startup may rotate CAs. Healthy pending
generations are loaded unchanged rather than reissued by normal initialization.

MicroShift then refreshes staged `resources/` from the active
`/var/lib/microshift/resources/` directory, which may have changed since
preparation or been restored by startup data management. This preserves resources
renewal does not regenerate, such as `kube-apiserver/secrets/service-account-key/`,
rather than restoring their old preparation-time copies. Generated kubeconfigs
are rebuilt using the current configuration and staged certificates. The etcd
database is never staged or replaced by certificate renewal.

Activation writes the recovery marker before renaming active `certs/` and
`resources/` into `previous/` and moving their staged replacements into place.
These are recoverable sequential renames, not a single atomic swap of both
directories. Originals are retained until post-replacement validation succeeds.
Replacement or validation failures trigger rollback; an interrupted activation
or rollback is recovered before startup uses the data. A failed recovery blocks
startup and preserves the transaction for troubleshooting.
