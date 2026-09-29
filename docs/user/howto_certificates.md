# Managing MicroShift Certificates

The `microshift certs` commands inspect and renew locally managed certificates.
All commands require root, including status and renewal dry-runs.

## Inspecting Certificate Status

Run `microshift certs status` as root to inspect the certificates managed by
MicroShift, including CAs, serving certificates, client certificates, and peer
certificates:

```bash
sudo microshift certs status
```

The command reads the local configuration and existing certificates under
`/var/lib/microshift/certs`. During normal operation it does not change certificate
material. It never restarts services or requires the Kubernetes API to be available.
An interrupted renewal is recovered before reporting status; see [Recovery](#recovery).
MicroShift must
have initialized its certificates first. This is not an inventory of certificates
managed by workloads or supplied externally by users.

## Output Formats

Without an output flag, the command prints a table with `SERVICE`, `CERTIFICATE`,
`STATUS`, `EXPIRY`, and `MESSAGE` columns. Expiry times are UTC. Status is
`Healthy`, `ExpiresSoon`, `ExpirationImminent`, `Expired`, or `NotYetValid`;
the message distinguishes certificates that are not expiring, expiring, expired,
or not yet valid.

Healthy certificates use a concise message, for example `Valid for 300 days`.
`ExpiresSoon` reports that remaining validity is at or below the warning
threshold; `ExpirationImminent`
reports the critical threshold instead. Thresholds use the certificate's actual
lifetime and rotation policy, not a fixed number of days. Displayed day counts
are rounded up; status comparisons use the unrounded values. Expired and
not-yet-valid messages do not include thresholds.

For automation, select JSON or YAML with `-o` or `--output`:

```bash
sudo microshift certs status -o json
sudo microshift certs status --output=yaml
```

Both formats emit one `CertificateStatusList` document on stdout, with
`apiVersion: microshift.openshift.io/v1alpha1`. They contain the same fields:

- `generatedAt`: the timestamp used for the report's calculations.
- `config`: the effective certificate policy, including
  `forceRestartOnExpirationImminent` and default serving/CA validity durations.
- `items`: certificates sorted by service and then name. Each item contains
  `service`, `name`, `role`, `rotationPolicy`, `status`, `notBefore`, `notAfter`,
  and `remainingSeconds`. The `status` field uses the same five state names as the
  table. Remaining seconds are zero at expiry and negative afterward.
- `warnings`: configuration warnings, or an empty array when none apply.

For currently valid certificates, the `standard` rotation policy is `Healthy`
above 58.3% remaining validity, `ExpiresSoon` above 33.3%, and
`ExpirationImminent` otherwise. The `extended` policy uses 15% and 10% thresholds.
At or after `notAfter`, certificates are `Expired`. Before `notBefore`, they are
`NotYetValid`, with a `Valid in ...` message in the table. `NotYetValid` is an
unhealthy state: check clock synchronization and certificate issuance before
deciding whether renewal is needed. It does not mean expiration is imminent.
A successful report exits with code 0 regardless of certificate state;
automation should inspect the reported `status` values.

The report contains certificate metadata, not certificate PEM data or private
keys. Go consumers can use the exported types and `AddToScheme` in
`github.com/openshift/microshift/pkg/apis/certificates/v1alpha1` to decode the
versioned documents.

## Renewing Certificates

Select exactly one renewal mode:

- `--serving` renews all managed leaf certificates: serving, client, and peer.
  CA certificates and keys remain unchanged.
- `--ca` renews every managed CA and cascades to all of its descendants.

Preview either operation while MicroShift is running:

```bash
sudo microshift certs renew --serving --dry-run
sudo microshift certs renew --ca --dry-run -o yaml
```

Dry-run validates the existing certificate/key pairs and signing relationships,
then reports current and proposed expiry times without generating keys or staging
new material. Renewal uses the existing per-certificate validity periods and
caps each descendant's expiry at the earliest expiry in its signing chain. If a
CA is expired or not yet valid, leaf-only renewal is refused; use `--ca` to renew
the chain.
Configurable validity periods are not yet implemented by this command.

For actual renewal, schedule a maintenance window, stop MicroShift, renew, then
start it explicitly:

```bash
sudo systemctl stop microshift
sudo microshift certs renew --serving
sudo systemctl start microshift
```

Replace `--serving` with `--ca` to renew the whole chain. Renewal refuses to apply
while MicroShift or its etcd scope is active. It never stops or starts a service
for you. Successful table output includes status for the renewed certificates.

Generated kubeconfigs under `/var/lib/microshift/resources` are updated in the
same transaction. After **CA renewal**, redistribute those kubeconfigs to any
external locations where you previously copied them. Existing external
kubeconfigs remain trusted after leaf-only renewal until their own certificates
expire. Applications that cache certificates or CA bundles may need a reload or
restart after either operation.

Both modes accept `-o json` and `-o yaml`. A successful command emits one versioned
`CertificateRenewalResult` with deterministically ordered `items` and an `impact`
summary. Dry-run reports `status: validated`, `dryRun: true`, and `changed: false`
for every item. Applied renewal reports `status: completed`, `dryRun: false`, and
`changed: true`, with expiry dates read back from the committed certificates.
Running renewal twice issues fresh certificates each time; it is not a no-op.

### Recovery

Renewal stages certificates, keys, bundles, and generated resources under
`/var/lib/microshift/.cert-renewal`, on the same filesystem as the active data.
It validates staging before replacement and keeps recoverable originals until
post-commit validation succeeds. The etcd database is not copied or replaced.
Allow enough free space for staging the `certs` and `resources` trees.
Renewal rejects symlinks and special files in these trees instead of following
them outside staging.

Failures before replacement leave active material untouched. Failed replacement
or validation restores the originals. If the process or host is interrupted
during replacement or rollback, the next `certs status` or `certs renew` command
recovers the interrupted transaction before proceeding. This recovery also
applies to a subsequent dry-run and requires MicroShift to be stopped:

```bash
sudo systemctl stop microshift
sudo microshift certs status
```

Startup refuses to use an incomplete transaction. Do not delete the transaction
directory manually: it may contain the only recoverable copies. If recovery
fails, keep MicroShift stopped and preserve that directory for troubleshooting.
Status and dry-run share a lock at `/var/lib/microshift-backups/certs.lock` with
the running service; startup writes, renewal, and recovery exclude concurrent
certificate commands. The lock file stays outside the data directory so renewal
and backup restoration cannot replace it. Do not delete it: the file persists,
but its lock is released automatically when the owning process closes it or exits.

## Warnings and Errors

In table mode, configuration warnings are written to stderr with a `WARNING:`
prefix. In JSON/YAML mode, warnings are included in the document's `warnings`
array, leaving stderr empty on success.

Failures exit non-zero. In table mode, stderr contains a human-readable error.
With `-o json` or `-o yaml`, failures leave stdout empty and write exactly one
versioned `Error` document to stderr, without usage text or additional diagnostics.
For example, invalid configuration produces this JSON shape:

```json
{
  "apiVersion": "microshift.openshift.io/v1alpha1",
  "kind": "Error",
  "generatedAt": "2026-09-25T12:00:00Z",
  "code": "InvalidConfiguration",
  "message": "failed to load MicroShift configuration; check /etc/microshift/config.yaml and /etc/microshift/config.d",
  "details": null
}
```

Use `code`, rather than parsing `message`, to classify failures. Certificate error
codes include `InvalidArguments`, `InsufficientPrivileges`,
`InvalidConfiguration`, `CertificateInventoryFailed`, `MicroShiftRunning`,
`RenewalFailed`, `RecoveryFailed`, and `InternalError`.
Configuration errors deliberately omit the underlying loader diagnostic because
it can contain raw configuration or credentials. Inspect the configuration files
locally without copying sensitive values into logs.

Only `json` and `yaml` are accepted values for the output flag; omit the flag for
the table. An unsupported output format is rejected with a human-readable error.
