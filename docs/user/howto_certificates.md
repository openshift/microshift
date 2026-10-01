# Inspecting MicroShift Certificate Status

Run `microshift certs status` as root to inspect the certificates managed by
MicroShift, including CAs, serving certificates, client certificates, and peer
certificates:

```bash
sudo microshift certs status
```

The command reads the local configuration and existing certificates under
`/var/lib/microshift/certs`. It does not create, renew, or repair certificates,
restart services, or require the Kubernetes API to be available. MicroShift must
have initialized its certificates first. This is not an inventory of certificates
managed by workloads or supplied externally by users.

## Output Formats

Without an output flag, the command prints a table with `SERVICE`, `CERTIFICATE`,
`STATUS`, `EXPIRY`, and `MESSAGE` columns. Expiry times are UTC. Status is
`Healthy`, `ExpiresSoon`, `ExpirationImminent`, or `Expired`; the message
distinguishes certificates that are not expiring, expiring, expired, or not yet
valid.

For currently valid certificates, the message also explains the applicable
threshold, for example `Valid for 300 days`. `ExpiresSoon` reports that
remaining validity is at or below the warning threshold; `ExpirationImminent`
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
  and `remainingSeconds`. The `status` field uses the same four state names as the
  table. Remaining seconds are zero at expiry and negative afterward.
- `warnings`: configuration warnings, or an empty array when none apply.

For currently valid certificates, the `standard` rotation policy is `Healthy`
above 58.3% remaining validity, `ExpiresSoon` above 33.3%, and
`ExpirationImminent` otherwise. The `extended` policy uses 15% and 10% thresholds.
At or after `notAfter`, certificates are `Expired`. Not-yet-valid certificates
remain in the urgent `ExpirationImminent` category, with a `Valid in ...` message
in the table. A successful report exits with code 0 regardless of certificate
state; automation should inspect the reported `status` values.

The report contains certificate metadata, not certificate PEM data or private
keys. Go consumers can use the exported types and `AddToScheme` in
`github.com/openshift/microshift/pkg/apis/certificates/v1alpha1` to decode the
versioned documents.

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

Use `code`, rather than parsing `message`, to classify failures. Status error
codes include `InvalidArguments`, `InsufficientPrivileges`,
`InvalidConfiguration`, `CertificateInventoryFailed`, and `InternalError`.
Configuration errors deliberately omit the underlying loader diagnostic because
it can contain raw configuration or credentials. Inspect the configuration files
locally without copying sensitive values into logs.

Only `json` and `yaml` are accepted values for the output flag; omit the flag for
the table. An unsupported output format is rejected with a human-readable error.
